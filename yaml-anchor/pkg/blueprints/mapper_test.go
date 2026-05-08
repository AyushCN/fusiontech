package blueprints

import (
	"strings"
	"testing"

	"yaml-anchor/pkg/detector"
)

func TestMapToYAML_GoStack(t *testing.T) {
	profile := &detector.ProjectProfile{
		Stack:   "go",
		Version: "1.21",
	}
	result := MapToYAML(profile)

	if !strings.Contains(result, `name: "go Pipeline"`) {
		t.Errorf("Expected Go pipeline name, got:\n%s", result)
	}
	if !strings.Contains(result, `backend-test:`) || !strings.Contains(result, `go test ./...`) {
		t.Errorf("Expected Go backend test job, got:\n%s", result)
	}
	if !strings.Contains(result, "# Detected version: 1.21") {
		t.Errorf("Expected version comment, got:\n%s", result)
	}
}

func TestMapToYAML_NodeStack(t *testing.T) {
	profile := &detector.ProjectProfile{
		Stack:     "node",
		Version:   "18",
		Framework: "next",
	}
	result := MapToYAML(profile)

	if !strings.Contains(result, `name: "node Pipeline"`) {
		t.Errorf("Expected Node pipeline name, got:\n%s", result)
	}
	if !strings.Contains(result, `frontend-build:`) || !strings.Contains(result, `npm run build`) {
		t.Errorf("Expected Node frontend build job, got:\n%s", result)
	}
	if !strings.Contains(result, "# Detected framework: next") {
		t.Errorf("Expected framework comment, got:\n%s", result)
	}
}

func TestMapToYAML_UnknownStack_FallsBackToGeneric(t *testing.T) {
	profile := &detector.ProjectProfile{
		Stack: "rust",
	}
	result := MapToYAML(profile)

	if !strings.Contains(result, `runs-on: "ubuntu-latest"`) {
		t.Errorf("Expected generic fallback with runs-on, got:\n%s", result)
	}
	if !strings.Contains(result, "actions/checkout@v4") {
		t.Errorf("Expected checkout step in generic fallback, got:\n%s", result)
	}
}

func TestMapToYAML_WithModuleName(t *testing.T) {
	profile := &detector.ProjectProfile{
		Stack:      "go",
		ModuleName: "github.com/myorg/myapp",
	}
	result := MapToYAML(profile)

	if !strings.Contains(result, "# Detected module: github.com/myorg/myapp") {
		t.Errorf("Expected module name comment, got:\n%s", result)
	}
}

func TestMapToYAML_WithInferredScripts(t *testing.T) {
	profile := &detector.ProjectProfile{
		Stack:           "node",
		InferredScripts: []string{"build", "test", "lint"},
	}
	result := MapToYAML(profile)

	if !strings.Contains(result, "# Available scripts:") {
		t.Errorf("Expected scripts comment, got:\n%s", result)
	}
}

func TestMapToYAML_EmptyProfile_UsesDefaults(t *testing.T) {
	profile := &detector.ProjectProfile{}
	result := MapToYAML(profile)

	if !strings.Contains(result, `name: "My Project"`) {
		t.Errorf("Expected default 'My Project' name, got:\n%s", result)
	}
}
