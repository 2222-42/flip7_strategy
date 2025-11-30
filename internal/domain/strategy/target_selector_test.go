package strategy_test

import (
	"testing"

	"flip7_strategy/internal/domain"
	"flip7_strategy/internal/domain/strategy"
)

func TestRiskBasedTargetSelector_ChooseTarget_FlipThree(t *testing.T) {
	// Setup
	self := domain.NewPlayer("Self", nil)
	self.TotalScore = 100
	self.CurrentHand = domain.NewPlayerHand()
	self.CurrentHand.NumberCards[domain.NumberValue(1)] = struct{}{} // Self has 1

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

	// Test Case 1: High Threshold (0.8) - Default behavior
	// Op2 risk is 1.0 (guaranteed bust). 1.0 > 0.8, so Op2 should be targeted.
	t.Run("High Threshold (0.8)", func(t *testing.T) {
		selector := strategy.NewRiskBasedTargetSelector(0.8)
		selector.SetDeck(deck)
		target := selector.ChooseTarget(domain.ActionFlipThree, candidates, self)

		if target.ID != op2.ID {
			t.Errorf("Expected target to be Op2 (High Risk), got %s", target.Name)
		}
	})

	// Test Case 2: Very High Threshold (1.1) - Should fallback to Leader
	// Op2 risk is 1.0. 1.0 < 1.1, so Op2 is NOT targeted by risk.
	// Fallback to Leader -> Op1 (150 points)
	t.Run("Very High Threshold (1.1)", func(t *testing.T) {
		selector := strategy.NewRiskBasedTargetSelector(1.1)
		selector.SetDeck(deck)
		target := selector.ChooseTarget(domain.ActionFlipThree, candidates, self)

		if target.ID != op1.ID {
			t.Errorf("Expected target to be Op1 (Leader), got %s", target.Name)
		}
	})

	// Test Case 3: Low Threshold (0.5)
	// Op2 risk is 1.0 > 0.5. Op2 targeted.
	t.Run("Low Threshold (0.5)", func(t *testing.T) {
		selector := strategy.NewRiskBasedTargetSelector(0.5)
		selector.SetDeck(deck)
		target := selector.ChooseTarget(domain.ActionFlipThree, candidates, self)

		if target.ID != op2.ID {
			t.Errorf("Expected target to be Op2 (High Risk), got %s", target.Name)
		}
	})
}

func TestRiskBasedTargetSelector_ChooseTarget_Freeze(t *testing.T) {
	// Setup
	self := domain.NewPlayer("Self", nil)
	self.TotalScore = 150
	self.CurrentHand = domain.NewPlayerHand()

	// Opponent 1: Highest score
	op1 := domain.NewPlayer("Op1", nil)
	op1.TotalScore = 180
	op1.CurrentHand = domain.NewPlayerHand()

	// Opponent 2: Lower score
	op2 := domain.NewPlayer("Op2", nil)
	op2.TotalScore = 100
	op2.CurrentHand = domain.NewPlayerHand()

	candidates := []*domain.Player{self, op1, op2}

	// Test: RiskBasedTargetSelector should delegate to DefaultTargetSelector for Freeze
	// Expected behavior: Target opponent with highest score (Op1)
	t.Run("Freeze delegates to default behavior", func(t *testing.T) {
		selector := strategy.NewRiskBasedTargetSelector(0.8)
		target := selector.ChooseTarget(domain.ActionFreeze, candidates, self)

		if target.ID != op1.ID {
			t.Errorf("Expected target to be Op1 (highest score), got %s", target.Name)
		}
	})

	// Test: When self is winning with high risk, should target self
	t.Run("Freeze self when winning with high risk", func(t *testing.T) {
		// Create a deck that makes continuing risky for self
		// Self has no cards yet, so any duplicate in deck creates risk
		selfWinning := domain.NewPlayer("SelfWinning", nil)
		selfWinning.TotalScore = 200 // Winning
		selfWinning.CurrentHand = domain.NewPlayerHand()
		selfWinning.CurrentHand.NumberCards[domain.NumberValue(1)] = struct{}{}
		selfWinning.CurrentHand.NumberCards[domain.NumberValue(2)] = struct{}{}

		// Create deck with duplicates
		cards := []domain.Card{
			{Type: domain.CardTypeNumber, Value: 1},
			{Type: domain.CardTypeNumber, Value: 2},
		}
		deck := domain.NewDeckFromCards(cards)

		candidatesWinning := []*domain.Player{selfWinning, op1, op2}

		selector := strategy.NewRiskBasedTargetSelector(0.8)
		selector.SetDeck(deck)
		target := selector.ChooseTarget(domain.ActionFreeze, candidatesWinning, selfWinning)

		// Should target self to secure the win (risk > 0.5)
		if target.ID != selfWinning.ID {
			t.Errorf("Expected target to be Self (winning with high risk), got %s", target.Name)
		}
	})
}

func TestRiskBasedTargetSelector_ChooseTarget_GiveSecondChance(t *testing.T) {
	// Setup
	self := domain.NewPlayer("Self", nil)
	self.TotalScore = 150
	self.CurrentHand = domain.NewPlayerHand()

	// Opponent 1: Highest score, no second chance
	op1 := domain.NewPlayer("Op1", nil)
	op1.TotalScore = 180
	op1.CurrentHand = domain.NewPlayerHand()

	// Opponent 2: Lowest score, no second chance
	op2 := domain.NewPlayer("Op2", nil)
	op2.TotalScore = 50
	op2.CurrentHand = domain.NewPlayerHand()

	// Opponent 3: Mid score, already has second chance
	op3 := domain.NewPlayer("Op3", nil)
	op3.TotalScore = 100
	op3.CurrentHand = domain.NewPlayerHand()
	op3.CurrentHand.ActionCards = []domain.Card{
		{Type: domain.CardTypeAction, ActionType: domain.ActionSecondChance},
	}

	candidates := []*domain.Player{self, op1, op2, op3}

	// Test: RiskBasedTargetSelector should delegate to DefaultTargetSelector for GiveSecondChance
	// Expected behavior: Target weakest opponent without second chance (Op2)
	t.Run("GiveSecondChance delegates to default behavior", func(t *testing.T) {
		selector := strategy.NewRiskBasedTargetSelector(0.8)
		target := selector.ChooseTarget(domain.ActionGiveSecondChance, candidates, self)

		if target.ID != op2.ID {
			t.Errorf("Expected target to be Op2 (weakest opponent), got %s", target.Name)
		}
	})

	// Test: Should skip opponents who already have second chance
	t.Run("GiveSecondChance skips opponents with second chance", func(t *testing.T) {
		// Only op3 has second chance, so should target op2 (lowest score without it)
		selector := strategy.NewRiskBasedTargetSelector(0.5)
		target := selector.ChooseTarget(domain.ActionGiveSecondChance, candidates, self)

		if target.ID == op3.ID {
			t.Errorf("Should not target Op3 who already has second chance")
		}
		if target.ID != op2.ID {
			t.Errorf("Expected target to be Op2 (weakest without second chance), got %s", target.Name)
		}
	})
}
