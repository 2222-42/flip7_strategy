package strategy

import (
	"testing"

	"flip7_strategy/internal/vengeance/domain"
)

func TestMustHitOverridesStay(t *testing.T) {
	h := domain.NewPlayerHand()
	h.ReceiveNumberLike(domain.NewSpecialCard(domain.SpecialZero))
	ctx := domain.DecisionContext{Hand: h, Deck: domain.NewUnshuffledDeck()}
	for _, s := range []domain.Strategy{
		NewCautiousStrategy(),
		NewAggressiveStrategy(),
		NewHeuristicStrategy(10),
		NewExpectedValueStrategy(),
		NewAdaptiveStrategy(),
	} {
		if s.Decide(ctx) != domain.TurnChoiceHit {
			t.Errorf("%s should hit with Zero", s.Name())
		}
	}
}

func TestHeuristicStopsAtThreshold(t *testing.T) {
	h := domain.NewPlayerHand()
	h.ReceiveNumberLike(domain.NewNumberCard(12))
	h.ReceiveNumberLike(domain.NewNumberCard(13))
	ctx := domain.DecisionContext{Hand: h, Deck: domain.NewUnshuffledDeck()}
	low := NewHeuristicStrategy(30)
	high := NewHeuristicStrategy(20)
	if low.Decide(ctx) != domain.TurnChoiceHit {
		t.Fatal("sum 25 should hit at threshold 30")
	}
	if high.Decide(ctx) != domain.TurnChoiceStay {
		t.Fatal("sum 25 should stay at threshold 20")
	}
}

func TestExpectedValueStaysWhenOnlyBustCards(t *testing.T) {
	h := domain.NewPlayerHand()
	h.ReceiveNumberLike(domain.NewNumberCard(5))
	h.ReceiveNumberLike(domain.NewNumberCard(9))
	deck := domain.DeckWithCards([]domain.TableCard{
		domain.NewNumberCard(5),
		domain.NewNumberCard(9),
	})
	ctx := domain.DecisionContext{Hand: h, Deck: deck}
	if NewExpectedValueStrategy().Decide(ctx) != domain.TurnChoiceStay {
		t.Fatal("EV should stay when every remaining card busts")
	}
}

func TestExpectedValueHitsWhenSafeGain(t *testing.T) {
	h := domain.NewPlayerHand()
	h.ReceiveNumberLike(domain.NewNumberCard(2))
	deck := domain.DeckWithCards([]domain.TableCard{
		domain.NewNumberCard(12),
		domain.NewNumberCard(11),
	})
	ctx := domain.DecisionContext{Hand: h, Deck: deck}
	if NewExpectedValueStrategy().Decide(ctx) != domain.TurnChoiceHit {
		t.Fatal("EV should hit on safe high numbers")
	}
}

func TestModifierGoesToHighestRoundScore(t *testing.T) {
	self := domain.NewPlayer("me", nil)
	low := domain.NewPlayer("low", nil)
	high := domain.NewPlayer("high", nil)
	self.CurrentHand = domain.NewPlayerHand()
	low.CurrentHand = domain.NewPlayerHand()
	high.CurrentHand = domain.NewPlayerHand()
	low.CurrentHand.ReceiveNumberLike(domain.NewNumberCard(3))
	high.CurrentHand.ReceiveNumberLike(domain.NewNumberCard(12))
	sel := NewDefaultTargetSelector()
	got := sel.ChooseModifierTarget(domain.ModifierMinus10, []*domain.Player{self, low, high}, self)
	if got == nil || got.ID != high.ID {
		t.Fatalf("got %v, want high", got)
	}
}

