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

	// Create a deck for risk estimation
	// High Risk Deck: Only duplicates of what self has
	highRiskDeck := domain.NewDeckFromCards([]domain.Card{
		{Type: domain.CardTypeNumber, Value: 1},
	})
	// Low Risk Deck: Only new cards
	lowRiskDeck := domain.NewDeckFromCards([]domain.Card{
		{Type: domain.CardTypeNumber, Value: 10},
	})

	// Self has 1
	self.CurrentHand.NumberCards[domain.NumberValue(1)] = struct{}{}

	tests := []struct {
		name           string
		action         domain.ActionType
		candidates     []*domain.Player
		selfScore      int
		deck           *domain.Deck
		expectedTarget *domain.Player
		description    string
	}{
		{
			name:           "Freeze targets leader opponent (Losing)",
			action:         domain.ActionFreeze,
			candidates:     []*domain.Player{self, opponent1, opponent2},
			selfScore:      100,
			deck:           lowRiskDeck,
			expectedTarget: opponent2, // Opponent2 has 120
			description:    "Freeze should target opponent with highest score when losing",
		},
		{
			name:           "Freeze targets self (Winning + High Risk)",
			action:         domain.ActionFreeze,
			candidates:     []*domain.Player{self, opponent1},
			selfScore:      150, // Winning vs 80
			deck:           highRiskDeck,
			expectedTarget: self,
			description:    "Freeze should target self when winning and risk is high",
		},
		{
			name:           "Freeze targets opponent (Winning + Low Risk)",
			action:         domain.ActionFreeze,
			candidates:     []*domain.Player{self, opponent1},
			selfScore:      150, // Winning vs 80
			deck:           lowRiskDeck,
			expectedTarget: opponent1,
			description:    "Freeze should target opponent when winning but risk is low",
		},
		{
			name:           "Freeze targets self if no opponents",
			action:         domain.ActionFreeze,
			candidates:     []*domain.Player{self},
			selfScore:      100,
			deck:           lowRiskDeck,
			expectedTarget: self,
			description:    "Freeze should fallback to self if no opponents",
		},
		{
			name:           "FlipThree targets leader opponent",
			action:         domain.ActionFlipThree,
			candidates:     []*domain.Player{self, opponent1, opponent2},
			selfScore:      100,
			deck:           lowRiskDeck,
			expectedTarget: opponent2,
			description:    "FlipThree should target opponent with highest score (120)",
		},
		{
			name:           "GiveSecondChance targets weakest opponent",
			action:         domain.ActionGiveSecondChance,
			candidates:     []*domain.Player{self, opponent1, opponent2},
			selfScore:      100,
			deck:           lowRiskDeck,
			expectedTarget: opponent1,
			description:    "GiveSecondChance should target opponent with lowest score (80)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			self.TotalScore = tt.selfScore
			s.SetDeck(tt.deck)
			target := s.ChooseTarget(tt.action, tt.candidates, self)
			if target.ID != tt.expectedTarget.ID {
				t.Errorf("%s: Expected target %s (ID: %v), got %s (ID: %v)",
					tt.description, tt.expectedTarget.Name, tt.expectedTarget.ID,
					target.Name, target.ID)
			}
		})
	}
}

func TestNewAggressiveStrategyWithSelector(t *testing.T) {
	// Test that the custom selector is properly initialized
	customSelector := strategy.NewRiskBasedTargetSelector(0.9)
	strat := strategy.NewAggressiveStrategyWithSelector(customSelector)

	if strat == nil {
		t.Fatal("Expected strategy to be initialized")
	}

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
		TotalScore: 50,
		CurrentHand: &domain.PlayerHand{
			NumberCards: make(map[domain.NumberValue]struct{}),
		},
	}

	// Setup opponent2 to have high risk
	opponent2.CurrentHand.NumberCards[domain.NumberValue(0)] = struct{}{}
	opponent2.CurrentHand.NumberCards[domain.NumberValue(1)] = struct{}{}
	opponent2.CurrentHand.NumberCards[domain.NumberValue(2)] = struct{}{}

	candidates := []*domain.Player{self, opponent1, opponent2}

	// Create a deck that guarantees high risk for opponent2
	cards := []domain.Card{
		{Type: domain.CardTypeNumber, Value: 0},
		{Type: domain.CardTypeNumber, Value: 1},
		{Type: domain.CardTypeNumber, Value: 2},
	}
	deck := domain.NewDeckFromCards(cards)
	strat.SetDeck(deck)

	// Test that ChooseTarget uses the custom selector (should target high-risk opponent)
	target := strat.ChooseTarget(domain.ActionFlipThree, candidates, self)

	// With RiskBasedTargetSelector(0.9), it should target the high-risk opponent2
	if target.ID != opponent2.ID {
		t.Errorf("Expected strategy to use custom selector and target high-risk opponent (opponent2), got %s", target.Name)
	}
}

