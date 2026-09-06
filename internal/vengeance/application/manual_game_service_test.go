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
	p := domain.NewPlayer("Me", strategy.NewCautiousStrategy())
	p.StartNewRound()
	p.CurrentHand.ReceiveNumberLike(domain.NewNumberCard(12))
	p.CurrentHand.ReceiveNumberLike(domain.NewNumberCard(13))
	p.CurrentHand.ReceiveNumberLike(domain.NewNumberCard(11))
	ctx := domain.DecisionContext{
		Deck: domain.NewUnshuffledDeck(),
		Hand: p.CurrentHand,
	}
	if strategy.NewCautiousStrategy().Decide(ctx) != domain.TurnChoiceStay {
		t.Fatal("Cautious should stay on number sum 36")
	}
}

func TestPromptStayDoesNotBank(t *testing.T) {
	p := domain.NewPlayer("Me", strategy.NewAdaptiveStrategy())
	game := domain.NewGame([]*domain.Player{p})
	game.CurrentRound = domain.NewRound([]*domain.Player{p}, p, domain.NewUnshuffledDeck())
	p.CurrentHand.ReceiveNumberLike(domain.NewNumberCard(8))
	svc := NewManualGameService(bufio.NewReader(strings.NewReader("S\n")))
	svc.Game = game
	svc.promptAndProcess(p, true)
	if p.CurrentHand.Status != domain.HandStatusStayed {
		t.Fatalf("status=%s", p.CurrentHand.Status)
	}
	if p.TotalScore != 0 {
		t.Fatalf("Stay must not bank yet, total=%d", p.TotalScore)
	}
}

func TestJustOneMoreForceStayWithoutBanking(t *testing.T) {
	p1 := domain.NewPlayer("Me", strategy.NewAdaptiveStrategy())
	p2 := domain.NewPlayer("Bob", strategy.NewAdaptiveStrategy())
	game := domain.NewGame([]*domain.Player{p1, p2})
	game.Deck = domain.NewUnshuffledDeck()
	game.CurrentRound = domain.NewRound([]*domain.Player{p1, p2}, p1, game.Deck)
	jom, ok := game.CurrentRound.Deck.RemoveMatching(domain.NewActionCard(domain.ActionJustOneMore).Spec)
	if !ok {
		t.Fatal("expected Just One More in deck")
	}
	svc := NewManualGameService(bufio.NewReader(strings.NewReader("2\n5\n")))
	svc.Game = game
	svc.processCard(p1, jom)
	if p2.CurrentHand.Status != domain.HandStatusStayed {
		t.Fatalf("Bob status=%s, want stayed", p2.CurrentHand.Status)
	}
	if p2.TotalScore != 0 {
		t.Fatalf("JOM stay must not bank yet, total=%d", p2.TotalScore)
	}
}

func TestReadCardFromTableRetriesMissingCard(t *testing.T) {
	p := domain.NewPlayer("Me", strategy.NewAdaptiveStrategy())
	game := domain.NewGame([]*domain.Player{p})
	game.CurrentRound = domain.NewRound([]*domain.Player{p}, p, domain.NewUnshuffledDeck())
	for {
		if _, ok := game.CurrentRound.Deck.RemoveMatching(domain.NewNumberCard(5).Spec); !ok {
			break
		}
	}
	svc := NewManualGameService(bufio.NewReader(strings.NewReader("5\n6\n")))
	svc.Game = game
	card, err := svc.readCardFromTable()
	if err != nil {
		t.Fatal(err)
	}
	if card.Spec.Value != 6 {
		t.Fatalf("got %v, want 6 after retry", card)
	}
}

func TestBustRateUsesDiscardWhenDrawPileEmpty(t *testing.T) {
	p := domain.NewPlayer("Me", strategy.NewAdaptiveStrategy())
	game := domain.NewGame([]*domain.Player{p})
	game.CurrentRound = domain.NewRound([]*domain.Player{p}, p, domain.DeckWithCards(nil))
	p.CurrentHand.ReceiveNumberLike(domain.NewNumberCard(12))
	game.DiscardPile = []domain.TableCard{domain.NewNumberCard(12), domain.NewNumberCard(12)}
	svc := NewManualGameService(bufio.NewReader(strings.NewReader("")))
	svc.Game = game
	d := svc.riskDeck()
	if d == nil || d.EstimateHitRisk(p.CurrentHand) != 1 {
		t.Fatalf("expected certain bust from discard 12s, deck=%v", d)
	}
}

func TestSwapRejectsSameOwnerPair(t *testing.T) {
	p1 := domain.NewPlayer("Me", strategy.NewAdaptiveStrategy())
	p2 := domain.NewPlayer("Bob", strategy.NewAdaptiveStrategy())
	game := domain.NewGame([]*domain.Player{p1, p2})
	game.Deck = domain.NewUnshuffledDeck()
	game.CurrentRound = domain.NewRound([]*domain.Player{p1, p2}, p1, game.Deck)
	a := domain.NewNumberCard(8)
	b := domain.NewNumberCard(3)
	p1.CurrentHand.ReceiveNumberLike(a)
	p2.CurrentHand.ReceiveNumberLike(b)
	swap, ok := game.CurrentRound.Deck.RemoveMatching(domain.NewActionCard(domain.ActionSwap).Spec)
	if !ok {
		t.Fatal("expected Swap in deck")
	}
	// 1 then 1 is the same card; then 1 and 2 is a legal pair.
	svc := NewManualGameService(bufio.NewReader(strings.NewReader("1\n1\n1\n2\n")))
	svc.Game = game
	svc.executeSwap(p1, swap)
	if len(p1.CurrentHand.NumberLine) != 1 || p1.CurrentHand.NumberLine[0].ID != b.ID {
		t.Fatalf("Me should have Bob's 3, line=%v", p1.CurrentHand.NumberLine)
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
