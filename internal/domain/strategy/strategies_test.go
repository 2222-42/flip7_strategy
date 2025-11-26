package strategy_test

import (
	"testing"

	"flip7_strategy/internal/domain"
	"flip7_strategy/internal/domain/strategy"
)

func TestHeuristicStrategy_Decide(t *testing.T) {
	s := strategy.NewHeuristicStrategy(22)
	deck := &domain.Deck{} // Deck state doesn't matter for HeuristicStrategy

	tests := []struct {
		name           string
		handNumbers    []int
		expectedChoice domain.TurnChoice
	}{
		{
			name:           "Sum < 22 (0)",
			handNumbers:    []int{},
			expectedChoice: domain.TurnChoiceHit,
		},
		{
			name:           "Sum < 22 (21)",
			handNumbers:    []int{10, 11},
			expectedChoice: domain.TurnChoiceHit,
		},
		{
			name:           "Sum = 22",
			handNumbers:    []int{10, 12},
			expectedChoice: domain.TurnChoiceStay,
		},
		{
			name:           "Sum > 22 (23)",
			handNumbers:    []int{11, 12},
			expectedChoice: domain.TurnChoiceStay,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hand := &domain.PlayerHand{
				NumberCards: make(map[domain.NumberValue]struct{}),
			}
			for _, n := range tt.handNumbers {
				hand.NumberCards[domain.NumberValue(n)] = struct{}{}
			}

			choice := s.Decide(deck, hand, 0, nil)
			if choice != tt.expectedChoice {
				t.Errorf("Expected %v, got %v", tt.expectedChoice, choice)
			}
		})
	}
}
