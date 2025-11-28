package domain_test

import (
	"testing"

	"flip7_strategy/internal/domain"
)

func TestEstimateFlipThreeRisk(t *testing.T) {
	// Scenario 1: High Risk
	// Hand has 0, 1, 2. Deck has duplicates of 0, 1, 2.
	// Drawing 3 cards from a deck of duplicates guarantees a bust.
	t.Run("High Risk (Guaranteed Bust)", func(t *testing.T) {
		handNumbers := map[domain.NumberValue]struct{}{
			0: {}, 1: {}, 2: {},
		}

		// Create a deck with only cards that are in the hand
		cards := []domain.Card{
			{Type: domain.CardTypeNumber, Value: 0},
			{Type: domain.CardTypeNumber, Value: 1},
			{Type: domain.CardTypeNumber, Value: 2},
		}
		deck := domain.NewDeckFromCards(cards)

		risk := deck.EstimateFlipThreeRisk(handNumbers, false)
		if risk < 0.99 {
			t.Errorf("Expected risk ~1.0, got %f", risk)
		}
	})

	// Scenario 2: Low Risk
	// Hand has 0, 1, 2. Deck has only safe cards (3, 4, 5).
	t.Run("Low Risk (Safe)", func(t *testing.T) {
		handNumbers := map[domain.NumberValue]struct{}{
			0: {}, 1: {}, 2: {},
		}

		cards := []domain.Card{
			{Type: domain.CardTypeNumber, Value: 3},
			{Type: domain.CardTypeNumber, Value: 4},
			{Type: domain.CardTypeNumber, Value: 5},
		}
		deck := domain.NewDeckFromCards(cards)

		risk := deck.EstimateFlipThreeRisk(handNumbers, false)
		if risk > 0.01 {
			t.Errorf("Expected risk ~0.0, got %f", risk)
		}
	})

	// Scenario 3: Second Chance Protection
	// Hand has 0. Deck has 0, 1, 2.
	// Drawing 0 would bust, but Second Chance saves it.
	// Drawing 0, 1, 2 -> Safe (0 uses SC, 1, 2 added).
	// Drawing 0, 0, 1 -> Bust (0 uses SC, second 0 busts).
	t.Run("Second Chance Protection", func(t *testing.T) {
		handNumbers := map[domain.NumberValue]struct{}{
			0: {},
		}

		// Deck with one duplicate (0) and two safe cards (1, 2)
		cards := []domain.Card{
			{Type: domain.CardTypeNumber, Value: 0},
			{Type: domain.CardTypeNumber, Value: 1},
			{Type: domain.CardTypeNumber, Value: 2},
		}
		deck := domain.NewDeckFromCards(cards)

		// With Second Chance, drawing 0 should NOT bust.
		// Since deck has only 3 cards and we draw 3, we will definitely draw 0, 1, 2.
		// Order doesn't matter for the set, but 0 triggers SC.
		// Result: Safe.
		risk := deck.EstimateFlipThreeRisk(handNumbers, true)
		if risk > 0.01 {
			t.Errorf("Expected risk ~0.0 with Second Chance, got %f", risk)
		}

		// Without Second Chance, drawing 0 busts.
		riskNoSC := deck.EstimateFlipThreeRisk(handNumbers, false)
		if riskNoSC < 0.99 {
			t.Errorf("Expected risk ~1.0 without Second Chance, got %f", riskNoSC)
		}
	})
}
