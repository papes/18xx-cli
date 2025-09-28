package security

import "fmt"

// SecurityError represents a security-related error that should block execution
type SecurityError struct {
	Type    string
	Message string
}

// Error implements the error interface
func (e *SecurityError) Error() string {
	return fmt.Sprintf("security error [%s]: %s", e.Type, e.Message)
}

// IsSecurityError checks if an error is a SecurityError
func IsSecurityError(err error) bool {
	_, ok := err.(*SecurityError)
	return ok
}

// NewSecurityError creates a new security error
func NewSecurityError(errorType, message string, args ...interface{}) *SecurityError {
	return &SecurityError{
		Type:    errorType,
		Message: fmt.Sprintf(message, args...),
	}
}

// SecurityWarning represents a security warning that doesn't block execution but should be logged
type SecurityWarning struct {
	Type    string
	Message string
}

// Error implements the error interface
func (w *SecurityWarning) Error() string {
	return fmt.Sprintf("security warning [%s]: %s", w.Type, w.Message)
}

// IsSecurityWarning checks if an error is a SecurityWarning
func IsSecurityWarning(err error) bool {
	_, ok := err.(*SecurityWarning)
	return ok
}

// NewSecurityWarning creates a new security warning
func NewSecurityWarning(warningType, message string, args ...interface{}) *SecurityWarning {
	return &SecurityWarning{
		Type:    warningType,
		Message: fmt.Sprintf(message, args...),
	}
}