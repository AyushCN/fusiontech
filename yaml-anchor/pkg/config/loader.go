// Package config provides utilities for loading pipeline definitions from
// an anchor.yaml configuration file into the schema.Pipeline IR.
package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
	"yaml-anchor/pkg/schema"
)

// Load reads an anchor.yaml file from the given path and unmarshals it
// into a *schema.Pipeline ready for generation or local execution.
func Load(configPath string) (*schema.Pipeline, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("could not read config file %q: %w", configPath, err)
	}

	var pipeline schema.Pipeline
	if err := yaml.Unmarshal(data, &pipeline); err != nil {
		return nil, fmt.Errorf("invalid anchor.yaml syntax: %w", err)
	}

	if err := validate(&pipeline); err != nil {
		return nil, err
	}

	return &pipeline, nil
}

// validate performs basic structural validation on the loaded pipeline.
func validate(p *schema.Pipeline) error {
	if p.Name == "" {
		return fmt.Errorf("pipeline must have a 'name' field")
	}
	if len(p.Jobs) == 0 {
		return fmt.Errorf("pipeline %q must define at least one job", p.Name)
	}
	for jobName, job := range p.Jobs {
		if job.RunsOn == "" {
			return fmt.Errorf("job %q must specify 'runs-on'", jobName)
		}
		if len(job.Steps) == 0 {
			return fmt.Errorf("job %q must have at least one step", jobName)
		}
		for i, step := range job.Steps {
			if step.Run == "" && step.Uses == "" {
				return fmt.Errorf("step %d in job %q must have either 'run' or 'uses'", i+1, jobName)
			}
		}
	}
	return nil
}
