package application

import (
	"testing"

	"flip7_strategy/internal/vengeance/domain"
)

func newSilentGame(n int, deck *domain.Deck) *GameService {
	players := make([]*domain.Player, n)
	for i := 0; i < n; i++ {
		players[i] = domain.NewPlayer(string(rune('A'+i)), domain.NewStubStrategy())
	}
	g := domain.NewGame(players)
	g.Deck = deck
	svc := NewGameService(g)
	svc.Silent = true
	return svc
}

func stacked(cards ...domain.TableCard) *domain.Deck {
	return domain.DeckWithCards(cards)
}

func TestAutoPlayCompletes(t *testing.T) {
	svc := newSilentGame(3, domain.NewDeck())
	svc.RunGame()
	if !svc.Game.IsCompleted {
		t.Fatal("game should complete")
	}
	if svc.Game.RoundCount == 0 {
		t.Fatal("no rounds played")
	}
	if len(svc.Game.Winners) == 0 {
		t.Fatalf("expected a winner, rounds=%d", svc.Game.RoundCount)
	}
}

func TestInitialDealStartsLeftOfDealer(t *testing.T) {
	deck := stacked(
		domain.NewNumberCard(1),
		domain.NewNumberCard(2),
		domain.NewNumberCard(3),
	)
	svc := newSilentGame(3, deck)
	dealer := svc.Game.Players[0]
	svc.Game.CurrentRound = domain.NewRound(svc.Game.Players, dealer, deck)
	svc.PlayRound()

	left := svc.Game.Players[1]
	if len(left.CurrentHand.NumberLine) == 0 || left.CurrentHand.NumberLine[0].Spec.Value != 1 {
		t.Fatalf("left of dealer should receive first card 1, got %+v", left.CurrentHand.NumberLine)
	}
}

func TestStayDoesNotBankUntilRoundEnd(t *testing.T) {
	deck := stacked(
		domain.NewNumberCard(12),
		domain.NewNumberCard(11),
	)
	svc := newSilentGame(2, deck)
	svc.Game.CurrentRound = domain.NewRound(svc.Game.Players, svc.Game.Players[0], deck)
	if svc.Game.Players[0].TotalScore != 0 {
		t.Fatal("pre-round score")
	}
	svc.PlayRound()
	if svc.Game.Players[1].TotalScore == 0 && svc.Game.Players[0].TotalScore == 0 {
		t.Fatal("someone should have banked at round end")
	}
}

func TestSwapStealDiscardWithEmptyTableAreDiscarded(t *testing.T) {
	for _, a := range []domain.ActionType{domain.ActionSwap, domain.ActionSteal, domain.ActionDiscard} {
		deck := stacked(domain.NewActionCard(a), domain.NewNumberCard(1), domain.NewNumberCard(2))
		svc := newSilentGame(2, deck)
		svc.Game.CurrentRound = domain.NewRound(svc.Game.Players, svc.Game.Players[0], deck)
		svc.PlayRound()
		if svc.Game.CurrentRound.EndReason == domain.RoundEndReasonAborted {
			t.Fatalf("%s stalled the round", a)
		}
	}
}

func TestJustOneMoreForcesStayIncludingZero(t *testing.T) {
	// Left of dealer (P1) is dealt Zero. Dealer (P0) is dealt Just One More and
	// stub targets the first opponent (P1). P1 takes one more card then stays.
	deck := stacked(
		domain.NewSpecialCard(domain.SpecialZero),
		domain.NewActionCard(domain.ActionJustOneMore),
		domain.NewNumberCard(2),
	)
	svc := newSilentGame(2, deck)
	svc.Game.CurrentRound = domain.NewRound(svc.Game.Players, svc.Game.Players[0], deck)
	svc.PlayRound()

	p1 := svc.Game.Players[1]
	if p1.CurrentHand.Status != domain.HandStatusStayed {
		t.Fatalf("Zero holder after JOM status=%s, want stayed", p1.CurrentHand.Status)
	}
	if !p1.CurrentHand.HasZero() {
		t.Fatal("Zero should still be in the line")
	}
}

