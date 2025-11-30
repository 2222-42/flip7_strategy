package strategy

import (
	"flip7_strategy/internal/domain"
	"math/rand"
)

// TargetSelector defines the logic for selecting a target for an action.
type TargetSelector interface {
	ChooseTarget(action domain.ActionType, candidates []*domain.Player, self *domain.Player) *domain.Player
	SetDeck(deck *domain.Deck)
}

// DefaultTargetSelector implements the standard target selection logic.
type DefaultTargetSelector struct {
	deck *domain.Deck
}

func NewDefaultTargetSelector() *DefaultTargetSelector {
	return &DefaultTargetSelector{}
}

func (s *DefaultTargetSelector) SetDeck(d *domain.Deck) {
	s.deck = d
}

func (s *DefaultTargetSelector) ChooseTarget(action domain.ActionType, candidates []*domain.Player, self *domain.Player) *domain.Player {
	// Shared logic:
	// Freeze -> Self (if winning and high risk) or opponent with highest score.
	// FlipThree -> High risk opponent (bust probability > 0.8) -> Leader opponent.
	// GiveSecondChance -> Weakest opponent (least threat).

	if action == domain.ActionFreeze {
		return chooseFreezeTarget(candidates, self, s.deck)
	}

	if action == domain.ActionFlipThree {
		// Check for high-risk opponents
		var bestTarget *domain.Player
		highestScore := -1

		// Filter opponents
		opponents := filterOpponents(candidates, self)

		if len(opponents) == 0 {
			return self // Should not happen usually
		}

		// Check risk for each opponent
		for _, p := range opponents {
			risk := 0.0
			if s.deck != nil {
				risk = s.deck.EstimateFlipThreeRisk(p.CurrentHand.NumberCards, p.CurrentHand.HasSecondChance())
			}
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

		// Fallback to Leader logic
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

	return selectLeader(candidates, self)
}

// RiskBasedTargetSelector allows configuring the risk threshold for Flip Three targeting.
type RiskBasedTargetSelector struct {
	DefaultTargetSelector
	FlipThreeRiskThreshold float64
}

func NewRiskBasedTargetSelector(threshold float64) *RiskBasedTargetSelector {
	return &RiskBasedTargetSelector{
		FlipThreeRiskThreshold: threshold,
	}
}

func (s *RiskBasedTargetSelector) ChooseTarget(action domain.ActionType, candidates []*domain.Player, self *domain.Player) *domain.Player {
	if action == domain.ActionFlipThree {
		// Check for high-risk opponents using custom threshold
		var bestTarget *domain.Player
		highestScore := -1

		// Filter opponents
		opponents := filterOpponents(candidates, self)

		if len(opponents) == 0 {
			return self
		}

		// Check risk for each opponent
		for _, p := range opponents {
			risk := 0.0
			if s.deck != nil {
				risk = s.deck.EstimateFlipThreeRisk(p.CurrentHand.NumberCards, p.CurrentHand.HasSecondChance())
			}
			if risk > s.FlipThreeRiskThreshold {
				if p.TotalScore > highestScore {
					highestScore = p.TotalScore
					bestTarget = p
				}
			}
		}

		if bestTarget != nil {
			return bestTarget
		}
		// Fallback to default logic (Leader)
		return selectLeader(candidates, self)
	}

	return s.DefaultTargetSelector.ChooseTarget(action, candidates, self)
}

// selectLeader selects the opponent with the highest score.
func selectLeader(candidates []*domain.Player, self *domain.Player) *domain.Player {
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

// filterOpponents filters out the self player from the candidates list.
func filterOpponents(candidates []*domain.Player, self *domain.Player) []*domain.Player {
	var opponents []*domain.Player
	for _, p := range candidates {
		if p.ID != self.ID {
			opponents = append(opponents, p)
		}
	}
	return opponents
}

// RandomTargetSelector selects targets randomly (for Aggressive strategy).
type RandomTargetSelector struct {
	deck *domain.Deck
}

func NewRandomTargetSelector() *RandomTargetSelector {
	return &RandomTargetSelector{}
}

func (s *RandomTargetSelector) SetDeck(d *domain.Deck) {
	s.deck = d
}

func (s *RandomTargetSelector) ChooseTarget(action domain.ActionType, candidates []*domain.Player, self *domain.Player) *domain.Player {
	if action == domain.ActionFreeze {
		return chooseFreezeTarget(candidates, self, s.deck)
	}

	if action == domain.ActionGiveSecondChance {
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
