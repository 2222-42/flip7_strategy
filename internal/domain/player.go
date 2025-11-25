package domain

import (
	"github.com/google/uuid"
)

// HandStatus represents the state of a player's hand.
type HandStatus string

const (
	HandStatusActive HandStatus = "active"
	HandStatusStayed HandStatus = "stayed"
	HandStatusBusted HandStatus = "busted"
	HandStatusFrozen HandStatus = "frozen"
)

// PointValue represents the calculated score of a hand.
type PointValue struct {
	BaseSum   int   `json:"base_sum"`
	Modifiers []int `json:"modifiers"`
	Bonus     int   `json:"bonus"`
	Total     int   `json:"total"`
}

// PlayerHand represents a player's cards in a single round.
type PlayerHand struct {
	ID               uuid.UUID                `json:"id"`
	NumberCards      map[NumberValue]struct{} `json:"number_cards"`     // Set for uniqueness check
	RawNumberCards   []NumberValue            `json:"raw_number_cards"` // For display/calculation
	ModifierCards    []Card                   `json:"modifier_cards"`
	ActionCards      []Card                   `json:"action_cards"`
	SecondChanceUsed bool                     `json:"second_chance_used"`
	Status           HandStatus               `json:"status"`
}

// HasSecondChance checks if the hand contains an unused Second Chance card.
func (h *PlayerHand) HasSecondChance() bool {
	for _, c := range h.ActionCards {
		if c.ActionType == ActionSecondChance {
			return true
		}
	}
	return false
}

// NewPlayerHand creates a new empty hand.
func NewPlayerHand() *PlayerHand {
	return &PlayerHand{
		ID:          uuid.New(),
		NumberCards: make(map[NumberValue]struct{}),
		Status:      HandStatusActive,
	}
}

// AddCard adds a card to the hand and checks for bust.
// Returns busted=true if the card caused a bust (duplicate number).
// Returns discarded cards if Second Chance was used.
func (h *PlayerHand) AddCard(card Card) (busted bool, flip7 bool, discarded []Card) {
	if h.Status != HandStatusActive {
		return false, false, nil
	}

	switch card.Type {
	case CardTypeNumber:
		if _, exists := h.NumberCards[card.Value]; exists {
			if !h.SecondChanceUsed {
				// If player has a Second Chance card, use it to avoid the bust.
				hasSecondChance := false
				scIndex := -1
				for i, c := range h.ActionCards {
					if c.ActionType == ActionSecondChance {
						hasSecondChance = true
						scIndex = i
						break
					}
				}

				if hasSecondChance {
					// Use Second Chance: Discard the duplicate (don't add it), discard the Second Chance card.
					h.SecondChanceUsed = true

					// Collect discarded cards
					scCard := h.ActionCards[scIndex]
					discarded = append(discarded, scCard)
					discarded = append(discarded, card)

					// Remove the Second Chance card
					h.ActionCards = append(h.ActionCards[:scIndex], h.ActionCards[scIndex+1:]...)
					return false, false, discarded
				}
			}
			h.Status = HandStatusBusted
			// Add the busting card to the hand so it stays on the table until round end
			h.RawNumberCards = append(h.RawNumberCards, card.Value)
			h.NumberCards[card.Value] = struct{}{}
			return true, false, nil
		}
		h.NumberCards[card.Value] = struct{}{}
		h.RawNumberCards = append(h.RawNumberCards, card.Value)

	case CardTypeModifier:
		h.ModifierCards = append(h.ModifierCards, card)

	case CardTypeAction:
		h.ActionCards = append(h.ActionCards, card)
		// Note: Immediate actions like FlipThree need to be handled by the caller/Round.
	}

	// Flip 7 condition: triggers when the player has 7 or more cards in hand (of any type) without busting.
	totalCards := len(h.RawNumberCards) + len(h.ModifierCards) + len(h.ActionCards)
	if totalCards >= 7 && h.Status == HandStatusActive {
		return false, true, nil
	}

	return false, false, nil
}

// Player represents a participant in the game.
type Player struct {
	ID          uuid.UUID   `json:"id"`
	Name        string      `json:"name"`
	TotalScore  int         `json:"total_score"`
	CurrentHand *PlayerHand `json:"current_hand"`
	Strategy    Strategy    `json:"-"` // AI Strategy
}

// NewPlayer creates a new player.
func NewPlayer(name string, strategy Strategy) *Player {
	return &Player{
		ID:       uuid.New(),
		Name:     name,
		Strategy: strategy,
	}
}

// StartNewRound prepares the player for a new round.
func (p *Player) StartNewRound() {
	p.CurrentHand = NewPlayerHand()
}

// BankScore adds the current hand's score to the total.
func (p *Player) BankScore(score int) {
	p.TotalScore += score
}
