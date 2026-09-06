package application

import (
	"bufio"
	"strings"
	"testing"

	"flip7_strategy/internal/vengeance/domain"
	"flip7_strategy/internal/vengeance/domain/strategy"
)

func TestParseManualInput(t *testing.T) {
	cases := map[string]domain.CardSpec{
		"7":   domain.NewNumberCard(7).Spec,
		"13":  domain.NewNumberCard(13).Spec,
		"0":   domain.NewSpecialCard(domain.SpecialZero).Spec,
		"z":   domain.NewSpecialCard(domain.SpecialZero).Spec,
		"U":   domain.NewSpecialCard(domain.SpecialUnlucky7).Spec,
		"L13": domain.NewSpecialCard(domain.SpecialLucky13).Spec,
		"-10": domain.NewModifierCard(domain.ModifierMinus10).Spec,
		"/2":  domain.NewModifierCard(domain.ModifierDivide2).Spec,
		"J":   domain.NewActionCard(domain.ActionJustOneMore).Spec,
		"F4":  domain.NewActionCard(domain.ActionFlipFour).Spec,
		"ST":  domain.NewActionCard(domain.ActionSteal).Spec,
		"SW":  domain.NewActionCard(domain.ActionSwap).Spec,
		"DI":  domain.NewActionCard(domain.ActionDiscard).Spec,
	}
	for in, want := range cases {
		got, err := ParseManualInput(in)
		if err != nil {
			t.Fatalf("%s: %v", in, err)
		}
		if !got.Equal(want) {
			t.Fatalf("%s: got %+v want %+v", in, got, want)
		}
	}
	if _, err := ParseManualInput("S"); err == nil {
		t.Fatal("S is stay, not a card")
	}
	if _, err := ParseManualInput("14"); err == nil {
		t.Fatal("14 is out of range")
	}
}

func TestManualStayBanksAtRoundEnd(t *testing.T) {
	in := strings.NewReader("2\nBob\n1\n5\n11\nS\nS\n")
	svc := NewManualGameService(bufio.NewReader(in))
	svc.setupPlayers()
	svc.Game.RoundCount = 1
	svc.playRound()
	if svc.Game.Players[0].TotalScore == 0 && svc.Game.Players[1].TotalScore == 0 {
		t.Fatal("someone should have banked at round end")
	}
}

func TestAdvisorSuggestsStayOnHighSum(t *testing.T) {
	p := domain.NewPlayer("Me", strategy.NewAdaptiveStrategy())
	p.StartNewRound()
	p.CurrentHand.ReceiveNumberLike(domain.NewNumberCard(12))
	p.CurrentHand.ReceiveNumberLike(domain.NewNumberCard(13))
	p.CurrentHand.ReceiveNumberLike(domain.NewNumberCard(11))
	ctx := domain.DecisionContext{
		Deck: domain.NewUnshuffledDeck(),
		Hand: p.CurrentHand,
	}
	choice := strategy.NewAdaptiveStrategy().Decide(ctx)
	if choice != domain.TurnChoiceStay && choice != domain.TurnChoiceHit {
		t.Fatalf("unexpected choice %s", choice)
	}
}

func TestRemoveMatchingUpdatesRemaining(t *testing.T) {
	d := domain.NewUnshuffledDeck()
	before := d.Remaining.ByNumber[5]
	_, ok := d.RemoveMatching(domain.NewNumberCard(5).Spec)
	if !ok {
		t.Fatal("expected a 5 in the deck")
	}
	if d.Remaining.ByNumber[5] != before-1 {
		t.Fatalf("remaining 5 = %d", d.Remaining.ByNumber[5])
	}
}
