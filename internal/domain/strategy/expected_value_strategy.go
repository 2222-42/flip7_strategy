package strategy

import (
	"fmt"
	"flip7_strategy/internal/domain"
)

// ExpectedValueStrategy calculates the Expected Exploration Value (EEV) of the next hit.
// EEV = Upside (expected gain) - Downside (expected loss).
type ExpectedValueStrategy struct {
	TargetSelector
	RiskTolerance float64 // How much negative EEV we are willing to accept. Negative values mean more risky.
}

// NewExpectedValueStrategy returns a new ExpectedValueStrategy instance with default risk tolerance (0.0).
func NewExpectedValueStrategy(riskTolerance float64) *ExpectedValueStrategy {
	return &ExpectedValueStrategy{
		TargetSelector: NewDefaultTargetSelector(),
		RiskTolerance:  riskTolerance,
	}
}

// NewExpectedValueStrategyWithSelector returns a new ExpectedValueStrategy instance with a custom target selector.
func NewExpectedValueStrategyWithSelector(riskTolerance float64, selector TargetSelector) *ExpectedValueStrategy {
	return &ExpectedValueStrategy{
		TargetSelector: selector,
		RiskTolerance:  riskTolerance,
	}
}

func (s *ExpectedValueStrategy) Name() string {
	if s.RiskTolerance == 0 {
		return "ExpectedValue"
	}
	return fmt.Sprintf("ExpectedValue(%.1f)", s.RiskTolerance)
}

func (s *ExpectedValueStrategy) Decide(deck *domain.Deck, hand *domain.PlayerHand, _ int, _ []*domain.Player) domain.TurnChoice {
	if hand.HasSecondChance() {
		return domain.TurnChoiceHit
	}
	// If deck is empty, must stay (though game logic usually handles this)
	if len(deck.Cards) == 0 {
		return domain.TurnChoiceStay
	}

	// Calculate current score
	calc := domain.NewScoreCalculator()
	currentScore := calc.Compute(hand).Total

	totalCards := len(deck.Cards)
	bustCards := 0
	expectedGainSum := 0.0

	// We iterate over all remaining cards in the deck to calculate probabilities exactly.
	// This simulates perfect card counting and expected value calculation based on the remaining deck.
	for _, card := range deck.Cards {
		clonedHand := hand.Clone()
		busted, _, _ := clonedHand.AddCard(card)

		if busted {
			bustCards++
		} else {
			// Calculate new score and gain
			newScore := calc.Compute(clonedHand).Total
			gain := newScore - currentScore
			expectedGainSum += float64(gain)
		}
	}

	if totalCards == 0 {
		return domain.TurnChoiceStay
	}

	// Upside: Expected average gain from drawing a non-busting card.
	upside := expectedGainSum / float64(totalCards)

	// Downside: Expected loss from busting.
	downside := float64(currentScore) * (float64(bustCards) / float64(totalCards))

	// Expected Exploration Value (EEV)
	eev := upside - downside

	// Hit if EEV plus our risk tolerance is greater than 0.
	// Positive RiskTolerance makes us more conservative (forces EEV > positive threshold).
	// Negative RiskTolerance makes us more risky (allows hitting even when EEV is slightly negative).
	if eev > s.RiskTolerance {
		return domain.TurnChoiceHit
	}

	return domain.TurnChoiceStay
}

