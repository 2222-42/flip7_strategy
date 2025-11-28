package domain

// TurnChoice represents the decision a player makes on their turn.
type TurnChoice string

const (
	TurnChoiceHit  TurnChoice = "hit"
	TurnChoiceStay TurnChoice = "stay"
)

// Strategy defines the behavior for an AI player.
type Strategy interface {
	Decide(deck *Deck, hand *PlayerHand, playerScore int, otherPlayers []*Player) TurnChoice
	ChooseTarget(action ActionType, candidates []*Player, self *Player, deck *Deck) *Player
	Name() string
}
