package parser

import (
	"fmt"
	"strconv"
	"strings"
)

// Token represents a parsed token with its type and value
type Token struct {
	Type  TokenType
	Value string
}

// TokenType represents the type of a parsed token
type TokenType int

const (
	TokenString TokenType = iota
	TokenNumber
	TokenEOF
)

// Tokenizer handles parsing command strings into tokens
type Tokenizer struct {
	input    string
	position int
	current  rune
}

// NewTokenizer creates a new tokenizer for the given input
func NewTokenizer(input string) *Tokenizer {
	t := &Tokenizer{
		input:    strings.TrimSpace(input),
		position: 0,
	}
	if len(t.input) > 0 {
		t.current = rune(t.input[0])
	}
	return t
}

// nextChar advances to the next character
func (t *Tokenizer) nextChar() {
	t.position++
	if t.position >= len(t.input) {
		t.current = 0 // EOF
	} else {
		t.current = rune(t.input[t.position])
	}
}

// skipWhitespace skips whitespace characters
func (t *Tokenizer) skipWhitespace() {
	for t.current != 0 && (t.current == ' ' || t.current == '\t' || t.current == '\n' || t.current == '\r') {
		t.nextChar()
	}
}

// readString reads a string token, handling quotes and escaping
func (t *Tokenizer) readString() string {
	var result strings.Builder

	// Check if it starts with a quote
	if t.current == '"' || t.current == '\'' {
		quote := t.current
		t.nextChar() // skip opening quote

		for t.current != 0 && t.current != quote {
			if t.current == '\\' {
				t.nextChar()
				if t.current != 0 {
					// Handle escape sequences
					switch t.current {
					case 'n':
						result.WriteRune('\n')
					case 't':
						result.WriteRune('\t')
					case 'r':
						result.WriteRune('\r')
					case '\\':
						result.WriteRune('\\')
					case '"':
						result.WriteRune('"')
					case '\'':
						result.WriteRune('\'')
					default:
						result.WriteRune(t.current)
					}
					t.nextChar()
				}
			} else {
				result.WriteRune(t.current)
				t.nextChar()
			}
		}

		if t.current == quote {
			t.nextChar() // skip closing quote
		}
	} else {
		// Read until whitespace
		for t.current != 0 && t.current != ' ' && t.current != '\t' && t.current != '\n' && t.current != '\r' {
			result.WriteRune(t.current)
			t.nextChar()
		}
	}

	return result.String()
}

// readNumber reads a numeric token
func (t *Tokenizer) readNumber() string {
	var result strings.Builder

	for t.current != 0 && (t.current >= '0' && t.current <= '9') {
		result.WriteRune(t.current)
		t.nextChar()
	}

	return result.String()
}

// isDigit checks if the current character is a digit
func (t *Tokenizer) isDigit() bool {
	return t.current >= '0' && t.current <= '9'
}

// NextToken returns the next token from the input
func (t *Tokenizer) NextToken() Token {
	for t.current != 0 {
		t.skipWhitespace()

		if t.current == 0 {
			break
		}

		if t.isDigit() {
			return Token{TokenNumber, t.readNumber()}
		}

		return Token{TokenString, t.readString()}
	}

	return Token{TokenEOF, ""}
}

// TokenizeAll returns all tokens from the input
func (t *Tokenizer) TokenizeAll() []Token {
	var tokens []Token

	for {
		token := t.NextToken()
		if token.Type == TokenEOF {
			break
		}
		tokens = append(tokens, token)
	}

	return tokens
}

// ParsedCommand represents a parsed command with its arguments
type ParsedCommand struct {
	Command string
	Args    []interface{}
}

// ParseCommand tokenizes a command string and returns a structured representation
func ParseCommand(input string) (*ParsedCommand, error) {
	if strings.TrimSpace(input) == "" {
		return nil, fmt.Errorf("empty command")
	}

	tokenizer := NewTokenizer(input)
	tokens := tokenizer.TokenizeAll()

	if len(tokens) == 0 {
		return nil, fmt.Errorf("no command found")
	}

	// First token is the command
	command := tokens[0].Value
	args := make([]interface{}, 0, len(tokens)-1)

	// Convert tokens based on command context
	for i := 1; i < len(tokens); i++ {
		token := tokens[i]
		if token.Type == TokenNumber {
			// Convert numeric tokens to integers, except for certain command contexts
			shouldKeepAsString := false

			// Keep game names as strings for setup commands
			if command == "setup" && i == 2 && len(tokens) > 2 && tokens[1].Value == "game" {
				shouldKeepAsString = true // "setup game 1830" - keep "1830" as string
			}

			if shouldKeepAsString {
				args = append(args, token.Value)
			} else {
				// Convert to integer for numeric arguments
				if num, err := strconv.Atoi(token.Value); err == nil {
					args = append(args, num)
				} else {
					args = append(args, token.Value) // Fallback to string
				}
			}
		} else {
			args = append(args, token.Value)
		}
	}

	return &ParsedCommand{
		Command: command,
		Args:    args,
	}, nil
}
