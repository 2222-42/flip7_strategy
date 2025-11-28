package strategy

import (
	"flip7_strategy/internal/domain"
)

// AdaptiveStrategy switches behavior based on game state.
// If any opponent has reached the winning threshold (200 points), it becomes Aggressive.
// Otherwise, it plays conservatively using Expected Value.
type AdaptiveStrategy struct {
	CommonTargetChooser
	Aggressive    *AggressiveStrategy
	ExpectedValue *ExpectedValueStrategy
}

func NewAdaptiveStrategy() *AdaptiveStrategy {
	return &AdaptiveStrategy{
		Aggressive:    NewAggressiveStrategy(),
		ExpectedValue: NewExpectedValueStrategy(),
	}
}

func (s *AdaptiveStrategy) Name() string {
	return "Adaptive"
}

func (s *AdaptiveStrategy) Decide(deck *domain.Deck, hand *domain.PlayerHand, playerScore int, otherPlayers []*domain.Player) domain.TurnChoice {
	// Check if any opponent has reached the winning threshold
	opponentThreat := false
	for _, p := range otherPlayers {
		if p.TotalScore >= domain.WinningThreshold {
			opponentThreat = true
			break
		}
	}

	if opponentThreat {
		// Switch to Aggressive mode
		return s.Aggressive.Decide(deck, hand, playerScore, otherPlayers)
	}

	// Default to Expected Value mode
	return s.ExpectedValue.Decide(deck, hand, playerScore, otherPlayers)
}
