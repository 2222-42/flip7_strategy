package domain

import (
	"testing"
)

func TestScoreCalculator_Compute(t *testing.T) {
	tests := []struct {
		name     string
		hand     *PlayerHand
		expected int
	}{
		{
			name: "User Example: 5, 8, 9, plus_4, multiply_2",
			hand: &PlayerHand{
				Status:         HandStatusStayed,
				RawNumberCards: []NumberValue{5, 8, 9},
				ModifierCards: []Card{
					{Type: CardTypeModifier, ModifierType: ModifierPlus4},
					{Type: CardTypeModifier, ModifierType: ModifierX2},
				},
				NumberCards: map[NumberValue]struct{}{
					5: {}, 8: {}, 9: {},
				},
			},
			// Expected: (5 + 8 + 9) * 2 + 4 = 22 * 2 + 4 = 48
			expected: 48,
		},
		{
			name: "Only Numbers: 5, 8, 9",
			hand: &PlayerHand{
				Status:         HandStatusStayed,
				RawNumberCards: []NumberValue{5, 8, 9},
				NumberCards: map[NumberValue]struct{}{
					5: {}, 8: {}, 9: {},
				},
			},
			// Expected: 5 + 8 + 9 = 22
			expected: 22,
		},
		{
			name: "Numbers + Multiple Additive Modifiers: 5, plus_2, plus_4",
			hand: &PlayerHand{
				Status:         HandStatusStayed,
				RawNumberCards: []NumberValue{5},
				ModifierCards: []Card{
					{Type: CardTypeModifier, ModifierType: ModifierPlus2},
					{Type: CardTypeModifier, ModifierType: ModifierPlus4},
				},
				NumberCards: map[NumberValue]struct{}{
					5: {},
				},
			},
			// Expected: 5 + 2 + 4 = 11
			expected: 11,
		},
		{
			name: "Numbers + Multiple Multipliers: 5, multiply_2, multiply_2",
			hand: &PlayerHand{
				Status:         HandStatusStayed,
				RawNumberCards: []NumberValue{5},
				ModifierCards: []Card{
					{Type: CardTypeModifier, ModifierType: ModifierX2},
					{Type: CardTypeModifier, ModifierType: ModifierX2},
				},
				NumberCards: map[NumberValue]struct{}{
					5: {},
				},
			},
			// Expected: 5 * 2 * 2 = 20
			expected: 20,
		},
		{
			name: "Mixed: 5, plus_4, multiply_2, multiply_2",
			hand: &PlayerHand{
				Status:         HandStatusStayed,
				RawNumberCards: []NumberValue{5},
				ModifierCards: []Card{
					{Type: CardTypeModifier, ModifierType: ModifierPlus4},
					{Type: CardTypeModifier, ModifierType: ModifierX2},
					{Type: CardTypeModifier, ModifierType: ModifierX2},
				},
				NumberCards: map[NumberValue]struct{}{
					5: {},
				},
			},
			// Expected: (5 * 4) + 4 = 24
			expected: 24,
		},
	}

	calc := NewScoreCalculator()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calc.Compute(tt.hand)
			if result.Total != tt.expected {
				t.Errorf("Compute() total = %v, want %v", result.Total, tt.expected)
			}
		})
	}
}
