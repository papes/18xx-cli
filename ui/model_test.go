package ui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/papes/18xxCli/state"
)

func TestNewModel(t *testing.T) {
	model := NewModel()

	// Check initial state
	if model == nil {
		t.Fatal("NewModel returned nil")
	}

	if model.gameState.Players == nil {
		t.Error("Game state players map not initialized")
	}

	if model.gameState.Companies == nil {
		t.Error("Game state companies map not initialized")
	}

	if model.history == nil {
		t.Error("History not initialized")
	}

	if model.commandExecutor == nil {
		t.Error("Command executor not initialized")
	}

	if model.input != "" {
		t.Error("Initial input should be empty")
	}

	if model.message == "" {
		t.Error("Initial message should not be empty")
	}

	if model.err != nil {
		t.Error("Initial error should be nil")
	}
}

func TestModelInit(t *testing.T) {
	model := NewModel()
	cmd := model.Init()

	if cmd != nil {
		t.Error("Init should return nil command")
	}
}

func TestExecuteCommand(t *testing.T) {
	tests := []struct {
		name            string
		command         string
		expectError     bool
		expectedMessage string
	}{
		{
			name:            "empty command",
			command:         "",
			expectError:     false,
			expectedMessage: "Welcome to 1830 Banking Tool! Type 'help' for commands.",
		},
		{
			name:        "help command",
			command:     "help",
			expectError: false,
		},
		{
			name:        "invalid command",
			command:     "invalid_command_that_does_not_exist",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := NewModel()
			newModel := model.executeCommand(tt.command)

			if tt.expectError && newModel.err == nil {
				t.Error("Expected error but got none")
			}

			if !tt.expectError && newModel.err != nil {
				t.Errorf("Expected no error but got: %v", newModel.err)
			}

			if tt.expectedMessage != "" && newModel.message != tt.expectedMessage {
				t.Errorf("Expected message %q, got %q", tt.expectedMessage, newModel.message)
			}

			// Input should always be cleared after command execution
			if newModel.input != "" {
				t.Error("Input should be cleared after command execution")
			}
		})
	}
}

func TestUndo(t *testing.T) {
	modelPtr := NewModel()
	model := *modelPtr // Get value type for undo/redo methods

	// Test undo with no history
	undoModel := model.undo()
	if undoModel.message != "Nothing to undo" {
		t.Errorf("Expected 'Nothing to undo', got %q", undoModel.message)
	}

	// Add some state changes using direct history manipulation
	newState := state.NewGameState()
	newState.BankMoney = 5000
	model.history.Push(newState)
	model.gameState = newState

	// Test successful undo
	undoModel = model.undo()
	if undoModel.message != "Undid last action" {
		t.Errorf("Expected 'Undid last action', got %q", undoModel.message)
	}

	if undoModel.gameState.BankMoney != 12000 { // Original bank money
		t.Errorf("Expected bank money 12000 after undo, got %d", undoModel.gameState.BankMoney)
	}

	if undoModel.err != nil {
		t.Errorf("Error should be nil after successful undo, got %v", undoModel.err)
	}

	if undoModel.input != "" {
		t.Error("Input should be cleared after undo")
	}
}

func TestRedo(t *testing.T) {
	modelPtr := NewModel()
	model := *modelPtr // Get value type for undo/redo methods

	// Test redo with no redo history
	redoModel := model.redo()
	if redoModel.message != "Nothing to redo" {
		t.Errorf("Expected 'Nothing to redo', got %q", redoModel.message)
	}

	// Add some state changes and undo to create redo history
	newState := state.NewGameState()
	newState.BankMoney = 5000
	model.history.Push(newState)
	model.gameState = newState

	// Undo to create redo opportunity
	model = model.undo()

	// Test successful redo
	redoModel = model.redo()
	if redoModel.message != "Redid action" {
		t.Errorf("Expected 'Redid action', got %q", redoModel.message)
	}

	if redoModel.gameState.BankMoney != 5000 {
		t.Errorf("Expected bank money 5000 after redo, got %d", redoModel.gameState.BankMoney)
	}

	if redoModel.err != nil {
		t.Errorf("Error should be nil after successful redo, got %v", redoModel.err)
	}

	if redoModel.input != "" {
		t.Error("Input should be cleared after redo")
	}
}

func TestIntToString(t *testing.T) {
	tests := []struct {
		input    int
		expected string
	}{
		{0, "0"},
		{1, "1"},
		{42, "42"},
		{123, "123"},
		{-1, "-1"},
		{-42, "-42"},
		{-123, "-123"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := intToString(tt.input)
			if result != tt.expected {
				t.Errorf("intToString(%d) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestBoolToString(t *testing.T) {
	tests := []struct {
		input    bool
		expected string
	}{
		{true, "true"},
		{false, "false"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := boolToString(tt.input)
			if result != tt.expected {
				t.Errorf("boolToString(%t) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestModelStateImmutability(t *testing.T) {
	modelPtr := NewModel()
	model := *modelPtr
	originalBankMoney := model.gameState.BankMoney

	// Execute a command (using direct execution to avoid parser dependencies)
	newModel := model.executeCommand("")

	// Original model should be unchanged
	if model.gameState.BankMoney != originalBankMoney {
		t.Error("Original model was modified during command execution")
	}

	// Models should be different instances (we can't easily test this with value types)
	// Just verify the behavior is correct
	_ = newModel
}

func TestModelHistoryIntegration(t *testing.T) {
	modelPtr := NewModel()
	model := *modelPtr

	// Simulate successful command execution by manually updating state
	newState := state.NewGameState()
	newState.BankMoney = 5000

	// Simulate what executeCommand does for a successful command
	model.history.Push(newState)
	model.gameState = newState
	model.message = "Test action completed"
	model.err = nil

	// Test that undo works
	undoModel := model.undo()
	if undoModel.gameState.BankMoney != 12000 {
		t.Errorf("Expected bank money 12000 after undo, got %d", undoModel.gameState.BankMoney)
	}

	// Test that redo works
	redoModel := undoModel.redo()
	if redoModel.gameState.BankMoney != 5000 {
		t.Errorf("Expected bank money 5000 after redo, got %d", redoModel.gameState.BankMoney)
	}
}

func TestModelErrorHandling(t *testing.T) {
	modelPtr := NewModel()
	model := *modelPtr

	// Simulate command execution with error
	model.err = tea.ErrProgramKilled
	model.message = "Error: something went wrong"

	// Verify error state
	if model.err == nil {
		t.Error("Expected error to be set")
	}

	if model.message == "" {
		t.Error("Expected error message to be set")
	}

	// Test that undo works even with error state
	undoModel := model.undo()
	if undoModel.message != "Nothing to undo" {
		t.Errorf("Expected undo to work, got message: %s", undoModel.message)
	}

	// Test that redo works
	redoModel := model.redo()
	if redoModel.message != "Nothing to redo" {
		t.Errorf("Expected redo to work, got message: %s", redoModel.message)
	}
}