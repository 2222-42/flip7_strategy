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

func TestHeuristicStrategy_ChooseTarget(t *testing.T) {
	s := strategy.NewHeuristicStrategy(22)

	// Create test players
	self := &domain.Player{
		ID:         domain.NewPlayer("Self", nil).ID,
		Name:       "Self",
		TotalScore: 100,
		CurrentHand: &domain.PlayerHand{
			NumberCards: make(map[domain.NumberValue]struct{}),
		},
	}

	opponent1 := &domain.Player{
		ID:         domain.NewPlayer("Opponent1", nil).ID,
		Name:       "Opponent1",
		TotalScore: 80,
		CurrentHand: &domain.PlayerHand{
			NumberCards: make(map[domain.NumberValue]struct{}),
		},
	}

	opponent2 := &domain.Player{
		ID:         domain.NewPlayer("Opponent2", nil).ID,
		Name:       "Opponent2",
		TotalScore: 120,
		CurrentHand: &domain.PlayerHand{
			NumberCards: make(map[domain.NumberValue]struct{}),
		},
	}

	opponent3 := &domain.Player{
		ID:         domain.NewPlayer("Opponent3", nil).ID,
		Name:       "Opponent3",
		TotalScore: 90,
		CurrentHand: &domain.PlayerHand{
			NumberCards:  make(map[domain.NumberValue]struct{}),
			ActionCards:  []domain.Card{{Type: domain.CardTypeAction, ActionType: domain.ActionSecondChance}},
		},
	}

	tests := []struct {
		name           string
		action         domain.ActionType
		candidates     []*domain.Player
		expectedTarget *domain.Player
		description    string
	}{
		{
			name:           "Freeze targets self",
			action:         domain.ActionFreeze,
			candidates:     []*domain.Player{self, opponent1, opponent2},
			expectedTarget: self,
			description:    "Freeze should always target self",
		},
		{
			name:           "FlipThree targets leader opponent",
			action:         domain.ActionFlipThree,
			candidates:     []*domain.Player{self, opponent1, opponent2},
			expectedTarget: opponent2,
			description:    "FlipThree should target opponent with highest score (120)",
		},
		{
			name:           "FlipThree targets any opponent if self is leader",
			action:         domain.ActionFlipThree,
			candidates:     []*domain.Player{self, opponent1},
			expectedTarget: opponent1,
			description:    "FlipThree should target opponent even if self has higher score",
		},
		{
			name:           "GiveSecondChance targets weakest opponent",
			action:         domain.ActionGiveSecondChance,
			candidates:     []*domain.Player{self, opponent1, opponent2},
			expectedTarget: opponent1,
			description:    "GiveSecondChance should target opponent with lowest score (80)",
		},
		{
			name:           "GiveSecondChance skips opponents with SecondChance",
			action:         domain.ActionGiveSecondChance,
			candidates:     []*domain.Player{self, opponent1, opponent3, opponent2},
			expectedTarget: opponent1,
			description:    "GiveSecondChance should skip opponent3 (has SecondChance) and target opponent1 (80)",
		},
		{
			name:           "GiveSecondChance fallback when all have SecondChance",
			action:         domain.ActionGiveSecondChance,
			candidates:     []*domain.Player{opponent3},
			expectedTarget: opponent3,
			description:    "GiveSecondChance should fallback to first candidate if all have SecondChance",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			target := s.ChooseTarget(tt.action, tt.candidates, self)
			if target.ID != tt.expectedTarget.ID {
				t.Errorf("%s: Expected target %s (ID: %v), got %s (ID: %v)",
					tt.description, tt.expectedTarget.Name, tt.expectedTarget.ID,
					target.Name, target.ID)
			}
		})
	}
}
