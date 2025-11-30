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

	// Test Freeze -> Leader (p2)
	target := s.ChooseTarget(domain.ActionFreeze, candidates, self)
	if target.ID != p2.ID {
		t.Errorf("Expected Freeze target to be Leader (p2), got %v", target.ID)
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

func TestAdaptiveStrategy_ChooseTarget_DynamicDelegation(t *testing.T) {
	s := strategy.NewAdaptiveStrategy()

	self := &domain.Player{
		ID:         domain.NewPlayer("Self", nil).ID,
		Name:       "Self",
		TotalScore: 100,
		CurrentHand: &domain.PlayerHand{
			NumberCards: make(map[domain.NumberValue]struct{}),
		},
	}

	// Create opponents - one with high score (threat)
	opponentThreat := &domain.Player{
		ID:         domain.NewPlayer("OpponentThreat", nil).ID,
		Name:       "OpponentThreat",
		TotalScore: 205, // Above winning threshold (200)
		CurrentHand: &domain.PlayerHand{
			NumberCards: make(map[domain.NumberValue]struct{}),
		},
	}

	opponentNormal := &domain.Player{
		ID:         domain.NewPlayer("OpponentNormal", nil).ID,
		Name:       "OpponentNormal",
		TotalScore: 150, // Below winning threshold
		CurrentHand: &domain.PlayerHand{
			NumberCards: make(map[domain.NumberValue]struct{}),
		},
	}

	opponentWeak := &domain.Player{
		ID:         domain.NewPlayer("OpponentWeak", nil).ID,
		Name:       "OpponentWeak",
		TotalScore: 50,
		CurrentHand: &domain.PlayerHand{
			NumberCards: make(map[domain.NumberValue]struct{}),
		},
	}

	deck := domain.NewDeck()
	s.SetDeck(deck)

	t.Run("Uses Aggressive targeting when opponent >= 200", func(t *testing.T) {
		// With a threat opponent (>= 200), should use Aggressive strategy targeting
		// Aggressive uses RandomTargetSelector which targets random opponents
		candidates := []*domain.Player{self, opponentThreat, opponentNormal}
		target := s.ChooseTarget(domain.ActionFlipThree, candidates, self)

		// Should target one of the opponents (not self)
		if target.ID == self.ID {
			t.Errorf("Expected to target an opponent, got self")
		}
	})

	t.Run("Uses ExpectedValue targeting when opponent < 200", func(t *testing.T) {
		// Without a threat opponent (all < 200), should use ExpectedValue strategy targeting
		// ExpectedValue uses DefaultTargetSelector which targets the leader
		candidates := []*domain.Player{self, opponentNormal, opponentWeak}
		target := s.ChooseTarget(domain.ActionFlipThree, candidates, self)

		// Should target the leader (opponentNormal with 150)
		if target.ID != opponentNormal.ID {
			t.Errorf("Expected to target leader opponent (OpponentNormal), got %s", target.Name)
		}
	})

	t.Run("GiveSecondChance targets weakest without threat", func(t *testing.T) {
		// Without threat, should use ExpectedValue targeting logic
		// GiveSecondChance targets weakest opponent
		candidates := []*domain.Player{self, opponentNormal, opponentWeak}
		target := s.ChooseTarget(domain.ActionGiveSecondChance, candidates, self)

		// Should target the weakest opponent
		if target.ID != opponentWeak.ID {
			t.Errorf("Expected to target weakest opponent (OpponentWeak), got %s", target.Name)
		}
	})
}
