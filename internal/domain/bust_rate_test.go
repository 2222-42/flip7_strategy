package domain

import (
	"testing"
)

// TestEstimateHitRisk_NegativeCountProtection verifies that EstimateHitRisk
// handles negative RemainingCounts gracefully (defensive programming).
func TestEstimateHitRisk_NegativeCountProtection(t *testing.T) {
	// Create a minimal deck
	cards := []Card{
		{Type: CardTypeNumber, Value: 1},
		{Type: CardTypeNumber, Value: 2},
		{Type: CardTypeNumber, Value: 3},
	}
	deck := NewDeckFromCards(cards)

	// Artificially set a negative count (simulating a bug scenario)
	deck.RemainingCounts[NumberValue(1)] = -1

	// Create a hand with card 1
	handNumbers := map[NumberValue]struct{}{
		NumberValue(1): {},
	}

	// Calculate risk - should not include the negative count
	risk := deck.EstimateHitRisk(handNumbers)

	// Expected: 0 / 3 = 0 (negative count should be treated as 0)
	if risk != 0 {
		t.Errorf("EstimateHitRisk with negative count = %f, want 0", risk)
	}
}

// TestEstimateHitRisk_ZeroCount verifies correct handling of zero counts.
func TestEstimateHitRisk_ZeroCount(t *testing.T) {
	cards := []Card{
		{Type: CardTypeNumber, Value: 2},
		{Type: CardTypeNumber, Value: 3},
	}
	deck := NewDeckFromCards(cards)

	// Set card 1 count to 0 (all drawn)
	deck.RemainingCounts[NumberValue(1)] = 0

	handNumbers := map[NumberValue]struct{}{
		NumberValue(1): {},
	}

	risk := deck.EstimateHitRisk(handNumbers)

	// Expected: 0 / 2 = 0
	if risk != 0 {
		t.Errorf("EstimateHitRisk with zero count = %f, want 0", risk)
	}
}

// TestEstimateHitRisk_PositiveCount verifies normal operation with positive counts.
func TestEstimateHitRisk_PositiveCount(t *testing.T) {
	cards := []Card{
		{Type: CardTypeNumber, Value: 1},
		{Type: CardTypeNumber, Value: 1},
		{Type: CardTypeNumber, Value: 2},
		{Type: CardTypeNumber, Value: 3},
	}
	deck := NewDeckFromCards(cards)

	handNumbers := map[NumberValue]struct{}{
		NumberValue(1): {},
	}

	risk := deck.EstimateHitRisk(handNumbers)

	// Expected: 2 / 4 = 0.5
	expected := 0.5
	if risk != expected {
		t.Errorf("EstimateHitRisk = %f, want %f", risk, expected)
	}
}
