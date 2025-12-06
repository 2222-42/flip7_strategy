package domain_test

import (
	"testing"

	"flip7_strategy/internal/domain"
)

func TestEstimateHitRisk(t *testing.T) {
	// Test case from issue: Deck has only Number 2 and Number 5
	// Hand has [5, 10, 7, 11, freeze]
	// Only Number 5 can cause a bust, so bust rate should be 50% (1/2)
	t.Run("Issue: Deck with 2 number cards and hand with one duplicate", func(t *testing.T) {
		handNumbers := map[domain.NumberValue]struct{}{
			5: {}, 10: {}, 7: {}, 11: {},
		}

		// Create a deck with Number 2 and Number 5
		cards := []domain.Card{
			{Type: domain.CardTypeNumber, Value: 2},
			{Type: domain.CardTypeNumber, Value: 5},
		}
		deck := domain.NewDeckFromCards(cards)

		risk := deck.EstimateHitRisk(handNumbers)
		// Expected: 1 risky card out of 2 total cards = 0.5 (50%)
		expectedRisk := 0.5
		if risk != expectedRisk {
			t.Errorf("Expected risk %.2f (50%%), got %.2f (%.2f%%)", expectedRisk, risk, risk*100)
		}
	})

	// Test with deck containing modifiers and actions
	// Hand has [5, 10]
	// Deck has [Number 5, Number 7, Modifier +2, Action Freeze]
	// Only Number 5 can bust, so bust rate should be 1/2 = 50% (only counting number cards)
	t.Run("Deck with number cards, modifiers and actions", func(t *testing.T) {
		handNumbers := map[domain.NumberValue]struct{}{
			5: {}, 10: {},
		}

		cards := []domain.Card{
			{Type: domain.CardTypeNumber, Value: 5},
			{Type: domain.CardTypeNumber, Value: 7},
			{Type: domain.CardTypeModifier, ModifierType: domain.ModifierPlus2},
			{Type: domain.CardTypeAction, ActionType: domain.ActionFreeze},
		}
		deck := domain.NewDeckFromCards(cards)

		risk := deck.EstimateHitRisk(handNumbers)
		// Expected: 1 risky number card out of 2 total number cards = 0.5 (50%)
		expectedRisk := 0.5
		if risk != expectedRisk {
			t.Errorf("Expected risk %.2f (50%%), got %.2f (%.2f%%)", expectedRisk, risk, risk*100)
		}
	})

	// Test with no number cards in deck (only modifiers and actions)
	// Should have 0% bust rate since no number cards can cause a bust
	t.Run("Deck with only modifiers and actions", func(t *testing.T) {
		handNumbers := map[domain.NumberValue]struct{}{
			5: {}, 10: {},
		}

		cards := []domain.Card{
			{Type: domain.CardTypeModifier, ModifierType: domain.ModifierPlus2},
			{Type: domain.CardTypeAction, ActionType: domain.ActionFreeze},
		}
		deck := domain.NewDeckFromCards(cards)

		risk := deck.EstimateHitRisk(handNumbers)
		expectedRisk := 0.0
		if risk != expectedRisk {
			t.Errorf("Expected risk %.2f (0%%), got %.2f (%.2f%%)", expectedRisk, risk, risk*100)
		}
	})

	// Test with all number cards being risky
	// Hand has [2, 3]
	// Deck has only [Number 2, Number 3]
	// All number cards can bust, so bust rate should be 100%
	t.Run("All number cards are risky", func(t *testing.T) {
		handNumbers := map[domain.NumberValue]struct{}{
			2: {}, 3: {},
		}

		cards := []domain.Card{
			{Type: domain.CardTypeNumber, Value: 2},
			{Type: domain.CardTypeNumber, Value: 3},
		}
		deck := domain.NewDeckFromCards(cards)

		risk := deck.EstimateHitRisk(handNumbers)
		expectedRisk := 1.0
		if risk != expectedRisk {
			t.Errorf("Expected risk %.2f (100%%), got %.2f (%.2f%%)", expectedRisk, risk, risk*100)
		}
	})
}

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
