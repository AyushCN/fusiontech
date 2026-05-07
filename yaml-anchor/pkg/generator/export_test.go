package generator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"yaml-anchor/pkg/schema"
)

func TestExportYAML_ValidPipeline(t *testing.T) {
	pipeline := &schema.Pipeline{
		Name: "Test Pipeline",
		On:   map[string]interface{}{"push": map[string]interface{}{"branches": []string{"main"}}},
		Jobs: map[string]*schema.Job{
			"build": {
				RunsOn: "ubuntu-latest",
				Steps: []*schema.Step{
					{Name: "Checkout", Uses: "actions/checkout@v4"},
				},
			},
		},
	}

	dir := t.TempDir()
	outPath := filepath.Join(dir, ".github", "workflows", "main.yml")

	err := ExportYAML(pipeline, outPath)
	if err != nil {
		t.Fatalf("ExportYAML() unexpected error: %v", err)
	}

	data, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("Could not read output file: %v", err)
	}
	if !strings.Contains(string(data), "Test Pipeline") {
		t.Error("Output YAML missing pipeline name")
	}
}

func TestExportYAML_MissingName_Error(t *testing.T) {
	pipeline := &schema.Pipeline{
		Jobs: map[string]*schema.Job{
			"build": {
				RunsOn: "ubuntu-latest",
				Steps:  []*schema.Step{{Name: "x", Run: "echo hi"}},
			},
		},
	}
	err := ExportYAML(pipeline, "/tmp/test-out.yml")
	if err == nil {
		t.Error("Expected validation error for pipeline with no name, got nil")
	}
}

func TestScanForSecrets_Clean(t *testing.T) {
	pipeline := &schema.Pipeline{
		Name: "Clean Pipeline",
		Jobs: map[string]*schema.Job{
			"build": {
				RunsOn: "ubuntu-latest",
				Steps:  []*schema.Step{{Name: "test", Run: "go test ./..."}},
			},
		},
	}
	issues := ScanForSecrets(pipeline)
	if len(issues) != 0 {
		t.Errorf("Expected 0 issues in clean pipeline, got %d: %v", len(issues), issues)
	}
}

func TestScanForSecrets_WithAwsKey(t *testing.T) {
	pipeline := &schema.Pipeline{
		Name: "Insecure Pipeline",
		Jobs: map[string]*schema.Job{
			"deploy": {
				RunsOn: "ubuntu-latest",
				Steps: []*schema.Step{
					{Name: "bad", Run: "aws configure set key AKIAIOSFODNN7EXAMPLE"},
				},
			},
		},
	}
	issues := ScanForSecrets(pipeline)
	if len(issues) == 0 {
		t.Error("Expected security issues for pipeline with hardcoded AWS key")
	}
}
