package detector

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetect_GoProject(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module myapp\n\ngo 1.21\n"), 0644)

	profile, err := Detect(dir)
	if err != nil {
		t.Fatalf("Detect() error: %v", err)
	}
	if profile.Stack != "go" {
		t.Errorf("Expected stack 'go', got %q", profile.Stack)
	}
	if profile.Version != "1.21" {
		t.Errorf("Expected version '1.21', got %q", profile.Version)
	}
	if profile.ModuleName != "myapp" {
		t.Errorf("Expected module 'myapp', got %q", profile.ModuleName)
	}
}

func TestDetect_NodeProject(t *testing.T) {
	dir := t.TempDir()
	pkgJSON := `{
		"name": "my-app",
		"engines": {"node": ">=18"},
		"scripts": {"test": "jest", "build": "vite build"},
		"dependencies": {"react": "^18.0.0"}
	}`
	os.WriteFile(filepath.Join(dir, "package.json"), []byte(pkgJSON), 0644)

	profile, err := Detect(dir)
	if err != nil {
		t.Fatalf("Detect() error: %v", err)
	}
	if profile.Stack != "node" {
		t.Errorf("Expected stack 'node', got %q", profile.Stack)
	}
	if profile.Framework != "React" {
		t.Errorf("Expected framework 'React', got %q", profile.Framework)
	}
	if profile.ModuleName != "my-app" {
		t.Errorf("Expected module 'my-app', got %q", profile.ModuleName)
	}
}

func TestDetect_PythonProject(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "requirements.txt"), []byte("flask==2.0.0\n"), 0644)

	profile, err := Detect(dir)
	if err != nil {
		t.Fatalf("Detect() error: %v", err)
	}
	if profile.Stack != "python" {
		t.Errorf("Expected stack 'python', got %q", profile.Stack)
	}
}

func TestDetect_RustProject(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "Cargo.toml"), []byte("[package]\nname = \"myapp\"\n"), 0644)

	profile, err := Detect(dir)
	if err != nil {
		t.Fatalf("Detect() error: %v", err)
	}
	if profile.Stack != "rust" {
		t.Errorf("Expected stack 'rust', got %q", profile.Stack)
	}
}

func TestDetect_WithDocker(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module myapp\ngo 1.21\n"), 0644)
	os.WriteFile(filepath.Join(dir, "Dockerfile"), []byte("FROM golang:1.21\n"), 0644)

	profile, err := Detect(dir)
	if err != nil {
		t.Fatalf("Detect() error: %v", err)
	}
	if !profile.HasDocker {
		t.Error("Expected HasDocker=true, got false")
	}
}

func TestDetect_WithExistingCI(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module myapp\ngo 1.21\n"), 0644)
	os.MkdirAll(filepath.Join(dir, ".github", "workflows"), 0755)

	profile, err := Detect(dir)
	if err != nil {
		t.Fatalf("Detect() error: %v", err)
	}
	if !profile.HasExistingCI {
		t.Error("Expected HasExistingCI=true, got false")
	}
}

func TestDetect_UnknownProject(t *testing.T) {
	dir := t.TempDir()

	profile, err := Detect(dir)
	if err != nil {
		t.Fatalf("Detect() unexpected error: %v", err)
	}
	// Unknown project should have empty stack
	if profile.Stack != "" {
		t.Errorf("Expected empty stack for unknown project, got %q", profile.Stack)
	}
}
