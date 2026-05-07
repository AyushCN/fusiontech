package generator

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"yaml-anchor/pkg/schema"
	"gopkg.in/yaml.v3"
)

// ExportYAML converts a pipeline to YAML and writes it to a file
func ExportYAML(pipeline *schema.Pipeline, outputPath string) error {
	if err := pipeline.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Scan for secrets before export
	if issues := ScanForSecrets(pipeline); len(issues) > 0 {
		return fmt.Errorf("security scan failed - found potential secrets:\n%v", issues)
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
func ScanForSecrets(pipeline *schema.Pipeline) []string {
	var issues []string

	// Patterns to detect
	patterns := map[string]*regexp.Regexp{
		"AWS_SECRET": regexp.MustCompile(`(?i)aws_secret|aws.*key|aws.*secret`),
		"GITHUB_TOKEN": regexp.MustCompile(`(?i)github.*token|gh_token|github_token`),
		"BEARER_TOKEN": regexp.MustCompile(`(?i)bearer\s+[a-z0-9]{20,}`),
		"PASSWORD": regexp.MustCompile(`(?i)password\s*=\s*['\"]*[a-z0-9]{8,}`),
	}

	for jobID, job := range pipeline.Jobs {
		for stepIdx, step := range job.Steps {
			if step.Run != "" {
				for secretType, pattern := range patterns {
					if pattern.MatchString(step.Run) {
						issues = append(issues, fmt.Sprintf(
							"[%s] Potential %s in job %q step %d",
							jobID, secretType, jobID, stepIdx,
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
					if pattern.MatchString(key) || pattern.MatchString(val) {
						issues = append(issues, fmt.Sprintf(
							"[%s] Potential %s in env var %q",
							jobID, secretType, key,
						))
					}
				}
			}
		}
	}

	return issues
}

// ValidatePipeline performs comprehensive validation
func ValidatePipeline(pipeline *schema.Pipeline) []string {
	var errors []string

	// Check name
	if pipeline.Name == "" {
		errors = append(errors, "pipeline must have a name")
	}

	// Check jobs
	if len(pipeline.Jobs) == 0 {
		errors = append(errors, "pipeline must have at least one job")
		return errors // Exit early
	}

	for jobID, job := range pipeline.Jobs {
		// Validate runner
		if job.RunsOn == "" {
			errors = append(errors, fmt.Sprintf("job %q missing runs-on", jobID))
		}

		// Validate steps
		if len(job.Steps) == 0 {
			errors = append(errors, fmt.Sprintf("job %q has no steps", jobID))
		}

		for stepIdx, step := range job.Steps {
			if step.Uses == "" && step.Run == "" {
				errors = append(errors, fmt.Sprintf(
					"job %q step %d missing 'uses' or 'run'",
					jobID, stepIdx,
				))
			}

			// Dangerous patterns
			if strings.Contains(step.Run, "curl") && strings.Contains(step.Run, "| bash") {
				errors = append(errors, fmt.Sprintf(
					"job %q step %d has dangerous curl | bash pattern",
					jobID, stepIdx,
				))
			}
		}

		// Validate dependencies
		for _, need := range job.Needs {
			if _, exists := pipeline.Jobs[need]; !exists {
				errors = append(errors, fmt.Sprintf(
					"job %q depends on non-existent job %q",
					jobID, need,
				))
			}
		}
	}

	return errors
}
