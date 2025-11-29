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
			NumberCards: make(map[domain.NumberValue]struct{}),
			ActionCards: []domain.Card{{Type: domain.CardTypeAction, ActionType: domain.ActionSecondChance}},
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
			name:           "Freeze targets leader opponent",
			action:         domain.ActionFreeze,
			candidates:     []*domain.Player{self, opponent1, opponent2},
			expectedTarget: opponent2,
			description:    "Freeze should target opponent with highest score (120)",
		},
		{
			name:           "Freeze targets self if no opponents",
			action:         domain.ActionFreeze,
			candidates:     []*domain.Player{self},
			expectedTarget: self,
			description:    "Freeze should fallback to self if no opponents",
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

func TestChooseTarget_FlipThree_HighRisk(t *testing.T) {
	s := strategy.CommonTargetChooser{}
	self := domain.NewPlayer("Self", nil)

	// Opponent 1: High Score (150), Low Risk (Hand empty)
	op1 := domain.NewPlayer("Op1", nil)
	op1.TotalScore = 150
	op1.CurrentHand = domain.NewPlayerHand()

	// Opponent 2: Low Score (50), High Risk (Hand has 0, 1, 2; Deck has duplicates)
	op2 := domain.NewPlayer("Op2", nil)
	op2.TotalScore = 50
	op2.CurrentHand = domain.NewPlayerHand()
	op2.CurrentHand.NumberCards[domain.NumberValue(0)] = struct{}{}
	op2.CurrentHand.NumberCards[domain.NumberValue(1)] = struct{}{}
	op2.CurrentHand.NumberCards[domain.NumberValue(2)] = struct{}{}

	candidates := []*domain.Player{self, op1, op2}

	// Create a deck that guarantees a bust for Op2
	// Deck has 0, 1, 2. Op2 has 0, 1, 2.
	// Drawing 3 cards will definitely hit a duplicate.
	cards := []domain.Card{
		{Type: domain.CardTypeNumber, Value: 0},
		{Type: domain.CardTypeNumber, Value: 1},
		{Type: domain.CardTypeNumber, Value: 2},
	}
	deck := domain.NewDeckFromCards(cards)

	// Normal logic (without risk check) would target Op1 (Leader).
	// New logic should target Op2 (High Risk).

	s.SetDeck(deck) // Inject deck
	target := s.ChooseTarget(domain.ActionFlipThree, candidates, self)

	if target.ID != op2.ID {
		t.Errorf("Expected target to be Op2 (High Risk), got %s (Score: %d)", target.Name, target.TotalScore)
	}
}
