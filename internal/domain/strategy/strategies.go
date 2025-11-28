package strategy

import (
	"flip7_strategy/internal/domain"
	"fmt"
	"math/rand"
)

// CautiousStrategy stays if the risk is even slightly elevated.
type CautiousStrategy struct{}

func (s *CautiousStrategy) Name() string {
	return "Cautious"
}

func (s *CautiousStrategy) Decide(deck *domain.Deck, hand *domain.PlayerHand, playerScore int, otherPlayers []*domain.Player) domain.TurnChoice {
	if hand.HasSecondChance() {
		return domain.TurnChoiceHit
	}
	if len(hand.NumberCards) == 0 {
		return domain.TurnChoiceHit
	}
	calc := domain.NewScoreCalculator()
	score := calc.Compute(hand)
	if score.Total > 30 {
		return domain.TurnChoiceStay
	}
	risk := deck.EstimateHitRisk(hand.NumberCards)
	if risk > 0.10 {
		return domain.TurnChoiceStay
	}
	return domain.TurnChoiceHit
}

func (s *CautiousStrategy) ChooseTarget(action domain.ActionType, candidates []*domain.Player, self *domain.Player, deck *domain.Deck) *domain.Player {
	// Cautious:
	// Freeze -> Self (secure points)
	// FlipThree -> Opponent (avoid risk)
	// GiveSecondChance -> Player with lowest score (keep game balanced)

	if action == domain.ActionFreeze {
		for _, p := range candidates {
			if p.ID == self.ID {
				return p
			}
		}
	}

	if action == domain.ActionGiveSecondChance {
		// GiveSecondChance: Give to player with lowest score to prolong game
		var bestTarget *domain.Player
		minScore := -1 // Initialize with -1 to indicate no target selected yet

		for _, p := range candidates {
			if p.ID != self.ID {
				// Skip players who already have a Second Chance card
				if p.CurrentHand.HasSecondChance() {
					continue
				}
				if minScore == -1 || p.TotalScore < minScore {
					minScore = p.TotalScore
					bestTarget = p
				}
			}
		}

		if bestTarget != nil {
			return bestTarget
		}
		// Fallback: if everyone has one or no opponents, give to first candidate (even if they have one, game service handles discard)
		return candidates[0]
	}

	// For FlipThree (or if self not available for Freeze), target random opponent
	var opponents []*domain.Player
	for _, p := range candidates {
		if p.ID != self.ID {
			opponents = append(opponents, p)
		}
	}
	if len(opponents) > 0 {
		return opponents[rand.Intn(len(opponents))]
	}
	return self
}

// AggressiveStrategy pushes luck until high risk.
type AggressiveStrategy struct{}

// NewAggressiveStrategy returns a new AggressiveStrategy instance.
func NewAggressiveStrategy() *AggressiveStrategy {
	return &AggressiveStrategy{}
}

func (s *AggressiveStrategy) Name() string {
	return "Aggressive"
}

func (s *AggressiveStrategy) Decide(deck *domain.Deck, hand *domain.PlayerHand, playerScore int, otherPlayers []*domain.Player) domain.TurnChoice {
	if hand.HasSecondChance() {
		return domain.TurnChoiceHit
	}
	risk := deck.EstimateHitRisk(hand.NumberCards)
	totalCards := len(hand.NumberCards)
	if totalCards == 6 && risk < 0.5 {
		return domain.TurnChoiceHit
	}
	if risk > 0.30 {
		return domain.TurnChoiceStay
	}
	return domain.TurnChoiceHit
}

func (s *AggressiveStrategy) ChooseTarget(action domain.ActionType, candidates []*domain.Player, self *domain.Player, deck *domain.Deck) *domain.Player {
	// Aggressive:
	// Freeze -> Self
	// FlipThree -> Opponent
	// GiveSecondChance -> Random player

	if action == domain.ActionFreeze {
		for _, p := range candidates {
			if p.ID == self.ID {
				return p
			}
		}
	}

	if action == domain.ActionGiveSecondChance {
		// GiveSecondChance: Give to random player
		return candidates[rand.Intn(len(candidates))]
	}

	var opponents []*domain.Player
	for _, p := range candidates {
		if p.ID != self.ID {
			opponents = append(opponents, p)
		}
	}
	if len(opponents) > 0 {
		return opponents[rand.Intn(len(opponents))]
	}
	return self
}

