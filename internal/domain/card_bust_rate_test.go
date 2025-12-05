package domain_test

import (
	"flip7_strategy/internal/domain"
	"testing"
)

func TestEstimateHitRiskWithAllCardsDrawn(t *testing.T) {
	// Test that bust rate is 0% when all copies of a card are drawn
	deck := domain.NewDeck()

	// Draw all 6 copies of card 6
	for i := 0; i < 6; i++ {
		for j, c := range deck.Cards {
			if c.Type == domain.CardTypeNumber && c.Value == domain.NumberValue(6) {
				deck.Cards = append(deck.Cards[:j], deck.Cards[j+1:]...)
				deck.RemainingCounts[domain.NumberValue(6)]--
				break
			}
		}
	}

	// Verify card 6 is completely drawn
	if deck.RemainingCounts[domain.NumberValue(6)] != 0 {
		t.Errorf("Expected 0 card 6s remaining, got %d", deck.RemainingCounts[domain.NumberValue(6)])
	}

	// Create a hand with card 6
	handNumbers := make(map[domain.NumberValue]struct{})
	handNumbers[domain.NumberValue(6)] = struct{}{}

	// Calculate bust rate - should be 0% since no 6s remain
	risk := deck.EstimateHitRisk(handNumbers)
	if risk != 0.0 {
		t.Errorf("Expected bust rate 0%%, got %.2f%%", risk*100)
	}
}

func TestEstimateHitRiskWithHighValueCards(t *testing.T) {
	tests := []struct {
		name         string
		cardValue    domain.NumberValue
		copiesInDeck int
		copiesToDraw int
		expectedRisk float64
		tolerance    float64
	}{
		{
			name:         "Card 12 with 1 copy remaining out of 12",
			cardValue:    domain.NumberValue(12),
			copiesInDeck: 12,
			copiesToDraw: 11,
			expectedRisk: 0.0, // Will calculate based on actual deck size
			tolerance:    0.01,
		},
		{
			name:         "Card 12 with all copies drawn",
			cardValue:    domain.NumberValue(12),
			copiesInDeck: 12,
			copiesToDraw: 12,
			expectedRisk: 0.0,
			tolerance:    0.0,
		},
		{
			name:         "Card 7 with all copies drawn",
			cardValue:    domain.NumberValue(7),
			copiesInDeck: 7,
			copiesToDraw: 7,
			expectedRisk: 0.0,
			tolerance:    0.0,
		},
		{
			name:         "Card 8 with half drawn",
			cardValue:    domain.NumberValue(8),
			copiesInDeck: 8,
			copiesToDraw: 4,
			expectedRisk: 0.0, // Will calculate
			tolerance:    0.01,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deck := domain.NewDeck()
			originalTotal := len(deck.Cards)

			// Draw specified number of copies
			for i := 0; i < tt.copiesToDraw; i++ {
				for j, c := range deck.Cards {
					if c.Type == domain.CardTypeNumber && c.Value == tt.cardValue {
						deck.Cards = append(deck.Cards[:j], deck.Cards[j+1:]...)
						deck.RemainingCounts[tt.cardValue]--
						break
					}
				}
			}

			// Verify the count
			expectedRemaining := tt.copiesInDeck - tt.copiesToDraw
			if deck.RemainingCounts[tt.cardValue] != expectedRemaining {
				t.Errorf("Expected %d cards remaining, got %d", expectedRemaining, deck.RemainingCounts[tt.cardValue])
			}

			// Create hand with the card
			handNumbers := make(map[domain.NumberValue]struct{})
			handNumbers[tt.cardValue] = struct{}{}

			// Calculate bust rate
			risk := deck.EstimateHitRisk(handNumbers)

			// Calculate expected risk if not specified
			if tt.expectedRisk == 0.0 && tt.copiesToDraw < tt.copiesInDeck {
				cardsDrawn := tt.copiesToDraw
				newTotal := originalTotal - cardsDrawn
				tt.expectedRisk = float64(expectedRemaining) / float64(newTotal)
			}

			// Allow for floating point tolerance
			if tt.copiesToDraw == tt.copiesInDeck {
				// All drawn, must be exactly 0
				if risk != 0.0 {
					t.Errorf("Expected bust rate 0%% (all cards drawn), got %.2f%%", risk*100)
				}
			} else {
				// Compare with tolerance
				diff := risk - tt.expectedRisk
				if diff < 0 {
					diff = -diff
				}
				if diff > tt.tolerance {
					t.Errorf("Expected bust rate %.2f%%, got %.2f%% (diff: %.2f%%)", tt.expectedRisk*100, risk*100, diff*100)
				}
			}
		})
	}
}