func TestStealPrefersLucky13(t *testing.T) {
	self := domain.NewPlayer("me", nil)
	opp := domain.NewPlayer("opp", nil)
	lucky := domain.NewSpecialCard(domain.SpecialLucky13)
	low := domain.NewNumberCard(4)
	refs := []domain.CardRef{
		{ID: low.ID, OwnerID: opp.ID, Spec: low.Spec},
		{ID: lucky.ID, OwnerID: opp.ID, Spec: lucky.Spec},
	}
	sel := NewDefaultTargetSelector()
	got := sel.ChooseCardTarget(domain.ActionSteal, refs, self)
	if got == nil || got.ID != lucky.ID {
		t.Fatalf("got %+v, want Lucky 13", got)
	}
}

func TestSwapDumpsZero(t *testing.T) {
	self := domain.NewPlayer("me", nil)
	opp := domain.NewPlayer("opp", nil)
	zero := domain.NewSpecialCard(domain.SpecialZero)
	prize := domain.NewNumberCard(13)
	refs := []domain.CardRef{
		{ID: zero.ID, OwnerID: self.ID, Spec: zero.Spec},
		{ID: prize.ID, OwnerID: opp.ID, Spec: prize.Spec},
	}
	pair := NewDefaultTargetSelector().ChooseSwapPair(refs, self)
	if pair == nil || pair.A.ID != zero.ID || pair.B.ID != prize.ID {
		t.Fatalf("got %+v", pair)
	}
}

func TestExpectedValueHitsWhenDrawPileEmptyButDiscardHasGain(t *testing.T) {
	h := domain.NewPlayerHand()
	h.ReceiveNumberLike(domain.NewNumberCard(2))
	ctx := domain.DecisionContext{
		Hand: h,
		Deck: domain.DeckWithCards(nil),
		DiscardPile: []domain.TableCard{
			domain.NewNumberCard(12),
			domain.NewNumberCard(11),
		},
	}
	if NewExpectedValueStrategy().Decide(ctx) != domain.TurnChoiceHit {
		t.Fatal("EV should hit into a reshuffle that still has safe high cards")
	}
}

func TestDiscardOwnZero(t *testing.T) {
	self := domain.NewPlayer("me", nil)
	opp := domain.NewPlayer("opp", nil)
	zero := domain.NewSpecialCard(domain.SpecialZero)
	low := domain.NewNumberCard(4)
	refs := []domain.CardRef{
		{ID: zero.ID, OwnerID: self.ID, Spec: zero.Spec},
		{ID: low.ID, OwnerID: opp.ID, Spec: low.Spec},
	}
	got := NewDefaultTargetSelector().ChooseCardTarget(domain.ActionDiscard, refs, self)
	if got == nil || got.ID != zero.ID {
		t.Fatalf("got %+v, want own Zero", got)
	}
}

func TestSwapFallsBackWhenNoBurden(t *testing.T) {
	self := domain.NewPlayer("me", nil)
	opp := domain.NewPlayer("opp", nil)
	mine := domain.NewNumberCard(8)
	theirs := domain.NewNumberCard(3)
	refs := []domain.CardRef{
		{ID: mine.ID, OwnerID: self.ID, Spec: mine.Spec},
		{ID: theirs.ID, OwnerID: opp.ID, Spec: theirs.Spec},
	}
	pair := NewDefaultTargetSelector().ChooseSwapPair(refs, self)
	if pair == nil {
		t.Fatal("legal two-owner pair must not be skipped")
	}
	if pair.A.OwnerID == pair.B.OwnerID {
		t.Fatal("swap must be between two players")
	}
}

func TestAdaptiveSwitchesWhenBehind(t *testing.T) {
	h := domain.NewPlayerHand()
	h.ReceiveNumberLike(domain.NewNumberCard(12))
	h.ReceiveNumberLike(domain.NewNumberCard(11))
	leader := domain.NewPlayer("lead", nil)
	leader.TotalScore = 180
	ctx := domain.DecisionContext{
		Hand:         h,
		Deck:         domain.NewUnshuffledDeck(),
		PlayerScore:  40,
		OtherPlayers: []*domain.Player{leader},
	}
	if !isBehind(ctx) {
		t.Fatal("should be behind")
	}
	_ = NewAdaptiveStrategy().Decide(ctx)
}
