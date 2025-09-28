package security

import (
	"errors"
	"testing"
)

func TestSecurityError(t *testing.T) {
	err := NewSecurityError("test_type", "test message with %s", "parameter")

	if err.Type != "test_type" {
		t.Errorf("Expected type 'test_type', got %s", err.Type)
	}

	expectedMessage := "test message with parameter"
	if err.Message != expectedMessage {
		t.Errorf("Expected message %q, got %q", expectedMessage, err.Message)
	}

	expectedError := "security error [test_type]: test message with parameter"
	if err.Error() != expectedError {
		t.Errorf("Expected error string %q, got %q", expectedError, err.Error())
	}
}

func TestIsSecurityError(t *testing.T) {
	securityErr := NewSecurityError("test", "test message")
	if !IsSecurityError(securityErr) {
		t.Error("Expected IsSecurityError to return true for SecurityError")
	}

	// Test with a regular error
	regularErr := errors.New("regular error")
	if IsSecurityError(regularErr) {
		t.Error("Expected IsSecurityError to return false for regular error")
	}

	// Test with nil
	if IsSecurityError(nil) {
		t.Error("Expected IsSecurityError to return false for nil")
	}
}

func TestSecurityWarning(t *testing.T) {
	warning := NewSecurityWarning("test_type", "test warning with %d", 42)

	if warning.Type != "test_type" {
		t.Errorf("Expected type 'test_type', got %s", warning.Type)
	}

	expectedMessage := "test warning with 42"
	if warning.Message != expectedMessage {
		t.Errorf("Expected message %q, got %q", expectedMessage, warning.Message)
	}

	expectedError := "security warning [test_type]: test warning with 42"
	if warning.Error() != expectedError {
		t.Errorf("Expected error string %q, got %q", expectedError, warning.Error())
	}
}

func TestIsSecurityWarning(t *testing.T) {
	securityWarning := NewSecurityWarning("test", "test message")
	if !IsSecurityWarning(securityWarning) {
		t.Error("Expected IsSecurityWarning to return true for SecurityWarning")
	}

	// Test with a security error (should return false)
	securityErr := NewSecurityError("test", "test message")
	if IsSecurityWarning(securityErr) {
		t.Error("Expected IsSecurityWarning to return false for SecurityError")
	}

	// Test with a regular error
	regularErr := errors.New("regular error")
	if IsSecurityWarning(regularErr) {
		t.Error("Expected IsSecurityWarning to return false for regular error")
	}

	// Test with nil
	if IsSecurityWarning(nil) {
		t.Error("Expected IsSecurityWarning to return false for nil")
	}
}

func TestSecurityErrorFormatting(t *testing.T) {
	tests := []struct {
		name           string
		errorType      string
		message        string
		args           []interface{}
		expectedOutput string
	}{
		{
			name:           "simple message",
			errorType:      "input_validation",
			message:        "Invalid input",
			args:           nil,
			expectedOutput: "security error [input_validation]: Invalid input",
		},
		{
			name:           "formatted message",
			errorType:      "length_check",
			message:        "Input too long: %d characters (max: %d)",
			args:           []interface{}{150, 100},
			expectedOutput: "security error [length_check]: Input too long: 150 characters (max: 100)",
		},
		{
			name:           "string formatting",
			errorType:      "forbidden_command",
			message:        "Command '%s' not allowed",
			args:           []interface{}{"dangerous_command"},
			expectedOutput: "security error [forbidden_command]: Command 'dangerous_command' not allowed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var err *SecurityError
			if tt.args == nil {
				err = &SecurityError{Type: tt.errorType, Message: tt.message}
			} else {
				err = NewSecurityError(tt.errorType, tt.message, tt.args...)
			}

			if err.Error() != tt.expectedOutput {
				t.Errorf("Expected error output %q, got %q", tt.expectedOutput, err.Error())
			}
		})
	}
}

func TestSecurityWarningFormatting(t *testing.T) {
	tests := []struct {
		name           string
		warningType    string
		message        string
		args           []interface{}
		expectedOutput string
	}{
		{
			name:           "simple warning",
			warningType:    "suspicious_pattern",
			message:        "Suspicious input detected",
			args:           nil,
			expectedOutput: "security warning [suspicious_pattern]: Suspicious input detected",
		},
		{
			name:           "formatted warning",
			warningType:    "whitespace_cleanup",
			message:        "Removed %d excessive spaces",
			args:           []interface{}{5},
			expectedOutput: "security warning [whitespace_cleanup]: Removed 5 excessive spaces",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var warning *SecurityWarning
			if tt.args == nil {
				warning = &SecurityWarning{Type: tt.warningType, Message: tt.message}
			} else {
				warning = NewSecurityWarning(tt.warningType, tt.message, tt.args...)
			}

			if warning.Error() != tt.expectedOutput {
				t.Errorf("Expected warning output %q, got %q", tt.expectedOutput, warning.Error())
			}
		})
	}
}