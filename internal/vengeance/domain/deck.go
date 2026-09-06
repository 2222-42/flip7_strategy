package domain

import (
	"errors"
	"math/rand"
	"time"
)

type RemainingCounts struct {
	ByNumber   map[NumberValue]int  `json:"by_number"`
	BySpecial  map[SpecialKind]int  `json:"by_special"`
	ByModifier map[ModifierType]int `json:"by_modifier"`
	ByAction   map[ActionType]int   `json:"by_action"`
}

func newRemainingCounts() RemainingCounts {
	return RemainingCounts{
		ByNumber:   make(map[NumberValue]int),
		BySpecial:  make(map[SpecialKind]int),
		ByModifier: make(map[ModifierType]int),
		ByAction:   make(map[ActionType]int),
	}
}

func countsFromCards(cards []TableCard) RemainingCounts {
	c := newRemainingCounts()
	for _, card := range cards {
		c.add(card.Spec)
	}
	return c
}

func (c RemainingCounts) add(spec CardSpec) {
	switch spec.Type {
	case CardTypeNumber:
		c.ByNumber[spec.Value]++
	case CardTypeSpecialNumber:
		c.BySpecial[spec.SpecialKind]++
	case CardTypeModifier:
		c.ByModifier[spec.ModifierType]++
	case CardTypeAction:
		c.ByAction[spec.ActionType]++
	}
}

func (c RemainingCounts) remove(spec CardSpec) {
	switch spec.Type {
	case CardTypeNumber:
		c.ByNumber[spec.Value]--
	case CardTypeSpecialNumber:
		c.BySpecial[spec.SpecialKind]--
	case CardTypeModifier:
		c.ByModifier[spec.ModifierType]--
	case CardTypeAction:
		c.ByAction[spec.ActionType]--
	}
}

type Deck struct {
	Cards     []TableCard     `json:"cards"`
	Remaining RemainingCounts `json:"remaining"`
}

func buildStandardCards() []TableCard {
	cards := make([]TableCard, 0, StandardDeckSize)

	regularCounts := map[NumberValue]int{
		1: 1, 2: 2, 3: 3, 4: 4, 5: 5, 6: 6, 7: 6,
		8: 8, 9: 9, 10: 10, 11: 11, 12: 12, 13: 12,
	}
	for v := NumberValue(1); v <= 13; v++ {
		for i := 0; i < regularCounts[v]; i++ {
			cards = append(cards, NewNumberCard(v))
		}
	}

	cards = append(cards,
		NewSpecialCard(SpecialZero),
		NewSpecialCard(SpecialUnlucky7),
		NewSpecialCard(SpecialLucky13),
	)

	for _, m := range []ModifierType{
		ModifierMinus2, ModifierMinus4, ModifierMinus6,
		ModifierMinus8, ModifierMinus10, ModifierDivide2,
	} {
		cards = append(cards, NewModifierCard(m))
	}

	for _, a := range []ActionType{
		ActionJustOneMore, ActionFlipFour, ActionSwap, ActionSteal, ActionDiscard,
	} {
		cards = append(cards, NewActionCard(a), NewActionCard(a))
	}

	return cards
}

func NewDeck() *Deck {
	d := NewUnshuffledDeck()
	d.Shuffle()
	return d
}

func NewUnshuffledDeck() *Deck {
	cards := buildStandardCards()
	return &Deck{Cards: cards, Remaining: countsFromCards(cards)}
}

func NewDeckFromCards(cards []TableCard) *Deck {
	d := DeckWithCards(cards)
	d.Shuffle()
	return d
}

// DeckWithCards builds a deck in the given order (no shuffle). For tests.
func DeckWithCards(cards []TableCard) *Deck {
	cp := make([]TableCard, len(cards))
	copy(cp, cards)
	return &Deck{Cards: cp, Remaining: countsFromCards(cp)}
}

func (d *Deck) Shuffle() {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	r.Shuffle(len(d.Cards), func(i, j int) {
		d.Cards[i], d.Cards[j] = d.Cards[j], d.Cards[i]
	})
}

func (d *Deck) Draw() (TableCard, error) {
	if len(d.Cards) == 0 {
		return TableCard{}, errors.New("deck is empty")
	}
	card := d.Cards[0]
	d.Cards = d.Cards[1:]
	d.Remaining.remove(card.Spec)
	return card, nil
}

func (d *Deck) EstimateHitRisk(hand *PlayerHand) float64 {
	total := len(d.Cards)
	if total == 0 || hand == nil {
		return 0
	}
	risk := 0
	for rank, n := range hand.RankCounts() {
		maxAllowed := 1
		if rank == 13 && hand.HasLucky13() && n < 2 {
			maxAllowed = 2
		}
		if n >= maxAllowed {
			risk += d.Remaining.ByNumber[rank]
		}
	}
	return float64(risk) / float64(total)
}

const flipFourRiskTrials = 500

func (d *Deck) EstimateFlipFourRisk(hand *PlayerHand) float64 {
	if hand == nil || len(d.Cards) == 0 {
		return 0
	}
	drawN := FlipFourCardCount
	if drawN > len(d.Cards) {
		drawN = len(d.Cards)
	}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	busts := 0
	for trial := 0; trial < flipFourRiskTrials; trial++ {
		order := r.Perm(len(d.Cards))
		cl := hand.Clone()
		busted := false
		for i := 0; i < drawN; i++ {
			card := d.Cards[order[i]]
			if !card.Spec.IsNumberLike() {
				continue
			}
			res := cl.ReceiveNumberLike(card)
			if res.Busted {
				busted = true
				break
			}
			if res.Flip7 {
				break
			}
		}
		if busted {
			busts++
		}
	}
	return float64(busts) / float64(flipFourRiskTrials)
}
