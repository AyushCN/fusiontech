package config

import (
	"fmt"
	"os"
	"yaml-anchor/pkg/schema"
	"gopkg.in/yaml.v3"
)

// Load reads a YAML configuration file and returns a validated Pipeline
func Load(filepath string) (*schema.Pipeline, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %q: %w", filepath, err)
	}

	return ParseYAML(string(data))
}

// ParseYAML parses YAML string and returns a validated Pipeline
func ParseYAML(content string) (*schema.Pipeline, error) {
	var pipeline schema.Pipeline

	if err := yaml.Unmarshal([]byte(content), &pipeline); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	if err := pipeline.Validate(); err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
	}

	return &pipeline, nil
}

// Write saves a pipeline to a YAML file
func Write(pipeline *schema.Pipeline, filepath string) error {
	if err := pipeline.Validate(); err != nil {
		return err
	}

	data, err := yaml.Marshal(pipeline)
	if err != nil {
		return fmt.Errorf("failed to marshal to YAML: %w", err)
	}

	if err := os.WriteFile(filepath, data, 0644); err != nil {
		return fmt.Errorf("failed to write file %q: %w", filepath, err)
	}

	return nil
}
