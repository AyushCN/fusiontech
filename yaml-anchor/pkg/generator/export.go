package generator

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
	"yaml-anchor/pkg/errors"
	"yaml-anchor/pkg/schema"
)

// ExportYAML converts a pipeline to YAML and writes it to a file
func ExportYAML(pipeline *schema.Pipeline, outputPath string) error {
	if err := pipeline.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Validate Pipeline thoroughly
	valErrs := ValidatePipeline(pipeline)
	if len(valErrs) > 0 {
		var msgs []string
		for _, e := range valErrs {
			msgs = append(msgs, e.Error())
		}
		return fmt.Errorf("validation failed:\n  - %s", strings.Join(msgs, "\n  - "))
	}

	// Scan for secrets before export
	secErrs := ScanForSecrets(pipeline)
	var blockErrors []string
	for _, err := range secErrs {
		if secErr, ok := err.(*errors.SecurityError); ok {
			if secErr.Severity == "HIGH" || secErr.Severity == "CRITICAL" {
				blockErrors = append(blockErrors, err.Error())
			}
		}
	}

	if len(blockErrors) > 0 {
		return fmt.Errorf("security scan failed - HIGH/CRITICAL issues blocked export:\n  - %s", strings.Join(blockErrors, "\n  - "))
	}

	// Convert to YAML
	data, err := yaml.Marshal(pipeline)
	if err != nil {
		return fmt.Errorf("failed to marshal YAML: %w", err)
	}

	// Create directory if needed
	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %q: %w", dir, err)
	}

	// Write file
	if err := os.WriteFile(outputPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// ScanForSecrets checks for hardcoded secrets in pipeline
func ScanForSecrets(pipeline *schema.Pipeline) []error {
	var issues []error

	// Stronger patterns to detect secrets
	// Avoid matching purely uppercase variables with no values, unless it's a known format like AKIA.
	patterns := map[string]struct {
		Regex    *regexp.Regexp
		Severity string
		Suggest  string
	}{
		"AWS_SECRET": {
			Regex:    regexp.MustCompile(`AKIA[0-9A-Z]{16}`),
			Severity: "CRITICAL",
			Suggest:  "Use injected secrets (e.g. ${{ secrets.AWS_ACCESS_KEY_ID }})",
		},
		"GITHUB_TOKEN": {
			Regex:    regexp.MustCompile(`(ghp|gho|ghu|ghs|ghr)_[A-Za-z0-9_]{36}`),
			Severity: "CRITICAL",
			Suggest:  "Use GitHub's automatic token: ${{ secrets.GITHUB_TOKEN }}",
		},
		"BEARER_TOKEN": {
			Regex:    regexp.MustCompile(`(?i)bearer\s+[A-Za-z0-9\-\._~\+/]{20,}=*`),
			Severity: "HIGH",
			Suggest:  "Store API tokens in environment secrets",
		},
		"PASSWORD_ASSIGNMENT": {
			Regex:    regexp.MustCompile(`(?i)(password|passwd|secret)\s*[:=]\s*['"]?[a-zA-Z0-9!@#\$%\^&\*\(\)_\+-=\[\]\{\};:,.<>/?]{8,}['"]?`),
			Severity: "HIGH",
			Suggest:  "Do not hardcode passwords in scripts",
		},
		"AZURE_TOKEN": {
			Regex:    regexp.MustCompile(`(?i)(eyJ[a-zA-Z0-9_-]{10,}\.eyJ[a-zA-Z0-9_-]{10,}\.[a-zA-Z0-9_-]{10,})`), // Basic JWT check
			Severity: "HIGH",
			Suggest:  "Extract JWTs and store securely",
		},
		"SLACK_TOKEN": {
			Regex:    regexp.MustCompile(`xox[baprs]-[0-9]{10,}-[a-zA-Z0-9]{20,}`),
			Severity: "CRITICAL",
			Suggest:  "Use Slack incoming webhooks or bot tokens via secrets",
		},
		"SSH_PRIVATE_KEY": {
			Regex:    regexp.MustCompile(`-----BEGIN (RSA|OPENSSH|DSA|EC|PGP) PRIVATE KEY-----`),
			Severity: "CRITICAL",
			Suggest:  "Never embed SSH private keys in CI files",
		},
	}

	for jobID, job := range pipeline.Jobs {
		for stepIdx, step := range job.Steps {
			if step.Run != "" {
				for secretType, pattern := range patterns {
					if pattern.Regex.MatchString(step.Run) {
						issues = append(issues, errors.NewSecurityError(
							pattern.Severity,
							fmt.Sprintf("Potential %s in job %q step %d 'run' block", secretType, jobID, stepIdx),
							pattern.Suggest,
						))
					}
				}
			}

			// Check environment variables
			for key, val := range step.Env {
				if val == "" {
					continue
				}
				for secretType, pattern := range patterns {
					if pattern.Regex.MatchString(key) || pattern.Regex.MatchString(val) {
						issues = append(issues, errors.NewSecurityError(
							pattern.Severity,
							fmt.Sprintf("Potential %s in job %q step %d env var %q", secretType, jobID, stepIdx, key),
							pattern.Suggest,
						))
					}
				}
			}
		}
	}

	return issues
}

// ValidatePipeline performs comprehensive validation
func ValidatePipeline(pipeline *schema.Pipeline) []error {
	var errs []error

	// Check name
	if pipeline.Name == "" {
		errs = append(errs, errors.NewValidationError("pipeline.name", "pipeline must have a name"))
	}

	// Check jobs
	if len(pipeline.Jobs) == 0 {
		errs = append(errs, errors.NewValidationError("pipeline.jobs", "pipeline must have at least one job"))
		return errs // Exit early
	}

	for jobID, job := range pipeline.Jobs {
		// Validate runner
		if job.RunsOn == "" {
			errs = append(errs, errors.NewValidationError(fmt.Sprintf("jobs.%s.runs-on", jobID), "missing runs-on"))
		}

		// Validate steps
		if len(job.Steps) == 0 {
			errs = append(errs, errors.NewValidationError(fmt.Sprintf("jobs.%s.steps", jobID), "has no steps"))
		}

		for stepIdx, step := range job.Steps {
			if step.Uses == "" && step.Run == "" {
				errs = append(errs, errors.NewValidationError(
					fmt.Sprintf("jobs.%s.steps[%d]", jobID, stepIdx),
					"missing 'uses' or 'run'",
				))
			}

			// Dangerous patterns
			if step.Run != "" {
				// Improved curl | bash detection
				dangerousCmds := []string{
					`curl.*\|.*bash`,
					`wget.*\|.*sh`,
					`bash.*<\(.*curl`,
					`bash.*<\(.*wget`,
					`curl.*\|.*sh`,
				}
				for _, pattern := range dangerousCmds {
					matched, _ := regexp.MatchString(pattern, step.Run)
					if matched {
						errs = append(errs, errors.NewValidationError(
							fmt.Sprintf("jobs.%s.steps[%d].run", jobID, stepIdx),
							"has dangerous remote script execution pattern (e.g. curl | bash)",
						))
					}
				}
			}
		}

		// Validate dependencies (circular dependency is handled by pipeline.Validate(), but we check for non-existent here)
		for _, need := range job.Needs {
			if _, exists := pipeline.Jobs[need]; !exists {
				errs = append(errs, errors.NewValidationError(
					fmt.Sprintf("jobs.%s.needs", jobID),
					fmt.Sprintf("depends on non-existent job %q", need),
				))
			}
		}
	}

	return errs
}
