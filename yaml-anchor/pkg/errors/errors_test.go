package errors

import (
	"strings"
	"testing"
)

func TestConfigError_WithPath(t *testing.T) {
	err := NewConfigError("anchor.yaml", "missing required field 'name'")
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	msg := err.Error()
	if !strings.Contains(msg, "anchor.yaml") {
		t.Errorf("Expected path in error message, got: %q", msg)
	}
	if !strings.Contains(msg, "missing required field") {
		t.Errorf("Expected message in error, got: %q", msg)
	}
}

func TestConfigError_WithoutPath(t *testing.T) {
	err := NewConfigError("", "file not found")
	msg := err.Error()
	if !strings.Contains(msg, "file not found") {
		t.Errorf("Expected message in error, got: %q", msg)
	}
}

func TestValidationError(t *testing.T) {
	err := NewValidationError("runs-on", "must specify a valid runner")
	msg := err.Error()
	if !strings.Contains(msg, "runs-on") {
		t.Errorf("Expected field name in error, got: %q", msg)
	}
	if !IsValidationError(err) {
		t.Error("IsValidationError() should return true")
	}
}

func TestSecurityError(t *testing.T) {
	err := NewSecurityError("HIGH", "Hardcoded AWS key detected", "Use ${{ secrets.AWS_KEY }} instead")
	msg := err.Error()
	if !strings.Contains(msg, "HIGH") {
		t.Errorf("Expected severity in error, got: %q", msg)
	}
	if !strings.Contains(msg, "AWS key") {
		t.Errorf("Expected message in error, got: %q", msg)
	}
	if !IsSecurityError(err) {
		t.Error("IsSecurityError() should return true")
	}
}

func TestIsConfigError(t *testing.T) {
	err := NewConfigError("path", "msg")
	if !IsConfigError(err) {
		t.Error("IsConfigError() should return true for ConfigError")
	}
	if IsConfigError(NewValidationError("f", "m")) {
		t.Error("IsConfigError() should return false for ValidationError")
	}
}
