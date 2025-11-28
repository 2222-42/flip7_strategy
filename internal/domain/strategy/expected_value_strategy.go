package strategy

import (
	"flip7_strategy/internal/domain"
)

// ExpectedValueStrategy calculates the expected value of the next hit.
type ExpectedValueStrategy struct {
	CommonTargetChooser
}

func (s *ExpectedValueStrategy) Name() string {
	return "ExpectedValue"
}

func (s *ExpectedValueStrategy) Decide(deck *domain.Deck, hand *domain.PlayerHand, _ int, _ []*domain.Player) domain.TurnChoice {
	// If deck is empty, must stay (though game logic usually handles this)
	if len(deck.Cards) == 0 {
		return domain.TurnChoiceStay
	}

	// Calculate current score
	calc := domain.NewScoreCalculator()
	currentScore := calc.Compute(hand).Total

	// Calculate Expected Value of the next card
	totalEV := 0.0
	totalCards := 0

	// We need to iterate over all possible next cards.
	// Since we don't have direct access to the deck's remaining cards list efficiently without peeking,
	// we rely on RemainingCounts for numbers and knowledge of deck composition for others.
	// However, the Deck struct exposes Cards slice, so we can iterate over it.
	// But iterating over the shuffled deck is cheating? No, the strategy receives the *Deck.
	// In a real game, you count cards. The Deck struct has RemainingCounts which simulates card counting.
	// But for modifiers and actions, we might need to know how many are left if we want to be precise.
	// The Deck struct in `card.go` has `RemainingCounts` only for numbers.
	// For this implementation, let's assume we can see the remaining cards in the deck (perfect counting)
	// or we should approximate. The issue description says "based on the left deck".
	// Let's use the `Cards` slice in the deck which represents the actual remaining cards.
	// This is equivalent to perfect card counting.

	for _, card := range deck.Cards {
		totalCards++

		// Simulate adding this card
		clonedHand := hand.Clone()
		busted, _, _ := clonedHand.AddCard(card)

		if busted {
			totalEV += 0 // Score becomes 0 if busted
		} else {
			// Calculate new score
			newScore := calc.Compute(clonedHand).Total
			totalEV += float64(newScore)
		}
	}

	if totalCards == 0 {
		return domain.TurnChoiceStay
	}

	averageEV := totalEV / float64(totalCards)

	// Decision logic:
	// If EV > Current Score, it means on average we improve our score.
	// However, we should also consider the risk of busting vs the reward.
	// Basic EV strategy: Hit if EV > Current Score.

	// Refinement from issue description: "affordable to be busted".
	// If we are far behind, we might need to take risks even if EV is slightly lower?
	// But strictly speaking, if EV < Current Score, hitting is a bad move on average.

	if averageEV > float64(currentScore) {
		return domain.TurnChoiceHit
	}

	return domain.TurnChoiceStay
}
