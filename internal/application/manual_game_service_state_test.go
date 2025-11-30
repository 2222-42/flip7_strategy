package application_test

import (
	"bufio"
	"encoding/base64"
	"strings"
	"testing"

	"flip7_strategy/internal/application"
	"flip7_strategy/internal/domain"
	"flip7_strategy/internal/domain/strategy"
)

func TestGameStateSerialization(t *testing.T) {
	// 1. Setup a game state
	p1 := domain.NewPlayer("P1", &strategy.ProbabilisticStrategy{})
	p2 := domain.NewPlayer("P2", &strategy.ProbabilisticStrategy{})
	players := []*domain.Player{p1, p2}
	game := domain.NewGame(players)

	deck := domain.NewDeck()
	dealer := p1
	game.CurrentRound = domain.NewRound(players, dealer, deck)
	game.DealerIndex = 0

	// Modify state to ensure it's captured
	p1.TotalScore = 50
	p1.CurrentHand.AddCard(domain.Card{Type: domain.CardTypeNumber, Value: 5})
	game.CurrentRound.ActivePlayers = []*domain.Player{p1, p2} // Both active
	game.CurrentRound.CurrentTurnIndex = 1                     // Set to test serialization

	// 2. Create service and use public RelinkPointers via SaveState/LoadState
	reader := bufio.NewReader(strings.NewReader(""))
	service := application.NewManualGameService(reader)
	service.Game = game

	// 3. Save state
	saveCode, err := service.SaveState()
	if err != nil {
		t.Fatalf("SaveState failed: %v", err)
	}

	// 4. Load state into a new service
	newService := application.NewManualGameService(reader)
	err = newService.LoadState(saveCode)
	if err != nil {
		t.Fatalf("LoadState failed: %v", err)
	}

	loadedGame := newService.Game

	// 5. Verify
	if game.ID != loadedGame.ID {
		t.Errorf("ID mismatch: got %v, want %v", loadedGame.ID, game.ID)
	}
	if len(game.Players) != len(loadedGame.Players) {
		t.Errorf("Players length mismatch: got %d, want %d", len(loadedGame.Players), len(game.Players))
	}
	if game.Players[0].Name != loadedGame.Players[0].Name {
		t.Errorf("Player 0 Name mismatch: got %s, want %s", loadedGame.Players[0].Name, game.Players[0].Name)
	}
	if game.Players[0].TotalScore != loadedGame.Players[0].TotalScore {
		t.Errorf("Player 0 TotalScore mismatch: got %d, want %d", loadedGame.Players[0].TotalScore, game.Players[0].TotalScore)
	}
	if len(game.Players[0].CurrentHand.RawNumberCards) != len(loadedGame.Players[0].CurrentHand.RawNumberCards) {
		t.Errorf("Player 0 Hand mismatch")
	}

	// Verify pointers are consistent
	if loadedGame.Players[0] != loadedGame.CurrentRound.Players[0] {
		t.Errorf("Round players should point to Game players")
	}
	if loadedGame.Players[0] != loadedGame.CurrentRound.Dealer {
		t.Errorf("Dealer should point to Game player")
	}

	// Verify CurrentTurnIndex is preserved
	if game.CurrentRound.CurrentTurnIndex != loadedGame.CurrentRound.CurrentTurnIndex {
		t.Errorf("CurrentTurnIndex mismatch: got %d, want %d", loadedGame.CurrentRound.CurrentTurnIndex, game.CurrentRound.CurrentTurnIndex)
	}
}

