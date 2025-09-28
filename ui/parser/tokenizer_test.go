package parser

import (
	"testing"
)

func TestTokenizer_NextToken(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []Token
	}{
		{
			name:  "simple command",
			input: "help",
			expected: []Token{
				{TokenString, "help"},
			},
		},
		{
			name:  "command with string arguments",
			input: "add player Alice",
			expected: []Token{
				{TokenString, "add"},
				{TokenString, "player"},
				{TokenString, "Alice"},
			},
		},
		{
			name:  "command with number arguments",
			input: "add player Bob 500",
			expected: []Token{
				{TokenString, "add"},
				{TokenString, "player"},
				{TokenString, "Bob"},
				{TokenNumber, "500"},
			},
		},
		{
			name:  "quoted strings",
			input: `add player "Alice Smith" 500`,
			expected: []Token{
				{TokenString, "add"},
				{TokenString, "player"},
				{TokenString, "Alice Smith"},
				{TokenNumber, "500"},
			},
		},
		{
			name:  "single quoted strings",
			input: `add player 'Bob Jones' 750`,
			expected: []Token{
				{TokenString, "add"},
				{TokenString, "player"},
				{TokenString, "Bob Jones"},
				{TokenNumber, "750"},
			},
		},
		{
			name:  "escaped characters",
			input: `add player "Alice \"Boss\" Smith" 1000`,
			expected: []Token{
				{TokenString, "add"},
				{TokenString, "player"},
				{TokenString, `Alice "Boss" Smith`},
				{TokenNumber, "1000"},
			},
		},
		{
			name:  "multiple spaces",
			input: "  add   player    Charlie   250  ",
			expected: []Token{
				{TokenString, "add"},
				{TokenString, "player"},
				{TokenString, "Charlie"},
				{TokenNumber, "250"},
			},
		},
		{
			name:     "empty input",
			input:    "",
			expected: []Token{},
		},
		{
			name:     "whitespace only",
			input:    "   \t\n  ",
			expected: []Token{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokenizer := NewTokenizer(tt.input)
			tokens := tokenizer.TokenizeAll()

			if len(tokens) != len(tt.expected) {
				t.Errorf("expected %d tokens, got %d", len(tt.expected), len(tokens))
				return
			}

			for i, expected := range tt.expected {
				if i >= len(tokens) {
					t.Errorf("missing token at index %d", i)
					continue
				}

				actual := tokens[i]
				if actual.Type != expected.Type {
					t.Errorf("token %d: expected type %v, got %v", i, expected.Type, actual.Type)
				}
				if actual.Value != expected.Value {
					t.Errorf("token %d: expected value %q, got %q", i, expected.Value, actual.Value)
				}
			}
		})
	}
}

func TestParseCommand(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectError bool
		expected    *ParsedCommand
	}{
		{
			name:        "simple command",
			input:       "help",
			expectError: false,
			expected: &ParsedCommand{
				Command: "help",
				Args:    []interface{}{},
			},
		},
		{
			name:        "command with string args",
			input:       "add player Alice",
			expectError: false,
			expected: &ParsedCommand{
				Command: "add",
				Args:    []interface{}{"player", "Alice"},
			},
		},
		{
			name:        "command with mixed args",
			input:       "add player Bob 500",
			expectError: false,
			expected: &ParsedCommand{
				Command: "add",
				Args:    []interface{}{"player", "Bob", 500},
			},
		},
		{
			name:        "command with quoted string",
			input:       `set par "Baltimore & Ohio" 76`,
			expectError: false,
			expected: &ParsedCommand{
				Command: "set",
				Args:    []interface{}{"par", "Baltimore & Ohio", 76},
			},
		},
		{
			name:        "empty input",
			input:       "",
			expectError: true,
			expected:    nil,
		},
		{
			name:        "whitespace only",
			input:       "   \t  ",
			expectError: true,
			expected:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseCommand(tt.input)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if result.Command != tt.expected.Command {
				t.Errorf("expected command %q, got %q", tt.expected.Command, result.Command)
			}

			if len(result.Args) != len(tt.expected.Args) {
				t.Errorf("expected %d args, got %d", len(tt.expected.Args), len(result.Args))
				return
			}

			for i, expected := range tt.expected.Args {
				if result.Args[i] != expected {
					t.Errorf("arg %d: expected %v, got %v", i, expected, result.Args[i])
				}
			}
		})
	}
}
