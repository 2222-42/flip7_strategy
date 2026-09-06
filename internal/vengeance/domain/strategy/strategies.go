package strategy

import (
	"fmt"

	"flip7_strategy/internal/vengeance/domain"
)

const DefaultHeuristicThreshold = 26

type CautiousStrategy struct {
	*DefaultTargetSelector
}

func NewCautiousStrategy() *CautiousStrategy {
	return &CautiousStrategy{DefaultTargetSelector: NewDefaultTargetSelector()}
}

func (s *CautiousStrategy) Name() string { return "Cautious" }

func (s *CautiousStrategy) Decide(ctx domain.DecisionContext) domain.TurnChoice {
	if shouldAlwaysHit(ctx) {
		return domain.TurnChoiceHit
	}
	if ctx.Hand.NumberSum() > 18 {
		return domain.TurnChoiceStay
	}
	if ctx.Deck != nil && ctx.Deck.EstimateHitRisk(ctx.Hand) > 0.10 {
		return domain.TurnChoiceStay
	}
	return domain.TurnChoiceHit
}

type AggressiveStrategy struct {
	*DefaultTargetSelector
}

func NewAggressiveStrategy() *AggressiveStrategy {
	sel := NewDefaultTargetSelector()
	sel.Flip7Bonus = Flip7BonusAttackLeader
	return &AggressiveStrategy{DefaultTargetSelector: sel}
}

func (s *AggressiveStrategy) Name() string { return "Aggressive" }

func (s *AggressiveStrategy) Decide(ctx domain.DecisionContext) domain.TurnChoice {
	if shouldAlwaysHit(ctx) {
		return domain.TurnChoiceHit
	}
	if ctx.Hand.DistinctRanks() >= 6 {
		return domain.TurnChoiceHit
	}
	if ctx.Deck != nil && ctx.Deck.EstimateHitRisk(ctx.Hand) > 0.40 {
		return domain.TurnChoiceStay
	}
	if ctx.Hand.NumberSum() > 50 {
		return domain.TurnChoiceStay
	}
	return domain.TurnChoiceHit
}

type HeuristicStrategy struct {
	*DefaultTargetSelector
	Threshold int
}

func NewHeuristicStrategy(threshold int) *HeuristicStrategy {
	if threshold <= 0 {
		threshold = DefaultHeuristicThreshold
	}
	return &HeuristicStrategy{
		DefaultTargetSelector: NewDefaultTargetSelector(),
		Threshold:             threshold,
	}
}

func (s *HeuristicStrategy) Name() string {
	return fmt.Sprintf("Heuristic-%d", s.Threshold)
}

func (s *HeuristicStrategy) Decide(ctx domain.DecisionContext) domain.TurnChoice {
	if shouldAlwaysHit(ctx) {
		return domain.TurnChoiceHit
	}
	if ctx.Hand.NumberSum() >= s.Threshold {
		return domain.TurnChoiceStay
	}
	return domain.TurnChoiceHit
}

type ExpectedValueStrategy struct {
	*DefaultTargetSelector
}

func NewExpectedValueStrategy() *ExpectedValueStrategy {
	sel := NewDefaultTargetSelector()
	sel.Flip7Bonus = Flip7BonusAttackLeader
	return &ExpectedValueStrategy{DefaultTargetSelector: sel}
}

func NewExpectedValueStrategyTakeBonus() *ExpectedValueStrategy {
	return &ExpectedValueStrategy{DefaultTargetSelector: NewDefaultTargetSelector()}
}

func NewExpectedValueStrategyWithRisk(flipFourRisk float64) *ExpectedValueStrategy {
	sel := NewDefaultTargetSelectorWithRisk(flipFourRisk)
	sel.Flip7Bonus = Flip7BonusAttackLeader
	return &ExpectedValueStrategy{DefaultTargetSelector: sel}
}

func (s *ExpectedValueStrategy) Name() string { return "ExpectedValue" }

func (s *ExpectedValueStrategy) Decide(ctx domain.DecisionContext) domain.TurnChoice {
	if shouldAlwaysHit(ctx) {
		return domain.TurnChoiceHit
	}
	if len(ctx.DrawUniverse()) == 0 {
		return domain.TurnChoiceStay
	}
	if expectedExplorationValue(ctx) > 0 {
		return domain.TurnChoiceHit
	}
	return domain.TurnChoiceStay
}

