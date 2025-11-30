package application

import (
	"encoding/base64"
	"encoding/json"
	"testing"

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

	// 2. Serialize
	data, err := json.Marshal(game)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}
	encoded := base64.StdEncoding.EncodeToString(data)

	// 3. Deserialize
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		t.Fatalf("DecodeString failed: %v", err)
	}

	var loadedGame domain.Game
	err = json.Unmarshal(decoded, &loadedGame)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	// 4. Relink pointers (Simulate the logic we need to implement)
	relinkPointers(&loadedGame)

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
}

func relinkPointers(g *domain.Game) {
	playerMap := make(map[string]*domain.Player)
	for _, p := range g.Players {
		playerMap[p.ID.String()] = p
	}

	if g.CurrentRound != nil {
		// Relink Round Players
		for i, p := range g.CurrentRound.Players {
			if existing, ok := playerMap[p.ID.String()]; ok {
				g.CurrentRound.Players[i] = existing
			}
		}
		// Relink Active Players
		for i, p := range g.CurrentRound.ActivePlayers {
			if existing, ok := playerMap[p.ID.String()]; ok {
				g.CurrentRound.ActivePlayers[i] = existing
			}
		}
		// Relink Dealer
		if g.CurrentRound.Dealer != nil {
			if existing, ok := playerMap[g.CurrentRound.Dealer.ID.String()]; ok {
				g.CurrentRound.Dealer = existing
			}
		}
	}

	// Relink Winners
	for i, p := range g.Winners {
		if existing, ok := playerMap[p.ID.String()]; ok {
			g.Winners[i] = existing
		}
	}
}
