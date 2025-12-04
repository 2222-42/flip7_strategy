package application_test

import (
	"testing"

	"flip7_strategy/internal/application"
	"flip7_strategy/internal/domain"
)

// MockStrategy for predictable testing
type MockStrategy struct {
	DecideResult       domain.TurnChoice
	ChooseTargetResult *domain.Player
}

func (s *MockStrategy) Name() string { return "Mock" }
func (s *MockStrategy) Decide(deck *domain.Deck, hand *domain.PlayerHand, score int, others []*domain.Player) domain.TurnChoice {
	return s.DecideResult
}
func (s *MockStrategy) ChooseTarget(action domain.ActionType, candidates []*domain.Player, self *domain.Player) *domain.Player {
	if s.ChooseTargetResult != nil {
		return s.ChooseTargetResult
	}
	return candidates[0]
}

func TestFlipThreeLogic(t *testing.T) {
	// Setup
	p1 := domain.NewPlayer("P1", &MockStrategy{DecideResult: domain.TurnChoiceStay})
	p2 := domain.NewPlayer("P2", &MockStrategy{DecideResult: domain.TurnChoiceStay})
	players := []*domain.Player{p1, p2}
	game := domain.NewGame(players)
	svc := application.NewGameService(game)
	svc.Silent = true

	// Manually rig the deck to test Flip Three sequence
	// We want P1 to Flip Three on P2.
	// P2 should draw 3 cards.
	// Let's force the deck to have specific cards.
	// Since we can't easily inject a rigged deck into the service without modifying it,
	// we'll just test the ExecuteFlipThree method if possible, or simulate the flow.
	// But ExecuteFlipThree uses round.Deck.

	// Create a round
	deck := domain.NewDeck()
	// Rig deck: Top cards will be drawn by P2.
	// We need to access the internal slice of cards in Deck.
	// Since Deck.Cards is public, we can modify it.

	// Card 1: Number 5
	// Card 2: Second Chance (Action)
	// Card 3: Number 6
	deck.Cards = []domain.Card{
		{Type: domain.CardTypeNumber, Value: 5},
		{Type: domain.CardTypeAction, ActionType: domain.ActionSecondChance},
		{Type: domain.CardTypeNumber, Value: 6},
	}
	// Update RemainingCounts to match the rigged deck to avoid inconsistency
	deck.RemainingCounts = make(map[domain.NumberValue]int)
	deck.RemainingCounts[5] = 1
	deck.RemainingCounts[6] = 1

	game.CurrentRound = domain.NewRound(players, p1, deck)

	// Execute
	svc.ExecuteFlipThree(p2)

	// Assertions
	if len(p2.CurrentHand.RawNumberCards) != 2 {
		t.Errorf("Expected P2 to have 2 number cards, got %d", len(p2.CurrentHand.RawNumberCards))
	}
	if !p2.CurrentHand.HasSecondChance() {
		t.Errorf("Expected P2 to have a Second Chance card")
	}
	if p2.CurrentHand.Status != domain.HandStatusActive {
		t.Errorf("Expected P2 to be Active, got %s", p2.CurrentHand.Status)
	}
}

func TestFreezeLogic(t *testing.T) {
	p1 := domain.NewPlayer("P1", &MockStrategy{})
	players := []*domain.Player{p1}
	game := domain.NewGame(players)
	svc := application.NewGameService(game)
	svc.Silent = true

	game.CurrentRound = domain.NewRound(players, p1, domain.NewDeck())

	// Give P1 some points
	p1.CurrentHand.AddCard(domain.Card{Type: domain.CardTypeNumber, Value: 10})

	// Resolve Freeze on P1
	card := domain.Card{Type: domain.CardTypeAction, ActionType: domain.ActionFreeze}
	svc.ProcessCardDraw(p1, card)

	if p1.CurrentHand.Status != domain.HandStatusFrozen {
		t.Errorf("Expected P1 to be Frozen, got %s", p1.CurrentHand.Status)
	}
	if p1.TotalScore != 10 {
		t.Errorf("Expected P1 to bank 10 points, got %d", p1.TotalScore)
	}
}

