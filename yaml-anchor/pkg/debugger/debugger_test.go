package debugger

import (
	"strings"
	"testing"
)

func TestAnalyze_MissingGoDep(t *testing.T) {
	suggestions := Analyze("build", "go build ./...", "cannot find package \"github.com/foo/bar\"")
	if len(suggestions) == 0 {
		t.Fatal("Expected at least one suggestion for missing Go dependency")
	}
	if suggestions[0].Severity != "error" {
		t.Errorf("Expected severity 'error', got %q", suggestions[0].Severity)
	}
	if !strings.Contains(suggestions[0].Fix, "go mod tidy") {
		t.Errorf("Expected fix to mention 'go mod tidy', got: %q", suggestions[0].Fix)
	}
}

func TestAnalyze_NodeModuleNotFound(t *testing.T) {
	suggestions := Analyze("test", "npm test", "Cannot find module 'express'")
	if len(suggestions) == 0 {
		t.Fatal("Expected at least one suggestion for missing Node module")
	}
	if !strings.Contains(suggestions[0].Fix, "npm install") {
		t.Errorf("Expected fix to mention npm install, got: %q", suggestions[0].Fix)
	}
}

func TestAnalyze_PermissionDenied(t *testing.T) {
	suggestions := Analyze("deploy", "./deploy.sh", "permission denied: ./deploy.sh")
	if len(suggestions) == 0 {
		t.Fatal("Expected at least one suggestion for permission denied")
	}
	if !strings.Contains(suggestions[0].Fix, "chmod") {
		t.Errorf("Expected fix to mention chmod, got: %q", suggestions[0].Fix)
	}
}

func TestAnalyze_CommandNotFound(t *testing.T) {
	suggestions := Analyze("lint", "golangci-lint run", "golangci-lint: command not found")
	if len(suggestions) == 0 {
		t.Fatal("Expected at least one suggestion for command not found")
	}
	if suggestions[0].Title != "Command Not Found" {
		t.Errorf("Expected title 'Command Not Found', got %q", suggestions[0].Title)
	}
}

func TestAnalyze_OOMKill(t *testing.T) {
	suggestions := Analyze("build", "cargo build --release", "exit status 137")
	if len(suggestions) == 0 {
		t.Fatal("Expected at least one suggestion for OOM kill")
	}
	if !strings.Contains(strings.ToLower(suggestions[0].Title), "memory") {
		t.Errorf("Expected OOM title, got %q", suggestions[0].Title)
	}
}

func TestAnalyze_UnknownError_Fallback(t *testing.T) {
	suggestions := Analyze("mystery", "some-command", "an entirely unknown exotic failure")
	if len(suggestions) == 0 {
		t.Fatal("Expected fallback suggestion for unknown error")
	}
	if suggestions[0].Severity != "info" {
		t.Errorf("Expected fallback severity 'info', got %q", suggestions[0].Severity)
	}
	if !strings.Contains(suggestions[0].Fix, "anchor exec") {
		t.Errorf("Expected fix to mention anchor exec, got: %q", suggestions[0].Fix)
	}
}

func TestFormat(t *testing.T) {
	suggestions := []Suggestion{
		{Severity: "error", Title: "Test Error", Description: "Something broke.", Fix: "Fix it."},
	}
	output := Format(suggestions)
	if !strings.Contains(output, "Test Error") {
		t.Error("Format() output missing suggestion title")
	}
	if !strings.Contains(output, "Fix it.") {
		t.Error("Format() output missing fix text")
	}
	if !strings.Contains(output, "AI Debugger") {
		t.Error("Format() output missing header")
	}
}
