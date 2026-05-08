package aigen

import (
	"context"
	"testing"
)

func TestGenerateOffline_MultiStackPrompt(t *testing.T) {
	t.Setenv("YAML_ANCHOR_LLM", "off")

	pipeline, source, err := Generate(context.Background(), "Go API and React frontend with Docker image build", "")
	if err != nil {
		t.Fatalf("Generate returned error: %v", err)
	}
	if source != "offline" {
		t.Fatalf("source = %q, want offline", source)
	}
	if err := pipeline.Validate(); err != nil {
		t.Fatalf("generated pipeline is invalid: %v", err)
	}

	for _, id := range []string{"backend-test", "frontend-build", "docker-build"} {
		if _, ok := pipeline.Jobs[id]; !ok {
			t.Fatalf("missing generated job %q", id)
		}
	}
	if len(pipeline.Jobs["docker-build"].Needs) == 0 {
		t.Fatal("docker-build should depend on prior jobs")
	}
}

func TestInferFileType(t *testing.T) {
	tests := map[string]string{
		"React frontend with npm build": "package.json",
		"go test ./...":                 "go",
		"pytest for a flask app":        "python",
		"Dockerfile and docker build":   "dockerfile",
	}

	for input, want := range tests {
		if got := InferFileType(input); got != want {
			t.Fatalf("InferFileType(%q) = %q, want %q", input, got, want)
		}
	}
}