func TestReshuffleLogic(t *testing.T) {
	// Setup
	p1 := domain.NewPlayer("P1", &MockStrategy{})
	players := []*domain.Player{p1}
	game := domain.NewGame(players)
	svc := application.NewGameService(game)
	svc.Silent = true

	// Initialize DiscardPile with known cards
	discardedCards := []domain.Card{
		{Type: domain.CardTypeNumber, Value: 1},
		{Type: domain.CardTypeNumber, Value: 2},
	}
	game.DiscardPile = discardedCards

	// Initialize Deck with 1 card
	deck := domain.NewDeck()
	deck.Cards = []domain.Card{
		{Type: domain.CardTypeNumber, Value: 3},
	}
	deck.RemainingCounts = map[domain.NumberValue]int{3: 1}

	game.CurrentRound = domain.NewRound(players, p1, deck)

	// 1. Draw the last card from the deck
	card1, err := svc.DrawCard()
	if err != nil {
		t.Fatalf("Expected successful draw, got error: %v", err)
	}
	if card1.Value != 3 {
		t.Errorf("Expected card value 3, got %d", card1.Value)
	}

	// 2. Deck is now empty. Draw again -> should trigger reshuffle from discard pile
	card2, err := svc.DrawCard()
	if err != nil {
		t.Fatalf("Expected successful draw after reshuffle, got error: %v", err)
	}

	// Verify card comes from discard pile (1 or 2)
	if card2.Value != 1 && card2.Value != 2 {
		t.Errorf("Expected card value 1 or 2, got %d", card2.Value)
	}

	// Verify DiscardPile is empty
	if len(game.DiscardPile) != 0 {
		t.Errorf("Expected DiscardPile to be empty, got %d cards", len(game.DiscardPile))
	}

	// Verify Deck has the remaining card
	if len(game.CurrentRound.Deck.Cards) != 1 {
		t.Errorf("Expected Deck to have 1 remaining card, got %d", len(game.CurrentRound.Deck.Cards))
	}
}

func TestRoundCountIncrement(t *testing.T) {
	// We want to verify that RunGame increments RoundCount.
	// To do this without running a long game, we can use a rigged deck that ensures a quick win.
	// P1 needs 200 points.
	// Round 1: P1 draws 200 points worth of cards, stays, wins.
	// Game should end. RoundCount should be 1.

	p1 := domain.NewPlayer("P1", &MockStrategy{DecideResult: domain.TurnChoiceStay})
	players := []*domain.Player{p1}
	game := domain.NewGame(players)
	svc := application.NewGameService(game)
	svc.Silent = true

	// Rig the deck for the first round
	// Note: RunGame creates a NEW deck for the first round.
	// We cannot easily inject a pre-configured deck into RunGame's first round because it calls domain.NewDeck().
	// However, we can modify the GameService to accept a DeckFactory or similar, OR we can rely on the fact that
	// if we can't rig the deck easily, we might need to test PlayRound directly or accept that RunGame is hard to test deterministically without DI.

	// BUT, we can check if RoundCount increments by manually running PlayRound loop?
	// The PR comment suggested: "Add a test that runs a complete game and verifies final RoundCount"
	// Let's try to run a game where P1 wins quickly.
	// Since we can't rig the deck in RunGame easily (it calls NewDeck), we might have to rely on probability or just test that it's NOT 0 after a game.
	// A better approach for this specific test might be to manually set up the game state as if it were inside RunGame loop.

	// Actually, let's just test that PlayRound doesn't reset it, and we can infer RunGame works if we trust the loop.
	// But to satisfy the reviewer, let's try to make a test that runs RunGame.

	// If we can't rig the deck, maybe we can rig the player?
	// If P1 starts with 200 points?
	// NewGame takes players.
	p1.TotalScore = 200
	// RunGame loop checks IsCompleted.
	// If P1 has 200, DetermineWinners will find him.
	// But RunGame loop condition is !s.Game.IsCompleted.
	// Inside loop:
	// 1. NewRound
	// 2. PlayRound
	// 3. Check winners.

	// So if P1 has 200, PlayRound runs once, then winners found, loop breaks.
	// RoundCount should be 1.

	svc.RunGame()

	if game.RoundCount != 1 {
		t.Errorf("Expected RoundCount to be 1, got %d", game.RoundCount)
	}
	if !game.IsCompleted {
		t.Errorf("Expected game to be completed")
	}
}
