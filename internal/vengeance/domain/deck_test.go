package domain

import "testing"

func TestNewDeckComposition(t *testing.T) {
	d := NewUnshuffledDeck()
	if len(d.Cards) != StandardDeckSize {
		t.Fatalf("deck size = %d, want %d", len(d.Cards), StandardDeckSize)
	}

	numbers := map[NumberValue]int{}
	specials := map[SpecialKind]int{}
	mods := map[ModifierType]int{}
	actions := map[ActionType]int{}
	for _, c := range d.Cards {
		switch c.Spec.Type {
		case CardTypeNumber:
			numbers[c.Spec.Value]++
		case CardTypeSpecialNumber:
			specials[c.Spec.SpecialKind]++
		case CardTypeModifier:
			mods[c.Spec.ModifierType]++
		case CardTypeAction:
			actions[c.Spec.ActionType]++
		}
	}

	wantN := map[NumberValue]int{
		1: 1, 2: 2, 3: 3, 4: 4, 5: 5, 6: 6, 7: 6,
		8: 8, 9: 9, 10: 10, 11: 11, 12: 12, 13: 12,
	}
	for v, n := range wantN {
		if numbers[v] != n {
			t.Errorf("regular %d copies = %d, want %d", v, numbers[v], n)
		}
	}
	if numbers[0] != 0 {
		t.Errorf("regular 0 copies = %d, want 0", numbers[0])
	}
	if specials[SpecialZero] != 1 || specials[SpecialUnlucky7] != 1 || specials[SpecialLucky13] != 1 {
		t.Errorf("specials = %+v", specials)
	}
	if len(mods) != 6 {
		t.Errorf("modifier kinds = %d, want 6", len(mods))
	}
	for _, a := range []ActionType{ActionJustOneMore, ActionFlipFour, ActionSwap, ActionSteal, ActionDiscard} {
		if actions[a] != 2 {
			t.Errorf("%s copies = %d, want 2", a, actions[a])
		}
	}

	if d.Remaining.ByNumber[13] != 12 || d.Remaining.ByNumber[7] != 6 {
		t.Errorf("remaining 13=%d 7=%d", d.Remaining.ByNumber[13], d.Remaining.ByNumber[7])
	}
}

func TestDrawUpdatesRemaining(t *testing.T) {
	d := NewUnshuffledDeck()
	before := d.Remaining.ByNumber[1]
	card, err := d.Draw()
	if err != nil {
		t.Fatal(err)
	}
	if card.Spec.Type != CardTypeNumber || card.Spec.Value != 1 {
		t.Fatalf("first unshuffled card = %v, want 1", card)
	}
	if d.Remaining.ByNumber[1] != before-1 {
		t.Fatalf("remaining 1 = %d", d.Remaining.ByNumber[1])
	}
}
