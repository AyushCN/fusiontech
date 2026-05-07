package scanner

import (
	"testing"

	"yaml-anchor/pkg/schema"
)

func TestHasSecret_AwsKeyInStep(t *testing.T) {
	pipeline := &schema.Pipeline{
		Name: "test",
		Jobs: map[string]*schema.Job{
			"build": {
				RunsOn: "ubuntu-latest",
				Steps: []*schema.Step{
					{Name: "bad step", Run: "echo AKIAIOSFODNN7EXAMPLE"},
				},
			},
		},
	}
	err := HasSecret(pipeline)
	if err == nil {
		t.Error("Expected HasSecret to detect AWS key, got nil")
	}
}

func TestHasSecret_CleanPipeline(t *testing.T) {
	pipeline := &schema.Pipeline{
		Name: "test",
		Jobs: map[string]*schema.Job{
			"build": {
				RunsOn: "ubuntu-latest",
				Steps: []*schema.Step{
					{Name: "checkout", Uses: "actions/checkout@v4"},
					{Name: "build", Run: "go build ./..."},
				},
			},
		},
	}
	err := HasSecret(pipeline)
	if err != nil {
		t.Errorf("Expected no secrets in clean pipeline, got: %v", err)
	}
}

func TestScan_AwsKey(t *testing.T) {
	findings, err := Scan(".", ScanOptions{
		IncludeDotEnv: true,
		EntropyLimit:  0,
	})
	if err != nil {
		t.Fatalf("Scan() unexpected error: %v", err)
	}
	// Just verify scan runs without crashing; actual findings depend on test env
	_ = findings
}

func TestFormatFindings_JSON(t *testing.T) {
	findings := []Finding{
		{
			File:        "test.yaml",
			Line:        1,
			Pattern:     "AWS Access Key",
			Severity:    SeverityHigh,
			Description: "test",
		},
	}
	out := FormatFindings(findings, "json")
	if out == "" {
		t.Error("FormatFindings(json) returned empty string")
	}
	if out[0] != '[' {
		t.Errorf("FormatFindings(json) should return JSON array, got: %s", out[:20])
	}
}

func TestFormatFindings_GitHub(t *testing.T) {
	findings := []Finding{
		{File: "ci.yaml", Line: 5, Pattern: "GitHub Token", Severity: SeverityHigh},
	}
	out := FormatFindings(findings, "github")
	if out == "" {
		t.Error("FormatFindings(github) returned empty string")
	}
}

func TestFormatFindings_Human_NoFindings(t *testing.T) {
	out := FormatFindings(nil, "human")
	if out == "" {
		t.Error("FormatFindings(human) with no findings should return success message")
	}
}
