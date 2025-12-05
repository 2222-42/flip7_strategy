package domain

import (
	"testing"
)

// TestDeckInitialization_RemainingCounts verifies that RemainingCounts
// is properly initialized for all number cards 0-12.
func TestDeckInitialization_RemainingCounts(t *testing.T) {
	deck := NewDeck()

	// Expected counts for each number card
	expected := map[NumberValue]int{
		0:  1,
		1:  1,
		2:  2,
		3:  3,
		4:  4,
		5:  5,
		6:  6,
		7:  7,
		8:  8,
		9:  9,
		10: 10,
		11: 11,
		12: 12,
	}

	// Check that RemainingCounts has all expected entries
	for val := NumberValue(0); val <= 12; val++ {
		count, exists := deck.RemainingCounts[val]
		if !exists {
			t.Errorf("RemainingCounts missing entry for card value %d", val)
			continue
		}
		if count != expected[val] {
			t.Errorf("RemainingCounts[%d] = %d, want %d", val, count, expected[val])
		}
	}

	// Verify actual card counts match RemainingCounts
	actualCounts := make(map[NumberValue]int)
	for _, card := range deck.Cards {
		if card.Type == CardTypeNumber {
			actualCounts[card.Value]++
		}
	}

	for val := NumberValue(0); val <= 12; val++ {
		if actualCounts[val] != deck.RemainingCounts[val] {
			t.Errorf("Actual count for card %d (%d) doesn't match RemainingCounts (%d)",
				val, actualCounts[val], deck.RemainingCounts[val])
		}
	}
}

// TestDeckInitialization_TotalCards verifies the deck has the correct total number of cards.
func TestDeckInitialization_TotalCards(t *testing.T) {
	deck := NewDeck()

	// Count expected number of cards
	// Numbers: 0:1 + 1:1 + 2:2 + 3:3 + 4:4 + 5:5 + 6:6 + 7:7 + 8:8 + 9:9 + 10:10 + 11:11 + 12:12 = 79
	// Modifiers: 6 types * 2 = 12
	// Actions: 3 types * 3 = 9
	// Total = 100
	expectedTotal := 100

	if len(deck.Cards) != expectedTotal {
		t.Errorf("Deck has %d cards, expected %d", len(deck.Cards), expectedTotal)
	}
}
