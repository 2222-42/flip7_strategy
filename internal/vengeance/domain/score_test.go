package domain

import "testing"

func TestOfficialScoreExample(t *testing.T) {
	h := NewPlayerHand()
	for _, v := range []NumberValue{3, 11, 5, 7, 10, 8, 4} {
		h.ReceiveNumberLike(NewNumberCard(v))
	}
	h.ReceiveModifier(NewModifierCard(ModifierDivide2))
	h.ReceiveModifier(NewModifierCard(ModifierMinus4))

	pv := NewScoreCalculator().Compute(h)
	if !h.HasFlip7() {
		t.Fatal("example is Flip 7")
	}
	if pv.BaseSum != 48 || pv.AfterDivide != 24 || pv.AfterModifiers != 20 || pv.Bonus != 15 || pv.Total != 35 {
		t.Fatalf("got %+v, want 48→24→20+15=35", pv)
	}
}

func TestScoreWithoutFlip7(t *testing.T) {
	h := NewPlayerHand()
	for _, v := range []NumberValue{3, 11, 5, 7, 10, 8} {
		h.ReceiveNumberLike(NewNumberCard(v))
	}
	h.ReceiveModifier(NewModifierCard(ModifierDivide2))
	h.ReceiveModifier(NewModifierCard(ModifierMinus4))
	pv := NewScoreCalculator().Compute(h)
	if pv.BaseSum != 44 || pv.AfterDivide != 22 || pv.Total != 18 || pv.Bonus != 0 {
		t.Fatalf("got %+v, want 44→22→18 with no bonus", pv)
	}
}

func TestScoreFloorAtZero(t *testing.T) {
	h := NewPlayerHand()
	h.ReceiveNumberLike(NewNumberCard(1))
	h.ReceiveModifier(NewModifierCard(ModifierMinus10))
	pv := NewScoreCalculator().Compute(h)
	if pv.Total != 0 {
		t.Fatalf("total=%d, want 0", pv.Total)
	}
}

func TestBustedScoresZero(t *testing.T) {
	h := NewPlayerHand()
	h.ReceiveNumberLike(NewNumberCard(5))
	h.ReceiveNumberLike(NewNumberCard(5))
	pv := NewScoreCalculator().Compute(h)
	if pv.Total != 0 {
		t.Fatalf("busted total=%d", pv.Total)
	}
}

func TestZeroOverrideWithoutFlip7(t *testing.T) {
	h := NewPlayerHand()
	h.ReceiveNumberLike(NewSpecialCard(SpecialZero))
	h.ReceiveNumberLike(NewNumberCard(12))
	h.ReceiveModifier(NewModifierCard(ModifierMinus2))
	pv := NewScoreCalculator().Compute(h)
	if !pv.ZeroOverride || pv.Total != 0 {
		t.Fatalf("got %+v", pv)
	}
}

func TestZeroWithFlip7UsesFullFormula(t *testing.T) {
	h := NewPlayerHand()
	h.ReceiveNumberLike(NewSpecialCard(SpecialZero))
	for _, v := range []NumberValue{1, 2, 3, 4, 5, 6} {
		h.ReceiveNumberLike(NewNumberCard(v))
	}
	pv := NewScoreCalculator().Compute(h)
	if pv.ZeroOverride {
		t.Fatal("Flip 7 should skip zero override")
	}
	if pv.BaseSum != 21 || pv.Bonus != 15 || pv.Total != 36 {
		t.Fatalf("got %+v, want base 21 +15 = 36", pv)
	}
}

func TestLucky13BothScoreOneRank(t *testing.T) {
	h := NewPlayerHand()
	h.ReceiveNumberLike(NewSpecialCard(SpecialLucky13))
	h.ReceiveNumberLike(NewNumberCard(13))
	for _, v := range []NumberValue{1, 2, 3, 4, 5, 6} {
		h.ReceiveNumberLike(NewNumberCard(v))
	}
	if h.DistinctRanks() != 7 {
		t.Fatalf("ranks=%d want 7", h.DistinctRanks())
	}
	pv := NewScoreCalculator().Compute(h)
	if pv.BaseSum != 13+13+21 {
		t.Fatalf("base=%d", pv.BaseSum)
	}
	if pv.Bonus != 15 {
		t.Fatalf("bonus=%d", pv.Bonus)
	}
}

func TestDetermineWinnersTie(t *testing.T) {
	a := NewPlayer("A", nil)
	b := NewPlayer("B", nil)
	c := NewPlayer("C", nil)
	a.TotalScore = 210
	b.TotalScore = 210
	c.TotalScore = 180
	g := NewGame([]*Player{a, b, c})
	w := g.DetermineWinners()
	if len(w) != 2 {
		t.Fatalf("winners=%d, want 2", len(w))
	}
}
