package scanner

import (
	"fmt"
	"regexp"

	"yaml-anchor/pkg/schema"
)

// Common secret patterns
var secretPatterns = map[string]*regexp.Regexp{
	"AWS Access Key": regexp.MustCompile(`AKIA[0-9A-Z]{16}`),
	"GitHub Token":   regexp.MustCompile(`(ghp|gho|ghu|ghs|ghr)_[A-Za-z0-9_]{36}`),
	"Bearer Token":   regexp.MustCompile(`Bearer [a-zA-Z0-9\-\._~+/]+=*`),
}

// HasSecret scans the pipeline steps for hardcoded secrets and returns an error if found.
func HasSecret(pipeline *schema.Pipeline) error {
	for jobName, job := range pipeline.Jobs {
		for i, step := range job.Steps {
			for secretType, pattern := range secretPatterns {
				if pattern.MatchString(step.Run) || pattern.MatchString(step.Uses) {
					return fmt.Errorf("SECURITY RISK: Detected hardcoded %s in Job '%s' (Step %d). Please use injected secrets instead", secretType, jobName, i+1)
				}
			}
		}
	}
	return nil
}
