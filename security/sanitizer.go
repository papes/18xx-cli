package security

import (
	"regexp"
	"strings"
	"unicode"
)

// InputSanitizer provides input sanitization and validation for user commands
type InputSanitizer struct {
	maxLength       int
	allowedCommands map[string]bool
	forbiddenChars  *regexp.Regexp
}

// NewInputSanitizer creates a new input sanitizer with default settings
func NewInputSanitizer() *InputSanitizer {
	// Define forbidden characters that could be used for injection or cause issues
	forbiddenPattern := `[\x00-\x08\x0B\x0C\x0E-\x1F\x7F]` // Control characters except tab, newline, carriage return

	return &InputSanitizer{
		maxLength:       1000,                          // Maximum command length
		allowedCommands: createDefaultAllowedCommands(), // Whitelist of allowed commands
		forbiddenChars:  regexp.MustCompile(forbiddenPattern),
	}
}

// SetMaxLength sets the maximum allowed input length
func (s *InputSanitizer) SetMaxLength(length int) {
	s.maxLength = length
}

// SetAllowedCommands sets the whitelist of allowed commands
func (s *InputSanitizer) SetAllowedCommands(commands []string) {
	s.allowedCommands = make(map[string]bool)
	for _, cmd := range commands {
		s.allowedCommands[strings.ToLower(cmd)] = true
	}
}

// SanitizeInput cleans and validates user input, returning sanitized input and any errors
func (s *InputSanitizer) SanitizeInput(input string) (string, error) {
	// Step 1: Basic length check
	if len(input) > s.maxLength {
		return "", NewSecurityError("input too long", "Input exceeds maximum length of %d characters", s.maxLength)
	}

	// Step 2: Remove forbidden control characters
	if s.forbiddenChars.MatchString(input) {
		return "", NewSecurityError("forbidden characters", "Input contains forbidden control characters")
	}

	// Step 3: Trim whitespace
	sanitized := strings.TrimSpace(input)

	// Step 4: Check for empty input after trimming
	if sanitized == "" {
		return "", NewSecurityError("empty input", "Input is empty after sanitization")
	}

	// Step 5: Basic structure validation - check for reasonable command structure
	if err := s.validateCommandStructure(sanitized); err != nil {
		return "", err
	}

	// Step 6: Command whitelist validation
	if err := s.validateCommand(sanitized); err != nil {
		return "", err
	}

	return sanitized, nil
}

// validateCommandStructure checks for basic command structure issues
func (s *InputSanitizer) validateCommandStructure(input string) error {
	// Check for excessive consecutive whitespace (could indicate injection attempts)
	if strings.Contains(input, "  ") {
		// Normalize multiple spaces to single spaces
		spaceRegex := regexp.MustCompile(`\s+`)
		normalized := spaceRegex.ReplaceAllString(input, " ")
		if len(normalized) != len(input) {
			return NewSecurityWarning("excessive whitespace", "Input contains excessive whitespace")
		}
	}

	// Check for balanced quotes
	singleQuotes := strings.Count(input, "'")
	doubleQuotes := strings.Count(input, "\"")

	if singleQuotes%2 != 0 {
		return NewSecurityError("unbalanced quotes", "Input contains unbalanced single quotes")
	}

	if doubleQuotes%2 != 0 {
		return NewSecurityError("unbalanced quotes", "Input contains unbalanced double quotes")
	}

	// Check for suspicious patterns that might indicate injection attempts
	suspiciousPatterns := []string{
		"&&", "||", ";", "|", ">", "<", "&", "$", "`", "\\",
	}

	for _, pattern := range suspiciousPatterns {
		if strings.Contains(input, pattern) {
			return NewSecurityError("suspicious characters", "Input contains potentially dangerous characters: %s", pattern)
		}
	}

	return nil
}

// validateCommand checks if the command is in the allowed list
func (s *InputSanitizer) validateCommand(input string) error {
	// Extract the first word as the command
	parts := strings.Fields(input)
	if len(parts) == 0 {
		return NewSecurityError("no command", "No command found in input")
	}

	command := strings.ToLower(parts[0])

	// Check multi-word commands (e.g., "add player", "buy ipo")
	for i := 1; i < len(parts) && i < 3; i++ { // Check up to 3-word commands
		multiCommand := strings.Join(parts[:i+1], " ")
		if s.allowedCommands[multiCommand] {
			return nil // Found a valid multi-word command
		}
	}

	// Check single-word command
	if !s.allowedCommands[command] {
		return NewSecurityError("forbidden command", "Command '%s' is not allowed", command)
	}

	return nil
}

