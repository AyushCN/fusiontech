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

func TestParseYAML_MultiMatrixExpansion(t *testing.T) {
	yaml := `
name: "Multi Matrix Test"
jobs:
  test:
    runs-on: ubuntu-latest
    strategy:
      matrix:
        os: ["ubuntu-latest", "macos-latest"]
        version: ["1", "2"]
    steps:
      - name: "Run"
        run: "echo hello"
`
	pipeline, err := ParseYAML(yaml)
	if err != nil {
		t.Fatalf("ParseYAML() unexpected error: %v", err)
	}
	if len(pipeline.Jobs) != 4 {
		t.Errorf("Expected 4 matrix-expanded jobs, got %d", len(pipeline.Jobs))
	}
	expectedJobs := []string{
		"test (macos-latest, 1)",
		"test (macos-latest, 2)",
		"test (ubuntu-latest, 1)",
		"test (ubuntu-latest, 2)",
	}
	for _, expectedName := range expectedJobs {
		if _, ok := pipeline.Jobs[expectedName]; !ok {
			t.Errorf("Expected expanded job %q to exist", expectedName)
		}
	}
}

func TestParseYAML_InvalidYAML(t *testing.T) {
	_, err := ParseYAML("not: valid: yaml: at: all: {{{")
	if err == nil {
		t.Error("Expected error for invalid YAML, got nil")
	}
}

// --- Matrix include/exclude tests (ported from nektos/act semantics) ---

func TestGetMatrixes_ExcludeRemovesCombo(t *testing.T) {
	raw := map[string]interface{}{
		"os":      []interface{}{"ubuntu", "windows"},
		"version": []interface{}{"1", "2"},
		"exclude": []interface{}{
			map[string]interface{}{"os": "windows", "version": "1"},
		},
	}
	result, err := getMatrixes(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// 4 combos - 1 excluded = 3
	if len(result) != 3 {
		t.Errorf("expected 3 results after exclude, got %d: %v", len(result), result)
	}
	for _, r := range result {
		if r["os"] == "windows" && r["version"] == "1" {
			t.Error("excluded combo (windows, 1) should not be present")
		}
	}
}

func TestGetMatrixes_IncludeAddsNewCombo(t *testing.T) {
	raw := map[string]interface{}{
		"os": []interface{}{"ubuntu"},
		"include": []interface{}{
			map[string]interface{}{"os": "macos", "extra": "value"},
		},
	}
	result, err := getMatrixes(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// 1 original + 1 extra include = 2
	if len(result) != 2 {
		t.Errorf("expected 2 results with include, got %d: %v", len(result), result)
	}
}

func TestGetMatrixes_IncludeMergesIntoExisting(t *testing.T) {
	raw := map[string]interface{}{
		"os": []interface{}{"ubuntu"},
		"include": []interface{}{
			map[string]interface{}{"os": "ubuntu", "color": "blue"},
		},
	}
	result, err := getMatrixes(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Still 1 combo, but with "color" merged in
	if len(result) != 1 {
		t.Errorf("expected 1 result (merged), got %d: %v", len(result), result)
	}
	if result[0]["color"] != "blue" {
		t.Errorf("expected 'color' to be merged into combo, got: %v", result[0])
	}
}

func TestGetMatrixes_ExcludeUnknownKey_ReturnsError(t *testing.T) {
	raw := map[string]interface{}{
		"os": []interface{}{"ubuntu"},
		"exclude": []interface{}{
			map[string]interface{}{"nonexistent": "value"},
		},
	}
	_, err := getMatrixes(raw)
	if err == nil {
		t.Error("expected error for exclude with unknown key, got nil")
	}
}

func TestGetMatrixes_EmptyMatrix_ReturnsOneEmptyCombo(t *testing.T) {
	raw := map[string]interface{}{}
	result, err := getMatrixes(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Errorf("expected 1 empty combo for empty matrix, got %d", len(result))
	}
}