func TestFlipFourStopsOnBustAndSkipsDelayed(t *testing.T) {
	// P1 dealt 5. P0 dealt Flip Four targeting P1. P1 then draws 5 (bust) plus
	// filler cards that must not be resolved if we stop.
	deck := stacked(
		domain.NewNumberCard(5),
		domain.NewActionCard(domain.ActionFlipFour),
		domain.NewNumberCard(5),
		domain.NewModifierCard(domain.ModifierMinus10),
		domain.NewNumberCard(9),
		domain.NewNumberCard(8),
	)
	svc := newSilentGame(2, deck)
	svc.Game.CurrentRound = domain.NewRound(svc.Game.Players, svc.Game.Players[0], deck)
	svc.PlayRound()

	p1 := svc.Game.Players[1]
	if p1.CurrentHand.Status != domain.HandStatusBusted {
		t.Fatalf("status=%s, want busted", p1.CurrentHand.Status)
	}
	if len(p1.CurrentHand.ModifierLine) != 0 {
		t.Fatal("delayed modifier must not apply after bust")
	}
}

func TestFlipFourUnlucky7WipesThenContinues(t *testing.T) {
	deck := stacked(
		domain.NewNumberCard(3),
		domain.NewActionCard(domain.ActionFlipFour),
		domain.NewNumberCard(4),
		domain.NewSpecialCard(domain.SpecialUnlucky7),
		domain.NewNumberCard(2),
		domain.NewNumberCard(1),
	)
	svc := newSilentGame(2, deck)
	svc.Game.CurrentRound = domain.NewRound(svc.Game.Players, svc.Game.Players[0], deck)
	svc.PlayRound()

	p1 := svc.Game.Players[1]
	if p1.CurrentHand.Status == domain.HandStatusBusted {
		t.Fatal("Unlucky 7 must not bust")
	}
	hasU7 := false
	for _, c := range p1.CurrentHand.NumberLine {
		if c.Spec.SpecialKind == domain.SpecialUnlucky7 {
			hasU7 = true
		}
		if c.Spec.Value == 3 || c.Spec.Value == 4 {
			t.Fatalf("wiped card %v still in line", c)
		}
	}
	if !hasU7 {
		t.Fatal("Unlucky 7 should remain")
	}
}

func TestOnlyNonBustedKeepsModifier(t *testing.T) {
	h := domain.NewPlayerHand()
	p := domain.NewPlayer("solo", domain.NewStubStrategy())
	p.CurrentHand = h
	g := domain.NewGame([]*domain.Player{p})
	g.Deck = stacked(domain.NewModifierCard(domain.ModifierMinus2))
	svc := NewGameService(g)
	svc.Silent = true
	svc.Game.CurrentRound = domain.NewRound([]*domain.Player{p}, p, g.Deck)
	svc.PlayRound()
	if len(p.CurrentHand.ModifierLine) != 1 {
		t.Fatalf("solo player should keep modifier, got %d", len(p.CurrentHand.ModifierLine))
	}
}

func TestReshuffleLeavesTableCards(t *testing.T) {
	onTable := domain.NewSpecialCard(domain.SpecialZero)
	fromDiscard := domain.NewNumberCard(2)
	deck := stacked(onTable)
	svc := newSilentGame(1, deck)
	svc.Game.DiscardPile = []domain.TableCard{fromDiscard}
	svc.Game.CurrentRound = domain.NewRound(svc.Game.Players, svc.Game.Players[0], deck)
	svc.PlayRound()
	p := svc.Game.Players[0]
	foundZero := false
	for _, c := range p.CurrentHand.NumberLine {
		if c.ID == onTable.ID {
			foundZero = true
		}
	}
	if !foundZero {
		t.Fatal("Zero dealt before reshuffle must stay in the line")
	}
}

