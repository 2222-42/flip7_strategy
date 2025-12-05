package application

import (
	"bufio"
	"fmt"
	"strings"
	"testing"

	"flip7_strategy/internal/domain"
)

// TestRemoveCardFromDeckAcrossRounds tests that card validation works correctly
// when cards are drawn across multiple rounds with deck resets.
func TestRemoveCardFromDeckAcrossRounds(t *testing.T) {
	// Simulate the scenario described in the bug report:
	// 1. Two players
	// 2. Each draws card 4 twice (4 total draws of card 4)
	// 3. Start a new round
	// 4. Try to draw card 4 again - should work because new deck is created

	reader := bufio.NewReader(strings.NewReader(""))
	service := NewManualGameService(reader, nil)

	// Create a game with 2 players
	players := []*domain.Player{
		domain.NewPlayer("Player1", nil),
		domain.NewPlayer("Player2", nil),
	}
	service.Game = domain.NewGame(players)

	// Start Round 1
	deck1 := domain.NewDeck()
	service.Game.CurrentRound = domain.NewRound(players, players[0], deck1)

	// Debug: Check initial state
	fmt.Printf("Round 1 Initial: RemainingCounts[4] = %d, len(Cards) = %d\n",
		service.Game.CurrentRound.Deck.RemainingCounts[domain.NumberValue(4)],
		len(service.Game.CurrentRound.Deck.Cards))

	// Player 1 draws card 4 twice
	card4 := domain.Card{Type: domain.CardTypeNumber, Value: domain.NumberValue(4)}
	
	err := service.removeCardFromDeck(card4)
	if err != nil {
		t.Fatalf("Round 1, Draw 1: Expected no error, got %v", err)
	}
	fmt.Printf("After Draw 1: RemainingCounts[4] = %d\n", service.Game.CurrentRound.Deck.RemainingCounts[domain.NumberValue(4)])
	
	err = service.removeCardFromDeck(card4)
	if err != nil {
		t.Fatalf("Round 1, Draw 2: Expected no error, got %v", err)
	}
	fmt.Printf("After Draw 2: RemainingCounts[4] = %d\n", service.Game.CurrentRound.Deck.RemainingCounts[domain.NumberValue(4)])

	// Player 2 draws card 4 twice
	err = service.removeCardFromDeck(card4)
	if err != nil {
		t.Fatalf("Round 1, Draw 3: Expected no error, got %v", err)
	}
	fmt.Printf("After Draw 3: RemainingCounts[4] = %d\n", service.Game.CurrentRound.Deck.RemainingCounts[domain.NumberValue(4)])
	
	err = service.removeCardFromDeck(card4)
	if err != nil {
		t.Fatalf("Round 1, Draw 4: Expected no error, got %v", err)
	}
	fmt.Printf("After Draw 4: RemainingCounts[4] = %d\n", service.Game.CurrentRound.Deck.RemainingCounts[domain.NumberValue(4)])

	// All 4 copies of card 4 should be exhausted
	err = service.removeCardFromDeck(card4)
	if err == nil {
		t.Fatalf("Round 1, Draw 5: Expected error (all card 4s drawn), got nil")
	}
	fmt.Printf("After Draw 5 (should fail): RemainingCounts[4] = %d\n", service.Game.CurrentRound.Deck.RemainingCounts[domain.NumberValue(4)])

	// Verify the count is 0
	if service.Game.CurrentRound.Deck.RemainingCounts[domain.NumberValue(4)] != 0 {
		t.Errorf("Expected RemainingCounts[4] = 0, got %d", service.Game.CurrentRound.Deck.RemainingCounts[domain.NumberValue(4)])
	}

	// End round 1
	service.Game.CurrentRound.End(domain.RoundEndReasonNoActivePlayers)

	// Start Round 2 - This should create a NEW deck
	deck2 := domain.NewDeck()
	service.Game.CurrentRound = domain.NewRound(players, players[0], deck2)

	fmt.Printf("\nRound 2 Initial: RemainingCounts[4] = %d, len(Cards) = %d\n",
		service.Game.CurrentRound.Deck.RemainingCounts[domain.NumberValue(4)],
		len(service.Game.CurrentRound.Deck.Cards))

	// Now try to draw card 4 again - should work because it's a new deck
	err = service.removeCardFromDeck(card4)
	if err != nil {
		t.Fatalf("Round 2, Draw 1: Expected no error (new deck), got %v", err)
	}

	// Verify the new deck has card 4s
	if service.Game.CurrentRound.Deck.RemainingCounts[domain.NumberValue(4)] != 3 {
		t.Errorf("Round 2: Expected RemainingCounts[4] = 3 after 1 draw, got %d", service.Game.CurrentRound.Deck.RemainingCounts[domain.NumberValue(4)])
	}
}

