package domain

import (
	"fmt"

	"github.com/google/uuid"
)

type CardType string

const (
	CardTypeNumber        CardType = "number"
	CardTypeSpecialNumber CardType = "special_number"
	CardTypeModifier      CardType = "modifier"
	CardTypeAction        CardType = "action"
)

type NumberValue int

type SpecialKind string

const (
	SpecialNone     SpecialKind = ""
	SpecialZero     SpecialKind = "the_zero"
	SpecialUnlucky7 SpecialKind = "unlucky_7"
	SpecialLucky13  SpecialKind = "lucky_13"
)

type ModifierType string

const (
	ModifierMinus2  ModifierType = "minus_2"
	ModifierMinus4  ModifierType = "minus_4"
	ModifierMinus6  ModifierType = "minus_6"
	ModifierMinus8  ModifierType = "minus_8"
	ModifierMinus10 ModifierType = "minus_10"
	ModifierDivide2 ModifierType = "divide_2"
)

func (m ModifierType) Amount() int {
	switch m {
	case ModifierMinus2:
		return -2
	case ModifierMinus4:
		return -4
	case ModifierMinus6:
		return -6
	case ModifierMinus8:
		return -8
	case ModifierMinus10:
		return -10
	default:
		return 0
	}
}

func (m ModifierType) IsDivide() bool {
	return m == ModifierDivide2
}

type ActionType string

const (
	ActionJustOneMore ActionType = "just_one_more"
	ActionFlipFour    ActionType = "flip_four"
	ActionSwap        ActionType = "swap"
	ActionSteal       ActionType = "steal"
	ActionDiscard     ActionType = "discard"
)

const (
	WinningThreshold  = 200
	Flip7Bonus        = 15
	FlipFourCardCount = 4
	StandardDeckSize  = 108
)

type CardID = uuid.UUID

type CardSpec struct {
	Type         CardType     `json:"type"`
	Value        NumberValue  `json:"value,omitempty"`
	SpecialKind  SpecialKind  `json:"special_kind,omitempty"`
	ModifierType ModifierType `json:"modifier_type,omitempty"`
	ActionType   ActionType   `json:"action_type,omitempty"`
}

func (c CardSpec) IsNumberLike() bool {
	return c.Type == CardTypeNumber || c.Type == CardTypeSpecialNumber
}

func (c CardSpec) Rank() (NumberValue, bool) {
	if !c.IsNumberLike() {
		return 0, false
	}
	return c.Value, true
}

func (c CardSpec) CountsTowardFlip7() bool {
	return c.IsNumberLike()
}

func (c CardSpec) String() string {
	switch c.Type {
	case CardTypeNumber:
		return fmt.Sprintf("%d", c.Value)
	case CardTypeSpecialNumber:
		return string(c.SpecialKind)
	case CardTypeModifier:
		return string(c.ModifierType)
	case CardTypeAction:
		return string(c.ActionType)
	default:
		return "unknown"
	}
}

type TableCard struct {
	ID       CardID   `json:"id"`
	Spec     CardSpec `json:"spec"`
	FaceUp   bool     `json:"face_up"`
	Sideways bool     `json:"sideways"`
}

func (c TableCard) String() string {
	return c.Spec.String()
}

func newCard(spec CardSpec) TableCard {
	return TableCard{
		ID:     uuid.New(),
		Spec:   spec,
		FaceUp: true,
	}
}

func NewNumberCard(v NumberValue) TableCard {
	return newCard(CardSpec{Type: CardTypeNumber, Value: v})
}

func NewSpecialCard(kind SpecialKind) TableCard {
	var v NumberValue
	switch kind {
	case SpecialZero:
		v = 0
	case SpecialUnlucky7:
		v = 7
	case SpecialLucky13:
		v = 13
	}
	return newCard(CardSpec{Type: CardTypeSpecialNumber, Value: v, SpecialKind: kind})
}

func NewModifierCard(m ModifierType) TableCard {
	return newCard(CardSpec{Type: CardTypeModifier, ModifierType: m})
}

func NewActionCard(a ActionType) TableCard {
	return newCard(CardSpec{Type: CardTypeAction, ActionType: a})
}

type CardRef struct {
	ID      CardID    `json:"id"`
	OwnerID uuid.UUID `json:"owner_id"`
	Spec    CardSpec  `json:"spec"`
}

type SwapPair struct {
	A CardRef `json:"a"`
	B CardRef `json:"b"`
}
