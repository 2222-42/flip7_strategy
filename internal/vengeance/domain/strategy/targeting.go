package strategy

import (
	"flip7_strategy/internal/vengeance/domain"
)

const DefaultFlipFourRisk = 0.50

type DefaultTargetSelector struct {
	deck            *domain.Deck
	FlipFourRiskMin float64
}

func NewDefaultTargetSelector() *DefaultTargetSelector {
	return &DefaultTargetSelector{FlipFourRiskMin: DefaultFlipFourRisk}
}

func NewDefaultTargetSelectorWithRisk(threshold float64) *DefaultTargetSelector {
	return &DefaultTargetSelector{FlipFourRiskMin: threshold}
}

func (s *DefaultTargetSelector) SetDeck(d *domain.Deck) {
	s.deck = d
}

func (s *DefaultTargetSelector) ChoosePlayerTarget(action domain.ActionType, candidates []*domain.Player, self *domain.Player) *domain.Player {
	opponents := opponentsOf(candidates, self)
	if len(opponents) == 0 {
		return firstOrSelf(candidates, self)
	}

	switch action {
	case domain.ActionJustOneMore:
		var best *domain.Player
		bestRanks := -1
		for _, p := range opponents {
			r := 0
			if p.CurrentHand != nil {
				r = p.CurrentHand.DistinctRanks()
			}
			if r > bestRanks {
				bestRanks = r
				best = p
			}
		}
		if bestRanks >= 5 {
			return best
		}
		return highestTotal(opponents)
	case domain.ActionFlipFour:
		if s.deck != nil {
			var best *domain.Player
			bestRisk := -1.0
			for _, p := range opponents {
				if p.CurrentHand == nil {
					continue
				}
				risk := s.deck.EstimateFlipFourRisk(p.CurrentHand)
				if risk >= s.FlipFourRiskMin && risk > bestRisk {
					bestRisk = risk
					best = p
				}
			}
			if best != nil {
				return best
			}
		}
		return highestTotal(opponents)
	default:
		return highestTotal(opponents)
	}
}

func (s *DefaultTargetSelector) ChooseModifierTarget(_ domain.ModifierType, candidates []*domain.Player, self *domain.Player) *domain.Player {
	opponents := opponentsOf(candidates, self)
	if len(opponents) == 0 {
		return firstOrSelf(candidates, self)
	}
	var best *domain.Player
	bestScore := -1
	calc := domain.NewScoreCalculator()
	for _, p := range opponents {
		sc := 0
		if p.CurrentHand != nil {
			sc = calc.Compute(p.CurrentHand).Total
		}
		if sc > bestScore {
			bestScore = sc
			best = p
		}
	}
	if best != nil {
		return best
	}
	return opponents[0]
}

func (s *DefaultTargetSelector) ChooseCardTarget(action domain.ActionType, faceUp []domain.CardRef, self *domain.Player) *domain.CardRef {
	var best *domain.CardRef
	bestVal := -1 << 30
	for i := range faceUp {
		ref := &faceUp[i]
		if self != nil && ref.OwnerID == self.ID {
			if action != domain.ActionDiscard {
				continue
			}
		}
		val := 0
		switch action {
		case domain.ActionSteal:
			val = stealValue(ref.Spec)
		case domain.ActionDiscard:
			if self != nil && ref.OwnerID == self.ID {
				continue
			}
			val = discardValue(ref.Spec)
		default:
			val = stealValue(ref.Spec)
		}
		if val > bestVal {
			bestVal = val
			best = ref
		}
	}
	if bestVal <= 0 {
		return nil
	}
	return best
}

func (s *DefaultTargetSelector) ChooseSwapPair(faceUp []domain.CardRef, self *domain.Player) *domain.SwapPair {
	if self == nil {
		return nil
	}
	var mine *domain.CardRef
	mineBurden := -1
	var theirs *domain.CardRef
	theirPrize := -1
	for i := range faceUp {
		ref := &faceUp[i]
		if ref.OwnerID == self.ID {
			b := burdenValue(ref.Spec)
			if b > mineBurden {
				mineBurden = b
				mine = ref
			}
			continue
		}
		p := stealValue(ref.Spec)
		if p > theirPrize {
			theirPrize = p
			theirs = ref
		}
	}
	if mine != nil && theirs != nil && mineBurden > 0 && theirPrize > 0 {
		return &domain.SwapPair{A: *mine, B: *theirs}
	}
	return nil
}

func stealValue(spec domain.CardSpec) int {
	if spec.SpecialKind == domain.SpecialLucky13 {
		return 100
	}
	if spec.SpecialKind == domain.SpecialZero || spec.SpecialKind == domain.SpecialUnlucky7 {
		return -1
	}
	if spec.IsNumberLike() {
		return int(spec.Value)
	}
	return -1
}

func discardValue(spec domain.CardSpec) int {
	if spec.SpecialKind == domain.SpecialLucky13 {
		return 100
	}
	if spec.SpecialKind == domain.SpecialZero {
		return -50
	}
	if spec.Type == domain.CardTypeModifier {
		return -20
	}
	if spec.IsNumberLike() {
		return int(spec.Value)
	}
	return -1
}

func burdenValue(spec domain.CardSpec) int {
	if spec.SpecialKind == domain.SpecialZero {
		return 100
	}
	if spec.Type == domain.CardTypeModifier {
		if spec.ModifierType.IsDivide() {
			return 40
		}
		return -spec.ModifierType.Amount()
	}
	if spec.SpecialKind == domain.SpecialUnlucky7 {
		return 5
	}
	return 0
}

func opponentsOf(candidates []*domain.Player, self *domain.Player) []*domain.Player {
	var out []*domain.Player
	for _, p := range candidates {
		if self == nil || p.ID != self.ID {
			out = append(out, p)
		}
	}
	return out
}

func highestTotal(ps []*domain.Player) *domain.Player {
	var best *domain.Player
	bestScore := -1
	for _, p := range ps {
		if p.TotalScore > bestScore {
			bestScore = p.TotalScore
			best = p
		}
	}
	return best
}

func firstOrSelf(candidates []*domain.Player, self *domain.Player) *domain.Player {
	if len(candidates) > 0 {
		return candidates[0]
	}
	return self
}
