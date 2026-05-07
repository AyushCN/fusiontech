package scanner

import (
	"fmt"

	"yaml-anchor/pkg/schema"
)

// HasSecret scans the pipeline steps for hardcoded secrets and returns an error if found.
// This is used during 'anchor generate' to catch secrets in the IR before YAML export.
func HasSecret(pipeline *schema.Pipeline) error {
	for jobName, job := range pipeline.Jobs {
		for i, step := range job.Steps {
			// Check Run field
			for name, pattern := range defaultPatterns {
				if pattern.MatchString(step.Run) {
					return fmt.Errorf("SECURITY RISK: Detected hardcoded %s in Job '%s' (Step %d, 'run' field). Please use injected secrets instead", name, jobName, i+1)
				}
				if pattern.MatchString(step.Uses) {
					return fmt.Errorf("SECURITY RISK: Detected hardcoded %s in Job '%s' (Step %d, 'uses' field). Please use injected secrets instead", name, jobName, i+1)
				}
			}
			
			// Optional: We could also run entropy check on the IR here if desired, 
			// but regex is usually enough for the structured IR.
		}
	}
	return nil
}
