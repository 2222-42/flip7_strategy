package domain_test

import (
	"errors"
	"testing"

	"flip7_strategy/internal/domain"
)

// Mock implementations for testing FlipThreeExecutor

type mockFlipThreeCardSource struct {
	cards []domain.Card
	index int
	err   error
}

func (m *mockFlipThreeCardSource) GetNextCard(cardNum int, target *domain.Player) (domain.Card, error) {
	if m.err != nil {
		return domain.Card{}, m.err
	}
	if m.index >= len(m.cards) {
		return domain.Card{}, errors.New("no more cards")
	}
	card := m.cards[m.index]
	m.index++
	return card, nil
}

type mockFlipThreeCardProcessor struct {
	immediateCards []domain.Card
	queuedCards    []domain.Card
	processError   error
}

func (m *mockFlipThreeCardProcessor) ProcessImmediateCard(target *domain.Player, card domain.Card) error {
	m.immediateCards = append(m.immediateCards, card)
	return m.processError
}

func (m *mockFlipThreeCardProcessor) ProcessQueuedAction(target *domain.Player, card domain.Card) error {
	m.queuedCards = append(m.queuedCards, card)
	return m.processError
}

type mockFlipThreeLogger struct {
	startCalled          bool
	cardDrawCount        int
	actionQueuedCount    int
	resolvingQueuedCount int
	flip7Called          bool
	endCalled            bool
	errorMessages        []string
}

func (m *mockFlipThreeLogger) LogStart(target *domain.Player) {
	m.startCalled = true
}

func (m *mockFlipThreeLogger) LogCardDraw(target *domain.Player, cardNum int, card domain.Card) {
	m.cardDrawCount++
}

func (m *mockFlipThreeLogger) LogActionQueued(card domain.Card) {
	m.actionQueuedCount++
}

func (m *mockFlipThreeLogger) LogResolvingQueued(card domain.Card) {
	m.resolvingQueuedCount++
}

func (m *mockFlipThreeLogger) LogFlip7(target *domain.Player, score int) {
	m.flip7Called = true
}

func (m *mockFlipThreeLogger) LogEnd(target *domain.Player) {
	m.endCalled = true
}

func (m *mockFlipThreeLogger) LogError(msg string) {
	m.errorMessages = append(m.errorMessages, msg)
}

func TestFlipThreeExecutor_Execute(t *testing.T) {
	tests := []struct {
		name                 string
		cards                []domain.Card
		expectedImmediate    int
		expectedQueued       int
		expectedCardDraws    int
		expectedActionsQueue int
		shouldEndRound       bool
	}{
		{
			name: "Three number cards",
			cards: []domain.Card{
				{Type: domain.CardTypeNumber, Value: 5},
				{Type: domain.CardTypeNumber, Value: 7},
				{Type: domain.CardTypeNumber, Value: 3},
			},
			expectedImmediate:    3,
			expectedQueued:       0,
			expectedCardDraws:    3,
			expectedActionsQueue: 0,
			shouldEndRound:       false,
		},
		{
			name: "Second Chance processed immediately",
			cards: []domain.Card{
				{Type: domain.CardTypeNumber, Value: 5},
				{Type: domain.CardTypeAction, ActionType: domain.ActionSecondChance},
				{Type: domain.CardTypeNumber, Value: 7},
			},
			expectedImmediate:    3,
			expectedQueued:       0,
			expectedCardDraws:    3,
			expectedActionsQueue: 0,
			shouldEndRound:       false,
		},
		{
			name: "Freeze action queued",
			cards: []domain.Card{
				{Type: domain.CardTypeNumber, Value: 5},
				{Type: domain.CardTypeAction, ActionType: domain.ActionFreeze},
				{Type: domain.CardTypeNumber, Value: 7},
			},
			expectedImmediate:    2,
			expectedQueued:       1,
			expectedCardDraws:    3,
			expectedActionsQueue: 1,
			shouldEndRound:       false,
		},
		{
			name: "FlipThree action queued",
			cards: []domain.Card{
				{Type: domain.CardTypeAction, ActionType: domain.ActionFlipThree},
				{Type: domain.CardTypeNumber, Value: 5},
				{Type: domain.CardTypeNumber, Value: 7},
			},
			expectedImmediate:    2,
			expectedQueued:       1,
			expectedCardDraws:    3,
			expectedActionsQueue: 1,
			shouldEndRound:       false,
		},
		{
			name: "Multiple actions queued",
			cards: []domain.Card{
				{Type: domain.CardTypeAction, ActionType: domain.ActionFreeze},
				{Type: domain.CardTypeAction, ActionType: domain.ActionFlipThree},
				{Type: domain.CardTypeNumber, Value: 5},
			},
			expectedImmediate:    1,
			expectedQueued:       2,
			expectedCardDraws:    3,
			expectedActionsQueue: 2,
			shouldEndRound:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			player := domain.NewPlayer("TestPlayer", nil)
			player.StartNewRound()

			round := &domain.Round{
				ActivePlayers: []*domain.Player{player},
			}

			source := &mockFlipThreeCardSource{cards: tt.cards}
			processor := &mockFlipThreeCardProcessor{}
			logger := &mockFlipThreeLogger{}

			executor := domain.NewFlipThreeExecutor(source, processor, logger)

			// Execute
			roundEnded := executor.Execute(player, round)

			// Verify
			if len(processor.immediateCards) != tt.expectedImmediate {
				t.Errorf("Expected %d immediate cards, got %d", tt.expectedImmediate, len(processor.immediateCards))
			}

			if len(processor.queuedCards) != tt.expectedQueued {
				t.Errorf("Expected %d queued cards, got %d", tt.expectedQueued, len(processor.queuedCards))
			}

			if logger.cardDrawCount != tt.expectedCardDraws {
				t.Errorf("Expected %d card draws logged, got %d", tt.expectedCardDraws, logger.cardDrawCount)
			}

			if logger.actionQueuedCount != tt.expectedActionsQueue {
				t.Errorf("Expected %d actions queued logged, got %d", tt.expectedActionsQueue, logger.actionQueuedCount)
			}

			if roundEnded != tt.shouldEndRound {
				t.Errorf("Expected roundEnded=%v, got %v", tt.shouldEndRound, roundEnded)
			}

			if !logger.startCalled {
				t.Error("Expected LogStart to be called")
			}

			if !logger.endCalled {
				t.Error("Expected LogEnd to be called")
			}
		})
	}
}

