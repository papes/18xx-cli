package parser

import (
	"testing"

	"github.com/papes/18xxCli/config"
	"github.com/papes/18xxCli/state"
)

func TestUncoveredFunctions(t *testing.T) {
	registry := CreateDefaultRegistry()
	executor := NewCommandExecutor(registry)

	// Test specific commands that weren't covered
	gameState := state.NewGameState()
	gameState = gameState.AddPlayer("alice", "Alice", 1000)
	gameState = gameState.AddCompany("BO", "Baltimore & Ohio", 10)

	// Test Description method on BaseCommand
	helpCmd, exists := registry.Get("help")
	if !exists {
		t.Fatal("Help command should exist")
	}
	description := helpCmd.Description()
	if description == "" {
		t.Error("Help command should have description")
	}

	// Test more edge cases for parser
	result := executor.Execute("withhold BO", gameState, nil)
	if result.Error == nil {
		t.Error("withhold with insufficient args should fail")
	}

	result = executor.Execute("revenue BO", gameState, nil)
	if result.Error == nil {
		t.Error("revenue with insufficient args should fail")
	}

	result = executor.Execute("set president BO", gameState, nil)
	if result.Error == nil {
		t.Error("set president with insufficient args should fail")
	}

	result = executor.Execute("buy train BO", gameState, nil)
	if result.Error == nil {
		t.Error("buy train with insufficient args should fail")
	}

	result = executor.Execute("transfer alice", gameState, nil)
	if result.Error == nil {
		t.Error("transfer with insufficient args should fail")
	}

	// Test with game config
	gameConfig := &config.GameConfig{
		Title:                "Test Game",
		BankSharesPayCompany: true,
		IPOSharesPayCompany:  true,
	}

	// Test status with config
	result = executor.Execute("status", gameState, gameConfig)
	if result.Error != nil {
		t.Errorf("status with config should work, got: %v", result.Error)
	}

	// Test formatGameStatus with different states
	emptyState := state.NewGameState()
	statusText := formatGameStatus(emptyState)
	if statusText == "" {
		t.Error("formatGameStatus should return text for empty state")
	}

	// Test boolToString function
	trueStr := boolToString(true)
	if trueStr != "true" {
		t.Errorf("boolToString(true) should return 'true', got %q", trueStr)
	}

	falseStr := boolToString(false)
	if falseStr != "false" {
		t.Errorf("boolToString(false) should return 'false', got %q", falseStr)
	}
}

func TestMoreTokenizerCases(t *testing.T) {
	// Test tokenizer with more edge cases
	tokenizer := NewTokenizer("command 'arg with spaces' \"another arg\"")
	tokens := tokenizer.TokenizeAll()
	if len(tokens) != 3 {
		t.Errorf("Expected 3 tokens, got %d", len(tokens))
	}

	// Test NextToken directly
	tokenizer = NewTokenizer("word1 word2")
	token1 := tokenizer.NextToken()
	if token1.Value != "word1" {
		t.Errorf("Expected 'word1', got %q", token1.Value)
	}

	token2 := tokenizer.NextToken()
	if token2.Value != "word2" {
		t.Errorf("Expected 'word2', got %q", token2.Value)
	}

	// Should return EOF when no more tokens
	token3 := tokenizer.NextToken()
	if token3.Type != TokenEOF {
		t.Errorf("Expected EOF token, got %v", token3.Type)
	}
}

func TestParseCommandFunction(t *testing.T) {
	// Test the ParseCommand function
	parsed, err := ParseCommand("help command")
	if err != nil {
		t.Errorf("ParseCommand should not error on valid input: %v", err)
	}
	if parsed.Command != "help" {
		t.Errorf("Expected command 'help', got %q", parsed.Command)
	}
	if len(parsed.Args) != 1 || parsed.Args[0] != "command" {
		t.Errorf("Expected args ['command'], got %v", parsed.Args)
	}

	// Test empty command (should return error)
	parsed, err = ParseCommand("")
	if err == nil {
		t.Error("ParseCommand should error on empty input")
	}
	if parsed != nil {
		t.Error("ParseCommand should return nil for empty input")
	}
}