func TestStealDuplicateBustsThief(t *testing.T) {
	fiveA := domain.NewNumberCard(5)
	fiveB := domain.NewNumberCard(5)
	steal := domain.NewActionCard(domain.ActionSteal)
	// Deal order: left of dealer (P1) gets 5, dealer (P0) gets 5, then P1 stays,
	// P0 hits... stub stays so P0 won't hit. Need steal on the deal to P0.
	deck := stacked(fiveB, steal)
	svc := newSilentGame(2, deck)
	svc.Game.CurrentRound = domain.NewRound(svc.Game.Players, svc.Game.Players[0], deck)
	// Put a 5 already in P0's hand before PlayRound overwrites hands via NewRound.
	// NewRound already ran in the line above. Give P0 the first 5 then process steal from P1.
	p0 := svc.Game.Players[0]
	p1 := svc.Game.Players[1]
	p0.CurrentHand.ReceiveNumberLike(fiveA)
	p1.CurrentHand.ReceiveNumberLike(fiveB)
	svc.resolveAction(p0, steal)
	if p0.CurrentHand.Status != domain.HandStatusBusted {
		t.Fatalf("thief status=%s, want busted", p0.CurrentHand.Status)
	}
}

func TestFlip7EndsRound(t *testing.T) {
	h := domain.NewPlayerHand()
	p := domain.NewPlayer("solo", domain.NewStubStrategy())
	p.CurrentHand = h
	for _, v := range []domain.NumberValue{1, 2, 3, 4, 5, 6} {
		h.ReceiveNumberLike(domain.NewNumberCard(v))
	}
	g := domain.NewGame([]*domain.Player{p})
	g.Deck = stacked(domain.NewNumberCard(7))
	svc := NewGameService(g)
	svc.Silent = true
	svc.Game.CurrentRound = domain.NewRound([]*domain.Player{p}, p, g.Deck)
	p.CurrentHand = h
	svc.ProcessFlip(p, domain.NewNumberCard(7))
	if svc.Game.CurrentRound.EndReason != domain.RoundEndReasonFlip7 {
		t.Fatalf("reason=%s", svc.Game.CurrentRound.EndReason)
	}
}

type passStrategy struct {
	*domain.StubStrategy
}

func (s *passStrategy) ChooseCardTarget(domain.ActionType, []domain.CardRef, *domain.Player) *domain.CardRef {
	return nil
}

func (s *passStrategy) ChooseSwapPair([]domain.CardRef, *domain.Player) *domain.SwapPair {
	return nil
}

func TestSwapStillResolvesWhenStrategyPasses(t *testing.T) {
	a := domain.NewNumberCard(8)
	b := domain.NewNumberCard(3)
	swap := domain.NewActionCard(domain.ActionSwap)
	svc := newSilentGame(2, stacked())
	svc.Game.Players[0].Strategy = &passStrategy{StubStrategy: domain.NewStubStrategy()}
	svc.Game.CurrentRound = domain.NewRound(svc.Game.Players, svc.Game.Players[0], svc.Game.Deck)
	p0 := svc.Game.Players[0]
	p1 := svc.Game.Players[1]
	p0.CurrentHand.ReceiveNumberLike(a)
	p1.CurrentHand.ReceiveNumberLike(b)
	svc.executeSwap(p0, swap)
	if len(p0.CurrentHand.NumberLine) != 1 || p0.CurrentHand.NumberLine[0].ID != b.ID {
		t.Fatalf("p0 line=%v, want stolen/swapped 3", p0.CurrentHand.NumberLine)
	}
	if len(p1.CurrentHand.NumberLine) != 1 || p1.CurrentHand.NumberLine[0].ID != a.ID {
		t.Fatalf("p1 line=%v, want 8", p1.CurrentHand.NumberLine)
	}
}

func TestStealStillResolvesWhenStrategyPasses(t *testing.T) {
	mine := domain.NewNumberCard(2)
	theirs := domain.NewNumberCard(12)
	steal := domain.NewActionCard(domain.ActionSteal)
	svc := newSilentGame(2, stacked())
	svc.Game.Players[0].Strategy = &passStrategy{StubStrategy: domain.NewStubStrategy()}
	svc.Game.CurrentRound = domain.NewRound(svc.Game.Players, svc.Game.Players[0], svc.Game.Deck)
	p0 := svc.Game.Players[0]
	p1 := svc.Game.Players[1]
	p0.CurrentHand.ReceiveNumberLike(mine)
	p1.CurrentHand.ReceiveNumberLike(theirs)
	svc.executeSteal(p0, steal)
	found := false
	for _, c := range p0.CurrentHand.NumberLine {
		if c.ID == theirs.ID {
			found = true
		}
	}
	if !found {
		t.Fatalf("p0 should have stolen 12, line=%v", p0.CurrentHand.NumberLine)
	}
}
