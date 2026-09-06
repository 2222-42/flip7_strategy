package application

import (
	"bufio"
	"strings"
	"testing"

	"flip7_strategy/internal/vengeance/domain"
)

func TestManualUndoReplacesDealtCard(t *testing.T) {
	// Dealer is Me, so Bob is dealt first, then Me.
	in := strings.NewReader("2\nBob\n1\n5\n11\nUNDO\n7\nS\nS\n")
	svc := NewManualGameService(bufio.NewReader(in))
	svc.setupPlayers()
	svc.Game.RoundCount = 1
	svc.playRound()

	me, bob := svc.Game.Players[0], svc.Game.Players[1]
	if me.TotalScore != 7 {
		t.Fatalf("Me total=%d, want 7 (11 undone, 7 dealt)", me.TotalScore)
	}
	if bob.TotalScore != 5 {
		t.Fatalf("Bob total=%d, want 5", bob.TotalScore)
	}
	if countNumber(svc.Game.CurrentRound.Deck, 11) == 0 {
		t.Fatal("undone 11 should be back in the draw pile")
	}
}

func TestManualRedoRestoresUndoneCard(t *testing.T) {
	in := strings.NewReader("2\nBob\n1\n5\n11\nUNDO\nREDO\nS\nS\n")
	svc := NewManualGameService(bufio.NewReader(in))
	svc.setupPlayers()
	svc.Game.RoundCount = 1
	svc.playRound()

	me := svc.Game.Players[0]
	if me.TotalScore != 11 {
		t.Fatalf("Me total=%d, want 11 after redo", me.TotalScore)
	}
}

func TestManualUndoAbortsJustOneMore(t *testing.T) {
	in := strings.NewReader("2\nBob\n1\n5\n8\nS\nJ\n2\nUNDO\nS\n")
	svc := NewManualGameService(bufio.NewReader(in))
	svc.setupPlayers()
	svc.Game.RoundCount = 1
	svc.playRound()

	me, bob := svc.Game.Players[0], svc.Game.Players[1]
	if me.TotalScore != 8 {
		t.Fatalf("Me total=%d, want 8", me.TotalScore)
	}
	if bob.TotalScore != 5 {
		t.Fatalf("Bob total=%d, want 5 (JOM undone, no forced card)", bob.TotalScore)
	}
	if bob.CurrentHand.Status != domain.HandStatusStayed {
		t.Fatalf("Bob status=%s, want stayed from his own Stay", bob.CurrentHand.Status)
	}
	if countAction(svc.Game.CurrentRound.Deck, domain.ActionJustOneMore) < 2 {
		t.Fatal("Just One More should be back in the draw pile after undo")
	}
}

func TestManualUIsUnluckySevenNotUndo(t *testing.T) {
	in := strings.NewReader("2\nBob\n1\n5\nU\nS\nS\n")
	svc := NewManualGameService(bufio.NewReader(in))
	svc.setupPlayers()
	svc.Game.RoundCount = 1
	svc.playRound()

	me := svc.Game.Players[0]
	if me.TotalScore != 7 {
		t.Fatalf("Me total=%d, want 7 from Unlucky 7 (U is not undo)", me.TotalScore)
	}
}

func TestManualUndoLastStayFromNextRound(t *testing.T) {
	in := strings.NewReader("2\nBob\n1\n\n5\n11\nS\nS\nUNDO\n7\nS\n")
	svc := NewManualGameService(bufio.NewReader(in))
	svc.Run()
	me := svc.Game.Players[0]
	if me.TotalScore != 18 {
		t.Fatalf("Me total=%d, want 18 (last Stay undone, then 11+7)", me.TotalScore)
	}
}

func TestReadLineConsumesEOFPartialLine(t *testing.T) {
	in := strings.NewReader("2\nBob\n1\n5\n11\nS\nS")
	svc := NewManualGameService(bufio.NewReader(in))
	svc.setupPlayers()
	svc.Game.RoundCount = 1
	svc.playRound()
	me := svc.Game.Players[0]
	if me.TotalScore != 11 {
		t.Fatalf("Me total=%d, want 11 from last Stay without trailing newline", me.TotalScore)
	}
}

