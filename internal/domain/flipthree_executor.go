package domain

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

// FlipThreeLogger handles logging during Flip Three execution.
type FlipThreeLogger interface {
	LogStart(target *Player)
	LogCardDraw(target *Player, cardNum int, card Card)
	LogActionQueued(card Card)
	LogResolvingQueued(card Card)
	LogFlip7(target *Player, score int)
	LogEnd(target *Player)
	LogError(msg string)
}

// FlipThreeExecutor centralizes the Flip Three execution logic.
// This eliminates duplication between game_service.go and manual_game_service.go.
type FlipThreeExecutor struct {
	cardSource    FlipThreeCardSource
	cardProcessor FlipThreeCardProcessor
	logger        FlipThreeLogger
}

// NewFlipThreeExecutor creates a new FlipThreeExecutor.
func NewFlipThreeExecutor(source FlipThreeCardSource, processor FlipThreeCardProcessor, logger FlipThreeLogger) *FlipThreeExecutor {
	return &FlipThreeExecutor{
		cardSource:    source,
		cardProcessor: processor,
		logger:        logger,
	}
}

// Execute runs the Flip Three logic: draw 3 cards with specific handling rules.
// Per domain model (docs/domain_model.md lines 169-172):
// 1. Draw 3 cards one by one
// 2. If Second Chance is drawn: Process immediately
// 3. If Flip Three or Freeze is drawn: Queue and resolve AFTER all 3 cards drawn
//    (only if the target hasn't busted)
// 4. Number/Modifier cards: Process immediately
// Returns true if the round should end (e.g., Flip 7 achieved).
func (fte *FlipThreeExecutor) Execute(target *Player, round *Round) bool {
	fte.logger.LogStart(target)
	
	queuedActions := []Card{}
	
	for i := 0; i < 3; i++ {
		// Exit early if target is no longer active
		if target.CurrentHand.Status != HandStatusActive {
			break
		}
		
		// Get the next card
		card, err := fte.cardSource.GetNextCard(i+1, target)
		if err != nil {
			fte.logger.LogError(err.Error())
			round.IsEnded = true
			round.EndReason = RoundEndReasonAborted
			return true
		}
		
		fte.logger.LogCardDraw(target, i+1, card)
		
		// Handle cards according to Flip Three rules
		if card.Type == CardTypeAction {
			if card.ActionType == ActionSecondChance {
				// Second Chance: Process immediately
				if err := fte.cardProcessor.ProcessImmediateCard(target, card); err != nil {
					fte.logger.LogError(err.Error())
				}
				
				// Check if player became inactive after processing
				if target.CurrentHand.Status != HandStatusActive {
					break
				}
				continue
				
			} else if card.ActionType == ActionFlipThree || card.ActionType == ActionFreeze {
				// Flip Three/Freeze: Queue for later
				fte.logger.LogActionQueued(card)
				queuedActions = append(queuedActions, card)
				
				// Add action card to hand (without triggering immediate resolution)
				target.CurrentHand.ActionCards = append(target.CurrentHand.ActionCards, card)
				
				// Check for Flip 7 (7 unique number cards)
				if len(target.CurrentHand.NumberCards) >= 7 {
					target.CurrentHand.Status = HandStatusStayed
					score := target.BankCurrentHand()
					fte.logger.LogFlip7(target, score)
					
					round.RemoveActivePlayer(target)
					round.EndReason = RoundEndReasonFlip7
					round.IsEnded = true
					return true
				}
				continue
			}
		}
		
		// Number/Modifier cards: Process immediately
		if err := fte.cardProcessor.ProcessImmediateCard(target, card); err != nil {
			fte.logger.LogError(err.Error())
		}
		
		// Check if round ended or player became inactive
		if round.IsEnded || target.CurrentHand.Status != HandStatusActive {
			break
		}
	}
	
	// Resolve queued actions if player is still active
	if target.CurrentHand.Status == HandStatusActive {
		for _, actionCard := range queuedActions {
			fte.logger.LogResolvingQueued(actionCard)
			
			if err := fte.cardProcessor.ProcessQueuedAction(target, actionCard); err != nil {
				fte.logger.LogError(err.Error())
			}
			
			if target.CurrentHand.Status != HandStatusActive {
				break
			}
		}
	} else {
		// If player busted during draws, ensure queued action cards are in hand
		// (they were drawn, so they should be tracked)
		// Note: ActionCards were already added to hand above, so nothing to do here
	}
	
	fte.logger.LogEnd(target)
	return round.IsEnded
}
