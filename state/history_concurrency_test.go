package state

import (
	"sync"
	"testing"
	"time"
)

// TestHistoryConcurrency tests that History operations are thread-safe
func TestHistoryConcurrency(t *testing.T) {
	// Create initial state and history
	initialState := NewGameState()
	history := NewHistory(initialState, 50)

	// Number of concurrent operations
	numGoroutines := 10
	numOperationsPerGoroutine := 20

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	// Start multiple goroutines performing concurrent operations
	for i := 0; i < numGoroutines; i++ {
		go func(goroutineID int) {
			defer wg.Done()

			for j := 0; j < numOperationsPerGoroutine; j++ {
				// Create a unique state for each operation
				state := NewGameState()
				state.BankMoney = Money(goroutineID*1000 + j)

				// Push state
				history.Push(state)

				// Occasionally try undo/redo
				if j%3 == 0 {
					history.Undo()
				}
				if j%5 == 0 {
					history.Redo()
				}

				// Check current state (read operation)
				_ = history.Current()

				// Check counts (read operations)
				_ = history.UndoCount()
				_ = history.RedoCount()
				_ = history.CanUndo()
				_ = history.CanRedo()

				// Small delay to increase chance of race conditions
				time.Sleep(time.Microsecond)
			}
		}(i)
	}

	// Wait for all goroutines to complete
	wg.Wait()

	// Verify history is still functional after concurrent operations
	currentState := history.Current()
	if currentState.Players == nil {
		t.Error("History state corrupted after concurrent operations")
	}

	// Test basic operations still work
	testState := NewGameState()
	testState.BankMoney = 99999
	history.Push(testState)

	retrieved := history.Current()
	if retrieved.BankMoney != 99999 {
		t.Errorf("Expected bank money 99999, got %d", retrieved.BankMoney)
	}
}

// TestHistoryReadWriteConcurrency tests concurrent reads and writes
func TestHistoryReadWriteConcurrency(t *testing.T) {
	initialState := NewGameState()
	history := NewHistory(initialState, 100)

	var wg sync.WaitGroup
	numReaders := 5
	numWriters := 2
	duration := 100 * time.Millisecond

	// Start reader goroutines
	for i := 0; i < numReaders; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			start := time.Now()
			for time.Since(start) < duration {
				_ = history.Current()
				_ = history.CanUndo()
				_ = history.CanRedo()
				_ = history.UndoCount()
				_ = history.RedoCount()
				time.Sleep(time.Microsecond)
			}
		}()
	}

	// Start writer goroutines
	for i := 0; i < numWriters; i++ {
		wg.Add(1)
		go func(writerID int) {
			defer wg.Done()
			start := time.Now()
			counter := 0
			for time.Since(start) < duration {
				state := NewGameState()
				state.BankMoney = Money(writerID*10000 + counter)
				history.Push(state)

				if counter%3 == 0 {
					history.Undo()
				}
				if counter%4 == 0 {
					history.Redo()
				}

				counter++
				time.Sleep(time.Microsecond * 10)
			}
		}(i)
	}

	wg.Wait()

	// Verify history is still consistent
	currentState := history.Current()
	if currentState.Players == nil {
		t.Error("History corrupted after concurrent read/write operations")
	}
}