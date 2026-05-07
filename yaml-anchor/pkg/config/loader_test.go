package config

import (
	"testing"
)

func TestParseYAML_ValidConfig(t *testing.T) {
	yaml := `
name: "Test Pipeline"
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - name: "Checkout"
        uses: "actions/checkout@v4"
`
	pipeline, err := ParseYAML(yaml)
	if err != nil {
		t.Fatalf("ParseYAML() unexpected error: %v", err)
	}
	if pipeline.Name != "Test Pipeline" {
		t.Errorf("Expected name 'Test Pipeline', got %q", pipeline.Name)
	}
	if _, ok := pipeline.Jobs["build"]; !ok {
		t.Error("Expected 'build' job to exist")
	}
}

func TestParseYAML_MissingName(t *testing.T) {
	yaml := `
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - name: "Step"
        run: "echo hello"
`
	_, err := ParseYAML(yaml)
	if err == nil {
		t.Error("Expected error for pipeline with no name, got nil")
	}
}

func TestParseYAML_CircularDependency(t *testing.T) {
	yaml := `
name: "Circular"
jobs:
  job-a:
    runs-on: ubuntu-latest
    needs: [job-b]
    steps:
      - name: "A"
        run: "echo a"
  job-b:
    runs-on: ubuntu-latest
    needs: [job-a]
    steps:
      - name: "B"
        run: "echo b"
`
	_, err := ParseYAML(yaml)
	if err == nil {
		t.Error("Expected circular dependency error, got nil")
	}
}

func TestParseYAML_MatrixExpansion(t *testing.T) {
	yaml := `
name: "Matrix Test"
jobs:
  test:
    runs-on: ubuntu-latest
    strategy:
      matrix:
        version: ["1", "2", "3"]
    steps:
      - name: "Run"
        run: "echo $version"
`
	pipeline, err := ParseYAML(yaml)
	if err != nil {
		t.Fatalf("ParseYAML() unexpected error: %v", err)
	}
	if len(pipeline.Jobs) != 3 {
		t.Errorf("Expected 3 matrix-expanded jobs, got %d", len(pipeline.Jobs))
	}
	if _, ok := pipeline.Jobs["test (1)"]; !ok {
		t.Error("Expected expanded job 'test (1)' to exist")
	}
}

func TestParseYAML_InvalidYAML(t *testing.T) {
	_, err := ParseYAML("not: valid: yaml: at: all: {{{")
	if err == nil {
		t.Error("Expected error for invalid YAML, got nil")
	}
}
