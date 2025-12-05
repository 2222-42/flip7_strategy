package application

import (
	"testing"

	"flip7_strategy/internal/domain"
)

// TestRemoveCardFromDeck_AllCopiesDrawn tests that removeCardFromDeck correctly
// returns an error when trying to remove a card that has already been fully drawn.
func TestRemoveCardFromDeck_AllCopiesDrawn(t *testing.T) {
	tests := []struct {
		name          string
		cardValue     domain.NumberValue
		expectedCount int
	}{
		{
			name:          "Card 1 (1 copy)",
			cardValue:     domain.NumberValue(1),
			expectedCount: 1,
		},
		{
			name:          "Card 6 (6 copies)",
			cardValue:     domain.NumberValue(6),
			expectedCount: 6,
		},
		{
			name:          "Card 12 (12 copies)",
			cardValue:     domain.NumberValue(12),
			expectedCount: 12,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a manual game service with a fresh deck
			game := domain.NewGame([]*domain.Player{
				domain.NewPlayer("Player1", nil),
			})
			deck := domain.NewDeck()
			game.CurrentRound = domain.NewRound(game.Players, game.Players[0], deck)

			service := &ManualGameService{
				Game: game,
			}

			card := domain.Card{Type: domain.CardTypeNumber, Value: tt.cardValue}

			// Remove all copies of the card
			for i := 0; i < tt.expectedCount; i++ {
				err := service.removeCardFromDeck(card)
				if err != nil {
					t.Fatalf("Failed to remove card %d on attempt %d/%d: %v", tt.cardValue, i+1, tt.expectedCount, err)
				}
			}

			// Verify RemainingCounts is 0
			if deck.RemainingCounts[tt.cardValue] != 0 {
				t.Errorf("RemainingCounts[%d] = %d, want 0", tt.cardValue, deck.RemainingCounts[tt.cardValue])
			}

			// Try to remove one more copy - should fail with pre-validation check
			err := service.removeCardFromDeck(card)
			if err == nil {
				t.Errorf("Expected error when removing card %d after all copies drawn, but got nil", tt.cardValue)
			}

			// Verify RemainingCounts didn't go negative
			if deck.RemainingCounts[tt.cardValue] < 0 {
				t.Errorf("RemainingCounts[%d] = %d, should not be negative", tt.cardValue, deck.RemainingCounts[tt.cardValue])
			}
		})
	}
}

// TestRemoveCardFromDeck_RemainingCountsValidation tests that the pre-validation
// check prevents removing cards when RemainingCounts is already at 0.
func TestRemoveCardFromDeck_RemainingCountsValidation(t *testing.T) {
	game := domain.NewGame([]*domain.Player{
		domain.NewPlayer("Player1", nil),
	})
	deck := domain.NewDeck()
	game.CurrentRound = domain.NewRound(game.Players, game.Players[0], deck)

	service := &ManualGameService{
		Game: game,
	}

	// Manually set RemainingCounts to 0 for card 7
	deck.RemainingCounts[domain.NumberValue(7)] = 0

	card := domain.Card{Type: domain.CardTypeNumber, Value: domain.NumberValue(7)}

	// Try to remove card 7 - should fail immediately due to pre-validation
	err := service.removeCardFromDeck(card)
	if err == nil {
		t.Error("Expected error when RemainingCounts is 0, but got nil")
	}

	// Verify the error message
	expectedMsg := "card not found in deck (already drawn?)"
	if err.Error() != expectedMsg {
		t.Errorf("Error message = %q, want %q", err.Error(), expectedMsg)
	}
}
