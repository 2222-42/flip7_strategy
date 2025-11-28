package strategy_test

import (
	"testing"

	"flip7_strategy/internal/domain"
	"flip7_strategy/internal/domain/strategy"

	"github.com/google/uuid"
)

func TestAdaptiveStrategy_Decide(t *testing.T) {
	s := strategy.NewAdaptiveStrategy()

	// Scenario:
	// Hand has 6 cards. High Score (50).
	// Deck has 10 cards: 4 duplicates (bust), 6 new (safe, low value).
	// Risk = 0.4.
	// Aggressive Strategy: Hits (because totalCards in hand is 6 and risk < 0.5, per AggressiveStrategy's totalCards check).
	// Expected Value Strategy:
	//   Current Score = 50.
	//   Avg Safe Card Value = 3.5.
	//   New Score (Safe) = 50 + 3.5 + 15 (Bonus) = 68.5.
	//   EV = 0.6 * 68.5 = 41.1.
	//   EV (41.1) < Current (50) -> STAY.

	// Setup Hand
	hand := domain.NewPlayerHand()
	handNumbers := []int{0, 8, 9, 10, 11, 12} // Sum = 50
	for _, n := range handNumbers {
		hand.AddCard(domain.Card{Type: domain.CardTypeNumber, Value: domain.NumberValue(n)})
	}

	// Setup Deck
	// 4 bad cards (8, 9, 10, 11)
	// 6 good cards (1, 2, 3, 4, 5, 6)
	deckCards := []domain.Card{
		{Type: domain.CardTypeNumber, Value: 8},
		{Type: domain.CardTypeNumber, Value: 9},
		{Type: domain.CardTypeNumber, Value: 10},
		{Type: domain.CardTypeNumber, Value: 11},
		{Type: domain.CardTypeNumber, Value: 1},
		{Type: domain.CardTypeNumber, Value: 2},
		{Type: domain.CardTypeNumber, Value: 3},
		{Type: domain.CardTypeNumber, Value: 4},
		{Type: domain.CardTypeNumber, Value: 5},
		{Type: domain.CardTypeNumber, Value: 6},
	}
	deck := domain.NewDeckFromCards(deckCards)

	// Test Case 1: No opponent threat -> Should use EV -> STAY
	t.Run("No Opponent Threat (Use EV)", func(t *testing.T) {
		otherPlayers := []*domain.Player{
			{ID: uuid.New(), TotalScore: 100},
			{ID: uuid.New(), TotalScore: 150},
		}
		choice := s.Decide(deck, hand, 100, otherPlayers)
		if choice != domain.TurnChoiceStay {
			t.Errorf("Expected Stay (EV behavior), got %v", choice)
		}
	})

	// Test Case 2: Opponent threat -> Should use Aggressive -> HIT
	t.Run("Opponent Threat (Use Aggressive)", func(t *testing.T) {
		otherPlayers := []*domain.Player{
			{ID: uuid.New(), TotalScore: 100},
			{ID: uuid.New(), TotalScore: 205}, // Threat!
		}
		choice := s.Decide(deck, hand, 100, otherPlayers)
		if choice != domain.TurnChoiceHit {
			t.Errorf("Expected Hit (Aggressive behavior), got %v", choice)
		}
	})
}

func TestAdaptiveStrategy_ChooseTarget(t *testing.T) {
	s := strategy.NewAdaptiveStrategy()

	// AdaptiveStrategy embeds CommonTargetChooser, so it should behave like it.
	// CommonTargetChooser logic:
	// Freeze -> Self
	// FlipThree -> Leader opponent
	// GiveSecondChance -> Weakest opponent

	self := &domain.Player{ID: uuid.New(), TotalScore: 100, CurrentHand: domain.NewPlayerHand()}
	p2 := &domain.Player{ID: uuid.New(), TotalScore: 150, CurrentHand: domain.NewPlayerHand()} // Leader
	p3 := &domain.Player{ID: uuid.New(), TotalScore: 50, CurrentHand: domain.NewPlayerHand()}  // Weakest
	candidates := []*domain.Player{self, p2, p3}
	deck := domain.NewDeck()

	// Test Freeze -> Self
	target := s.ChooseTarget(domain.ActionFreeze, candidates, self)
	if target.ID != self.ID {
		t.Errorf("Expected Freeze target to be Self, got %v", target.ID)
	}

	// Test FlipThree -> Leader (p2)
	s.SetDeck(deck) // Inject deck for test
	target = s.ChooseTarget(domain.ActionFlipThree, candidates, self)
	if target.ID != p2.ID {
		t.Errorf("Expected FlipThree target to be Leader (p2), got %v", target.ID)
	}

	// Test GiveSecondChance -> Weakest (p3)
	target = s.ChooseTarget(domain.ActionGiveSecondChance, candidates, self)
	if target.ID != p3.ID {
		t.Errorf("Expected GiveSecondChance target to be Weakest (p3), got %v", target.ID)
	}
}
