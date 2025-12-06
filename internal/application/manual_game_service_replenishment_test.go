package application

import (
	"bufio"
	"flip7_strategy/internal/domain"
	"strings"
	"testing"
)

// MockLoggerRepl removed, using MockLogger from reproduce_issue_test.go

func TestManualGameService_Replenishment(t *testing.T) {
	// Setup Service
	reader := bufio.NewReader(strings.NewReader(""))
	// Reuse MockLogger from reproduce_issue_test.go (same package)
	svc := NewManualGameService(reader, &MockLogger{})

	// Setup Game manually
	p1 := domain.NewPlayer("P1", nil)
	svc.Game = domain.NewGame([]*domain.Player{p1})
	// Ensure Deck is initialized
	if svc.Game.Deck == nil {
		svc.Game.Deck = domain.NewDeck()
	}
	svc.Game.CurrentRound = domain.NewRound(svc.Game.Players, p1, svc.Game.Deck)
	// Sync deck references (PR feedback)
	svc.Game.Deck = svc.Game.CurrentRound.Deck

	// Rig Deck: Contains only "1"
	card1 := domain.Card{Type: domain.CardTypeNumber, Value: 1}
	svc.Game.CurrentRound.Deck.Cards = []domain.Card{card1}
	svc.Game.CurrentRound.Deck.RemainingCounts = make(map[domain.NumberValue]int)
	svc.Game.CurrentRound.Deck.RemainingCounts[1] = 1

	// Rig DiscardPile: Contains "2"
	card2 := domain.Card{Type: domain.CardTypeNumber, Value: 2}
	svc.Game.DiscardPile = []domain.Card{card2}

	// 1. Remove Card 1 (Should succeed)
	err := svc.removeCardFromDeck(card1)
	if err != nil {
		t.Fatalf("Failed to remove existing card 1: %v", err)
	}

	// Deck should now be empty
	if len(svc.Game.CurrentRound.Deck.Cards) != 0 {
		t.Errorf("Deck should be empty, but has %d cards", len(svc.Game.CurrentRound.Deck.Cards))
	}

	// 2. Remove Card 2 (Should succeed due to replenishment)
	err = svc.removeCardFromDeck(card2)
	if err != nil {
		t.Errorf("Replenishment Failed: Expected success when removing card 2 from replenished deck, but got error: %v", err)
	} else {
		t.Log("Replenishment Successful: Card 2 removed from deck.")

		// Verify Side Effects (PR feedback)
		// 1. Discard pile should be empty
		if len(svc.Game.DiscardPile) != 0 {
			t.Errorf("Expected DiscardPile to be empty after replenishment, got %d cards", len(svc.Game.DiscardPile))
		}

		// 2. Game.Deck should be updated to point to the new deck (CurrentRound.Deck)
		if svc.Game.Deck != svc.Game.CurrentRound.Deck {
			t.Error("Expected Game.Deck to be updated to CurrentRound.Deck after replenishment")
		}
	}
}