// SanitizeArgument sanitizes individual command arguments
func (s *InputSanitizer) SanitizeArgument(arg string) (string, error) {
	// Remove control characters
	if s.forbiddenChars.MatchString(arg) {
		return "", NewSecurityError("forbidden characters", "Argument contains forbidden control characters")
	}

	// Check for excessively long arguments
	if len(arg) > 200 { // Individual arguments shouldn't be too long
		return "", NewSecurityError("argument too long", "Argument exceeds maximum length of 200 characters")
	}

	// Check for null bytes and other dangerous characters
	if strings.ContainsRune(arg, '\x00') {
		return "", NewSecurityError("null byte", "Argument contains null byte")
	}

	return arg, nil
}

// ValidateNumericInput validates numeric inputs for safety
func (s *InputSanitizer) ValidateNumericInput(input string) (int, error) {
	// Remove any whitespace
	cleaned := strings.TrimSpace(input)

	if cleaned == "" {
		return 0, NewSecurityError("empty numeric input", "Numeric input is empty")
	}

	// Check that input contains only digits (positive integers only for safety)
	for _, r := range cleaned {
		if !unicode.IsDigit(r) {
			return 0, NewSecurityError("invalid numeric input", "Numeric input contains non-digit characters")
		}
	}

	// Check for reasonable length (prevent integer overflow attempts)
	if len(cleaned) > 10 { // 10 digits should be more than enough for game values
		return 0, NewSecurityError("numeric input too long", "Numeric input is too long")
	}

	// Convert to integer
	value := 0
	for _, r := range cleaned {
		digit := int(r - '0')
		value = value*10 + digit

		// Check for overflow during conversion
		if value < 0 {
			return 0, NewSecurityError("numeric overflow", "Numeric input causes overflow")
		}
	}

	return value, nil
}

// ValidatePlayerID validates player IDs for safety
func (s *InputSanitizer) ValidatePlayerID(id string) error {
	if id == "" {
		return NewSecurityError("empty player ID", "Player ID cannot be empty")
	}

	if len(id) > 50 {
		return NewSecurityError("player ID too long", "Player ID exceeds maximum length")
	}

	// Player IDs should contain only safe characters
	validIDPattern := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	if !validIDPattern.MatchString(id) {
		return NewSecurityError("invalid player ID", "Player ID contains invalid characters")
	}

	return nil
}

// ValidateCompanyID validates company IDs for safety
func (s *InputSanitizer) ValidateCompanyID(id string) error {
	if id == "" {
		return NewSecurityError("empty company ID", "Company ID cannot be empty")
	}

	if len(id) > 20 {
		return NewSecurityError("company ID too long", "Company ID exceeds maximum length")
	}

	// Company IDs should contain only safe characters
	validIDPattern := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	if !validIDPattern.MatchString(id) {
		return NewSecurityError("invalid company ID", "Company ID contains invalid characters")
	}

	return nil
}

// createDefaultAllowedCommands creates the default whitelist of allowed commands
func createDefaultAllowedCommands() map[string]bool {
	commands := []string{
		// Basic commands
		"help", "h", "?",
		"status", "st",
		"quit", "exit",
		"undo", "redo",
		"logs", "log",

		// Setup commands
		"add player", "ap",
		"add players", "aps",
		"add company", "add companies",
		"setup game",
		"list games",
		"load companies", "load game",

		// Stock operations
		"buy", "b", "buy ipo", "buy bank",
		"sell", "s", "sell bank",
		"set", "sp", "set par", "set price", "set president",

		// Company operations
		"revenue", "withhold",
		"pay dividend",
		"buy train",
		"float",
		"transfer",
	}

	allowed := make(map[string]bool)
	for _, cmd := range commands {
		allowed[strings.ToLower(cmd)] = true
	}

	return allowed
}

// NormalizeInput normalizes input by removing excessive whitespace and converting to consistent format
func (s *InputSanitizer) NormalizeInput(input string) string {
	// Remove leading/trailing whitespace
	normalized := strings.TrimSpace(input)

	// Replace multiple consecutive whitespace with single spaces
	spaceRegex := regexp.MustCompile(`\s+`)
	normalized = spaceRegex.ReplaceAllString(normalized, " ")

	return normalized
}