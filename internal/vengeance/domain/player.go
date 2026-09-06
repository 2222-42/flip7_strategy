package domain

import "github.com/google/uuid"

type HandStatus string

const (
	HandStatusActive HandStatus = "active"
	HandStatusStayed HandStatus = "stayed"
	HandStatusBusted HandStatus = "busted"
)

type ReceiveResult struct {
	Busted bool
	Flip7  bool
	Wiped  []TableCard
}

type PlayerHand struct {
	ID           uuid.UUID   `json:"id"`
	NumberLine   []TableCard `json:"number_line"`
	ModifierLine []TableCard `json:"modifier_line"`
	Status       HandStatus  `json:"status"`
}

func NewPlayerHand() *PlayerHand {
	return &PlayerHand{
		ID:     uuid.New(),
		Status: HandStatusActive,
	}
}

func (h *PlayerHand) RankCounts() map[NumberValue]int {
	counts := make(map[NumberValue]int)
	for _, c := range h.NumberLine {
		if rank, ok := c.Spec.Rank(); ok {
			counts[rank]++
		}
	}
	return counts
}

func (h *PlayerHand) NumberSum() int {
	sum := 0
	for _, c := range h.NumberLine {
		if rank, ok := c.Spec.Rank(); ok {
			sum += int(rank)
		}
	}
	return sum
}

func (h *PlayerHand) Clone() *PlayerHand {
	cp := &PlayerHand{
		ID:           h.ID,
		Status:       h.Status,
		NumberLine:   append([]TableCard(nil), h.NumberLine...),
		ModifierLine: append([]TableCard(nil), h.ModifierLine...),
	}
	return cp
}

func (h *PlayerHand) DistinctRanks() int {
	seen := make(map[NumberValue]struct{})
	for _, c := range h.NumberLine {
		if rank, ok := c.Spec.Rank(); ok {
			seen[rank] = struct{}{}
		}
	}
	return len(seen)
}

func (h *PlayerHand) HasFlip7() bool {
	return h.DistinctRanks() >= 7
}

func (h *PlayerHand) HasLucky13() bool {
	for _, c := range h.NumberLine {
		if c.Spec.SpecialKind == SpecialLucky13 {
			return true
		}
	}
	return false
}

func (h *PlayerHand) HasZero() bool {
	for _, c := range h.NumberLine {
		if c.Spec.SpecialKind == SpecialZero {
			return true
		}
	}
	return false
}

func (h *PlayerHand) MustHit() bool {
	return h.Status == HandStatusActive && h.HasZero()
}

func (h *PlayerHand) CanStay() bool {
	return h.Status == HandStatusActive && !h.HasZero()
}

func (h *PlayerHand) FaceUpCards() []TableCard {
	if h.Status == HandStatusBusted {
		return nil
	}
	out := make([]TableCard, 0, len(h.NumberLine)+len(h.ModifierLine))
	for _, c := range h.NumberLine {
		if c.FaceUp {
			out = append(out, c)
		}
	}
	for _, c := range h.ModifierLine {
		if c.FaceUp {
			out = append(out, c)
		}
	}
	return out
}

func (h *PlayerHand) Stay() {
	if h.Status != HandStatusActive {
		return
	}
	h.Status = HandStatusStayed
	if len(h.NumberLine) > 0 {
		h.NumberLine[0].Sideways = true
	}
}

func (h *PlayerHand) bust() {
	h.Status = HandStatusBusted
	for i := range h.NumberLine {
		h.NumberLine[i].FaceUp = false
		h.NumberLine[i].Sideways = false
	}
	for i := range h.ModifierLine {
		h.ModifierLine[i].FaceUp = false
		h.ModifierLine[i].Sideways = false
	}
}

func (h *PlayerHand) ReceiveNumberLike(card TableCard) ReceiveResult {
	card.FaceUp = true
	card.Sideways = false

	if card.Spec.SpecialKind == SpecialUnlucky7 {
		wiped := make([]TableCard, 0, len(h.NumberLine)+len(h.ModifierLine))
		wiped = append(wiped, h.NumberLine...)
		wiped = append(wiped, h.ModifierLine...)
		h.NumberLine = []TableCard{card}
		h.ModifierLine = nil
		return ReceiveResult{Wiped: wiped, Flip7: h.HasFlip7()}
	}

	rank, _ := card.Spec.Rank()
	counts := h.RankCounts()
	have := counts[rank]
	maxAllowed := 1
	if rank == 13 && (h.HasLucky13() || card.Spec.SpecialKind == SpecialLucky13) {
		maxAllowed = 2
	}
	if have >= maxAllowed {
		h.NumberLine = append(h.NumberLine, card)
		h.bust()
		return ReceiveResult{Busted: true}
	}

	h.NumberLine = append(h.NumberLine, card)
	return ReceiveResult{Flip7: h.HasFlip7()}
}

func (h *PlayerHand) ReceiveModifier(card TableCard) {
	card.FaceUp = true
	card.Sideways = false
	h.ModifierLine = append(h.ModifierLine, card)
}

func (h *PlayerHand) RemoveCard(id CardID) (TableCard, bool) {
	for i, c := range h.NumberLine {
		if c.ID == id {
			h.NumberLine = append(h.NumberLine[:i], h.NumberLine[i+1:]...)
			return c, true
		}
	}
	for i, c := range h.ModifierLine {
		if c.ID == id {
			h.ModifierLine = append(h.ModifierLine[:i], h.ModifierLine[i+1:]...)
			return c, true
		}
	}
	return TableCard{}, false
}

type Player struct {
	ID          uuid.UUID   `json:"id"`
	Name        string      `json:"name"`
	TotalScore  int         `json:"total_score"`
	CurrentHand *PlayerHand `json:"current_hand"`
	Strategy    Strategy    `json:"-"`
}

func NewPlayer(name string, strategy Strategy) *Player {
	return &Player{
		ID:       uuid.New(),
		Name:     name,
		Strategy: strategy,
	}
}

func (p *Player) StartNewRound() {
	p.CurrentHand = NewPlayerHand()
}

func (p *Player) BankScore(score int) {
	p.TotalScore += score
}

func (p *Player) BankCurrentHand() int {
	score := NewScoreCalculator().Compute(p.CurrentHand)
	p.BankScore(score.Total)
	return score.Total
}
