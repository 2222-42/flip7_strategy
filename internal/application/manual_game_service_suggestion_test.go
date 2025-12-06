package application_test

import (
	"bufio"
	"flip7_strategy/internal/application"
	"flip7_strategy/internal/domain"
	"strings"
	"testing"
)

// TestFormatCandidateOption tests the formatting of candidate options
// with various scenarios including suggested candidates and nil hands.
func TestFormatCandidateOption(t *testing.T) {
	service := &application.ManualGameService{}

	candidate := domain.NewPlayer("Candidate", nil)
	candidate.TotalScore = 150
	candidate.CurrentHand = &domain.PlayerHand{
		RawNumberCards: []domain.NumberValue{5, 8},
	}

	t.Run("Suggested", func(t *testing.T) {
		suggested := candidate // Same pointer
		output := service.FormatCandidateOption(candidate, suggested)

		if !strings.Contains(output, "[Suggested]") {
			t.Errorf("Expected output to contain '[Suggested]', got: %s", output)
		}
		if !strings.Contains(output, "Candidate") {
			t.Errorf("Expected output to contain name, got: %s", output)
		}
		if !strings.Contains(output, "Score: 150") {
			t.Errorf("Expected output to contain score, got: %s", output)
		}
	})

	t.Run("NotSuggested", func(t *testing.T) {
		other := domain.NewPlayer("Other", nil)
		output := service.FormatCandidateOption(candidate, other)

		if strings.Contains(output, "[Suggested]") {
			t.Errorf("Expected output NOT to contain '[Suggested]', got: %s", output)
		}
	})

	t.Run("NilSuggested", func(t *testing.T) {
		output := service.FormatCandidateOption(candidate, nil)

		if strings.Contains(output, "[Suggested]") {
			t.Errorf("Expected output NOT to contain '[Suggested]' for nil suggestion, got: %s", output)
		}
		if !strings.Contains(output, "Candidate") {
			t.Errorf("Expected output to contain name, got: %s", output)
		}
	})

	t.Run("NilCurrentHand", func(t *testing.T) {
		candidateWithoutHand := domain.NewPlayer("NoHand", nil)
		candidateWithoutHand.TotalScore = 100
		// NewPlayer creates players with nil CurrentHand

		output := service.FormatCandidateOption(candidateWithoutHand, nil)

		if !strings.Contains(output, "NoHand") {
			t.Errorf("Expected output to contain name, got: %s", output)
		}
		if !strings.Contains(output, "[]") {
			t.Errorf("Expected output to contain empty hand '[]', got: %s", output)
		}
	})
}

// TestPromptForTargetSuggestionLogic tests the integration of AdaptiveStrategy
// in the target selection process. Note: Full integration testing of promptForTarget
// is challenging due to its interactive nature (stdout/stdin), so we focus on
// verifying the suggestion logic works correctly when integrated into the service.
func TestPromptForTargetSuggestionLogic(t *testing.T) {
	// This test verifies that the suggestion logic in promptForTarget works correctly
	// by testing the FormatCandidateOption method with realistic game scenarios.
	// The actual AdaptiveStrategy.ChooseTarget behavior is tested in the strategy package.

	t.Run("WithCurrentRound", func(t *testing.T) {
		// Setup a game with a current round and deck
		reader := bufio.NewReader(strings.NewReader("1\n"))
		service := application.NewManualGameService(reader, nil)

		// Initialize game with players
		p1 := domain.NewPlayer("Player1", nil)
		p2 := domain.NewPlayer("Player2", nil)
		players := []*domain.Player{p1, p2}
		service.Game = domain.NewGame(players)

		// Start a new round to initialize CurrentRound and Deck
		deck := domain.NewDeck()
		service.Game.CurrentRound = domain.NewRound(players, p1, deck)

		// Verify that with a CurrentRound, the deck is available for AdaptiveStrategy
		if service.Game.CurrentRound == nil {
			t.Fatal("Expected CurrentRound to be initialized")
		}
		if service.Game.CurrentRound.Deck == nil {
			t.Fatal("Expected Deck to be initialized in CurrentRound")
		}

		// The actual suggestion logic is exercised in promptForTarget,
		// but we can't easily test it without mocking stdout/stdin.
		// We verify the components are in place and FormatCandidateOption
		// works correctly with suggestions.
		player1 := service.Game.Players[0]
		player2 := service.Game.Players[1]

		output := service.FormatCandidateOption(player1, player2)
		if strings.Contains(output, "[Suggested]") {
			t.Error("Expected player1 NOT to be marked as suggested when player2 is suggested")
		}

		output = service.FormatCandidateOption(player2, player2)
		if !strings.Contains(output, "[Suggested]") {
			t.Error("Expected player2 to be marked as suggested")
		}
	})

	t.Run("WithoutCurrentRound", func(t *testing.T) {
		// Verify that the service handles nil CurrentRound gracefully
		reader := bufio.NewReader(strings.NewReader("1\n"))
		service := application.NewManualGameService(reader, nil)

		p1 := domain.NewPlayer("Player1", nil)
		p2 := domain.NewPlayer("Player2", nil)
		players := []*domain.Player{p1, p2}
		service.Game = domain.NewGame(players)
		// Don't start a round, so CurrentRound is nil

		if service.Game.CurrentRound != nil {
			t.Fatal("Expected CurrentRound to be nil")
		}

		// The service should still work and AdaptiveStrategy should handle nil deck
		player1 := service.Game.Players[0]
		player1.CurrentHand = &domain.PlayerHand{
			RawNumberCards: []domain.NumberValue{3, 7},
		}

		// FormatCandidateOption should work even without CurrentRound
		output := service.FormatCandidateOption(player1, nil)
		if !strings.Contains(output, "Player1") {
			t.Errorf("Expected output to contain player name, got: %s", output)
		}
		if !strings.Contains(output, "[3, 7]") {
			t.Errorf("Expected output to contain hand, got: %s", output)
		}
	})
}