func TestNewProbabilisticStrategyWithSelector(t *testing.T) {
	customSelector := strategy.NewRiskBasedTargetSelector(0.7)
	strat := strategy.NewProbabilisticStrategyWithSelector(customSelector)

	if strat == nil {
		t.Fatal("Expected strategy to be initialized")
	}

	// Test targeting behavior
	self := domain.NewPlayer("Self", nil)
	self.CurrentHand = domain.NewPlayerHand()
	opponent := domain.NewPlayer("Opponent", nil)
	opponent.TotalScore = 150
	opponent.CurrentHand = domain.NewPlayerHand()

	candidates := []*domain.Player{self, opponent}
	deck := domain.NewDeck()
	strat.SetDeck(deck)

	target := strat.ChooseTarget(domain.ActionFlipThree, candidates, self)

	// Should target the opponent (leader)
	if target.ID != opponent.ID {
		t.Errorf("Expected strategy to target opponent, got %v", target.ID)
	}
}

func TestNewHeuristicStrategyWithSelector(t *testing.T) {
	threshold := 25
	customSelector := strategy.NewRiskBasedTargetSelector(0.85)
	strat := strategy.NewHeuristicStrategyWithSelector(threshold, customSelector)

	if strat == nil {
		t.Fatal("Expected strategy to be initialized")
	}

	// Verify threshold is set correctly
	if strat.Name() != "Heuristic-25" {
		t.Errorf("Expected strategy name to be 'Heuristic-25', got %s", strat.Name())
	}

	// Test targeting behavior
	self := domain.NewPlayer("Self", nil)
	self.CurrentHand = domain.NewPlayerHand()
	opponent := domain.NewPlayer("Opponent", nil)
	opponent.TotalScore = 120
	opponent.CurrentHand = domain.NewPlayerHand()

	candidates := []*domain.Player{self, opponent}
	deck := domain.NewDeck()
	strat.SetDeck(deck)

	target := strat.ChooseTarget(domain.ActionFreeze, candidates, self)

	// Should target the opponent (leader)
	if target.ID != opponent.ID {
		t.Errorf("Expected strategy to target opponent, got %v", target.ID)
	}
}

func TestNewExpectedValueStrategyWithSelector(t *testing.T) {
	customSelector := strategy.NewRiskBasedTargetSelector(0.75)
	strat := strategy.NewExpectedValueStrategyWithSelector(customSelector)

	if strat == nil {
		t.Fatal("Expected strategy to be initialized")
	}

	// Test targeting behavior
	self := domain.NewPlayer("Self", nil)
	self.TotalScore = 100
	self.CurrentHand = domain.NewPlayerHand()
	opponent := domain.NewPlayer("Opponent", nil)
	opponent.TotalScore = 150
	opponent.CurrentHand = domain.NewPlayerHand()

	candidates := []*domain.Player{self, opponent}
	deck := domain.NewDeck()
	strat.SetDeck(deck)

	target := strat.ChooseTarget(domain.ActionFlipThree, candidates, self)

	// Should target the opponent (leader)
	if target.ID != opponent.ID {
		t.Errorf("Expected strategy to target opponent, got %v", target.ID)
	}
}

func TestChooseTarget_FlipThree_HighRisk(t *testing.T) {
	s := strategy.DefaultTargetSelector{}
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
