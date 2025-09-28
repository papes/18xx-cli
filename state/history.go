package state

import "sync"

// History manages a stack of game states for undo/redo functionality
// All methods are thread-safe
type History struct {
	mutex   sync.RWMutex
	states  []GameState
	current int
	maxSize int
}

// NewHistory creates a new history manager with an initial state
func NewHistory(initialState GameState, maxSize int) *History {
	return &History{
		states:  []GameState{initialState},
		current: 0,
		maxSize: maxSize,
	}
}

// Current returns the current game state
func (h *History) Current() GameState {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	return h.states[h.current]
}

// Push adds a new state to history and moves current pointer
// This clears any redo states (states after current)
func (h *History) Push(state GameState) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	// Clear any redo states
	h.states = h.states[:h.current+1]

	// Add new state
	h.states = append(h.states, state)
	h.current++

	// Enforce max size by removing oldest states
	if len(h.states) > h.maxSize {
		removeCount := len(h.states) - h.maxSize
		h.states = h.states[removeCount:]
		h.current -= removeCount
	}
}

// CanUndo returns true if there are states to undo to
func (h *History) CanUndo() bool {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	return h.current > 0
}

// Undo moves back one state in history
// Returns the new current state and true if successful
func (h *History) Undo() (GameState, bool) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	if h.current <= 0 {
		return GameState{}, false
	}

	h.current--
	return h.states[h.current], true
}

// CanRedo returns true if there are states to redo to
func (h *History) CanRedo() bool {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	return h.current < len(h.states)-1
}

// Redo moves forward one state in history
// Returns the new current state and true if successful
func (h *History) Redo() (GameState, bool) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	if h.current >= len(h.states)-1 {
		return GameState{}, false
	}

	h.current++
	return h.states[h.current], true
}

// UndoCount returns the number of states that can be undone
func (h *History) UndoCount() int {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	return h.current
}

// RedoCount returns the number of states that can be redone
func (h *History) RedoCount() int {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	return len(h.states) - 1 - h.current
}
