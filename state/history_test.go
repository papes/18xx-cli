package state

import (
	"testing"
)

func TestNewHistory(t *testing.T) {
	initialState := NewGameState()
	history := NewHistory(initialState, 50)

	current := history.Current()
	if len(current.Players) != 0 || len(current.Companies) != 0 {
		t.Error("History should return initial state as current")
	}

	if history.CanUndo() {
		t.Error("Should not be able to undo with only initial state")
	}

	if history.CanRedo() {
		t.Error("Should not be able to redo initially")
	}
}

func TestHistoryPush(t *testing.T) {
	initialState := NewGameState()
	history := NewHistory(initialState, 50)

	// Push a new state
	state2 := initialState.AddPlayer("alice", "Alice", 600)
	history.Push(state2)

	current := history.Current()
	if len(current.Players) != 1 {
		t.Error("Current should return the pushed state with 1 player")
	}

	if !history.CanUndo() {
		t.Error("Should be able to undo after pushing")
	}

	if history.CanRedo() {
		t.Error("Should not be able to redo after pushing new state")
	}
}

func TestHistoryUndo(t *testing.T) {
	initialState := NewGameState()
	history := NewHistory(initialState, 50)

	// Push some states
	state2 := initialState.AddPlayer("alice", "Alice", 600)
	state3 := state2.AddPlayer("bob", "Bob", 600)

	history.Push(state2)
	history.Push(state3)

	// Test undo
	undoState, ok := history.Undo()
	if !ok {
		t.Fatal("Undo should succeed")
	}

	if len(undoState.Players) != 1 {
		t.Errorf("Undo state should have 1 player, got %d", len(undoState.Players))
	}

	current := history.Current()
	if len(current.Players) != 1 {
		t.Error("Current should be state2 after undo (1 player)")
	}

	if !history.CanUndo() {
		t.Error("Should still be able to undo to initial state")
	}

	if !history.CanRedo() {
		t.Error("Should be able to redo after undo")
	}

	// Undo to initial state
	undoState, ok = history.Undo()
	if !ok {
		t.Fatal("Second undo should succeed")
	}

	if len(undoState.Players) != 0 {
		t.Errorf("Undo state should have 0 players, got %d", len(undoState.Players))
	}

	current = history.Current()
	if len(current.Players) != 0 {
		t.Error("Current should be initial state after second undo (0 players)")
	}

	if history.CanUndo() {
		t.Error("Should not be able to undo past initial state")
	}

	// Test undo when none available
	_, ok = history.Undo()
	if ok {
		t.Error("Undo should fail when at initial state")
	}
}

func TestHistoryRedo(t *testing.T) {
	initialState := NewGameState()
	history := NewHistory(initialState, 50)

	// Push some states
	state2 := initialState.AddPlayer("alice", "Alice", 600)
	state3 := state2.AddPlayer("bob", "Bob", 600)

	history.Push(state2)
	history.Push(state3)

	// Undo twice
	history.Undo()
	history.Undo()

	// Test redo
	redoState, ok := history.Redo()
	if !ok {
		t.Fatal("Redo should succeed")
	}

	if len(redoState.Players) != 1 {
		t.Errorf("Redo state should have 1 player, got %d", len(redoState.Players))
	}

	current := history.Current()
	if len(current.Players) != 1 {
		t.Error("Current should be state2 after redo (1 player)")
	}

	// Second redo
	redoState, ok = history.Redo()
	if !ok {
		t.Fatal("Second redo should succeed")
	}

	if len(redoState.Players) != 2 {
		t.Errorf("Redo state should have 2 players, got %d", len(redoState.Players))
	}

	current = history.Current()
	if len(current.Players) != 2 {
		t.Error("Current should be state3 after second redo (2 players)")
	}

	if history.CanRedo() {
		t.Error("Should not be able to redo past latest state")
	}

	// Test redo when none available
	_, ok = history.Redo()
	if ok {
		t.Error("Redo should fail when at latest state")
	}
}

func TestHistoryPushClearsRedoStack(t *testing.T) {
	initialState := NewGameState()
	history := NewHistory(initialState, 50)

	// Push states and undo
	state2 := initialState.AddPlayer("alice", "Alice", 600)
	state3 := state2.AddPlayer("bob", "Bob", 600)

	history.Push(state2)
	history.Push(state3)
	history.Undo() // Now we can redo

	if !history.CanRedo() {
		t.Error("Should be able to redo before push")
	}

	// Push new state - should clear redo stack
	state4 := state2.AddCompany("BO", "Baltimore & Ohio", 10)
	history.Push(state4)

	if history.CanRedo() {
		t.Error("Should not be able to redo after pushing new state")
	}

	if len(history.Current().Companies) != 1 {
		t.Error("Current state should have the company from state4")
	}
}

func TestHistoryMaxSize(t *testing.T) {
	initialState := NewGameState()
	history := NewHistory(initialState, 3) // Small max size

	// Push more states than max size
	state1 := initialState.AddPlayer("alice", "Alice", 600)
	state2 := state1.AddPlayer("bob", "Bob", 600)
	state3 := state2.AddPlayer("charlie", "Charlie", 600)
	state4 := state3.AddPlayer("dave", "Dave", 600)

	history.Push(state1)
	history.Push(state2)
	history.Push(state3)
	history.Push(state4)

	// Should only be able to undo 2 times with max size 3 (states: [state2, state3, state4])
	undoCount := 0
	for history.CanUndo() {
		history.Undo()
		undoCount++
		if undoCount > 5 { // Safety check
			t.Fatal("Infinite undo loop")
		}
	}

	if undoCount != 2 {
		t.Errorf("Expected to undo 2 times with max size 3, got %d", undoCount)
	}

	// Should be at state2 (oldest remaining state after size limiting)
	if len(history.Current().Players) != 2 {
		t.Errorf("Expected 2 players after undoing to limit, got %d", len(history.Current().Players))
	}
}
