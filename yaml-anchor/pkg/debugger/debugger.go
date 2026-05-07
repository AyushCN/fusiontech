package debugger

import (
	"fmt"
	"strings"
)

// Suggestion represents a fix suggestion from the debugger engine.
type Suggestion struct {
	Severity    string // "error", "warning", "info"
	Title       string
	Description string
	Fix         string
}

// Analyze takes a step name, a run command, and the raw error string and returns actionable suggestions.
func Analyze(stepName, runCmd, errMsg string) []Suggestion {
	var suggestions []Suggestion
	lower := strings.ToLower(errMsg)

	// --- Go-specific patterns ---
	if strings.Contains(lower, "cannot find package") || strings.Contains(lower, "no required module provides") {
		suggestions = append(suggestions, Suggestion{
			Severity:    "error",
			Title:       "Missing Go Dependency",
			Description: fmt.Sprintf("Step '%s' failed because a required Go package could not be found.", stepName),
			Fix:         "Run `go mod tidy` to resolve missing dependencies, then commit go.mod and go.sum.",
		})
	}

	if strings.Contains(lower, "build constraints exclude all go files") {
		suggestions = append(suggestions, Suggestion{
			Severity:    "error",
			Title:       "Build Constraint Mismatch",
			Description: "No Go files match the current build constraints (OS/arch).",
			Fix:         "Check your `//go:build` tags or GOOS/GOARCH environment variables.",
		})
	}

	// --- Node-specific patterns ---
	if strings.Contains(lower, "cannot find module") || strings.Contains(lower, "module not found") {
		suggestions = append(suggestions, Suggestion{
			Severity:    "error",
			Title:       "Missing Node Module",
			Description: fmt.Sprintf("Step '%s' failed because a required Node module was not found.", stepName),
			Fix:         "Add `npm install` or `npm ci` as a step before this one in your anchor.yaml.",
		})
	}

	if strings.Contains(lower, "npm err") || strings.Contains(lower, "npm error") {
		suggestions = append(suggestions, Suggestion{
			Severity:    "error",
			Title:       "npm Error",
			Description: "An npm command failed. This is often caused by a lockfile mismatch.",
			Fix:         "Try using `npm ci` instead of `npm install` for reproducible installs.",
		})
	}

	// --- Shell / general patterns ---
	if strings.Contains(lower, "permission denied") {
		suggestions = append(suggestions, Suggestion{
			Severity:    "error",
			Title:       "Permission Denied",
			Description: fmt.Sprintf("The command `%s` was denied execution permissions.", runCmd),
			Fix:         "Add `chmod +x <script>` as a prior step, or check that you are not attempting to write to a read-only path.",
		})
	}

	if strings.Contains(lower, "command not found") {
		// Extract the command name if possible
		cmd := runCmd
		if parts := strings.Fields(runCmd); len(parts) > 0 {
			cmd = parts[0]
		}
		suggestions = append(suggestions, Suggestion{
			Severity:    "error",
			Title:       "Command Not Found",
			Description: fmt.Sprintf("The command `%s` is not available in this container image.", cmd),
			Fix:         fmt.Sprintf("Add an installation step before this one: e.g., `apt-get install -y %s` or use a Docker image that includes it.", cmd),
		})
	}

	if strings.Contains(lower, "no space left on device") {
		suggestions = append(suggestions, Suggestion{
			Severity:    "error",
			Title:       "Disk Full",
			Description: "The container ran out of disk space.",
			Fix:         "Run `anchor clean` to prune old Dagger/Docker caches, then try again.",
		})
	}

	if strings.Contains(lower, "exit status 137") || strings.Contains(lower, "oom") {
		suggestions = append(suggestions, Suggestion{
			Severity:    "error",
			Title:       "Out of Memory (OOM Kill)",
			Description: "The container was killed because it ran out of memory.",
			Fix:         "Increase Docker Desktop's memory limit in Settings → Resources.",
		})
	}

	// --- Generic fallback ---
	if len(suggestions) == 0 {
		suggestions = append(suggestions, Suggestion{
			Severity:    "info",
			Title:       "Unrecognized Error",
			Description: fmt.Sprintf("Step '%s' failed with an unrecognized error.", stepName),
			Fix:         "Use `anchor exec <job>` to drop into an interactive shell and reproduce the failure manually.",
		})
	}

	return suggestions
}

// Format returns a human-readable string representation of suggestions.
func Format(suggestions []Suggestion) string {
	var sb strings.Builder
	sb.WriteString("\n🧠 AI Debugger Suggestions:\n")
	sb.WriteString(strings.Repeat("─", 40) + "\n")
	for i, s := range suggestions {
		icon := "ℹ️"
		if s.Severity == "error" {
			icon = "❌"
		} else if s.Severity == "warning" {
			icon = "⚠️"
		}
		sb.WriteString(fmt.Sprintf("%d. %s %s\n", i+1, icon, s.Title))
		sb.WriteString(fmt.Sprintf("   %s\n", s.Description))
		sb.WriteString(fmt.Sprintf("   💡 Fix: %s\n\n", s.Fix))
	}
	return sb.String()
}
