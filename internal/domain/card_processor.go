package domain

// CardProcessResult contains the result of processing a card.
type CardProcessResult struct {
	Busted         bool
	Flip7          bool
	DiscardedCards []Card
	RemovedPlayer  bool // Whether the player should be removed from active players
}

// CardProcessor handles the logic of processing card draws.
type CardProcessor struct{}

// NewCardProcessor creates a new CardProcessor.
func NewCardProcessor() *CardProcessor {
	return &CardProcessor{}
}

// ProcessCard processes a card draw for a player and returns the result.
// This method handles:
// - Adding the card to the player's hand
// - Checking for bust conditions
// - Checking for Flip 7
// - Handling discarded cards (e.g., from Second Chance usage)
func (cp *CardProcessor) ProcessCard(p *Player, card Card) CardProcessResult {
	result := CardProcessResult{}

	// Add card to hand
	busted, flip7, discarded := p.CurrentHand.AddCard(card)
	result.Busted = busted
	result.Flip7 = flip7
	result.DiscardedCards = discarded

	// Determine if player should be removed from active players
	if busted {
		p.CurrentHand.Status = HandStatusBusted
		result.RemovedPlayer = true
	} else if flip7 {
		p.CurrentHand.Status = HandStatusStayed
		p.BankCurrentHand()
		result.RemovedPlayer = true
	}

	return result
}
