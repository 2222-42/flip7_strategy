package strategy

import (
	"flip7_strategy/internal/domain"
	"math/rand"
)

// CautiousStrategy stays if the risk is even slightly elevated.
type CautiousStrategy struct{}

func (s *CautiousStrategy) Name() string {
	return "Cautious"
}

func (s *CautiousStrategy) Decide(deck *domain.Deck, hand *domain.PlayerHand, playerScore int, otherPlayers []*domain.Player) domain.TurnChoice {
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

func (s *CautiousStrategy) ChooseTarget(action domain.ActionType, candidates []*domain.Player, self *domain.Player) *domain.Player {
	// Cautious:
	// Freeze -> Self (secure points)
	// FlipThree -> Opponent (avoid risk)
	// GiveSecondChance -> Teammate? (No teams). Give to weakest player? Or random?
	// Let's give to the player with lowest score to keep game balanced/prolonged?

	if action == domain.ActionFreeze {
		for _, p := range candidates {
			if p.ID == self.ID {
				return p
			}
		}
	}

	if action == domain.ActionGiveSecondChance {
		// GiveSecondChance: Give to player with lowest score to prolong game
		bestTarget := candidates[0] // Default, will be self if no opponents
		minScore := 10000
		for _, p := range candidates {
			if p.ID != self.ID {
				if p.TotalScore < minScore {
					minScore = p.TotalScore
					bestTarget = p
				}
			}
		}
		return bestTarget
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

func (s *AggressiveStrategy) Name() string {
	return "Aggressive"
}

func (s *AggressiveStrategy) Decide(deck *domain.Deck, hand *domain.PlayerHand, playerScore int, otherPlayers []*domain.Player) domain.TurnChoice {
	risk := deck.EstimateHitRisk(hand.NumberCards)
	totalCards := len(hand.RawNumberCards) + len(hand.ModifierCards) + len(hand.ActionCards)
	if totalCards == 6 && risk < 0.5 {
		return domain.TurnChoiceHit
	}
	if risk > 0.30 {
		return domain.TurnChoiceStay
	}
	return domain.TurnChoiceHit
}

func (s *AggressiveStrategy) ChooseTarget(action domain.ActionType, candidates []*domain.Player, self *domain.Player) *domain.Player {
	// Aggressive:
	// Freeze -> Self
	// FlipThree -> Opponent
	// GiveSecondChance -> Random? Or Strongest to keep them in check? (Doesn't make sense)
	// Give to random.

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

// ProbabilisticStrategy uses expected value (simplified).
type ProbabilisticStrategy struct{}

func (s *ProbabilisticStrategy) Name() string {
	return "Probabilistic"
}

func (s *ProbabilisticStrategy) Decide(deck *domain.Deck, hand *domain.PlayerHand, playerScore int, otherPlayers []*domain.Player) domain.TurnChoice {
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

func (s *ProbabilisticStrategy) ChooseTarget(action domain.ActionType, candidates []*domain.Player, self *domain.Player) *domain.Player {
	// Probabilistic:
	// Freeze -> Self.
	// FlipThree -> Leader opponent.
	// GiveSecondChance -> Weakest opponent (least threat).

	if action == domain.ActionFreeze {
		for _, p := range candidates {
			if p.ID == self.ID {
				return p
			}
		}
	}

	if action == domain.ActionGiveSecondChance {
		bestTarget := candidates[0] // Default, will be self if no opponents
		minScore := 10000
		for _, p := range candidates {
			if p.ID != self.ID { // Only consider opponents
				if p.TotalScore < minScore {
					minScore = p.TotalScore
					bestTarget = p
				}
			}
		}
		return bestTarget
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