// TestRemoveCardFromDeckBugRepro tests the exact scenario from the bug report:
// 1. Set number of players to 2
// 2. Type 4 for 4 times (each player drew 4 two times and busted)
// 3. At next round, type 4
// Expected: In round 2, typing 4 should work (new deck)
func TestRemoveCardFromDeckBugRepro(t *testing.T) {
	reader := bufio.NewReader(strings.NewReader(""))
	service := NewManualGameService(reader, nil)

	// 1. Set number of players to 2
	players := []*domain.Player{
		domain.NewPlayer("Player1", nil),
		domain.NewPlayer("Player2", nil),
	}
	service.Game = domain.NewGame(players)
	service.Game.DealerIndex = 0

	// Start Round 1
	service.Game.RoundCount = 1
	deck1 := domain.NewDeck()
	service.Game.CurrentRound = domain.NewRound(players, players[0], deck1)

	fmt.Printf("\n=== Round 1 ===\n")
	fmt.Printf("Initial: RemainingCounts[4] = %d, Cards in deck = %d\n",
		service.Game.CurrentRound.Deck.RemainingCounts[domain.NumberValue(4)],
		len(service.Game.CurrentRound.Deck.Cards))

	// 2. Type 4 for 4 times
	card4 := domain.Card{Type: domain.CardTypeNumber, Value: domain.NumberValue(4)}
	
	for i := 1; i <= 4; i++ {
		err := service.removeCardFromDeck(card4)
		if err != nil {
			t.Fatalf("Round 1, Draw %d: Expected no error, got %v", i, err)
		}
		fmt.Printf("After draw %d: RemainingCounts[4] = %d\n", i, service.Game.CurrentRound.Deck.RemainingCounts[domain.NumberValue(4)])
	}

	// Verify all card 4s are exhausted
	err := service.removeCardFromDeck(card4)
	if err == nil {
		t.Errorf("Round 1: Expected error after drawing all card 4s, got nil")
	} else {
		fmt.Printf("Round 1: Correctly rejected 5th draw of card 4: %v\n", err)
	}

	// End Round 1
	service.Game.CurrentRound.End(domain.RoundEndReasonNoActivePlayers)
	
	// 3. At next round, type 4
	service.Game.RoundCount++
	deck2 := domain.NewDeck()
	service.Game.CurrentRound = domain.NewRound(players, players[0], deck2)

	fmt.Printf("\n=== Round 2 ===\n")
	fmt.Printf("Initial: RemainingCounts[4] = %d, Cards in deck = %d\n",
		service.Game.CurrentRound.Deck.RemainingCounts[domain.NumberValue(4)],
		len(service.Game.CurrentRound.Deck.Cards))

	// This should work because it's a new deck
	err = service.removeCardFromDeck(card4)
	if err != nil {
		t.Errorf("Round 2: Expected no error (new deck has card 4s), got %v", err)
	} else {
		fmt.Printf("Round 2: Successfully drew card 4 from new deck\n")
		fmt.Printf("After draw: RemainingCounts[4] = %d\n", service.Game.CurrentRound.Deck.RemainingCounts[domain.NumberValue(4)])
	}
}
