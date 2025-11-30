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
