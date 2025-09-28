package parser

import (
	"testing"

	"github.com/papes/18xxCli/state"
)

func TestTokenizerEdgeCases(t *testing.T) {
	// Test tokenizer with edge cases
	tests := []struct {
		input    string
		expected int // Expected number of tokens
	}{
		{"", 0},
		{"   ", 0},
		{"\"unterminated string", 1}, // Should handle unterminated strings
		{"'single quoted string'", 1},
		{"123.45", 2}, // Number with decimal gets split
		{"command arg1 'quoted arg' 123", 4},
	}

	for _, tt := range tests {
		t.Run("tokenize_"+tt.input, func(t *testing.T) {
			tokenizer := NewTokenizer(tt.input)
			tokens := tokenizer.TokenizeAll()
			if len(tokens) != tt.expected {
				t.Errorf("Expected %d tokens for %q, got %d", tt.expected, tt.input, len(tokens))
			}
		})
	}
}

func TestTokenizerSpecialCases(t *testing.T) {
	// Test reading strings with escape sequences
	tokenizer := NewTokenizer(`"quoted \"escaped\" string"`)
	tokens := tokenizer.TokenizeAll()
	if len(tokens) != 1 {
		t.Errorf("Expected 1 token for escaped string, got %d", len(tokens))
	}

	// Test mixed quotes
	tokenizer = NewTokenizer(`'single' "double"`)
	tokens = tokenizer.TokenizeAll()
	if len(tokens) != 2 {
		t.Errorf("Expected 2 tokens for mixed quotes, got %d", len(tokens))
	}
}

func TestRegistryListCommands(t *testing.T) {
	registry := CreateDefaultRegistry()
	commands := registry.List()

	// Should have a reasonable number of commands
	if len(commands) < 10 {
		t.Errorf("Expected at least 10 commands, got %d", len(commands))
	}

	// Should include basic commands
	foundHelp := false
	foundStatus := false
	for _, cmdName := range commands {
		if cmdName == "help" {
			foundHelp = true
		}
		if cmdName == "status" {
			foundStatus = true
		}
	}

	if !foundHelp {
		t.Error("Expected to find 'help' command in list")
	}
	if !foundStatus {
		t.Error("Expected to find 'status' command in list")
	}
}

func TestCommandAliasesCheck(t *testing.T) {
	registry := CreateDefaultRegistry()

	// Test that aliases work for commands
	helpCmd, exists := registry.Get("help")
	if !exists {
		t.Fatal("Help command should exist")
	}

	aliases := helpCmd.Aliases()
	if len(aliases) == 0 {
		t.Error("Help command should have aliases")
	}

	// Test that we can access command by alias
	for _, alias := range aliases {
		cmd, exists := registry.Get(alias)
		if !exists {
			t.Errorf("Should be able to access help command by alias %q", alias)
		}
		if cmd.Name() != "help" {
			t.Errorf("Alias %q should point to help command", alias)
		}
	}
}

func TestGetSuggestionsEdgeCases(t *testing.T) {
	registry := CreateDefaultRegistry()

	// Test empty prefix
	suggestions := registry.GetSuggestions("")
	if len(suggestions) == 0 {
		t.Error("Empty prefix should return some suggestions")
	}

	// Test non-matching prefix
	suggestions = registry.GetSuggestions("xyz")
	if len(suggestions) != 0 {
		t.Error("Non-matching prefix should return no suggestions")
	}

	// Test prefix that matches multiple commands
	suggestions = registry.GetSuggestions("a")
	if len(suggestions) == 0 {
		t.Error("Prefix 'a' should match some commands")
	}
}

func TestCommandUsage(t *testing.T) {
	registry := CreateDefaultRegistry()

	// Test that we can get usage for specific commands
	helpCmd, exists := registry.Get("help")
	if !exists {
		t.Fatal("Help command should exist")
	}

	usage := helpCmd.Usage()
	if usage == "" {
		t.Error("Help command should have usage string")
	}

	// Test status command too
	statusCmd, exists := registry.Get("status")
	if !exists {
		t.Fatal("Status command should exist")
	}

	usage = statusCmd.Usage()
	if usage == "" {
		t.Error("Status command should have usage string")
	}
}

func TestExecutorWithNilConfig(t *testing.T) {
	registry := CreateDefaultRegistry()
	executor := NewCommandExecutor(registry)
	gameState := state.NewGameState()

	// Test commands that should work without config
	result := executor.Execute("help", gameState, nil)
	if result.Error != nil {
		t.Errorf("Help command should work without config, got error: %v", result.Error)
	}

	result = executor.Execute("status", gameState, nil)
	if result.Error != nil {
		t.Errorf("Status command should work without config, got error: %v", result.Error)
	}
}

func TestIntToStringHelper(t *testing.T) {
	// Test the intToString helper function for edge cases
	tests := []struct {
		input    int
		expected string
	}{
		{0, "0"},
		{1, "1"},
		{-1, "-1"},
		{999, "999"},
		{1000, "1000"},
		{-999, "-999"},
		{123456789, "123456789"},
	}

	for _, tt := range tests {
		result := intToString(tt.input)
		if result != tt.expected {
			t.Errorf("intToString(%d) = %q, expected %q", tt.input, result, tt.expected)
		}
	}
}

func TestParseStringAsIntEdgeCases(t *testing.T) {
	// Test parseStringAsInt with various inputs
	tests := []struct {
		input     string
		expectErr bool
	}{
		{"0", false},
		{"123", false},
		{"-123", true}, // Negative numbers not supported
		{"abc", true},
		{"", false}, // Empty string returns 0
		{"123abc", true},
		{"12.34", true},
	}

	for _, tt := range tests {
		val, err := parseStringAsInt(tt.input)
		if tt.expectErr && err == nil {
			t.Errorf("parseStringAsInt(%q) should return error", tt.input)
		}
		if !tt.expectErr && err != nil {
			t.Errorf("parseStringAsInt(%q) should not return error, got: %v", tt.input, err)
		}
		if !tt.expectErr && tt.input != "" {
			// For valid inputs, check that conversion worked
			expectedVal := 0
			if tt.input == "123" {
				expectedVal = 123
			}
			if tt.input != "0" && val != expectedVal {
				t.Errorf("parseStringAsInt(%q) = %d, expected %d", tt.input, val, expectedVal)
			}
		}
	}
}