// CommonTargetChooser implements shared target selection logic.
type CommonTargetChooser struct{}

func (c *CommonTargetChooser) ChooseTarget(action domain.ActionType, candidates []*domain.Player, self *domain.Player, deck *domain.Deck) *domain.Player {
	// Shared logic:
	// Freeze -> Self.
	// FlipThree -> High risk opponent (bust probability > 0.8) -> Leader opponent.
	// GiveSecondChance -> Weakest opponent (least threat).

	if action == domain.ActionFreeze {
		for _, p := range candidates {
			if p.ID == self.ID {
				return p
			}
		}
	}

	if action == domain.ActionFlipThree {
		// Check for high-risk opponents
		var bestTarget *domain.Player
		highestScore := -1

		// Filter opponents
		var opponents []*domain.Player
		for _, p := range candidates {
			if p.ID != self.ID {
				opponents = append(opponents, p)
			}
		}

		if len(opponents) == 0 {
			return self // Should not happen usually
		}

		// Check risk for each opponent
		for _, p := range opponents {
			risk := deck.EstimateFlipThreeRisk(p.CurrentHand.NumberCards, p.CurrentHand.HasSecondChance())
			if risk > 0.8 {
				if p.TotalScore > highestScore {
					highestScore = p.TotalScore
					bestTarget = p
				}
			}
		}

		if bestTarget != nil {
			return bestTarget
		}

		// Fallback to Leader logic (existing logic below)
	}

	if action == domain.ActionGiveSecondChance {
		var bestTarget *domain.Player
		minScore := -1

		for _, p := range candidates {
			if p.ID != self.ID { // Only consider opponents
				if p.CurrentHand.HasSecondChance() {
					continue
				}
				if minScore == -1 || p.TotalScore < minScore {
					minScore = p.TotalScore
					bestTarget = p
				}
			}
		}
		if bestTarget != nil {
			return bestTarget
		}
		return candidates[0]
	}

	bestTarget := self
	maxScore := -1
	for _, p := range candidates {
		if p.ID != self.ID {
			if p.TotalScore > maxScore {
				maxScore = p.TotalScore
				bestTarget = p
			}
		}
	}

	if bestTarget.ID == self.ID && len(candidates) > 1 {
		for _, p := range candidates {
			if p.ID != self.ID {
				return p
			}
		}
	}

	return bestTarget
}

// ProbabilisticStrategy uses expected value (simplified).
type ProbabilisticStrategy struct {
	CommonTargetChooser
}

func (s *ProbabilisticStrategy) Name() string {
	return "Probabilistic"
}

func (s *ProbabilisticStrategy) Decide(deck *domain.Deck, hand *domain.PlayerHand, playerScore int, otherPlayers []*domain.Player) domain.TurnChoice {
	if hand.HasSecondChance() {
		return domain.TurnChoiceHit
	}
	risk := deck.EstimateHitRisk(hand.NumberCards)
	maxOpponentScore := 0
	for _, p := range otherPlayers {
		if p.TotalScore > maxOpponentScore {
			maxOpponentScore = p.TotalScore
		}
	}
	threshold := 0.20
	if playerScore < maxOpponentScore-50 {
		threshold = 0.40
	} else if playerScore > 180 {
		threshold = 0.05
	}
	if risk > threshold {
		return domain.TurnChoiceStay
	}
	return domain.TurnChoiceHit
}

const DefaultHeuristicThreshold = 27

// HeuristicStrategy stops when sum of number cards >= Threshold.
type HeuristicStrategy struct {
	CommonTargetChooser
	Threshold int
}

func NewHeuristicStrategy(threshold int) *HeuristicStrategy {
	return &HeuristicStrategy{Threshold: threshold}
}

func (s *HeuristicStrategy) Name() string {
	return fmt.Sprintf("Heuristic-%d", s.Threshold)
}

func (s *HeuristicStrategy) Decide(deck *domain.Deck, hand *domain.PlayerHand, playerScore int, otherPlayers []*domain.Player) domain.TurnChoice {
	if hand.HasSecondChance() {
		return domain.TurnChoiceHit
	}
	sum := 0
	for val := range hand.NumberCards {
		sum += int(val)
	}

	if sum >= s.Threshold {
		return domain.TurnChoiceStay
	}
	return domain.TurnChoiceHit
}