type AdaptiveStrategy struct {
	ev  *ExpectedValueStrategy
	agg *AggressiveStrategy
}

func NewAdaptiveStrategy() *AdaptiveStrategy {
	return &AdaptiveStrategy{
		ev:  NewExpectedValueStrategy(),
		agg: NewAggressiveStrategy(),
	}
}

func (s *AdaptiveStrategy) Name() string { return "Adaptive" }

func (s *AdaptiveStrategy) SetDeck(d *domain.Deck) {
	s.ev.SetDeck(d)
	s.agg.SetDeck(d)
}

func (s *AdaptiveStrategy) SetRules(r domain.GameRules) {
	s.ev.SetRules(r)
	s.agg.SetRules(r)
}

func (s *AdaptiveStrategy) active(ctx domain.DecisionContext) domain.Strategy {
	if isBehind(ctx) {
		return s.agg
	}
	return s.ev
}

func (s *AdaptiveStrategy) Decide(ctx domain.DecisionContext) domain.TurnChoice {
	return s.active(ctx).Decide(ctx)
}

func (s *AdaptiveStrategy) ChoosePlayerTarget(action domain.ActionType, candidates []*domain.Player, self *domain.Player) *domain.Player {
	return s.ev.ChoosePlayerTarget(action, candidates, self)
}

func (s *AdaptiveStrategy) ChooseModifierTarget(mod domain.ModifierType, candidates []*domain.Player, self *domain.Player) *domain.Player {
	return s.ev.ChooseModifierTarget(mod, candidates, self)
}

func (s *AdaptiveStrategy) ChooseCardTarget(action domain.ActionType, faceUp []domain.CardRef, self *domain.Player) *domain.CardRef {
	return s.ev.ChooseCardTarget(action, faceUp, self)
}

func (s *AdaptiveStrategy) ChooseSwapPair(faceUp []domain.CardRef, self *domain.Player) *domain.SwapPair {
	return s.ev.ChooseSwapPair(faceUp, self)
}

func (s *AdaptiveStrategy) ChooseFlip7Bonus(self *domain.Player, opponents []*domain.Player) domain.Flip7BonusChoice {
	return s.ev.ChooseFlip7Bonus(self, opponents)
}

func shouldAlwaysHit(ctx domain.DecisionContext) bool {
	if ctx.Hand == nil {
		return true
	}
	if ctx.Hand.MustHit() {
		return true
	}
	if len(ctx.Hand.NumberLine) == 0 {
		return true
	}
	return false
}

func isBehind(ctx domain.DecisionContext) bool {
	for _, o := range ctx.OtherPlayers {
		if o.TotalScore >= ctx.PlayerScore+25 {
			return true
		}
		if o.TotalScore >= 160 && o.TotalScore > ctx.PlayerScore {
			return true
		}
	}
	return false
}

func expectedExplorationValue(ctx domain.DecisionContext) float64 {
	calc := domain.NewScoreCalculatorFor(ctx.Rules)
	current := calc.Compute(ctx.Hand).Total
	cards := ctx.DrawUniverse()
	total := len(cards)
	if total == 0 {
		return 0
	}

	hasOthers := false
	for _, o := range ctx.OtherPlayers {
		if o.CurrentHand != nil && o.CurrentHand.Status != domain.HandStatusBusted {
			hasOthers = true
			break
		}
	}

	gainSum := 0.0
	for _, card := range cards {
		cl := ctx.Hand.Clone()
		switch card.Spec.Type {
		case domain.CardTypeNumber, domain.CardTypeSpecialNumber:
			res := cl.ReceiveNumberLike(card)
			if res.Busted {
				gainSum += float64(calc.Compute(cl).Total - current)
				continue
			}
			pv := calc.Compute(cl)
			gain := pv.Total - current
			if ctx.Rules.Flip7AsAttack && pv.Bonus > 0 && hasOthers {
				gain -= pv.Bonus
			}
			gainSum += float64(gain)
		case domain.CardTypeModifier:
			if !hasOthers {
				cl.ReceiveModifier(card)
				gainSum += float64(calc.Compute(cl).Total - current)
			}
		}
	}

	return gainSum / float64(total)
}
