package parser

import (
	"testing"

	"github.com/papes/18xxCli/state"
)

func TestDefaultRegistry_Integration(t *testing.T) {
	registry := CreateDefaultRegistry()
	executor := NewCommandExecutor(registry)
	gameState := state.NewGameState()

	tests := []struct {
		name           string
		command        string
		expectError    bool
		expectedStates func(before, after state.GameState) bool
		description    string
	}{
		{
			name:        "help command",
			command:     "help",
			expectError: false,
			expectedStates: func(before, after state.GameState) bool {
				// Help command shouldn't change state
				return before.BankMoney == after.BankMoney &&
					len(before.Players) == len(after.Players) &&
					len(before.Companies) == len(after.Companies)
			},
			description: "help command should not modify game state",
		},
		{
			name:        "status command",
			command:     "status",
			expectError: false,
			expectedStates: func(before, after state.GameState) bool {
				// Status command shouldn't change state
				return before.BankMoney == after.BankMoney &&
					len(before.Players) == len(after.Players) &&
					len(before.Companies) == len(after.Companies)
			},
			description: "status command should not modify game state",
		},
		{
			name:        "add player command",
			command:     "add player Alice 500",
			expectError: false,
			expectedStates: func(before, after state.GameState) bool {
				// Should add one player
				return len(after.Players) == len(before.Players)+1 &&
					after.Players["Alice"].Cash == 500
			},
			description: "add player should add a new player with specified cash",
		},
		{
			name:        "add multiple players",
			command:     "add players Bob Charlie 750",
			expectError: false,
			expectedStates: func(before, after state.GameState) bool {
				// Should add two players
				return len(after.Players) == len(before.Players)+2 &&
					after.Players["Bob"].Cash == 750 &&
					after.Players["Charlie"].Cash == 750
			},
			description: "add players should add multiple players with same cash",
		},
		{
			name:        "invalid command",
			command:     "nonexistent command",
			expectError: true,
			expectedStates: func(before, after state.GameState) bool {
				// State should remain unchanged
				return before.BankMoney == after.BankMoney &&
					len(before.Players) == len(after.Players)
			},
			description: "invalid command should not change state",
		},
		{
			name:        "invalid arguments",
			command:     "add player",
			expectError: true,
			expectedStates: func(before, after state.GameState) bool {
				// State should remain unchanged
				return len(before.Players) == len(after.Players)
			},
			description: "invalid arguments should not change state",
		},
		{
			name:        "quoted player name",
			command:     `add player "Dave Jones" 1000`,
			expectError: false,
			expectedStates: func(before, after state.GameState) bool {
				// Should add player with quoted name
				return len(after.Players) == len(before.Players)+1 &&
					after.Players["Dave Jones"].Cash == 1000
			},
			description: "should handle quoted player names",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			beforeState := gameState
			result := executor.Execute(tt.command, gameState, nil)

			if tt.expectError {
				if result.Error == nil {
					t.Errorf("expected error for command %q, got nil", tt.command)
				}
			} else {
				if result.Error != nil {
					t.Errorf("unexpected error for command %q: %v", tt.command, result.Error)
				}
			}

			// Check state changes
			if tt.expectedStates != nil {
				if !tt.expectedStates(beforeState, result.NewState) {
					t.Errorf("state validation failed for command %q: %s", tt.command, tt.description)
				}
			}

			// Update gameState for next test (if command succeeded)
			if result.Error == nil {
				gameState = result.NewState
			}
		})
	}
}

func TestCommandAliases(t *testing.T) {
	registry := CreateDefaultRegistry()
	executor := NewCommandExecutor(registry)
	gameState := state.NewGameState()

	tests := []struct {
		name    string
		command string
		alias   string
		args    string
	}{
		{
			name:    "help alias",
			command: "help",
			alias:   "h",
			args:    "",
		},
		{
			name:    "status alias",
			command: "status",
			alias:   "st",
			args:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fullCommand := tt.command + " " + tt.args
			aliasCommand := tt.alias + " " + tt.args

			result1 := executor.Execute(fullCommand, gameState, nil)
			result2 := executor.Execute(aliasCommand, gameState, nil)

			// Both should have same outcome
			if (result1.Error == nil) != (result2.Error == nil) {
				t.Errorf("command and alias had different error states")
			}

			if result1.Error == nil && result2.Error == nil {
				// For read-only commands, messages should be identical
				if result1.Message != result2.Message {
					t.Errorf("command and alias returned different messages")
				}
			}
		})
	}
}

func TestErrorHandling(t *testing.T) {
	registry := CreateDefaultRegistry()
	executor := NewCommandExecutor(registry)
	gameState := state.NewGameState()

	errorTests := []struct {
		name        string
		command     string
		expectError string
	}{
		{
			name:        "empty command",
			command:     "",
			expectError: "empty command",
		},
		{
			name:        "unknown command",
			command:     "invalid",
			expectError: "unknown command",
		},
		{
			name:        "insufficient arguments",
			command:     "add player Alice",
			expectError: "invalid arguments",
		},
		{
			name:        "invalid argument type",
			command:     "add player Alice notanumber",
			expectError: "invalid arguments",
		},
		{
			name:        "negative cash",
			command:     "add player Alice -100",
			expectError: "invalid arguments",
		},
	}

	for _, tt := range errorTests {
		t.Run(tt.name, func(t *testing.T) {
			result := executor.Execute(tt.command, gameState, nil)

			if result.Error == nil {
				t.Errorf("expected error for command %q, got nil", tt.command)
				return
			}

			// Check that error message contains expected text
			errorMsg := result.Error.Error()
			if !containsIgnoreCase(errorMsg, tt.expectError) {
				t.Errorf("expected error containing %q, got %q", tt.expectError, errorMsg)
			}

			// Ensure state wasn't modified
			if len(result.NewState.Players) != len(gameState.Players) {
				t.Errorf("state was modified despite error")
			}
		})
	}
}

// Helper function to check if a string contains another string (case insensitive)
func containsIgnoreCase(s, substr string) bool {
	s = toLowerCase(s)
	substr = toLowerCase(substr)
	return contains(s, substr)
}

func toLowerCase(s string) string {
	result := ""
	for _, r := range s {
		if r >= 'A' && r <= 'Z' {
			result += string(r + 32)
		} else {
			result += string(r)
		}
	}
	return result
}

func contains(s, substr string) bool {
	if len(substr) > len(s) {
		return false
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