func TestSaveStateAndLoadState(t *testing.T) {
	t.Run("Valid save and load", func(t *testing.T) {
		// Setup a game
		p1 := domain.NewPlayer("User", nil) // User-controlled
		p2 := domain.NewPlayer("AI", &strategy.ProbabilisticStrategy{})
		players := []*domain.Player{p1, p2}
		game := domain.NewGame(players)

		deck := domain.NewDeck()
		game.CurrentRound = domain.NewRound(players, p1, deck)
		game.DealerIndex = 0
		p1.TotalScore = 100

		reader := bufio.NewReader(strings.NewReader(""))
		service := application.NewManualGameService(reader)
		service.Game = game

		// Save
		code, err := service.SaveState()
		if err != nil {
			t.Fatalf("SaveState failed: %v", err)
		}
		if code == "" {
			t.Errorf("SaveState returned empty code")
		}

		// Load into new service
		newService := application.NewManualGameService(reader)
		err = newService.LoadState(code)
		if err != nil {
			t.Fatalf("LoadState failed: %v", err)
		}

		// Verify game state
		if newService.Game.ID != game.ID {
			t.Errorf("Game ID mismatch")
		}
		if newService.Game.Players[0].TotalScore != 100 {
			t.Errorf("Player score not restored correctly")
		}

		// Verify strategies are restored correctly
		if newService.Game.Players[0].Strategy != nil {
			t.Errorf("User-controlled player should have nil strategy, got %v", newService.Game.Players[0].Strategy)
		}
		if newService.Game.Players[1].Strategy == nil {
			t.Errorf("AI player should have non-nil strategy")
		}
	})

	t.Run("Invalid base64 code", func(t *testing.T) {
		reader := bufio.NewReader(strings.NewReader(""))
		service := application.NewManualGameService(reader)

		err := service.LoadState("invalid!@#$%")
		if err == nil {
			t.Errorf("LoadState should fail with invalid base64")
		}
	})

	t.Run("Invalid JSON in code", func(t *testing.T) {
		reader := bufio.NewReader(strings.NewReader(""))
		service := application.NewManualGameService(reader)

		invalidJSON := base64.StdEncoding.EncodeToString([]byte("{invalid json}"))
		err := service.LoadState(invalidJSON)
		if err == nil {
			t.Errorf("LoadState should fail with invalid JSON")
		}
	})

	t.Run("Completed game cannot be loaded", func(t *testing.T) {
		// Setup a completed game
		p1 := domain.NewPlayer("User", nil)
		p2 := domain.NewPlayer("AI", &strategy.ProbabilisticStrategy{})
		players := []*domain.Player{p1, p2}
		game := domain.NewGame(players)
		game.IsCompleted = true

		reader := bufio.NewReader(strings.NewReader(""))
		service := application.NewManualGameService(reader)
		service.Game = game

		// Save
		code, err := service.SaveState()
		if err != nil {
			t.Fatalf("SaveState failed: %v", err)
		}

		// Try to load
		newService := application.NewManualGameService(reader)
		err = newService.LoadState(code)
		if err == nil {
			t.Errorf("LoadState should reject completed game")
		}
		if !strings.Contains(err.Error(), "already completed") {
			t.Errorf("Error should mention game is completed, got: %v", err)
		}
	})

	t.Run("Ended round cannot be loaded", func(t *testing.T) {
		// Setup a game with ended round
		p1 := domain.NewPlayer("User", nil)
		p2 := domain.NewPlayer("AI", &strategy.ProbabilisticStrategy{})
		players := []*domain.Player{p1, p2}
		game := domain.NewGame(players)

		deck := domain.NewDeck()
		game.CurrentRound = domain.NewRound(players, p1, deck)
		game.CurrentRound.IsEnded = true

		reader := bufio.NewReader(strings.NewReader(""))
		service := application.NewManualGameService(reader)
		service.Game = game

		// Save
		code, err := service.SaveState()
		if err != nil {
			t.Fatalf("SaveState failed: %v", err)
		}

		// Try to load
		newService := application.NewManualGameService(reader)
		err = newService.LoadState(code)
		if err == nil {
			t.Errorf("LoadState should reject game with ended round")
		}
		if !strings.Contains(err.Error(), "already ended") {
			t.Errorf("Error should mention round is ended, got: %v", err)
		}
	})

	t.Run("Pointer relinking works correctly", func(t *testing.T) {
		// Setup game with multiple pointer references
		p1 := domain.NewPlayer("User", nil)
		p2 := domain.NewPlayer("AI", &strategy.ProbabilisticStrategy{})
		players := []*domain.Player{p1, p2}
		game := domain.NewGame(players)

		deck := domain.NewDeck()
		game.CurrentRound = domain.NewRound(players, p1, deck)
		game.Winners = []*domain.Player{p1}

		reader := bufio.NewReader(strings.NewReader(""))
		service := application.NewManualGameService(reader)
		service.Game = game

		// Save and load
		code, err := service.SaveState()
		if err != nil {
			t.Fatalf("SaveState failed: %v", err)
		}

		newService := application.NewManualGameService(reader)
		err = newService.LoadState(code)
		if err != nil {
			t.Fatalf("LoadState failed: %v", err)
		}

		// Verify all pointers point to same instances
		g := newService.Game
		if g.Players[0] != g.CurrentRound.Players[0] {
			t.Errorf("Round.Players should point to Game.Players instances")
		}
		if g.Players[0] != g.CurrentRound.Dealer {
			t.Errorf("Round.Dealer should point to Game.Players instance")
		}
		if g.Players[0] != g.CurrentRound.ActivePlayers[0] {
			t.Errorf("Round.ActivePlayers should point to Game.Players instances")
		}
		if g.Players[0] != g.Winners[0] {
			t.Errorf("Winners should point to Game.Players instances")
		}
	})
}