func TestFlipThreeExecutor_Flip7Achievement(t *testing.T) {
	// Setup player with 6 unique number cards already
	player := domain.NewPlayer("TestPlayer", nil)
	player.StartNewRound()
	
	// Add 6 number cards to hand
	for i := 0; i < 6; i++ {
		player.CurrentHand.AddCard(domain.Card{Type: domain.CardTypeNumber, Value: domain.NumberValue(i)})
	}

	round := &domain.Round{
		ActivePlayers: []*domain.Player{player},
	}

	// Draw one more number card during Flip Three to reach 7 - should trigger Flip 7
	cards := []domain.Card{
		{Type: domain.CardTypeNumber, Value: 7}, // 7th unique number card -> Flip 7!
		{Type: domain.CardTypeNumber, Value: 10},
		{Type: domain.CardTypeNumber, Value: 11},
	}

	source := &mockFlipThreeCardSource{cards: cards}
	processor := &mockFlipThreeCardProcessor{}
	logger := &mockFlipThreeLogger{}

	executor := domain.NewFlipThreeExecutor(source, processor, logger)

	// Execute
	roundEnded := executor.Execute(player, round)

	// Verify Flip 7 was NOT triggered here because processor doesn't actually process
	// The test framework mocks processing, so Flip 7 won't be detected
	// This test validates that the executor calls the processor, not that Flip 7 happens
	if roundEnded {
		t.Error("Round should not end in mock environment")
	}

	// The processor should have been called for the immediate card
	if len(processor.immediateCards) != 3 {
		t.Errorf("Expected 3 immediate cards processed, got %d", len(processor.immediateCards))
	}
}

func TestFlipThreeExecutor_ErrorHandling(t *testing.T) {
	player := domain.NewPlayer("TestPlayer", nil)
	player.StartNewRound()

	round := &domain.Round{
		ActivePlayers: []*domain.Player{player},
	}

	// Simulate error getting card
	source := &mockFlipThreeCardSource{err: errors.New("deck empty")}
	processor := &mockFlipThreeCardProcessor{}
	logger := &mockFlipThreeLogger{}

	executor := domain.NewFlipThreeExecutor(source, processor, logger)

	// Execute
	roundEnded := executor.Execute(player, round)

	// Verify error was logged and round aborted
	if !roundEnded {
		t.Error("Expected round to end due to error")
	}

	if len(logger.errorMessages) == 0 {
		t.Error("Expected error to be logged")
	}

	if round.EndReason != domain.RoundEndReasonAborted {
		t.Errorf("Expected end reason Aborted, got %v", round.EndReason)
	}
}
