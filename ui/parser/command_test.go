package parser

import (
	"testing"

	"github.com/papes/18xxCli/config"
	"github.com/papes/18xxCli/state"
)

func TestCommandRegistry(t *testing.T) {
	registry := NewCommandRegistry()

	// Create a test command
	testCmd := &TestCommand{
		BaseCommand: NewBaseCommand("test", "Test command", "test <arg>", "t"),
	}

	// Test registration
	registry.Register(testCmd)

	// Test retrieval by name
	cmd, exists := registry.Get("test")
	if !exists {
		t.Error("command not found by name")
	}
	if cmd != testCmd {
		t.Error("wrong command returned")
	}

	// Test retrieval by alias
	cmd, exists = registry.Get("t")
	if !exists {
		t.Error("command not found by alias")
	}
	if cmd != testCmd {
		t.Error("wrong command returned by alias")
	}

	// Test case insensitive
	cmd, exists = registry.Get("TEST")
	if !exists {
		t.Error("command not found with uppercase")
	}

	cmd, exists = registry.Get("T")
	if !exists {
		t.Error("alias not found with uppercase")
	}

	// Test non-existent command
	_, exists = registry.Get("nonexistent")
	if exists {
		t.Error("non-existent command found")
	}
}

func TestCommandExecutor_Execute(t *testing.T) {
	registry := NewCommandRegistry()
	testCmd := &TestCommand{
		BaseCommand: NewBaseCommand("test", "Test command", "test <arg>"),
	}
	registry.Register(testCmd)

	executor := NewCommandExecutor(registry)
	gameState := state.NewGameState()

	tests := []struct {
		name            string
		command         string
		expectError     bool
		expectedMessage string
	}{
		{
			name:            "valid command",
			command:         "test hello",
			expectError:     false,
			expectedMessage: "Test executed with: hello",
		},
		{
			name:        "invalid command",
			command:     "nonexistent",
			expectError: true,
		},
		{
			name:        "validation error",
			command:     "test",
			expectError: true,
		},
		{
			name:        "empty command",
			command:     "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := executor.Execute(tt.command, gameState, nil)

			if tt.expectError {
				if result.Error == nil {
					t.Error("expected error, got nil")
				}
			} else {
				if result.Error != nil {
					t.Errorf("unexpected error: %v", result.Error)
				}
				if result.Message != tt.expectedMessage {
					t.Errorf("expected message %q, got %q", tt.expectedMessage, result.Message)
				}
			}
		})
	}
}

func TestMultiWordCommands(t *testing.T) {
	registry := NewCommandRegistry()

	// Register multi-word command
	addPlayerCmd := &TestCommand{
		BaseCommand: NewBaseCommand("add player", "Add player command", "add player <name>"),
	}
	registry.Register(addPlayerCmd)

	executor := NewCommandExecutor(registry)
	gameState := state.NewGameState()

	// Test that "add player Alice" correctly identifies the command
	result := executor.Execute("add player Alice", gameState, nil)

	if result.Error != nil {
		t.Errorf("unexpected error: %v", result.Error)
	}

	expected := "Test executed with: Alice"
	if result.Message != expected {
		t.Errorf("expected message %q, got %q", expected, result.Message)
	}
}

func TestGetSuggestions(t *testing.T) {
	registry := NewCommandRegistry()

	registry.Register(&TestCommand{
		BaseCommand: NewBaseCommand("help", "Help command", "help"),
	})
	registry.Register(&TestCommand{
		BaseCommand: NewBaseCommand("history", "History command", "history"),
	})
	registry.Register(&TestCommand{
		BaseCommand: NewBaseCommand("status", "Status command", "status"),
	})

	tests := []struct {
		input       string
		expectedLen int
		contains    []string
	}{
		{
			input:       "h",
			expectedLen: 2,
			contains:    []string{"help", "history"},
		},
		{
			input:       "help",
			expectedLen: 1,
			contains:    []string{"help"},
		},
		{
			input:       "st",
			expectedLen: 1,
			contains:    []string{"status"},
		},
		{
			input:       "xyz",
			expectedLen: 0,
			contains:    []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			suggestions := registry.GetSuggestions(tt.input)

			if len(suggestions) != tt.expectedLen {
				t.Errorf("expected %d suggestions, got %d", tt.expectedLen, len(suggestions))
			}

			for _, expected := range tt.contains {
				found := false
				for _, suggestion := range suggestions {
					if suggestion == expected {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected suggestion %q not found", expected)
				}
			}
		})
	}
}

// TestCommand is a simple test implementation of Command
type TestCommand struct {
	BaseCommand
}

func (c *TestCommand) Validate(args []interface{}) error {
	if len(args) != 1 {
		return NewValidationError("requires exactly one argument")
	}
	if _, ok := args[0].(string); !ok {
		return NewValidationError("argument must be a string")
	}
	return nil
}

func (c *TestCommand) Execute(args []interface{}, gameState state.GameState, gameConfig *config.GameConfig) CommandResult {
	arg := args[0].(string)
	return CommandResult{
		NewState:      gameState,
		Message:       "Test executed with: " + arg,
		Error:         nil,
		NewGameConfig: gameConfig,
	}
}

func TestLogsCommand(t *testing.T) {
	registry := CreateDefaultRegistry()
	executor := NewCommandExecutor(registry)

	// Create a game state with some log entries
	gameState := state.NewGameState()
	gameState = gameState.AddPlayer("Alice", "Alice", 600)
	gameState = gameState.AddCompany("BO", "Baltimore & Ohio", 10)

	// Test logs command with empty log
	result := executor.Execute("logs", gameState, nil)
	if result.Error != nil {
		t.Errorf("unexpected error: %v", result.Error)
	}
	if result.Message != "No actions logged yet" {
		t.Errorf("expected 'No actions logged yet', got %q", result.Message)
	}

	// Add log entries to test the new dividend payment format
	// Add individual dividend payment entry
	gameState.Log = append(gameState.Log, state.LogEntry{
		Action: "dividend_payment",
		Details: map[string]interface{}{
			"company":         "BO",
			"recipient":       "Alice",
			"recipient_type":  "player",
			"shares_owned":    state.ShareCount(3),
			"amount_received": state.Money(50),
			"total_dividend":  state.Money(100),
		},
	})

	// Add summary entry
	gameState.Log = append(gameState.Log, state.LogEntry{
		Action: "pay_dividend_with_config",
		Details: map[string]interface{}{
			"company":             "BO",
			"total_amount":        state.Money(100),
			"individual_payments": 1,
		},
	})

	// Test logs command with log entries
	result = executor.Execute("logs", gameState, nil)
	if result.Error != nil {
		t.Errorf("unexpected error: %v", result.Error)
	}

	// Check that the message contains expected content
	if result.Message == "" {
		t.Error("expected non-empty log message")
	}

	// Should contain the expected log content
	expectedContents := []string{"BO", "dividend_payment", "Alice receives $50", "3 shares of BO", "pay_dividend_with_config"}
	for _, expected := range expectedContents {
		if !stringContains(result.Message, expected) {
			t.Errorf("expected log message to contain %q, got %q", expected, result.Message)
		}
	}
}

func stringContains(s, substr string) bool {
	return len(s) >= len(substr) && s[len(s)-len(substr):] == substr ||
		(len(s) > len(substr) && findSubstring(s, substr) >= 0)
}

func findSubstring(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