func TestManualUndoAbortsSwapAfterFirstCard(t *testing.T) {
	in := strings.NewReader("2\nBob\n1\n5\n8\nS\nSW\nUNDO\nS\n")
	svc := NewManualGameService(bufio.NewReader(in))
	svc.setupPlayers()
	svc.Game.RoundCount = 1
	svc.playRound()
	me, bob := svc.Game.Players[0], svc.Game.Players[1]
	if me.TotalScore != 8 || bob.TotalScore != 5 {
		t.Fatalf("scores Me=%d Bob=%d, want 8 and 5 (Swap aborted)", me.TotalScore, bob.TotalScore)
	}
	if len(me.CurrentHand.NumberLine) != 1 || me.CurrentHand.NumberLine[0].Spec.Value != 8 {
		t.Fatalf("Me should still have 8, line=%v", me.CurrentHand.NumberLine)
	}
}

func TestNestedFailedRedoDoesNotUseSuggested(t *testing.T) {
	p1 := domain.NewPlayer("Me", nil)
	p2 := domain.NewPlayer("Bob", nil)
	game := domain.NewGame([]*domain.Player{p1, p2})
	game.Deck = domain.NewUnshuffledDeck()
	game.CurrentRound = domain.NewRound([]*domain.Player{p1, p2}, p1, game.Deck)
	p1.CurrentHand.ReceiveNumberLike(domain.NewNumberCard(8))
	p2.CurrentHand.ReceiveNumberLike(domain.NewNumberCard(3))
	mod, ok := game.CurrentRound.Deck.RemoveMatching(domain.NewModifierCard(domain.ModifierMinus4).Spec)
	if !ok {
		t.Fatal("expected -4")
	}
	svc := NewManualGameService(bufio.NewReader(strings.NewReader("REDO\n1\n")))
	svc.Game = game
	svc.assignModifier(p1, mod)
	if len(p1.CurrentHand.ModifierLine) != 1 {
		t.Fatalf("failed REDO must re-prompt, then 1 assigns -4 to Me, mods=%v", p1.CurrentHand.ModifierLine)
	}
}

func TestUndoCommandTrimsWhitespace(t *testing.T) {
	if !isUndoCommand(" <") || !isRedoCommand(" >") {
		t.Fatal("UNDO/REDO aliases should match after TrimSpace")
	}
}

func TestSnapshotRestoreRelinksOfferOrder(t *testing.T) {
	p1 := domain.NewPlayer("Me", nil)
	p2 := domain.NewPlayer("Bob", nil)
	svc := NewManualGameService(bufio.NewReader(strings.NewReader("")))
	svc.Game = domain.NewGame([]*domain.Player{p1, p2})
	svc.Game.DealerIndex = 0
	svc.Game.Deck = domain.NewUnshuffledDeck()
	svc.Game.CurrentRound = domain.NewRound(svc.Game.Players, p1, svc.Game.Deck)
	svc.phase = phaseDeal
	svc.cursor = 0
	svc.PushState()
	p2.CurrentHand.ReceiveNumberLike(domain.NewNumberCard(5))
	svc.cursor = 1
	svc.PushState()

	p2.CurrentHand.ReceiveNumberLike(domain.NewNumberCard(9))
	svc.cursor = 2
	if !svc.tryAbortCurrent() {
		t.Fatal("expected abort of in-progress action")
	}
	if svc.cursor != 1 {
		t.Fatalf("cursor=%d, want 1", svc.cursor)
	}
	if svc.phase != phaseDeal {
		t.Fatalf("phase=%s, want deal", svc.phase)
	}
	if len(svc.Game.CurrentRound.OfferOrder) != 2 {
		t.Fatalf("offer order not rebuilt: %d", len(svc.Game.CurrentRound.OfferOrder))
	}
	if svc.Game.CurrentRound.OfferOrder[0] != svc.Game.Players[1] {
		t.Fatal("offer order should start at dealer's left")
	}
	if len(svc.Game.Players[1].CurrentHand.NumberLine) != 1 {
		t.Fatalf("Bob line=%v, want only the 5", svc.Game.Players[1].CurrentHand.NumberLine)
	}
	if svc.Game.Players[1].Strategy == nil {
		t.Fatal("strategy should be relinked")
	}
}

func countNumber(d *domain.Deck, n domain.NumberValue) int {
	if d == nil {
		return 0
	}
	c := 0
	for _, card := range d.Cards {
		if card.Spec.Type == domain.CardTypeNumber && card.Spec.Value == n {
			c++
		}
	}
	return c
}

func countAction(d *domain.Deck, a domain.ActionType) int {
	if d == nil {
		return 0
	}
	c := 0
	for _, card := range d.Cards {
		if card.Spec.Type == domain.CardTypeAction && card.Spec.ActionType == a {
			c++
		}
	}
	return c
}
