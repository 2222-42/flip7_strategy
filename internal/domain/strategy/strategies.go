package strategy

import (
	"flip7_strategy/internal/domain"
	"fmt"
	"math/rand"
)

// CautiousStrategy stays if the risk is even slightly elevated.
type CautiousStrategy struct {
	deck *domain.Deck
}

func (s *CautiousStrategy) SetDeck(d *domain.Deck) {
	s.deck = d
}

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

func (s *CautiousStrategy) ChooseTarget(action domain.ActionType, candidates []*domain.Player, self *domain.Player) *domain.Player {
	// Cautious:
	// Freeze -> Self (secure points)
	// FlipThree -> Opponent (avoid risk)
	// GiveSecondChance -> Player with lowest score (keep game balanced)

	if action == domain.ActionFreeze {
		return chooseFreezeTarget(candidates, self, s.deck)
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
type AggressiveStrategy struct {
	TargetSelector
}

func (s *AggressiveStrategy) SetDeck(d *domain.Deck) {
	s.TargetSelector.SetDeck(d)
}

// NewAggressiveStrategy returns a new AggressiveStrategy instance.
func NewAggressiveStrategy() *AggressiveStrategy {
	return &AggressiveStrategy{
		TargetSelector: &RandomTargetSelector{},
	}
}

// NewAggressiveStrategyWithSelector returns a new AggressiveStrategy instance with a custom target selector.
func NewAggressiveStrategyWithSelector(selector TargetSelector) *AggressiveStrategy {
	return &AggressiveStrategy{
		TargetSelector: selector,
	}
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

// CommonTargetChooser implements shared target selection logic.
// Deprecated: Use DefaultTargetSelector instead.
type CommonTargetChooser struct {
	TargetSelector
}

func (c *CommonTargetChooser) SetDeck(d *domain.Deck) {
	if c.TargetSelector == nil {
		c.TargetSelector = NewDefaultTargetSelector()
	}
	c.TargetSelector.SetDeck(d)
}

func (c *CommonTargetChooser) ChooseTarget(action domain.ActionType, candidates []*domain.Player, self *domain.Player) *domain.Player {
	if c.TargetSelector == nil {
		c.TargetSelector = NewDefaultTargetSelector()
	}
	return c.TargetSelector.ChooseTarget(action, candidates, self)
}

// ProbabilisticStrategy uses expected value (simplified).
type ProbabilisticStrategy struct {
	CommonTargetChooser
}

// NewProbabilisticStrategyWithSelector returns a new ProbabilisticStrategy instance with a custom target selector.
func NewProbabilisticStrategyWithSelector(selector TargetSelector) *ProbabilisticStrategy {
	return &ProbabilisticStrategy{
		CommonTargetChooser: CommonTargetChooser{TargetSelector: selector},
	}
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

// chooseFreezeTarget encapsulates the logic for selecting a target for ActionFreeze.
func chooseFreezeTarget(candidates []*domain.Player, self *domain.Player, deck *domain.Deck) *domain.Player {
	// Freeze -> Opponent with highest score
	var bestTarget *domain.Player
	maxScore := -1

	for _, p := range candidates {
		if p.ID != self.ID {
			if p.TotalScore > maxScore {
				maxScore = p.TotalScore
				bestTarget = p
			}
		}
	}

	// Refined Logic:
	// If we are winning (score > maxScore) AND risk is high, target Self to secure win.
	// Otherwise, target opponent to stop them.
	if self.TotalScore > maxScore {
		risk := 0.0
		if deck != nil {
			risk = deck.EstimateHitRisk(self.CurrentHand.NumberCards)
		}
		// If risk is high (> 50%), freeze self to be safe.
		if risk > 0.5 {
			return self
		}
		// If risk is low, we might want to continue (so freeze opponent).
	}

	if bestTarget != nil {
		return bestTarget
	}
	return self
}

const DefaultHeuristicThreshold = 27

// HeuristicStrategy stops when sum of number cards >= Threshold.
type HeuristicStrategy struct {
	CommonTargetChooser
	Threshold int
}

func NewHeuristicStrategy(threshold int) *HeuristicStrategy {
	return &HeuristicStrategy{
		CommonTargetChooser: CommonTargetChooser{TargetSelector: NewDefaultTargetSelector()},
		Threshold:           threshold,
	}
}

// NewHeuristicStrategyWithSelector returns a new HeuristicStrategy instance with a custom target selector.
func NewHeuristicStrategyWithSelector(threshold int, selector TargetSelector) *HeuristicStrategy {
	return &HeuristicStrategy{
		CommonTargetChooser: CommonTargetChooser{TargetSelector: selector},
		Threshold:           threshold,
	}
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
