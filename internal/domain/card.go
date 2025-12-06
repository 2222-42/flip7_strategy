package domain

import (
	"errors"
	"fmt"
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

// IsAdditive returns true if the modifier type adds points (rather than multiplying).
func (m ModifierType) IsAdditive() bool {
	return m == ModifierPlus2 || m == ModifierPlus4 || m == ModifierPlus6 || m == ModifierPlus8 || m == ModifierPlus10
}

// ActionType represents the type of action card.
type ActionType string

const (
	ActionFreeze           ActionType = "freeze"
	ActionFlipThree        ActionType = "flip_three"
	ActionSecondChance     ActionType = "second_chance"
	ActionGiveSecondChance ActionType = "give_second_chance" // Pseudo-action for target selection
)

// FlipThreeCardCount is the number of cards drawn during Flip Three action.
const FlipThreeCardCount = 3

// Card represents a single card in the game.
type Card struct {
	Type         CardType     `json:"type"`
	Value        NumberValue  `json:"value,omitempty"`         // For Number cards
	ModifierType ModifierType `json:"modifier_type,omitempty"` // For Modifier cards
	ActionType   ActionType   `json:"action_type,omitempty"`   // For Action cards
}

func (c Card) String() string {
	switch c.Type {
	case CardTypeNumber:
		return fmt.Sprintf("%d", c.Value)
	case CardTypeModifier:
		return string(c.ModifierType)
	case CardTypeAction:
		return string(c.ActionType)
	default:
		return "unknown"
	}
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
			count = 1 // Card 0 has 1 copy as per game rules ("0:x1").
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
// Only number cards can cause a bust, so we only count number cards in the total.
func (d *Deck) EstimateHitRisk(handNumbers map[NumberValue]struct{}) float64 {
	// Count total number cards in deck
	totalNumberCards := 0
	for _, count := range d.RemainingCounts {
		totalNumberCards += count
	}

	if totalNumberCards == 0 {
		return 0
	}

	// Count risky number cards (those matching hand)
	riskCards := 0
	for val := range handNumbers {
		riskCards += d.RemainingCounts[val]
	}

	return float64(riskCards) / float64(totalNumberCards)
}

// NewDeckFromCards creates a new deck from a list of cards (e.g., discard pile).
func NewDeckFromCards(cards []Card) *Deck {
	counts := make(map[NumberValue]int)
	for _, c := range cards {
		if c.Type == CardTypeNumber {
			counts[c.Value]++
		}
	}

	d := &Deck{
		Cards:           cards,
		RemainingCounts: counts,
	}
	d.Shuffle()
	return d
}

// EstimateFlipThreeRisk calculates the probability of busting when drawing 3 cards.
// It uses a Monte Carlo simulation.
func (d *Deck) EstimateFlipThreeRisk(handNumbers map[NumberValue]struct{}, hasSecondChance bool) float64 {
	if len(d.Cards) == 0 {
		return 0
	}

	trials := 1000
	busts := 0

	for i := 0; i < trials; i++ {
		// Clone deck state (simplified: just shuffle indices)
		// Actually, we need to simulate drawing 3 cards from the current remaining cards.
		// Since we don't want to modify the actual deck, we can just pick 3 random indices
		// from the remaining cards. Note: drawing changes probabilities for subsequent draws.
		// So we should shuffle a temporary slice of indices or just pick 3 unique indices.

		// Optimization: If deck has < 3 cards, we just draw all of them.
		// But for simulation, let's assume we can always draw (or deck reshuffles, but we only know current deck).
		// If deck < 3, probability is deterministic based on those cards.
		// Let's just simulate drawing up to 3 cards from the current deck.

		deckSize := len(d.Cards)
		drawCount := 3
		if deckSize < 3 {
			drawCount = deckSize
		}

		// Create a permutation of indices to simulate a shuffle
		perm := Perm(deckSize)

		// Simulation state
		currentHand := make(map[NumberValue]struct{})
		for k := range handNumbers {
			currentHand[k] = struct{}{}
		}
		simHasSecondChance := hasSecondChance
		busted := false

		for j := 0; j < drawCount; j++ {
			cardIdx := perm[j]
			card := d.Cards[cardIdx]

			if card.Type == CardTypeNumber {
				if _, exists := currentHand[card.Value]; exists {
					if simHasSecondChance {
						simHasSecondChance = false
						// Discard the duplicate and the second chance.
						// The duplicate is NOT added to hand.
					} else {
						busted = true
						break
					}
				} else {
					currentHand[card.Value] = struct{}{}
				}
			} else if card.Type == CardTypeAction && card.ActionType == ActionSecondChance {
				simHasSecondChance = true
			}
			// Modifiers and other actions don't cause bust directly (FlipThree/Freeze are queued, not resolved in this risk calc)
		}

		if busted {
			busts++
		}
	}

	return float64(busts) / float64(trials)
}
