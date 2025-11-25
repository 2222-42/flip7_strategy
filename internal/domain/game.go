package domain

import (
	"github.com/google/uuid"
)

// RoundEndReason explains why a round ended.
type RoundEndReason string

const (
	RoundEndReasonNoActivePlayers RoundEndReason = "no_active_players"
	RoundEndReasonFlip7           RoundEndReason = "flip7_achieved"
	RoundEndReasonAborted         RoundEndReason = "aborted"
)

// Round represents a single round of play.
type Round struct {
	ID            uuid.UUID      `json:"id"`
	Dealer        *Player        `json:"dealer"`
	Players       []*Player      `json:"players"`
	Deck          *Deck          `json:"deck"`
	ActivePlayers []*Player      `json:"active_players"`
	IsEnded       bool           `json:"is_ended"`
	EndReason     RoundEndReason `json:"end_reason"`
}

// NewRound creates a new round.
func NewRound(players []*Player, dealer *Player, deck *Deck) *Round {
	// Find dealer index
	dealerIdx := -1
	for i, p := range players {
		if p.ID == dealer.ID {
			dealerIdx = i
			break
		}
	}

	// Reorder players starting from dealer
	active := []*Player{}
	if dealerIdx != -1 {
		for i := 0; i < len(players); i++ {
			idx := (dealerIdx + i) % len(players)
			p := players[idx]
			p.StartNewRound()
			active = append(active, p)
		}
	} else {
		// Fallback if dealer not found (shouldn't happen)
		for _, p := range players {
			p.StartNewRound()
			active = append(active, p)
		}
	}

	return &Round{
		ID:            uuid.New(),
		Dealer:        dealer,
		Players:       players,
		Deck:          deck,
		ActivePlayers: active,
	}
}

// Game represents the entire game session.
type Game struct {
	ID           uuid.UUID `json:"id"`
	Players      []*Player `json:"players"`
	CurrentRound *Round    `json:"current_round"`
	DealerIndex  int       `json:"dealer_index"`
	IsCompleted  bool      `json:"is_completed"`
	Winner       *Player   `json:"winner"`
	DiscardPile  []Card    `json:"discard_pile"`
}

// NewGame creates a new game.
func NewGame(players []*Player) *Game {
	return &Game{
		ID:      uuid.New(),
		Players: players,
	}
}
