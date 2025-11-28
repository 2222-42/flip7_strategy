package strategy_test

import (
	"testing"

	"flip7_strategy/internal/domain"
	"flip7_strategy/internal/domain/strategy"
)

func TestExpectedValueStrategy_Decide(t *testing.T) {
	s := &strategy.ExpectedValueStrategy{}

	// Helper to create a deck with specific cards
	createDeck := func(cards []domain.Card) *domain.Deck {
		return domain.NewDeckFromCards(cards)
	}

	tests := []struct {
		name           string
		deckCards      []domain.Card
		handNumbers    []int
		expectedChoice domain.TurnChoice
	}{
		{
			name: "High EV: Deck has high value card, hand is empty",
			deckCards: []domain.Card{
				{Type: domain.CardTypeNumber, Value: 10},
			},
			handNumbers:    []int{},
			expectedChoice: domain.TurnChoiceHit,
		},
		{
			name: "Low EV: Deck causes bust",
			deckCards: []domain.Card{
				{Type: domain.CardTypeNumber, Value: 5},
			},
			handNumbers:    []int{5}, // Already have 5, drawing 5 causes bust (score 0)
			expectedChoice: domain.TurnChoiceStay,
		},
		{
			name: "Mixed EV: 50% chance to bust, but reward is high enough?",
			// Hand: 10. Score: 10.
			// Deck: 10 (Bust, 0), 5 (Score 15).
			// EV = (0 + 15) / 2 = 7.5.
			// Current Score = 10.
			// EV < Current, should Stay.
			deckCards: []domain.Card{
				{Type: domain.CardTypeNumber, Value: 10},
				{Type: domain.CardTypeNumber, Value: 5},
			},
			handNumbers:    []int{10},
			expectedChoice: domain.TurnChoiceStay,
		},
		{
			name: "Positive EV Hit: 50% chance to bust, but reward is very high",
			// Hand: 2. Score: 2.
			// Deck: 2 (Bust, 0), 12 (Score 14).
			// EV = (0 + 14) / 2 = 7.
			// Current Score = 2.
			// EV > Current, should Hit.
			deckCards: []domain.Card{
				{Type: domain.CardTypeNumber, Value: 2},
				{Type: domain.CardTypeNumber, Value: 12},
			},
			handNumbers:    []int{2},
			expectedChoice: domain.TurnChoiceHit,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deck := createDeck(tt.deckCards)
			hand := domain.NewPlayerHand()
			for _, n := range tt.handNumbers {
				hand.AddCard(domain.Card{Type: domain.CardTypeNumber, Value: domain.NumberValue(n)})
			}

			choice := s.Decide(deck, hand, 0, nil)
			if choice != tt.expectedChoice {
				t.Errorf("Expected %v, got %v", tt.expectedChoice, choice)
			}
		})
	}
}
