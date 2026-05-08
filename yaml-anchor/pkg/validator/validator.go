package validator

import (
	"fmt"
	"regexp"
	"strings"
)

// ValidationResult holds the outcome of a validation run.
type ValidationResult struct {
	Valid  bool
	Errors []string
}

func (r *ValidationResult) AddError(msg string) {
	r.Valid = false
	r.Errors = append(r.Errors, msg)
}

func (r *ValidationResult) Error() string {
	return strings.Join(r.Errors, "; ")
}

// jobIDRegex mirrors nektos/act's job name validation rule exactly.
// Job names must start with a letter or underscore, and contain only
// alphanumeric characters, hyphens, or underscores.
var jobIDRegex = regexp.MustCompile(`^([[:alpha:]_][[:alnum:]_\-]*)$`)

// ValidateJobID checks that a job ID conforms to GitHub Actions naming rules.
func ValidateJobID(id string) error {
	if id == "" {
		return fmt.Errorf("job ID cannot be empty")
	}
	if !jobIDRegex.MatchString(id) {
		return fmt.Errorf(
			"job ID %q is invalid: names must start with a letter or '_' and contain only alphanumeric characters, '-', or '_'",
			id,
		)
	}
	return nil
}


// ValidateStepName checks that a step name is valid.
func ValidateStepName(name string) error {
	if name == "" {
		return fmt.Errorf("step name cannot be empty")
	}
	if len(name) > 100 {
		return fmt.Errorf("step name %q too long (max 100 characters)", name)
	}
	return nil
}

// ValidateRunsOn checks that a runs-on value is a recognized GitHub-hosted or self-hosted runner.
func ValidateRunsOn(runsOn string) error {
	validRunners := map[string]bool{
		"ubuntu-latest":  true,
		"ubuntu-22.04":  true,
		"ubuntu-20.04":  true,
		"windows-latest": true,
		"windows-2022":  true,
		"windows-2019":  true,
		"macos-latest":  true,
		"macos-13":      true,
		"macos-12":      true,
		"self-hosted":   true,
	}
	if !validRunners[runsOn] && !strings.HasPrefix(runsOn, "self-hosted") {
		return fmt.Errorf("unrecognized runner %q — use a GitHub-hosted runner or prefix with 'self-hosted'", runsOn)
	}
	return nil
}

// ValidateCron checks that a cron expression has the correct 5-field format.
func ValidateCron(cron string) error {
	parts := strings.Fields(cron)
	if len(parts) != 5 {
		return fmt.Errorf("invalid cron expression %q (must have exactly 5 space-separated fields)", cron)
	}
	return nil
}

// ValidatePipelineName checks that a pipeline name is non-empty and not excessively long.
func ValidatePipelineName(name string) error {
	if name == "" {
		return fmt.Errorf("pipeline name cannot be empty")
	}
	if len(name) > 100 {
		return fmt.Errorf("pipeline name too long (max 100 characters)")
	}
	return nil
}
