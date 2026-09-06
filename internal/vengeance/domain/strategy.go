package domain

type DecisionContext struct {
	Deck         *Deck
	DiscardPile  []TableCard
	Hand         *PlayerHand
	PlayerScore  int
	OtherPlayers []*Player
}

// DrawUniverse is the card set the next Hit will sample: the draw pile, or the
// discard pile if a reshuffle would happen first.
func (ctx DecisionContext) DrawUniverse() []TableCard {
	if ctx.Deck != nil && len(ctx.Deck.Cards) > 0 {
		return ctx.Deck.Cards
	}
	return ctx.DiscardPile
}

type Strategy interface {
	Name() string
	Decide(ctx DecisionContext) TurnChoice
	ChoosePlayerTarget(action ActionType, candidates []*Player, self *Player) *Player
	ChooseModifierTarget(mod ModifierType, candidates []*Player, self *Player) *Player
	ChooseCardTarget(action ActionType, faceUp []CardRef, self *Player) *CardRef
	ChooseSwapPair(faceUp []CardRef, self *Player) *SwapPair
}

// StubStrategy stays unless The Zero forces a hit. Targeting is first legal option.
type StubStrategy struct{}

func NewStubStrategy() *StubStrategy {
	return &StubStrategy{}
}

func (s *StubStrategy) Name() string { return "Stub" }

func (s *StubStrategy) Decide(ctx DecisionContext) TurnChoice {
	if ctx.Hand != nil && ctx.Hand.MustHit() {
		return TurnChoiceHit
	}
	return TurnChoiceStay
}

func (s *StubStrategy) ChoosePlayerTarget(_ ActionType, candidates []*Player, self *Player) *Player {
	for _, p := range candidates {
		if self == nil || p.ID != self.ID {
			return p
		}
	}
	if len(candidates) > 0 {
		return candidates[0]
	}
	return self
}

func (s *StubStrategy) ChooseModifierTarget(_ ModifierType, candidates []*Player, self *Player) *Player {
	return s.ChoosePlayerTarget(ActionDiscard, candidates, self)
}

func (s *StubStrategy) ChooseCardTarget(action ActionType, faceUp []CardRef, self *Player) *CardRef {
	if action == ActionSteal || action == ActionDiscard {
		for i := range faceUp {
			if self == nil || faceUp[i].OwnerID != self.ID {
				return &faceUp[i]
			}
		}
	}
	if len(faceUp) > 0 {
		return &faceUp[0]
	}
	return nil
}

func (s *StubStrategy) ChooseSwapPair(faceUp []CardRef, self *Player) *SwapPair {
	if self == nil {
		return FirstLegalSwapPair(faceUp)
	}
	var mine *CardRef
	var theirs *CardRef
	for i := range faceUp {
		ref := &faceUp[i]
		if ref.OwnerID == self.ID && mine == nil {
			mine = ref
		}
		if ref.OwnerID != self.ID && theirs == nil {
			theirs = ref
		}
	}
	if mine != nil && theirs != nil {
		return &SwapPair{A: *mine, B: *theirs}
	}
	return FirstLegalSwapPair(faceUp)
}

func FirstLegalSwapPair(faceUp []CardRef) *SwapPair {
	for i := range faceUp {
		for j := i + 1; j < len(faceUp); j++ {
			if faceUp[i].OwnerID != faceUp[j].OwnerID {
				return &SwapPair{A: faceUp[i], B: faceUp[j]}
			}
		}
	}
	return nil
}
