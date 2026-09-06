package application

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	"flip7_strategy/internal/vengeance/domain"
)

const (
	phaseDeal  = "deal"
	phaseTurns = "turns"
)

type gameMemento string

type gameHistory struct {
	mementos     []gameMemento
	currentIndex int
}

func (h *gameHistory) Push(m gameMemento) {
	if h.currentIndex < len(h.mementos)-1 {
		h.mementos = h.mementos[:h.currentIndex+1]
	}
	h.mementos = append(h.mementos, m)
	h.currentIndex = len(h.mementos) - 1
}

func (h *gameHistory) Undo() (gameMemento, bool) {
	if h.currentIndex > 0 {
		h.currentIndex--
		return h.mementos[h.currentIndex], true
	}
	return "", false
}

func (h *gameHistory) Redo() (gameMemento, bool) {
	if h.currentIndex < len(h.mementos)-1 {
		h.currentIndex++
		return h.mementos[h.currentIndex], true
	}
	return "", false
}

func (h *gameHistory) Current() (gameMemento, bool) {
	if len(h.mementos) == 0 || h.currentIndex < 0 || h.currentIndex >= len(h.mementos) {
		return "", false
	}
	return h.mementos[h.currentIndex], true
}

type gameSnapshot struct {
	Game   *domain.Game `json:"game"`
	Phase  string       `json:"phase"`
	Cursor int          `json:"cursor"`
}

func (s *ManualGameService) snapshot() (string, error) {
	if s.Game == nil {
		return "", fmt.Errorf("no game")
	}
	data, err := json.Marshal(gameSnapshot{Game: s.Game, Phase: s.phase, Cursor: s.cursor})
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(data), nil
}

func (s *ManualGameService) restore(encoded string) error {
	raw, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return fmt.Errorf("invalid snapshot: %w", err)
	}
	var snap gameSnapshot
	if err := json.Unmarshal(raw, &snap); err != nil {
		return fmt.Errorf("parse snapshot: %w", err)
	}
	if snap.Game == nil {
		return fmt.Errorf("snapshot missing game")
	}
	s.relink(snap.Game)
	s.Game = snap.Game
	s.phase = snap.Phase
	s.cursor = snap.Cursor
	s.rewound = true
	return nil
}

func (s *ManualGameService) relink(g *domain.Game) {
	byID := make(map[string]*domain.Player, len(g.Players))
	for _, p := range g.Players {
		p.Strategy = s.Advisor
		byID[p.ID.String()] = p
	}
	if g.CurrentRound == nil {
		return
	}
	r := g.CurrentRound
	for i, p := range r.Players {
		if e, ok := byID[p.ID.String()]; ok {
			r.Players[i] = e
		}
	}
	if r.Dealer != nil {
		if e, ok := byID[r.Dealer.ID.String()]; ok {
			r.Dealer = e
		}
	}
	r.RebuildOfferOrder()
	if r.Deck != nil {
		r.Deck.RebuildRemaining()
		g.Deck = r.Deck
	} else if g.Deck != nil {
		g.Deck.RebuildRemaining()
	}
	for i, p := range g.Winners {
		if e, ok := byID[p.ID.String()]; ok {
			g.Winners[i] = e
		}
	}
}

func (s *ManualGameService) PushState() {
	enc, err := s.snapshot()
	if err != nil {
		fmt.Printf("Warning: failed to save undo state: %v\n", err)
		return
	}
	s.History.Push(gameMemento(enc))
}

func (s *ManualGameService) tryUndo() bool {
	m, ok := s.History.Undo()
	if !ok {
		fmt.Println("Cannot undo.")
		return false
	}
	if err := s.restore(string(m)); err != nil {
		fmt.Printf("Undo failed: %v\n", err)
		s.History.Redo()
		s.rewound = false
		return false
	}
	fmt.Println("Undid last action.")
	return true
}

// tryAbortCurrent restores the last committed snapshot without moving the
// history pointer. Nested prompts (JOM, Flip Four, targets) use this so UNDO
// cancels the in-progress action instead of also undoing the previous turn.
func (s *ManualGameService) tryAbortCurrent() bool {
	m, ok := s.History.Current()
	if !ok {
		fmt.Println("Cannot undo.")
		return false
	}
	if err := s.restore(string(m)); err != nil {
		fmt.Printf("Undo failed: %v\n", err)
		s.rewound = false
		return false
	}
	fmt.Println("Cancelled the current action.")
	return true
}

func (s *ManualGameService) tryRedo() bool {
	m, ok := s.History.Redo()
	if !ok {
		fmt.Println("Cannot redo.")
		return false
	}
	if err := s.restore(string(m)); err != nil {
		fmt.Printf("Redo failed: %v\n", err)
		s.History.Undo()
		s.rewound = false
		return false
	}
	fmt.Println("Redid action.")
	return true
}
