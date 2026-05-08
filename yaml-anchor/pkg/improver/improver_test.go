package improver

import (
	"context"
	"strings"
	"testing"

	"yaml-anchor/pkg/config"
)

func TestImproveYAML_ProducesValidPipeline(t *testing.T) {
	t.Setenv("YAML_ANCHOR_LLM", "off")

	current := `
name: broken
jobs:
  test:
    runs-on: ubuntu-latest
    needs: [missing]
    steps:
      - run: go test ./...
`
	improved, source, err := ImproveYAML(context.Background(), current, "job test depends on missing job missing")
	if err != nil {
		t.Fatalf("ImproveYAML returned error: %v", err)
	}
	if source == "" {
		t.Fatal("source should be reported")
	}
	if _, err := config.ParseYAML(improved); err != nil {
		t.Fatalf("improved YAML should parse and validate: %v\n%s", err, improved)
	}
}

func TestImproveYAML_RepairsMissingDependency(t *testing.T) {
	current := `
name: broken
jobs:
  test:
    runs-on: ubuntu-latest
    needs: [missing]
    steps:
      - run: go test ./...
`
	improved, source, err := ImproveYAML(context.Background(), current, "job test depends on missing job missing")
	if err != nil {
		t.Fatalf("ImproveYAML returned error: %v", err)
	}
	if source != "repair" {
		t.Fatalf("source = %q, want repair", source)
	}
	pipeline, err := config.ParseYAML(improved)
	if err != nil {
		t.Fatalf("improved YAML should parse and validate: %v\n%s", err, improved)
	}
	if got := pipeline.Jobs["test"].Needs; len(got) != 0 {
		t.Fatalf("needs = %v, want missing dependency removed", got)
	}
}

func TestImproveYAML_AddsSetupSteps(t *testing.T) {
	current := `
name: missing setup
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - run: npm ci
      - run: npm run build
`
	improved, source, err := ImproveYAML(context.Background(), current, "npm: command not found")
	if err != nil {
		t.Fatalf("ImproveYAML returned error: %v", err)
	}
	if source != "repair" {
		t.Fatalf("source = %q, want repair", source)
	}
	if _, err := config.ParseYAML(improved); err != nil {
		t.Fatalf("improved YAML should parse and validate: %v\n%s", err, improved)
	}
	assertContains(t, improved, "actions/checkout@v4")
	assertContains(t, improved, "actions/setup-node@v3")
}

func TestImproveYAML_RepairsNPMCIWithoutLockfile(t *testing.T) {
	current := `
name: npm lock
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v3
      - run: npm ci
`
	improved, source, err := ImproveYAML(context.Background(), current, "npm ci can only install packages with an existing package-lock.json")
	if err != nil {
		t.Fatalf("ImproveYAML returned error: %v", err)
	}
	if source != "repair" {
		t.Fatalf("source = %q, want repair", source)
	}
	assertContains(t, improved, "npm install")
}

func assertContains(t *testing.T, haystack, needle string) {
	t.Helper()
	if !strings.Contains(haystack, needle) {
		t.Fatalf("expected %q in:\n%s", needle, haystack)
	}
}
