package application

import (
	"bufio"
	"strings"
	"testing"

	"flip7_strategy/internal/domain"
)

// removeAllCardsOfValue removes all cards of a given value from the deck
func removeAllCardsOfValue(deck *domain.Deck, value domain.NumberValue, count int) {
	for removed := 0; removed < count; {
		for i := 0; i < len(deck.Cards); i++ {
			if deck.Cards[i].Type == domain.CardTypeNumber && deck.Cards[i].Value == value {
				deck.Cards = append(deck.Cards[:i], deck.Cards[i+1:]...)
				deck.RemainingCounts[value]--
				removed++
				break
			}
		}
	}
}

func TestRemoveCardFromDeck(t *testing.T) {
	tests := []struct {
		name         string
		setupDeck    func() *domain.Deck
		cardToRemove domain.Card
		wantError    bool
		errorMsg     string
	}{
		{
			name: "Remove valid number card from deck",
			setupDeck: func() *domain.Deck {
				return domain.NewDeck()
			},
			cardToRemove: domain.Card{Type: domain.CardTypeNumber, Value: domain.NumberValue(5)},
			wantError:    false,
		},
		{
			name: "Remove number card when all copies already drawn",
			setupDeck: func() *domain.Deck {
				deck := domain.NewDeck()
				// Card 1 has only 1 copy, remove it
				removeAllCardsOfValue(deck, domain.NumberValue(1), 1)
				return deck
			},
			cardToRemove: domain.Card{Type: domain.CardTypeNumber, Value: domain.NumberValue(1)},
			wantError:    true,
			errorMsg:     "card not found in deck (already drawn?)",
		},
		{
			name: "Remove number card 6 when all 6 copies already drawn",
			setupDeck: func() *domain.Deck {
				deck := domain.NewDeck()
				// Card 6 has 6 copies, remove all 6
				removeAllCardsOfValue(deck, domain.NumberValue(6), 6)
				return deck
			},
			cardToRemove: domain.Card{Type: domain.CardTypeNumber, Value: domain.NumberValue(6)},
			wantError:    true,
			errorMsg:     "card not found in deck (already drawn?)",
		},
		{
			name: "Remove number card 12 when all 12 copies already drawn",
			setupDeck: func() *domain.Deck {
				deck := domain.NewDeck()
				// Card 12 has 12 copies, remove all 12
				removeAllCardsOfValue(deck, domain.NumberValue(12), 12)
				return deck
			},
			cardToRemove: domain.Card{Type: domain.CardTypeNumber, Value: domain.NumberValue(12)},
			wantError:    true,
			errorMsg:     "card not found in deck (already drawn?)",
		},
		{
			name: "Remove number card 7 when 2 out of 7 copies remaining",
			setupDeck: func() *domain.Deck {
				deck := domain.NewDeck()
				// Card 7 has 7 copies, remove 5
				removeAllCardsOfValue(deck, domain.NumberValue(7), 5)
				return deck
			},
			cardToRemove: domain.Card{Type: domain.CardTypeNumber, Value: domain.NumberValue(7)},
			wantError:    false,
		},
		{
			name: "Remove modifier card when available",
			setupDeck: func() *domain.Deck {
				return domain.NewDeck()
			},
			cardToRemove: domain.Card{Type: domain.CardTypeModifier, ModifierType: domain.ModifierPlus2},
			wantError:    false,
		},
		{
			name: "Remove action card when available",
			setupDeck: func() *domain.Deck {
				return domain.NewDeck()
			},
			cardToRemove: domain.Card{Type: domain.CardTypeAction, ActionType: domain.ActionFreeze},
			wantError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup service with a game and round
			reader := bufio.NewReader(strings.NewReader(""))
			service := NewManualGameService(reader, nil)

			// Create a simple game with one player
			players := []*domain.Player{domain.NewPlayer("Test", nil)}
			service.Game = domain.NewGame(players)

			// Create a round with the test deck
			deck := tt.setupDeck()
			service.Game.CurrentRound = domain.NewRound(players, players[0], deck)

			// Try to remove the card
			err := service.removeCardFromDeck(tt.cardToRemove)

			if tt.wantError {
				if err == nil {
					t.Errorf("Expected error but got nil")
				} else if tt.errorMsg != "" && err.Error() != tt.errorMsg {
					t.Errorf("Expected error message '%s', got '%s'", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}
