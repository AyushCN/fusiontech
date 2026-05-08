// Package exporter defines the Platform interface for multi-CI export.
// Each exporter takes a validated YamlAnchor Pipeline and renders
// platform-specific CI configuration YAML.
package exporter

import "yaml-anchor/pkg/schema"

// Platform represents a CI/CD platform exporter.
type Platform interface {
	// Name returns the platform identifier (e.g. "github", "gitlab").
	Name() string

	// Export renders the pipeline as platform-specific YAML bytes.
	Export(pipeline *schema.Pipeline) ([]byte, error)

	// DefaultOutputPath returns the conventional file path for this platform.
	DefaultOutputPath() string
}
