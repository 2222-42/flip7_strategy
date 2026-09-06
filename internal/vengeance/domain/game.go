package domain

import "github.com/google/uuid"

type RoundEndReason string

const (
	RoundEndReasonNoActivePlayers RoundEndReason = "no_active_players"
	RoundEndReasonFlip7           RoundEndReason = "flip7_achieved"
	RoundEndReasonAborted         RoundEndReason = "aborted"
)

type TurnChoice string

const (
	TurnChoiceHit  TurnChoice = "hit"
	TurnChoiceStay TurnChoice = "stay"
)

type Round struct {
	ID         uuid.UUID      `json:"id"`
	Dealer     *Player        `json:"dealer"`
	Players    []*Player      `json:"players"`
	OfferOrder []*Player      `json:"-"`
	Deck       *Deck          `json:"deck"`
	IsEnded    bool           `json:"is_ended"`
	EndReason  RoundEndReason `json:"end_reason"`
}

func NewRound(players []*Player, dealer *Player, deck *Deck) *Round {
	dealerIdx := 0
	for i, p := range players {
		if p.ID == dealer.ID {
			dealerIdx = i
			break
		}
	}

	n := len(players)
	order := make([]*Player, 0, n)
	for i := 1; i <= n; i++ {
		p := players[(dealerIdx+i)%n]
		p.StartNewRound()
		order = append(order, p)
	}

	return &Round{
		ID:         uuid.New(),
		Dealer:     dealer,
		Players:    players,
		OfferOrder: order,
		Deck:       deck,
	}
}

// RebuildOfferOrder restores clockwise order from the dealer without resetting hands.
func (r *Round) RebuildOfferOrder() {
	if r == nil || r.Dealer == nil || len(r.Players) == 0 {
		return
	}
	dealerIdx := 0
	for i, p := range r.Players {
		if p.ID == r.Dealer.ID {
			dealerIdx = i
			break
		}
	}
	n := len(r.Players)
	order := make([]*Player, 0, n)
	for i := 1; i <= n; i++ {
		order = append(order, r.Players[(dealerIdx+i)%n])
	}
	r.OfferOrder = order
}

func (r *Round) End(reason RoundEndReason) {
	r.IsEnded = true
	r.EndReason = reason
}

func (r *Round) ActivePlayers() []*Player {
	var active []*Player
	for _, p := range r.OfferOrder {
		if p.CurrentHand != nil && p.CurrentHand.Status == HandStatusActive {
			active = append(active, p)
		}
	}
	return active
}

func (r *Round) NonBustedPlayers() []*Player {
	var out []*Player
	for _, p := range r.Players {
		if p.CurrentHand != nil && p.CurrentHand.Status != HandStatusBusted {
			out = append(out, p)
		}
	}
	return out
}

// GameRules is the Brutal Mode overlay. Zero value is standard Vengeance.
type GameRules struct {
	ScoreCanGoNegative    bool `json:"score_can_go_negative"`
	ModifiersTargetBusted bool `json:"modifiers_target_busted"`
	Flip7AsAttack         bool `json:"flip7_as_attack"`
}

func StandardRules() GameRules { return GameRules{} }

func BrutalRules() GameRules {
	return GameRules{
		ScoreCanGoNegative:    true,
		ModifiersTargetBusted: true,
		Flip7AsAttack:         true,
	}
}

type Game struct {
	ID           uuid.UUID   `json:"id"`
	Players      []*Player   `json:"players"`
	CurrentRound *Round      `json:"current_round"`
	DealerIndex  int         `json:"dealer_index"`
	IsCompleted  bool        `json:"is_completed"`
	Winners      []*Player   `json:"winners"`
	DiscardPile  []TableCard `json:"discard_pile"`
	RoundCount   int         `json:"round_count"`
	Deck         *Deck       `json:"deck"`
	Rules        GameRules   `json:"rules"`
}

func NewGame(players []*Player) *Game {
	return &Game{
		ID:      uuid.New(),
		Players: players,
	}
}

func (g *Game) DetermineWinners() []*Player {
	var candidates []*Player
	highest := 0
	for _, p := range g.Players {
		if p.TotalScore >= WinningThreshold {
			if p.TotalScore > highest {
				highest = p.TotalScore
				candidates = []*Player{p}
			} else if p.TotalScore == highest {
				candidates = append(candidates, p)
			}
		}
	}
	return candidates
}

func (g *Game) PlayerByID(id uuid.UUID) *Player {
	for _, p := range g.Players {
		if p.ID == id {
			return p
		}
	}
	return nil
}

func (g *Game) FaceUpRefs() []CardRef {
	if g.CurrentRound == nil {
		return nil
	}
	var refs []CardRef
	for _, p := range g.Players {
		if p.CurrentHand == nil || p.CurrentHand.Status == HandStatusBusted {
			continue
		}
		for _, c := range p.CurrentHand.FaceUpCards() {
			refs = append(refs, CardRef{ID: c.ID, OwnerID: p.ID, Spec: c.Spec})
		}
	}
	return refs
}
