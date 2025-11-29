package application_test

import (
	"testing"

	"flip7_strategy/internal/domain"
)

func TestCanPlayerStay(t *testing.T) {

	tests := []struct {
		name     string
		hand     *domain.PlayerHand
		expected bool
	}{
		{
			name:     "Empty hand",
			hand:     domain.NewPlayerHand(),
			expected: false,
		},
		{
			name: "Only Number cards",
			hand: func() *domain.PlayerHand {
				h := domain.NewPlayerHand()
				h.AddCard(domain.Card{Type: domain.CardTypeNumber, Value: 5})
				return h
			}(),
			expected: true,
		},
		{
			name: "Only Modifier cards (additive)",
			hand: func() *domain.PlayerHand {
				h := domain.NewPlayerHand()
				h.AddCard(domain.Card{Type: domain.CardTypeModifier, ModifierType: domain.ModifierPlus2})
				return h
			}(),
			expected: true,
		},
		{
			name: "Only X2 modifier (no numbers/additives)",
			hand: func() *domain.PlayerHand {
				h := domain.NewPlayerHand()
				h.AddCard(domain.Card{Type: domain.CardTypeModifier, ModifierType: domain.ModifierX2})
				return h
			}(),
			expected: false,
		},
		{
			name: "Only Second Chance (no other actions)",
			hand: func() *domain.PlayerHand {
				h := domain.NewPlayerHand()
				h.AddCard(domain.Card{Type: domain.CardTypeAction, ActionType: domain.ActionSecondChance})
				return h
			}(),
			expected: false,
		},
		{
			name: "Second Chance + Other Action",
			hand: func() *domain.PlayerHand {
				h := domain.NewPlayerHand()
				h.AddCard(domain.Card{Type: domain.CardTypeAction, ActionType: domain.ActionSecondChance})
				h.AddCard(domain.Card{Type: domain.CardTypeAction, ActionType: domain.ActionFreeze})
				return h
			}(),
			expected: true,
		},
		{
			name: "Two Second Chance cards",
			hand: func() *domain.PlayerHand {
				h := domain.NewPlayerHand()
				h.AddCard(domain.Card{Type: domain.CardTypeAction, ActionType: domain.ActionSecondChance})
				h.AddCard(domain.Card{Type: domain.CardTypeAction, ActionType: domain.ActionSecondChance})
				return h
			}(),
			expected: true,
		},
		{
			name: "Other Action only",
			hand: func() *domain.PlayerHand {
				h := domain.NewPlayerHand()
				h.AddCard(domain.Card{Type: domain.CardTypeAction, ActionType: domain.ActionFreeze})
				return h
			}(),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.hand.CanStay(); got != tt.expected {
				t.Errorf("CanStay() = %v, want %v", got, tt.expected)
			}
		})
	}
}
