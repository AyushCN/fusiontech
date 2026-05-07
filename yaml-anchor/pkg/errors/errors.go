package errors

import "fmt"

// ConfigError is returned when loading or parsing anchor.yaml fails.
type ConfigError struct {
	Message string
	Path    string
}

// ValidationError is returned when a pipeline field fails validation.
type ValidationError struct {
	Field   string
	Message string
}

// SecurityError is returned when a secret or dangerous pattern is detected.
type SecurityError struct {
	Severity   string // "low", "medium", "high", "critical"
	Message    string
	Suggestion string
}

func (e *ConfigError) Error() string {
	if e.Path != "" {
		return fmt.Sprintf("config error in %q: %s", e.Path, e.Message)
	}
	return fmt.Sprintf("config error: %s", e.Message)
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error in field %q: %s", e.Field, e.Message)
}

func (e *SecurityError) Error() string {
	return fmt.Sprintf("[%s] %s — Fix: %s", e.Severity, e.Message, e.Suggestion)
}

// Constructor functions

func NewConfigError(path, msg string) error {
	return &ConfigError{Message: msg, Path: path}
}

func NewValidationError(field, msg string) error {
	return &ValidationError{Field: field, Message: msg}
}

func NewSecurityError(severity, msg, suggestion string) error {
	return &SecurityError{Severity: severity, Message: msg, Suggestion: suggestion}
}

// IsConfigError returns true if the error is a ConfigError.
func IsConfigError(err error) bool {
	_, ok := err.(*ConfigError)
	return ok
}

// IsValidationError returns true if the error is a ValidationError.
func IsValidationError(err error) bool {
	_, ok := err.(*ValidationError)
	return ok
}

// IsSecurityError returns true if the error is a SecurityError.
func IsSecurityError(err error) bool {
	_, ok := err.(*SecurityError)
	return ok
}
