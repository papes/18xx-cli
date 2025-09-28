package security

import (
	"strings"
	"testing"
)

func TestNewInputSanitizer(t *testing.T) {
	sanitizer := NewInputSanitizer()

	if sanitizer.maxLength != 1000 {
		t.Errorf("Expected default max length 1000, got %d", sanitizer.maxLength)
	}

	if sanitizer.allowedCommands == nil {
		t.Error("Expected allowed commands to be initialized")
	}

	if sanitizer.forbiddenChars == nil {
		t.Error("Expected forbidden chars regex to be initialized")
	}
}

func TestSetMaxLength(t *testing.T) {
	sanitizer := NewInputSanitizer()
	sanitizer.SetMaxLength(500)

	if sanitizer.maxLength != 500 {
		t.Errorf("Expected max length 500, got %d", sanitizer.maxLength)
	}
}

func TestSetAllowedCommands(t *testing.T) {
	sanitizer := NewInputSanitizer()
	commands := []string{"help", "status", "quit"}
	sanitizer.SetAllowedCommands(commands)

	if !sanitizer.allowedCommands["help"] {
		t.Error("Expected 'help' to be allowed")
	}

	if !sanitizer.allowedCommands["status"] {
		t.Error("Expected 'status' to be allowed")
	}

	if sanitizer.allowedCommands["unknown"] {
		t.Error("Expected 'unknown' to not be allowed")
	}
}

func TestSanitizeInput_ValidInputs(t *testing.T) {
	sanitizer := NewInputSanitizer()

	validInputs := []string{
		"help",
		"status",
		"add player Alice 600",
		"buy ipo Alice BO 2",
		"set par BO 67",
		"transfer Alice Bob 100",
	}

	for _, input := range validInputs {
		t.Run(input, func(t *testing.T) {
			result, err := sanitizer.SanitizeInput(input)
			if err != nil {
				t.Errorf("Expected valid input %q to pass sanitization, got error: %v", input, err)
			}
			if result != input {
				t.Errorf("Expected result %q, got %q", input, result)
			}
		})
	}
}

func TestSanitizeInput_InvalidInputs(t *testing.T) {
	sanitizer := NewInputSanitizer()

	invalidInputs := []struct {
		input       string
		expectError string
	}{
		{"", "empty input"},
		{"   ", "empty input"},
		{strings.Repeat("a", 1001), "input too long"},
		{"help\x00", "forbidden characters"},
		{"help\x1F", "forbidden characters"},
		{"help\x7F", "forbidden characters"},
		{"unknown_command", "forbidden command"},
		{"help && rm -rf /", "suspicious characters"},
		{"help | cat", "suspicious characters"},
		{"help; exit", "suspicious characters"},
		{"help > file", "suspicious characters"},
		{"help < file", "suspicious characters"},
		{"help & background", "suspicious characters"},
		{"help $variable", "suspicious characters"},
		{"help `command`", "suspicious characters"},
		{"help\\escape", "suspicious characters"},
		{"help 'unbalanced", "unbalanced quotes"},
		{"help \"unbalanced", "unbalanced quotes"},
	}

	for _, tt := range invalidInputs {
		t.Run(tt.input, func(t *testing.T) {
			_, err := sanitizer.SanitizeInput(tt.input)
			if err == nil {
				t.Errorf("Expected input %q to fail sanitization", tt.input)
			}
			if !strings.Contains(err.Error(), tt.expectError) {
				t.Errorf("Expected error containing %q, got: %v", tt.expectError, err)
			}
		})
	}
}

func TestSanitizeInput_TrimWhitespace(t *testing.T) {
	sanitizer := NewInputSanitizer()

	input := "  help  "
	expected := "help"

	result, err := sanitizer.SanitizeInput(input)
	if err != nil {
		t.Errorf("Expected input to pass sanitization, got error: %v", err)
	}

	if result != expected {
		t.Errorf("Expected trimmed result %q, got %q", expected, result)
	}
}

func TestValidateCommandStructure(t *testing.T) {
	sanitizer := NewInputSanitizer()

	tests := []struct {
		name        string
		input       string
		expectError bool
	}{
		{"valid single command", "help", false},
		{"valid multi-word command", "add player Alice", false},
		{"valid quoted argument", "add player 'Alice Smith' 600", false},
		{"balanced quotes", "add player \"Alice\" 600", false},
		{"unbalanced single quotes", "add player 'Alice 600", true},
		{"unbalanced double quotes", "add player \"Alice 600", true},
		{"suspicious characters &&", "help && exit", true},
		{"suspicious characters ||", "help || exit", true},
		{"suspicious characters ;", "help; exit", true},
		{"suspicious characters |", "help | cat", true},
		{"suspicious characters >", "help > file", true},
		{"suspicious characters <", "help < file", true},
		{"suspicious characters &", "help & background", true},
		{"suspicious characters $", "help $var", true},
		{"suspicious characters `", "help `cmd`", true},
		{"suspicious characters \\", "help\\escape", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := sanitizer.validateCommandStructure(tt.input)
			if (err != nil) != tt.expectError {
				t.Errorf("validateCommandStructure(%q) error = %v, expectError = %v", tt.input, err, tt.expectError)
			}
		})
	}
}

func TestValidateCommand(t *testing.T) {
	sanitizer := NewInputSanitizer()

	tests := []struct {
		name        string
		input       string
		expectError bool
	}{
		{"valid single command", "help", false},
		{"valid alias", "h", false},
		{"valid multi-word command", "add player Alice 600", false},
		{"valid multi-word alias", "ap Alice 600", false},
		{"invalid command", "unknown_command", true},
		{"empty input", "", true},
		{"valid buy command", "buy ipo Alice BO 2", false},
		{"valid set command", "set par BO 67", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := sanitizer.validateCommand(tt.input)
			if (err != nil) != tt.expectError {
				t.Errorf("validateCommand(%q) error = %v, expectError = %v", tt.input, err, tt.expectError)
			}
		})
	}
}

