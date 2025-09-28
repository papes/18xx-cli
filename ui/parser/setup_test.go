package parser

import (
	"testing"
)

func TestSetupGameCommand_Validation(t *testing.T) {
	cmd := &SetupGameCommand{}

	// Test with string game name (should work)
	args := []interface{}{"1830", "alice", "bob", "charlie"}
	err := cmd.Validate(args)
	if err != nil {
		t.Errorf("Expected validation to pass with string game name, got error: %v", err)
	}

	// Test with invalid game name type
	args = []interface{}{1830, "alice", "bob"}
	err = cmd.Validate(args)
	if err == nil {
		t.Error("Expected validation to fail with int game name (should be string)")
	}
}

func TestLoadCompaniesCommand_Validation(t *testing.T) {
	cmd := &LoadCompaniesCommand{}

	// Test with string game name (should work)
	args := []interface{}{"1830"}
	err := cmd.Validate(args)
	if err != nil {
		t.Errorf("Expected validation to pass with string game name, got error: %v", err)
	}

	// Test with invalid game name type
	args = []interface{}{1830}
	err = cmd.Validate(args)
	if err == nil {
		t.Error("Expected validation to fail with int game name (should be string)")
	}
}

func TestParseCommand_AllArgumentsAsStrings(t *testing.T) {
	// Test that "setup game 1830 alice bob" is parsed with all args as strings
	result, err := ParseCommand("setup game 1830 alice bob")
	if err != nil {
		t.Fatalf("Failed to parse command: %v", err)
	}

	if result.Command != "setup" {
		t.Errorf("Expected command 'setup', got '%s'", result.Command)
	}

	if len(result.Args) != 4 {
		t.Fatalf("Expected 4 args, got %d", len(result.Args))
	}

	// All args should be strings now
	expectedArgs := []string{"game", "1830", "alice", "bob"}
	for i, expected := range expectedArgs {
		if result.Args[i] != expected {
			t.Errorf("Expected arg %d to be '%s', got %v", i, expected, result.Args[i])
		}
	}
}
