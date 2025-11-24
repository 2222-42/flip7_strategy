package domain

import (
	"errors"
	"math/rand"
	"time"
)

// CardType represents the category of a card.
type CardType string

const (
	CardTypeNumber   CardType = "number"
	CardTypeModifier CardType = "modifier"
	CardTypeAction   CardType = "action"
)

// NumberValue represents the value of a number card (0-12).
type NumberValue int

// ModifierType represents the type of modifier card.
type ModifierType string

const (
	ModifierPlus2  ModifierType = "plus_2"
	ModifierPlus4  ModifierType = "plus_4"
	ModifierPlus6  ModifierType = "plus_6"
	ModifierPlus8  ModifierType = "plus_8"
	ModifierPlus10 ModifierType = "plus_10"
	ModifierX2     ModifierType = "multiply_2"
)

// ActionType represents the type of action card.
type ActionType string

const (
	ActionFreeze           ActionType = "freeze"
	ActionFlipThree        ActionType = "flip_three"
	ActionSecondChance     ActionType = "second_chance"
	ActionGiveSecondChance ActionType = "give_second_chance" // Pseudo-action for target selection
)

// Card represents a single card in the game.
type Card struct {
	Type         CardType     `json:"type"`
	Value        NumberValue  `json:"value,omitempty"`         // For Number cards
	ModifierType ModifierType `json:"modifier_type,omitempty"` // For Modifier cards
	ActionType   ActionType   `json:"action_type,omitempty"`   // For Action cards
}

// Deck represents the deck of cards.
type Deck struct {
	Cards           []Card              `json:"cards"`
	RemainingCounts map[NumberValue]int `json:"remaining_counts"`
}

// NewDeck creates a new shuffled deck.
func NewDeck() *Deck {
	cards := []Card{}
	counts := make(map[NumberValue]int)

	// Add Number cards: 0 (1 copy), 1 (1 copy), ..., 12 (12 copies)
	// Wait, the rules say: "0-12 pts". Usually in these games, the count matches the number?
	// Checking the domain model doc: "Numbers: 12:x12, 11:x11, ..., 1:x1, 0:x1."
	// Wait, usually 1 has 1 copy, 2 has 2 copies...
	// Let's stick to the doc: "12:x12... 0:x1".
	// Actually, 0 usually has special rules or count. The doc says "0:x1".

	// 0 to 12
	for i := 0; i <= 12; i++ {
		count := i
		if i == 0 {
			count = 1 // Special case for 0? Or maybe 0 doesn't exist? Doc says "0-12".
			// Let's assume 0 has 1 copy based on doc "0:x1".
		}

		val := NumberValue(i)
		counts[val] = count
		for j := 0; j < count; j++ {
			cards = append(cards, Card{Type: CardTypeNumber, Value: val})
		}
	}

	// Add Modifiers: 2x each
	modifiers := []ModifierType{ModifierPlus2, ModifierPlus4, ModifierPlus6, ModifierPlus8, ModifierPlus10, ModifierX2}
	for _, mod := range modifiers {
		for j := 0; j < 2; j++ {
			cards = append(cards, Card{Type: CardTypeModifier, ModifierType: mod})
		}
	}

	// Add Actions: 3x each
	actions := []ActionType{ActionFreeze, ActionFlipThree, ActionSecondChance}
	for _, act := range actions {
		for j := 0; j < 3; j++ {
			cards = append(cards, Card{Type: CardTypeAction, ActionType: act})
		}
	}

	d := &Deck{
		Cards:           cards,
		RemainingCounts: counts,
	}
	d.Shuffle()
	return d
}

// Shuffle randomizes the deck order.
func (d *Deck) Shuffle() {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	r.Shuffle(len(d.Cards), func(i, j int) {
		d.Cards[i], d.Cards[j] = d.Cards[j], d.Cards[i]
	})
}

// Draw removes the top card from the deck.
func (d *Deck) Draw() (Card, error) {
	if len(d.Cards) == 0 {
		return Card{}, errors.New("deck is empty")
	}
	card := d.Cards[0]
	d.Cards = d.Cards[1:]

	// Update counts for strategy tracking
	if card.Type == CardTypeNumber {
		d.RemainingCounts[card.Value]--
	}

	return card, nil
}

// EstimateHitRisk calculates the probability of busting based on the current hand.
func (d *Deck) EstimateHitRisk(handNumbers map[NumberValue]struct{}) float64 {
	total := len(d.Cards)
	if total == 0 {
		return 0
	}

	riskCards := 0
	for val := range handNumbers {
		riskCards += d.RemainingCounts[val]
	}

	return float64(riskCards) / float64(total)
}
