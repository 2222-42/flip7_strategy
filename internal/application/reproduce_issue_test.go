package application

import (
	"bufio"
	"flip7_strategy/internal/domain"
	"strings"
	"testing"
)

type MockLogger struct{}

func (m *MockLogger) Log(gameID, roundID, playerID, eventType string, details map[string]interface{}) {
}
func (m *MockLogger) Close() {}

// TestReproduceDeckExhaustion verifies that the deck is NOT reset between rounds in manual mode.
func TestReproduceDeckExhaustion(t *testing.T) {
	// Inputs:
	// Setup:
	// "" (No resume)
	// "2" (Num players)
	// "Player2" (Name of P2)
	// "1" (Start player Me)
	//
	// Round 1 (Me is Dealer/First):
	// "4" (Me)
	// "4" (Me)
	// "4" (Me)
	// "4" (Me) -> All 4s gone.
	// "S" (Me Stay)
	// "S" (Player 2 Stay) -> Round 1 Ends.
	//
	// Round 2 (Player 2 is Dealer/First):
	// "4" (Try to draw 4. Should fail if deck persists.)
	// "5" (Fallback)
	inputLines := []string{
		"",        // No resume
		"2",       // 2 players
		"Player2", // P2 Name
		"1",       // Start with Me
		// Round 1
		"4", "4", "4", "4", // Me takes all 4s
		"S", // Me Stay
		"5", // P2 draws 5 (Valid hit)
		"S", // P2 Stay
		// Round 2
		"4", // P2 tries 4.
		"6", // P2 tries 6 (Fallback).
		"S", // P2 Stay
		"S", // Me Stay
	}
	input := strings.Join(inputLines, "\n") + "\n"
	reader := bufio.NewReader(strings.NewReader(input))

	svc := NewManualGameService(reader, &MockLogger{})

	// Manually drive the game to avoid infinite loops and race conditions
	svc.setupPlayers()

	if svc.Game == nil {
		t.Fatal("Game not initialized")
	}

	// Verify Setup
	if len(svc.Game.Players) != 2 {
		t.Fatalf("Expected 2 players, got %d", len(svc.Game.Players))
	}
	// Initial Dealer should be 0 (Me)
	if svc.Game.DealerIndex != 0 {
		t.Fatalf("Expected DealerIndex 0 (Me), got %d", svc.Game.DealerIndex)
	}

	// Play Round 1
	svc.playRound()

	// Verify Round 1 Ended
	if svc.Game.CurrentRound == nil || !svc.Game.CurrentRound.IsEnded {
		t.Fatal("Round 1 should have ended")
	}

	// Prepare for Round 2
	svc.Game.DealerIndex = (svc.Game.DealerIndex + 1) % len(svc.Game.Players)
	// CurrentTurnIndex resetting is handled in NewRound/playRound logic?
	// NewRound creates new ActivePlayers.

	// Play Round 2 (This is where the bug manifests)
	// The function playRound calls NewRound -> NewDeck().
	svc.playRound()

	// Analyze P2's hand (P2 should be the one who acted first in Round 2, as Dealer)
	// Dealer is P2. Round starts with Dealer.
	// P2 tries to draw "4".

	// Find P2
	var p2 *domain.Player
	for _, p := range svc.Game.Players {
		if p.Name == "Player2" {
			p2 = p
			break
		}
	}
	if p2 == nil {
		t.Fatal("Player 2 not found")
	}

	// In the BUG scenario: NewDeck() restored the 4s. P2 successfully drew a 4.
	// In the FIX scenario: 4s are gone. "4" input fails (caught by loop in playRound, prints error).
	// Next input "5" is consumed. P2 draws 5.

	has4 := false
	for _, val := range p2.CurrentHand.RawNumberCards {
		if val == 4 {
			has4 = true
			break
		}
	}

	if has4 {
		t.Errorf("BUG REPRODUCED: Player 2 was able to draw '4' even though all '4's were exhausted in Round 1.")
	} else {
		// Verify they got the '5' to ensure the test is valid
		has5 := false
		for _, val := range p2.CurrentHand.RawNumberCards {
			if val == 5 {
				has5 = true
				break
			}
		}
		if !has5 {
			t.Log("Player 2 does not have 5 either. Check inputs/logic.")
		}
	}
}