func TestSanitizeArgument(t *testing.T) {
	sanitizer := NewInputSanitizer()

	tests := []struct {
		name        string
		arg         string
		expected    string
		expectError bool
	}{
		{"valid argument", "Alice", "Alice", false},
		{"valid number", "600", "600", false},
		{"valid company ID", "BO", "BO", false},
		{"argument with null byte", "Alice\x00", "", true},
		{"argument with control char", "Alice\x1F", "", true},
		{"too long argument", strings.Repeat("a", 201), "", true},
		{"valid long argument", strings.Repeat("a", 200), strings.Repeat("a", 200), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := sanitizer.SanitizeArgument(tt.arg)
			if (err != nil) != tt.expectError {
				t.Errorf("SanitizeArgument(%q) error = %v, expectError = %v", tt.arg, err, tt.expectError)
			}
			if !tt.expectError && result != tt.expected {
				t.Errorf("Expected result %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestValidateNumericInput(t *testing.T) {
	sanitizer := NewInputSanitizer()

	tests := []struct {
		name        string
		input       string
		expected    int
		expectError bool
	}{
		{"valid number", "123", 123, false},
		{"zero", "0", 0, false},
		{"large number", "999999", 999999, false},
		{"empty input", "", 0, true},
		{"whitespace only", "   ", 0, true},
		{"negative number", "-123", 0, true},
		{"decimal number", "12.34", 0, true},
		{"non-numeric", "abc", 0, true},
		{"mixed alphanumeric", "12a34", 0, true},
		{"too long number", "12345678901", 0, true},
		{"valid with whitespace", " 123 ", 123, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := sanitizer.ValidateNumericInput(tt.input)
			if (err != nil) != tt.expectError {
				t.Errorf("ValidateNumericInput(%q) error = %v, expectError = %v", tt.input, err, tt.expectError)
			}
			if !tt.expectError && result != tt.expected {
				t.Errorf("Expected result %d, got %d", tt.expected, result)
			}
		})
	}
}

func TestValidatePlayerID(t *testing.T) {
	sanitizer := NewInputSanitizer()

	tests := []struct {
		name        string
		id          string
		expectError bool
	}{
		{"valid player ID", "alice", false},
		{"valid with numbers", "player123", false},
		{"valid with underscore", "player_1", false},
		{"valid with hyphen", "player-1", false},
		{"empty ID", "", true},
		{"too long ID", strings.Repeat("a", 51), true},
		{"with space", "alice bob", true},
		{"with special chars", "alice@bob", true},
		{"with dot", "alice.bob", true},
		{"max length valid", strings.Repeat("a", 50), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := sanitizer.ValidatePlayerID(tt.id)
			if (err != nil) != tt.expectError {
				t.Errorf("ValidatePlayerID(%q) error = %v, expectError = %v", tt.id, err, tt.expectError)
			}
		})
	}
}

func TestValidateCompanyID(t *testing.T) {
	sanitizer := NewInputSanitizer()

	tests := []struct {
		name        string
		id          string
		expectError bool
	}{
		{"valid company ID", "BO", false},
		{"valid with numbers", "NYC123", false},
		{"valid with underscore", "B_O", false},
		{"valid with hyphen", "B-O", false},
		{"empty ID", "", true},
		{"too long ID", strings.Repeat("a", 21), true},
		{"with space", "B O", true},
		{"with special chars", "B@O", true},
		{"with dot", "B.O", true},
		{"max length valid", strings.Repeat("a", 20), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := sanitizer.ValidateCompanyID(tt.id)
			if (err != nil) != tt.expectError {
				t.Errorf("ValidateCompanyID(%q) error = %v, expectError = %v", tt.id, err, tt.expectError)
			}
		})
	}
}

func TestNormalizeInput(t *testing.T) {
	sanitizer := NewInputSanitizer()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"trim whitespace", "  help  ", "help"},
		{"normalize spaces", "add  player    Alice", "add player Alice"},
		{"mixed whitespace", "add\t\tplayer\n\nAlice", "add player Alice"},
		{"already normalized", "help", "help"},
		{"empty input", "", ""},
		{"only whitespace", "   \t\n  ", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizer.NormalizeInput(tt.input)
			if result != tt.expected {
				t.Errorf("NormalizeInput(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestCreateDefaultAllowedCommands(t *testing.T) {
	commands := createDefaultAllowedCommands()

	// Test that essential commands are included
	essentialCommands := []string{
		"help", "h", "?",
		"status", "st",
		"quit", "exit",
		"add player", "ap",
		"buy", "sell",
		"set par", "set price",
		"transfer",
	}

	for _, cmd := range essentialCommands {
		if !commands[cmd] {
			t.Errorf("Expected essential command %q to be in allowed commands", cmd)
		}
	}
}

func TestMultiWordCommandValidation(t *testing.T) {
	sanitizer := NewInputSanitizer()

	tests := []struct {
		name        string
		input       string
		expectError bool
	}{
		{"add player valid", "add player Alice 600", false},
		{"buy ipo valid", "buy ipo Alice BO 2", false},
		{"set par valid", "set par BO 67", false},
		{"set price valid", "set price BO 75", false},
		{"pay dividend valid", "pay dividend BO 100", false},
		{"invalid multi-word", "invalid command here", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := sanitizer.validateCommand(tt.input)
			if (err != nil) != tt.expectError {
				t.Errorf("validateCommand(%q) error = %v, expectError = %v", tt.input, err, tt.expectError)
			}
		})
	}
}