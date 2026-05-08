package blueprints

import (
	"fmt"
	"sort"
	"strings"

	"yaml-anchor/pkg/detector"
)

// MapToYAML returns a recommended anchor.yaml content based on the project profile.
func MapToYAML(profile *detector.ProjectProfile) string {
	name := "My Project"
	if profile.ModuleName != "" {
		name = profile.ModuleName
	} else if profile.Stack != "" {
		name = fmt.Sprintf("%s Pipeline", profile.Stack)
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf("name: %q\n\n", name))
	writeDetectionComments(&b, profile)
	b.WriteString("\non:\n")
	b.WriteString("  push:\n")
	b.WriteString("    branches: [main]\n")
	b.WriteString("  pull_request:\n")
	b.WriteString("    branches: [main]\n\n")
	b.WriteString("jobs:\n")

	wrote := false
	needs := []string{}

	if profile.HasGo || profile.Stack == "go" {
		wrote = true
		needs = append(needs, "backend-test")
		b.WriteString(`  backend-test:
    runs-on: "ubuntu-latest"
    steps:
      - name: "Checkout"
        uses: "actions/checkout@v4"
      - name: "Setup Go"
        uses: "actions/setup-go@v4"
      - name: "Download dependencies"
        run: "go mod download"
      - name: "Run tests"
        run: "go test ./..."
      - name: "Build"
        run: "go build ./..."
`)
	}

	if profile.HasNode || profile.Stack == "node" {
		wrote = true
		needs = append(needs, "frontend-build")
		b.WriteString(`  frontend-build:
    runs-on: "ubuntu-latest"
    steps:
      - name: "Checkout"
        uses: "actions/checkout@v4"
      - name: "Setup Node"
        uses: "actions/setup-node@v3"
      - name: "Install dependencies"
        run: "npm ci"
`)
		if hasScript(profile, "lint") {
			b.WriteString(`      - name: "Lint"
        run: "npm run lint"
`)
		}
		if hasScript(profile, "test") {
			b.WriteString(`      - name: "Test"
        run: "npm test"
`)
		}
		if hasScript(profile, "build") || len(profile.ScriptCommands) == 0 {
			b.WriteString(`      - name: "Build"
        run: "npm run build"
`)
		}
	}

	if profile.HasPython || profile.Stack == "python" {
		wrote = true
		needs = append(needs, "python-test")
		b.WriteString(`  python-test:
    runs-on: "ubuntu-latest"
    steps:
      - name: "Checkout"
        uses: "actions/checkout@v4"
      - name: "Setup Python"
        uses: "actions/setup-python@v4"
      - name: "Install dependencies"
        run: "pip install -r requirements.txt"
      - name: "Run tests"
        run: "pytest"
`)
	}

	if profile.HasRust || profile.Stack == "rust" {
		wrote = true
		needs = append(needs, "rust-test")
		b.WriteString(`  rust-test:
    runs-on: "ubuntu-latest"
    steps:
      - name: "Checkout"
        uses: "actions/checkout@v4"
      - name: "Build"
        run: "cargo build --locked"
      - name: "Test"
        run: "cargo test --locked"
`)
	}

	if profile.HasDocker {
		wrote = true
		b.WriteString(`  docker-build:
    runs-on: "ubuntu-latest"
`)
		if len(needs) > 0 {
			b.WriteString(fmt.Sprintf("    needs: [%s]\n", quoteList(needs)))
		}
		b.WriteString(`    steps:
      - name: "Checkout"
        uses: "actions/checkout@v4"
      - name: "Build Docker image"
        run: "docker build -t app:latest ."
`)
	}

	if !wrote {
		b.WriteString(`  main:
    runs-on: "ubuntu-latest"
    steps:
      - name: "Checkout"
        uses: "actions/checkout@v4"
      - name: "Add project command"
        run: "echo 'Add your build or test command here'"
`)
	}

	return b.String()
}

func writeDetectionComments(b *strings.Builder, profile *detector.ProjectProfile) {
	if len(profile.Stacks) > 0 {
		b.WriteString(fmt.Sprintf("# Detected stacks: %s\n", strings.Join(profile.Stacks, ", ")))
	} else if profile.Stack != "" {
		b.WriteString(fmt.Sprintf("# Detected stack: %s\n", profile.Stack))
	}
	if profile.Version != "" {
		b.WriteString(fmt.Sprintf("# Detected version: %s\n", profile.Version))
	}
	if profile.ModuleName != "" {
		b.WriteString(fmt.Sprintf("# Detected module: %s\n", profile.ModuleName))
	}
	if profile.Framework != "" {
		b.WriteString(fmt.Sprintf("# Detected framework: %s\n", profile.Framework))
	}
	if len(profile.InferredScripts) > 0 {
		b.WriteString(fmt.Sprintf("# Available scripts: %v\n", profile.InferredScripts))
	}
	if profile.HasDocker {
		b.WriteString("# Dockerfile detected\n")
	}
	if len(profile.ExistingCI) > 0 {
		b.WriteString(fmt.Sprintf("# Existing CI detected: %s\n", strings.Join(profile.ExistingCI, ", ")))
	}
}

func hasScript(profile *detector.ProjectProfile, name string) bool {
	if profile.ScriptCommands != nil {
		if _, ok := profile.ScriptCommands[name]; ok {
			return true
		}
	}
	for _, script := range profile.InferredScripts {
		if script == name {
			return true
		}
	}
	return false
}

func quoteList(items []string) string {
	sort.Strings(items)
	quoted := make([]string, 0, len(items))
	for _, item := range items {
		quoted = append(quoted, fmt.Sprintf("%q", item))
	}
	return strings.Join(quoted, ", ")
}
