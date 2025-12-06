package domain

import "fmt"

// FlipThreeCardSource provides cards for Flip Three execution.
// Different implementations exist for AI mode (deck) and manual mode (user input).
type FlipThreeCardSource interface {
	// GetNextCard returns the next card for Flip Three.
	// Returns error if no card is available (deck empty, invalid input, etc.).
	GetNextCard(cardNum int, target *Player) (Card, error)
}

// FlipThreeCardProcessor handles processing of cards during Flip Three.
// Different implementations exist for AI mode and manual mode.
type FlipThreeCardProcessor interface {
	// ProcessImmediateCard processes cards that should be handled immediately
	// (Number, Modifier, Second Chance cards).
	ProcessImmediateCard(target *Player, card Card) error

	// ProcessQueuedAction processes a queued action card (Flip Three, Freeze)
	// after all 3 cards have been drawn.
	ProcessQueuedAction(target *Player, card Card) error
}

// FlipThreeLogger is an optional callback for logging during Flip Three execution.
// If nil, no logging occurs. This decouples the domain from specific logging implementations.
type FlipThreeLogger func(message string)

// FlipThreeExecutor centralizes the Flip Three execution logic.
// This eliminates duplication between game_service.go and manual_game_service.go.
type FlipThreeExecutor struct {
	cardSource    FlipThreeCardSource
	cardProcessor FlipThreeCardProcessor
	logger        FlipThreeLogger
}

// NewFlipThreeExecutor creates a new FlipThreeExecutor.
// logger can be nil if no logging is needed.
func NewFlipThreeExecutor(source FlipThreeCardSource, processor FlipThreeCardProcessor, logger FlipThreeLogger) *FlipThreeExecutor {
	return &FlipThreeExecutor{
		cardSource:    source,
		cardProcessor: processor,
		logger:        logger,
	}
}

// log is a helper to call the logger if it's not nil.
func (fte *FlipThreeExecutor) log(format string, args ...interface{}) {
	if fte.logger != nil {
		fte.logger(fmt.Sprintf(format, args...))
	}
}

// Execute runs the Flip Three logic: draw 3 cards with specific handling rules.
// Per domain model (docs/domain_model.md lines 169-172):
//  1. Draw 3 cards one by one
//  2. If Second Chance is drawn: Process immediately
//  3. If Flip Three or Freeze is drawn: Queue and resolve AFTER all 3 cards drawn
//     (only if the target hasn't busted)
//  4. Number/Modifier cards: Process immediately
//
// Returns true if the round should end (e.g., Flip 7 achieved).
func (fte *FlipThreeExecutor) Execute(target *Player, round *Round) bool {
	fte.log("--- %s must draw 3 cards! ---", target.Name)

	queuedActions := []Card{}

	for i := 0; i < FlipThreeCardCount; i++ {
		// Exit early if target is no longer active
		if target.CurrentHand.Status != HandStatusActive {
			break
		}

		// Get the next card
		card, err := fte.cardSource.GetNextCard(i+1, target)
		if err != nil {
			fte.log("Error: %s", err.Error())
			round.IsEnded = true
			round.EndReason = RoundEndReasonAborted
			return true
		}

		fte.log("%s forced draw (%d/3): %v", target.Name, i+1, card)

		// Handle cards according to Flip Three rules
		if card.Type == CardTypeAction {
			if card.ActionType == ActionSecondChance {
				// Second Chance: Process immediately
				if err := fte.cardProcessor.ProcessImmediateCard(target, card); err != nil {
					fte.log("Error: %s", err.Error())
				}

				// Check if player became inactive after processing
				if target.CurrentHand.Status != HandStatusActive {
					break
				}
				continue

			} else if card.ActionType == ActionFlipThree || card.ActionType == ActionFreeze {
				// Flip Three/Freeze: Queue for later
				fte.log("Action %s queued for after Flip Three", card.ActionType)
				queuedActions = append(queuedActions, card)

				// Add action card to hand WITHOUT triggering immediate resolution.
				// We bypass AddCard() here intentionally because:
				// 1. We don't want to trigger the action immediately (it's queued)
				// 2. AddCard() would check for Flip 7, which we do manually below
				// 3. Action cards don't add to NumberCards, so no bust risk
				target.CurrentHand.ActionCards = append(target.CurrentHand.ActionCards, card)

				// Manually check for Flip 7 (7 unique number cards).
				// We must check here because Flip 7 ends the round immediately,
				// preventing further card draws during Flip Three.
				// Note: Action cards don't add to NumberCards, but we check
				// because the player might have already had 7 number cards.
				if len(target.CurrentHand.NumberCards) >= 7 {
					target.CurrentHand.Status = HandStatusStayed
					score := target.BankCurrentHand()
					fte.log("%s FLIP 7! Bonus! Banked %d points! Total: %d", target.Name, score, target.TotalScore)

					round.RemoveActivePlayer(target)
					round.EndReason = RoundEndReasonFlip7
					round.IsEnded = true
					return true
				}
				continue
			}
		}

		// Number/Modifier cards: Process immediately.
		// Note: ProcessImmediateCard calls the full card processing logic
		// (ProcessCardDraw in AI mode, processCard in manual mode), which
		// handles Flip 7 detection automatically via AddCard().
		if err := fte.cardProcessor.ProcessImmediateCard(target, card); err != nil {
			fte.log("Error: %s", err.Error())
		}

		// Check if round ended or player became inactive
		if round.IsEnded || target.CurrentHand.Status != HandStatusActive {
			break
		}
	}

	// Resolve queued actions if player is still active
	if target.CurrentHand.Status == HandStatusActive {
		for _, actionCard := range queuedActions {
			fte.log("Resolving queued action %s...", actionCard.ActionType)

			if err := fte.cardProcessor.ProcessQueuedAction(target, actionCard); err != nil {
				fte.log("Error: %s", err.Error())
			}

			if target.CurrentHand.Status != HandStatusActive {
				break
			}
		}
	}
	// Note: If player busted during draws, queued action cards were already added to hand above

	fte.log("--- End of Flip Three for %s ---", target.Name)
	return round.IsEnded
}
