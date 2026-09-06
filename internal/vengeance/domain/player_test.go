package domain

import "testing"

func TestDuplicateRankBusts(t *testing.T) {
	h := NewPlayerHand()
	h.ReceiveNumberLike(NewNumberCard(5))
	res := h.ReceiveNumberLike(NewNumberCard(5))
	if !res.Busted || h.Status != HandStatusBusted {
		t.Fatalf("expected bust, status=%s", h.Status)
	}
	for _, c := range h.NumberLine {
		if c.FaceUp {
			t.Fatal("busted cards must be face down")
		}
	}
}

func TestLucky13AllowsSecond13(t *testing.T) {
	h := NewPlayerHand()
	h.ReceiveNumberLike(NewSpecialCard(SpecialLucky13))
	res := h.ReceiveNumberLike(NewNumberCard(13))
	if res.Busted {
		t.Fatal("second 13 should not bust with Lucky 13")
	}
	if h.DistinctRanks() != 1 {
		t.Fatalf("distinct ranks = %d, want 1", h.DistinctRanks())
	}
	res = h.ReceiveNumberLike(NewNumberCard(13))
	if !res.Busted {
		t.Fatal("third 13 should bust")
	}
}

func TestLucky13AfterRegular13(t *testing.T) {
	h := NewPlayerHand()
	h.ReceiveNumberLike(NewNumberCard(13))
	res := h.ReceiveNumberLike(NewSpecialCard(SpecialLucky13))
	if res.Busted {
		t.Fatal("Lucky 13 onto a 13 should not bust")
	}
}

func TestUnlucky7WipesWithoutBust(t *testing.T) {
	h := NewPlayerHand()
	h.ReceiveNumberLike(NewNumberCard(7))
	h.ReceiveNumberLike(NewNumberCard(3))
	h.ReceiveModifier(NewModifierCard(ModifierMinus4))
	res := h.ReceiveNumberLike(NewSpecialCard(SpecialUnlucky7))
	if res.Busted {
		t.Fatal("Unlucky 7 must not bust on receipt")
	}
	if len(res.Wiped) != 3 {
		t.Fatalf("wiped %d, want 3", len(res.Wiped))
	}
	if len(h.NumberLine) != 1 || h.NumberLine[0].Spec.SpecialKind != SpecialUnlucky7 {
		t.Fatalf("line = %+v", h.NumberLine)
	}
	if len(h.ModifierLine) != 0 {
		t.Fatal("modifiers should be wiped")
	}

	res = h.ReceiveNumberLike(NewNumberCard(7))
	if !res.Busted {
		t.Fatal("regular 7 after Unlucky 7 should bust")
	}
}

func TestZeroMustHitAndCountsTowardFlip7(t *testing.T) {
	h := NewPlayerHand()
	h.ReceiveNumberLike(NewSpecialCard(SpecialZero))
	if !h.MustHit() || h.CanStay() {
		t.Fatal("Zero holder must hit and cannot stay")
	}
	for _, v := range []NumberValue{1, 2, 3, 4, 5, 6} {
		h.ReceiveNumberLike(NewNumberCard(v))
	}
	if !h.HasFlip7() {
		t.Fatalf("ranks=%d, want 7 (0 plus 1-6)", h.DistinctRanks())
	}
}

func TestStayMarksLeftmostAndLeavesFaceUp(t *testing.T) {
	h := NewPlayerHand()
	h.ReceiveNumberLike(NewNumberCard(3))
	h.ReceiveNumberLike(NewNumberCard(8))
	h.Stay()
	if h.Status != HandStatusStayed {
		t.Fatalf("status=%s", h.Status)
	}
	if !h.NumberLine[0].Sideways || !h.NumberLine[0].FaceUp {
		t.Fatal("leftmost should be sideways and face up")
	}
	if len(h.FaceUpCards()) != 2 {
		t.Fatal("stayed cards remain targetable")
	}
}

func TestEmptyLineCanStay(t *testing.T) {
	h := NewPlayerHand()
	if !h.CanStay() {
		t.Fatal("empty line may stay")
	}
	h.Stay()
	if h.Status != HandStatusStayed {
		t.Fatalf("status=%s", h.Status)
	}
}

func TestStayDoesNotMarkModifierWhenNoNumbers(t *testing.T) {
	h := NewPlayerHand()
	h.ReceiveModifier(NewModifierCard(ModifierMinus4))
	h.Stay()
	if h.Status != HandStatusStayed {
		t.Fatalf("status=%s", h.Status)
	}
	if h.ModifierLine[0].Sideways {
		t.Fatal("Stay turns the leftmost number card sideways; modifiers stay upright when the number row is empty")
	}
}

func TestRemoveCardStopsZeroEffect(t *testing.T) {
	h := NewPlayerHand()
	z := NewSpecialCard(SpecialZero)
	h.ReceiveNumberLike(z)
	if _, ok := h.RemoveCard(z.ID); !ok {
		t.Fatal("remove failed")
	}
	if h.HasZero() || h.MustHit() {
		t.Fatal("Must Hit should end when Zero leaves")
	}
}
