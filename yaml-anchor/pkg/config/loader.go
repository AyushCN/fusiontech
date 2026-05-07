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

	if err := expandBlueprints(&pipeline); err != nil {
		return nil, err
	}

	if err := validate(&pipeline); err != nil {
		return nil, err
	}

	return &pipeline, nil
}

// expandBlueprints resolves high-level blueprints into actual steps
func expandBlueprints(p *schema.Pipeline) error {
	for jobName, job := range p.Jobs {
		if job.Blueprint != "" {
			switch job.Blueprint {
			case "go-app":
				job.RunsOn = "ubuntu-latest"
				job.Steps = []schema.Step{
					{Name: "Checkout Code", Uses: "actions/checkout@v4"},
					{Name: "Setup Go", Uses: "actions/setup-go@v4", Env: map[string]string{"go-version": "1.21"}},
					{Name: "Go Build", Run: "go build ./..."},
					{Name: "Go Test", Run: "go test ./..."},
				}
			case "node-app":
				job.RunsOn = "ubuntu-latest"
				job.Steps = []schema.Step{
					{Name: "Checkout Code", Uses: "actions/checkout@v4"},
					{Name: "Setup Node", Uses: "actions/setup-node@v3", Env: map[string]string{"node-version": "18"}},
					{Name: "NPM Install", Run: "npm ci"},
					{Name: "NPM Test", Run: "npm test"},
				}
			default:
				return fmt.Errorf("job %q references unknown blueprint: %q", jobName, job.Blueprint)
			}
			// Update the job back in the map
			p.Jobs[jobName] = job
		}
	}
	return nil
}

// validate performs basic structural and DAG validation on the loaded pipeline.
func validate(p *schema.Pipeline) error {
	if p.Name == "" {
		return fmt.Errorf("pipeline must have a 'name' field")
	}
	if len(p.Jobs) == 0 {
		return fmt.Errorf("pipeline %q must define at least one job", p.Name)
	}

	// 1. Structural Validation
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

	// 2. DAG Validation (Cycle Detection and Dependency Checks)
	visited := make(map[string]bool)
	recStack := make(map[string]bool)

	var checkCycle func(string) error
	checkCycle = func(jobName string) error {
		if recStack[jobName] {
			return fmt.Errorf("circular dependency detected involving job: %s", jobName)
		}
		if visited[jobName] {
			return nil
		}
		visited[jobName] = true
		recStack[jobName] = true

		job := p.Jobs[jobName]
		for _, dep := range job.Needs {
			if _, exists := p.Jobs[dep]; !exists {
				return fmt.Errorf("job %q needs non-existent job %q", jobName, dep)
			}
			if err := checkCycle(dep); err != nil {
				return err
			}
		}
		recStack[jobName] = false
		return nil
	}

	for jobName := range p.Jobs {
		if !visited[jobName] {
			if err := checkCycle(jobName); err != nil {
				return err
			}
		}
	}

	return nil
}
