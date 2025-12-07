package application

import (
	"testing"
)

func TestGameHistory_Push(t *testing.T) {
	h := &GameHistory{
		currentIndex: -1, // Simulate initialization if not using constructor
	}

	// Case 1: Push to empty history
	h.Push("state1")
	if len(h.mementos) != 1 {
		t.Errorf("Expected length 1, got %d", len(h.mementos))
	}
	if h.currentIndex != 0 {
		t.Errorf("Expected index 0, got %d", h.currentIndex)
	}
	if h.mementos[0] != "state1" {
		t.Errorf("Expected state1, got %s", h.mementos[0])
	}

	// Case 2: Push new state
	h.Push("state2")
	if len(h.mementos) != 2 {
		t.Errorf("Expected length 2, got %d", len(h.mementos))
	}
	if h.currentIndex != 1 {
		t.Errorf("Expected index 1, got %d", h.currentIndex)
	}

	// Case 3: Undo then Push (Truncation)
	h.Undo() // Index becomes 0 ("state1")
	h.Push("state3")
	if len(h.mementos) != 2 {
		t.Errorf("Expected length 2 (truncated), got %d", len(h.mementos))
	}
	if h.currentIndex != 1 {
		t.Errorf("Expected index 1, got %d", h.currentIndex)
	}
	if h.mementos[1] != "state3" {
		t.Errorf("Expected state3, got %s", h.mementos[1])
	}
}

func TestGameHistory_Undo(t *testing.T) {
	h := &GameHistory{
		mementos:     []GameMemento{"state1", "state2"},
		currentIndex: 1,
	}

	// Case 1: Successful Undo
	m, ok := h.Undo()
	if !ok {
		t.Error("Expected undo to succeed")
	}
	if m != "state1" {
		t.Errorf("Expected state1, got %s", m)
	}
	if h.currentIndex != 0 {
		t.Errorf("Expected index 0, got %d", h.currentIndex)
	}

	// Case 2: Undo at start of history
	_, ok = h.Undo()
	if ok {
		t.Error("Expected undo to fail at start of history")
	}
	if h.currentIndex != 0 {
		t.Errorf("Expected index to remain 0, got %d", h.currentIndex)
	}
}

func TestGameHistory_Redo(t *testing.T) {
	h := &GameHistory{
		mementos:     []GameMemento{"state1", "state2"},
		currentIndex: 0,
	}

	// Case 1: Successful Redo
	m, ok := h.Redo()
	if !ok {
		t.Error("Expected redo to succeed")
	}
	if m != "state2" {
		t.Errorf("Expected state2, got %s", m)
	}
	if h.currentIndex != 1 {
		t.Errorf("Expected index 1, got %d", h.currentIndex)
	}

	// Case 2: Redo at end of history
	_, ok = h.Redo()
	if ok {
		t.Error("Expected redo to fail at end of history")
	}
	if h.currentIndex != 1 {
		t.Errorf("Expected index to remain 1, got %d", h.currentIndex)
	}
}
