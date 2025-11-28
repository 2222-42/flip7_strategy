package domain

import (
	"testing"
)

func TestPlayerHand_AddCard_Flip7(t *testing.T) {
	tests := []struct {
		name          string
		initialCards  []Card
		newCard       Card
		wantFlip7     bool
		wantBusted    bool
		wantDiscarded int
	}{
		{
			name: "Flip 7 with 7 number cards",
			initialCards: []Card{
				{Type: CardTypeNumber, Value: 1},
				{Type: CardTypeNumber, Value: 2},
				{Type: CardTypeNumber, Value: 3},
				{Type: CardTypeNumber, Value: 4},
				{Type: CardTypeNumber, Value: 5},
				{Type: CardTypeNumber, Value: 6},
			},
			newCard:    Card{Type: CardTypeNumber, Value: 7},
			wantFlip7:  true,
			wantBusted: false,
		},
		{
			name: "Not Flip 7 with 6 number cards and 1 modifier",
			initialCards: []Card{
				{Type: CardTypeNumber, Value: 1},
				{Type: CardTypeNumber, Value: 2},
				{Type: CardTypeNumber, Value: 3},
				{Type: CardTypeNumber, Value: 4},
				{Type: CardTypeNumber, Value: 5},
				{Type: CardTypeModifier, ModifierType: ModifierPlus2},
			},
			newCard:    Card{Type: CardTypeNumber, Value: 6},
			wantFlip7:  false, // Should be false because only 6 number cards
			wantBusted: false,
		},
		{
			name: "Not Flip 7 with 6 number cards and 1 action",
			initialCards: []Card{
				{Type: CardTypeNumber, Value: 1},
				{Type: CardTypeNumber, Value: 2},
				{Type: CardTypeNumber, Value: 3},
				{Type: CardTypeNumber, Value: 4},
				{Type: CardTypeNumber, Value: 5},
				{Type: CardTypeAction, ActionType: ActionFreeze},
			},
			newCard:    Card{Type: CardTypeNumber, Value: 6},
			wantFlip7:  false, // Should be false because only 6 number cards
			wantBusted: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewPlayerHand()
			for _, c := range tt.initialCards {
				h.AddCard(c)
			}

			busted, flip7, discarded := h.AddCard(tt.newCard)

			if flip7 != tt.wantFlip7 {
				t.Errorf("Flip7 mismatch: got %v, want %v", flip7, tt.wantFlip7)
			}
			if busted != tt.wantBusted {
				t.Errorf("Busted mismatch: got %v, want %v", busted, tt.wantBusted)
			}
			if len(discarded) != tt.wantDiscarded {
				t.Errorf("Discarded count mismatch: got %d, want %d", len(discarded), tt.wantDiscarded)
			}
		})
	}
}
