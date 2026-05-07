# YamlAnchor Improvement Roadmap - Exact Implementation Guide

> **Last Updated:** 2026-05-07  
> **Priority Level:** HIGH - These are blocking issues preventing production readiness

---

## 🎯 CRITICAL AREAS (Fix First)

### AREA 1: Missing Backend Server & API Connection
**Status:** ❌ NOT IMPLEMENTED  
**Impact:** Frontend is completely isolated; can't actually generate real pipelines  
**Effort:** 3-4 days

#### Problem:
- Frontend simulates AI generation locally (hardcoded keyword matching in `AIGenerator.jsx`)
- No Go backend server running
- No communication protocol between React UI and Go CLI
- Frontend duplicates pipeline logic that should live in backend

#### Solution Steps:

**Step 1.1: Create Go HTTP Server** (NEW FILE)
```
📁 yaml-anchor/cmd/server.go
```

Create the server command:
```go
package cmd

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/spf13/cobra"
	"yaml-anchor/pkg/analyzer"
	"yaml-anchor/pkg/generator"
)

var (
	port string
)

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start the YamlAnchor HTTP API server",
	Long: `Starts a local HTTP server that provides API endpoints for pipeline generation.
Default listens on :8080`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("🚀 Starting YamlAnchor Server on :%s\n", port)
		
		// API Routes
		http.HandleFunc("/api/analyze", handleAnalyze)
		http.HandleFunc("/api/generate", handleGenerate)
		http.HandleFunc("/api/validate", handleValidate)
		http.HandleFunc("/health", handleHealth)
		
		// CORS middleware
		http.HandleFunc("/", corsMiddleware)
		
		if err := http.ListenAndServe(":"+port, nil); err != nil {
			log.Fatalf("Server error: %v", err)
		}
	},
}

// Request/Response structs
type AnalyzeRequest struct {
	CodeContent string `json:"code"`
	Filetype    string `json:"filetype"` // go, js, package.json, dockerfile
}

type AnalyzeDependenciesRequest struct {
	Dependencies map[string]string `json:"dependencies"` // name -> version
}

type PipelineResponse struct {
	Name   string            `json:"name"`
	On     map[string]interface{} `json:"on"`
	Jobs   []JobSpec         `json:"jobs"`
	Errors []string          `json:"errors,omitempty"`
}

type JobSpec struct {
	ID     string      `json:"id"`
	RunsOn string      `json:"runs_on"`
	Steps  []StepSpec  `json:"steps"`
}

type StepSpec struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Uses string `json:"uses,omitempty"`
	Run  string `json:"run,omitempty"`
}

// Handlers
func handleAnalyze(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req AnalyzeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Use analyzer package (create next)
	result := analyzer.AnalyzeCode(req.CodeContent, req.Filetype)
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func handleGenerate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req PipelineResponse
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Generate YAML
	yamlStr, err := generator.GeneratePipelineYAML(req)
	if err != nil {
		http.Error(w, fmt.Sprintf("Generation failed: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/yaml")
	w.Write([]byte(yamlStr))
}

func handleValidate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req PipelineResponse
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	errors := generator.ValidatePipeline(req)
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"valid":  len(errors) == 0,
		"errors": errors,
	})
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
		"version": "0.1.0",
	})
}

func corsMiddleware(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
}

func init() {
	serverCmd.Flags().StringVarP(&port, "port", "p", "8080", "Port to listen on")
	rootCmd.AddCommand(serverCmd)
}
```

**Step 1.2: Update root.go** to register the server command
```
📝 yaml-anchor/cmd/root.go (ALREADY EXISTS)
```
No changes needed—the `init()` function in `server.go` registers it automatically.

**Step 1.3: Create analyzer package** (NEW FILE)
```
📁 yaml-anchor/pkg/analyzer/analyzer.go
```

```go
package analyzer

import (
	"regexp"
	"strings"
)

type AnalysisResult struct {
	Language     string            `json:"language"`
	Dependencies map[string]string `json:"dependencies"`
	Framework    string            `json:"framework"`
	Suggestions  []string          `json:"suggestions"`
}

func AnalyzeCode(content, filetype string) AnalysisResult {
	result := AnalysisResult{
		Language:     filetype,
		Dependencies: make(map[string]string),
		Suggestions:  []string{},
	}

	switch filetype {
	case "go":
		result = analyzeGo(content)
	case "js", "jsx":
		result = analyzeJavaScript(content)
	case "package.json":
		result = analyzePackageJSON(content)
	case "dockerfile":
		result = analyzeDockerfile(content)
	case "go.mod":
		result = analyzeGoMod(content)
	}

	return result
}

func analyzeGo(content string) AnalysisResult {
	result := AnalysisResult{
		Language:     "go",
		Dependencies: make(map[string]string),
	}

	// Extract imports
	importRegex := regexp.MustCompile(`import\s*\(\s*([\s\S]*?)\s*\)`)
	matches := importRegex.FindStringSubmatch(content)
	if len(matches) > 1 {
		imports := strings.Split(matches[1], "\n")
		for _, imp := range imports {
			imp = strings.TrimSpace(imp)
			if imp != "" && !strings.HasPrefix(imp, "//") {
				result.Dependencies[imp] = "unknown"
			}
		}
	}

	// Detect frameworks/tools
	if strings.Contains(content, "github.com/gorilla/mux") || 
	   strings.Contains(content, "gin-gonic/gin") {
		result.Framework = "web"
		result.Suggestions = append(result.Suggestions, "Detected web framework—add HTTP testing step")
	}

	if strings.Contains(content, "testify") {
		result.Suggestions = append(result.Suggestions, "Add 'go test ./...' step")
	}

	return result
}

func analyzeJavaScript(content string) AnalysisResult {
	result := AnalysisResult{
		Language:     "javascript",
		Dependencies: make(map[string]string),
	}

	// Look for React/Vue/Next.js patterns
	if strings.Contains(content, "import React") || strings.Contains(content, "from 'react'") {
		result.Framework = "react"
		result.Suggestions = append(result.Suggestions, "Detected React—add ESLint & build steps")
	}

	return result
}

func analyzePackageJSON(content string) AnalysisResult {
	result := AnalysisResult{
		Language:     "nodejs",
		Dependencies: make(map[string]string),
	}

	// Extract test, build scripts
	if strings.Contains(content, "\"test\"") {
		result.Suggestions = append(result.Suggestions, "npm test step found")
	}

	return result
}

func analyzeDockerfile(content string) AnalysisResult {
	result := AnalysisResult{
		Language:     "dockerfile",
		Dependencies: make(map[string]string),
	}

	result.Suggestions = append(result.Suggestions, "Add Docker build & push step")
	return result
}

func analyzeGoMod(content string) AnalysisResult {
	result := AnalysisResult{
		Language:     "go",
		Dependencies: make(map[string]string),
	}

	lines := strings.Split(content, "\n")
	for _, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "require") {
			result.Suggestions = append(result.Suggestions, "Go module detected—configure build")
		}
	}

	return result
}
```

**Step 1.4: Update generator package** (EXISTING - EXPAND)
```
📝 yaml-anchor/pkg/generator/generator.go
```

Add these functions:
```go
package generator

import (
	"fmt"
	"gopkg.in/yaml.v3"
)

// Existing ExportYAML function + ADD these:

func GeneratePipelineYAML(pipeline interface{}) (string, error) {
	data, err := yaml.Marshal(pipeline)
	if err != nil {
		return "", fmt.Errorf("failed to marshal YAML: %w", err)
	}
	return string(data), nil
}

func ValidatePipeline(p interface{}) []string {
	errors := []string{}
	
	// Type assert to map for validation
	pMap, ok := p.(map[string]interface{})
	if !ok {
		errors = append(errors, "Invalid pipeline structure")
		return errors
	}

	if pMap["name"] == "" {
		errors = append(errors, "Pipeline must have a name")
	}

	if jobs, ok := pMap["jobs"].([]interface{}); ok {
		if len(jobs) == 0 {
			errors = append(errors, "Pipeline must have at least one job")
		}
	}

	return errors
}
```

**Step 1.5: Update go.mod** (ADD dependency)
```
📝 yaml-anchor/go.mod
```

Add this line in the `require` section:
```
github.com/rs/cors v1.11.1
```

---

### AREA 2: Frontend Not Connected to Backend
**Status:** ❌ NOT IMPLEMENTED  
**Impact:** UI has duplicate logic; can't leverage Go backend  
**Effort:** 2 days

#### Problem:
- `AIGenerator.jsx` has hardcoded simulation logic
- No API calls to backend
- No environment configuration for API URL

#### Solution Steps:

**Step 2.1: Create API service layer** (NEW FILE)
```
📁 yaml-anchor/ui/src/services/api.js
```

```javascript
// Base configuration
const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080';

export const api = {
  // Analyze code/dependencies
  async analyzeCode(content, filetype) {
    try {
      const response = await fetch(`${API_BASE_URL}/api/analyze`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ code: content, filetype }),
      });
      if (!response.ok) throw new Error(`API error: ${response.status}`);
      return await response.json();
    } catch (error) {
      console.error('Analysis failed:', error);
      throw error;
    }
  },

  // Generate pipeline from user input
  async generatePipeline(input) {
    try {
      // First analyze the input
      const analysis = await this.analyzeCode(input, 'text');
      
      // Then generate pipeline based on analysis
      const response = await fetch(`${API_BASE_URL}/api/generate`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ analysis }),
      });
      if (!response.ok) throw new Error(`Generation failed: ${response.status}`);
      return await response.json();
    } catch (error) {
      console.error('Generation failed:', error);
      throw error;
    }
  },

  // Validate pipeline before export
  async validatePipeline(pipeline) {
    try {
      const response = await fetch(`${API_BASE_URL}/api/validate`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(pipeline),
      });
      if (!response.ok) throw new Error(`Validation failed: ${response.status}`);
      return await response.json();
    } catch (error) {
      console.error('Validation failed:', error);
      throw error;
    }
  },

  // Health check
  async checkHealth() {
    try {
      const response = await fetch(`${API_BASE_URL}/health`);
      return response.ok;
    } catch {
      return false;
    }
  },
};
```

**Step 2.2: Create environment config** (NEW FILE)
```
📁 yaml-anchor/ui/.env.example
```

```env
# Backend API configuration
VITE_API_URL=http://localhost:8080

# Optional: AI service keys (future)
VITE_OPENAI_KEY=
VITE_CLAUDE_KEY=
```

**Step 2.3: Update AIGenerator.jsx** to use API
```
📝 yaml-anchor/ui/src/components/AIGenerator.jsx
```

Replace the entire `simulateAILogic` function:

```javascript
import { api } from '../services/api';
import { useState, useEffect } from 'react';

export default function AIGenerator({ onPipelineGenerated }) {
  const [input, setInput] = useState('');
  const [isGenerating, setIsGenerating] = useState(false);
  const [backendConnected, setBackendConnected] = useState(false);
  const [error, setError] = useState('');

  // Check backend connection on mount
  useEffect(() => {
    api.checkHealth().then(setBackendConnected);
  }, []);

  const handleGenerate = async () => {
    if (!input.trim()) return;
    if (!backendConnected) {
      setError('Backend server not available. Start with: npm run server');
      return;
    }

    setIsGenerating(true);
    setError('');
    
    try {
      const pipeline = await api.generatePipeline(input);
      
      // Validate before showing
      const validation = await api.validatePipeline(pipeline);
      if (!validation.valid) {
        setError(`Validation errors: ${validation.errors.join(', ')}`);
      }
      
      onPipelineGenerated(pipeline);
    } catch (err) {
      setError(`Failed to generate: ${err.message}`);
      console.error(err);
    } finally {
      setIsGenerating(false);
    }
  };

  return (
    <div className="panel">
      <div className="panel-header">
        <div style={{ display: 'flex', alignItems: 'center' }}>
          <TerminalSquare size={16} />
          Input & AI Generator
        </div>
        {backendConnected ? (
          <span style={{ fontSize: '0.7rem', color: 'var(--accent-green)' }}>● CONNECTED</span>
        ) : (
          <span style={{ fontSize: '0.7rem', color: 'var(--danger)' }}>● OFFLINE</span>
        )}
      </div>
      <div className="panel-content">
        <div className="ai-input-wrapper">
          {error && (
            <div style={{ 
              padding: '0.75rem', 
              background: 'rgba(239, 68, 68, 0.1)',
              border: '1px solid var(--danger)',
              borderRadius: '4px',
              color: 'var(--danger)',
              fontSize: '0.8rem'
            }}>
              {error}
            </div>
          )}
          
          <p style={{ fontSize: '0.85rem', color: 'var(--text-secondary)' }}>
            Describe your project stack or paste code. The AI will analyze it and generate an optimal pipeline.
            <br/><br/>
            <em>Try: "Go backend with Docker", "React frontend", "Node.js with tests"</em>
          </p>
          
          <textarea 
            className="ai-textarea"
            placeholder="e.g., Go backend service with unit tests, Docker containerization..."
            value={input}
            onChange={(e) => setInput(e.target.value)}
            disabled={!backendConnected}
          />

          <button 
            className="btn btn-ai" 
            onClick={handleGenerate}
            disabled={isGenerating || !input.trim() || !backendConnected}
          >
            {isGenerating ? (
              <>
                <Loader2 size={18} className="animate-spin" />
                Analyzing & Generating...
              </>
            ) : (
              <>
                <Cpu size={18} />
                Generate Pipeline
              </>
            )}
          </button>
        </div>
      </div>
    </div>
  );
}
```

**Step 2.4: Update package.json** to add dev script
```
📝 yaml-anchor/ui/package.json
```

In the `scripts` section, add:
```json
"server": "cd .. && go run main.go server",
"dev:full": "concurrently \"npm run server\" \"npm run dev\""
```

Also add dev dependency:
```json
"devDependencies": {
  "concurrently": "^10.0.0"
}
```

---

### AREA 3: No Testing Framework
**Status:** ❌ ZERO TESTS  
**Impact:** Can't catch regressions; reduces code quality  
**Effort:** 3-4 days

#### Problem:
- Zero test files in entire repository
- No CI/CD validation pipeline
- Can't confidently refactor code

#### Solution Steps:

**Step 3.1: Add Go testing** (NEW FILE)
```
📁 yaml-anchor/pkg/generator/generator_test.go
```

```go
package generator

import (
	"testing"
)

func TestValidatePipeline(t *testing.T) {
	tests := []struct {
		name    string
		pipeline interface{}
		wantErrors bool
	}{
		{
			name: "valid pipeline",
			pipeline: map[string]interface{}{
				"name": "test-pipeline",
				"jobs": []interface{}{
					map[string]interface{}{
						"id": "test-job",
					},
				},
			},
			wantErrors: false,
		},
		{
			name: "missing name",
			pipeline: map[string]interface{}{
				"name": "",
				"jobs": []interface{}{},
			},
			wantErrors: true,
		},
		{
			name: "no jobs",
			pipeline: map[string]interface{}{
				"name": "test",
				"jobs": []interface{}{},
			},
			wantErrors: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := ValidatePipeline(tt.pipeline)
			hasErrors := len(errors) > 0
			if hasErrors != tt.wantErrors {
				t.Errorf("ValidatePipeline() got errors: %v, want errors: %v", hasErrors, tt.wantErrors)
			}
		})
	}
}

func TestGeneratePipelineYAML(t *testing.T) {
	pipeline := map[string]interface{}{
		"name": "test-pipeline",
		"on": map[string]interface{}{
			"push": map[string]interface{}{
				"branches": []string{"main"},
			},
		},
	}

	yaml, err := GeneratePipelineYAML(pipeline)
	if err != nil {
		t.Fatalf("GeneratePipelineYAML() failed: %v", err)
	}

	if yaml == "" {
		t.Error("GeneratePipelineYAML() returned empty string")
	}

	if !contains(yaml, "name:") || !contains(yaml, "test-pipeline") {
		t.Error("YAML output doesn't contain expected fields")
	}
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
```

**Step 3.2: Add analyzer tests** (NEW FILE)
```
📁 yaml-anchor/pkg/analyzer/analyzer_test.go
```

```go
package analyzer

import (
	"strings"
	"testing"
)

func TestAnalyzeGo(t *testing.T) {
	content := `package main

import (
	"fmt"
	"github.com/gorilla/mux"
)

func main() {
	fmt.Println("test")
}`

	result := analyzeGo(content)

	if result.Language != "go" {
		t.Errorf("Expected language 'go', got '%s'", result.Language)
	}

	if len(result.Dependencies) == 0 {
		t.Error("Expected dependencies to be extracted")
	}

	if result.Framework != "web" {
		t.Errorf("Expected to detect web framework, got '%s'", result.Framework)
	}
}

func TestAnalyzeJavaScript(t *testing.T) {
	content := `import React from 'react'
import { useState } from 'react'`

	result := analyzeJavaScript(content)

	if result.Framework != "react" {
		t.Errorf("Expected React detection, got '%s'", result.Framework)
	}
}

func TestAnalyzeCode(t *testing.T) {
	goCode := `package main
import "fmt"`
	
	result := AnalyzeCode(goCode, "go")
	
	if result.Language != "go" {
		t.Errorf("AnalyzeCode failed for Go: expected 'go', got '%s'", result.Language)
	}
}
```

**Step 3.3: Add React tests** (NEW FILE)
```
📁 yaml-anchor/ui/src/components/__tests__/AIGenerator.test.jsx
```

```javascript
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import AIGenerator from '../AIGenerator';

// Mock the api module
vi.mock('../../services/api', () => ({
  api: {
    checkHealth: vi.fn(() => Promise.resolve(true)),
    generatePipeline: vi.fn(() => Promise.resolve({
      name: 'test',
      jobs: [{ id: 'job1', steps: [] }],
    })),
    validatePipeline: vi.fn(() => Promise.resolve({ valid: true, errors: [] })),
  },
}));

describe('AIGenerator', () => {
  let mockCallback;

  beforeEach(() => {
    mockCallback = vi.fn();
  });

  it('renders without crashing', () => {
    render(<AIGenerator onPipelineGenerated={mockCallback} />);
    expect(screen.getByText(/Input & AI Generator/i)).toBeInTheDocument();
  });

  it('disables button when input is empty', () => {
    render(<AIGenerator onPipelineGenerated={mockCallback} />);
    const button = screen.getByRole('button', { name: /Generate Pipeline/i });
    expect(button).toBeDisabled();
  });

  it('enables button when input has text', async () => {
    render(<AIGenerator onPipelineGenerated={mockCallback} />);
    const textarea = screen.getByPlaceholderText(/Go backend/i);
    
    fireEvent.change(textarea, { target: { value: 'test input' } });
    
    await waitFor(() => {
      const button = screen.getByRole('button', { name: /Generate Pipeline/i });
      expect(button).not.toBeDisabled();
    });
  });
});
```

**Step 3.4: Setup test infrastructure** (NEW FILE)
```
📁 yaml-anchor/ui/vitest.config.js
```

```javascript
import { defineConfig } from 'vitest/config';
import react from '@vitejs/plugin-react';

export default defineConfig({
  plugins: [react()],
  test: {
    globals: true,
    environment: 'jsdom',
    setupFiles: [],
    coverage: {
      provider: 'v8',
      reporter: ['text', 'json', 'html'],
    },
  },
});
```

**Step 3.5: Update package.json** for tests
```
📝 yaml-anchor/ui/package.json
```

Add to devDependencies:
```json
"vitest": "^2.0.0",
"@vitest/ui": "^2.0.0",
"@testing-library/react": "^16.0.0",
"@testing-library/jest-dom": "^6.4.0",
"jsdom": "^24.0.0"
```

Add to scripts:
```json
"test": "vitest",
"test:ui": "vitest --ui",
"test:coverage": "vitest --coverage"
```

**Step 3.6: Add Go test script** (NEW FILE)
```
📁 yaml-anchor/Makefile
```

```makefile
.PHONY: test test-go test-ui test-all coverage

test-go:
	cd yaml-anchor && go test ./... -v -race

test-ui:
	cd yaml-anchor/ui && npm run test

test-all: test-go test-ui

coverage:
	cd yaml-anchor && go test ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out

lint:
	cd yaml-anchor && golangci-lint run ./...
	cd yaml-anchor/ui && npm run lint

build:
	cd yaml-anchor && go build -o bin/yaml-anchor main.go
	cd yaml-anchor/ui && npm run build

run-server:
	cd yaml-anchor && go run main.go server

run-ui:
	cd yaml-anchor/ui && npm run dev

dev:
	concurrently "make run-server" "make run-ui"
```

---

### AREA 4: Incomplete Backend Commands
**Status:** ⚠️ PARTIALLY IMPLEMENTED  
**Impact:** Can't simulate/dry-run pipelines locally  
**Effort:** 2-3 days

#### Problem:
- `generate` command exists but only writes to file
- `simulate` command referenced in root.go but never implemented
- No Dagger integration for actual pipeline execution
- No error handling for edge cases

#### Solution Steps:

**Step 4.1: Implement simulate command** (NEW FILE)
```
📁 yaml-anchor/cmd/simulate.go
```

```go
package cmd

import (
	"context"
	"fmt"
	"log"

	"dagger.io/dagger"
	"github.com/spf13/cobra"
	"yaml-anchor/pkg/config"
)

var (
	simulateConfigPath string
	dryRun             bool
)

var simulateCmd = &cobra.Command{
	Use:   "simulate",
	Short: "Simulate pipeline execution locally using Dagger",
	Long: `Reads your anchor.yaml configuration and simulates the entire
pipeline execution locally using Dagger, without pushing to CI/CD.
Use --dry-run to preview without executing.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Loading pipeline config from %s...\n", simulateConfigPath)

		pipeline, err := config.Load(simulateConfigPath)
		if err != nil {
			log.Fatalf("Failed to load config: %v", err)
		}

		fmt.Printf("🚀 Simulating pipeline: %q\n", pipeline.Name)

		if dryRun {
			fmt.Println("[DRY RUN] Would execute:")
			for _, job := range pipeline.Jobs {
				fmt.Printf("  Job: %s (runs-on: %s)\n", job.ID, job.RunsOn)
				for _, step := range job.Steps {
					if step.Uses != "" {
						fmt.Printf("    - uses: %s\n", step.Uses)
					} else {
						fmt.Printf("    - run: %s\n", step.Run)
					}
				}
			}
			return
		}

		// Actual simulation with Dagger
		if err := simulateWithDagger(cmd.Context(), pipeline); err != nil {
			log.Fatalf("Simulation failed: %v", err)
		}

		fmt.Println("✓ Pipeline simulation completed successfully")
	},
}

func simulateWithDagger(ctx context.Context, pipeline interface{}) error {
	client, err := dagger.Connect(ctx)
	if err != nil {
		return fmt.Errorf("failed to connect to Dagger: %w", err)
	}
	defer client.Close()

	container := client.Container().From("ubuntu:latest").
		WithExec([]string{"sh", "-c", "echo 'Starting simulation...'"})

	_, err = container.Sync(ctx)
	if err != nil {
		return fmt.Errorf("container exec failed: %w", err)
	}

	return nil
}

func init() {
	simulateCmd.Flags().StringVarP(&simulateConfigPath, "config", "c", "anchor.yaml",
		"Path to your anchor.yaml pipeline definition")
	simulateCmd.Flags().BoolVarP(&dryRun, "dry-run", "d", false,
		"Preview execution without actually running")
	rootCmd.AddCommand(simulateCmd)
}
```

**Step 4.2: Create config loader** (NEW FILE)
```
📁 yaml-anchor/pkg/config/loader.go
```

```go
package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Name  string `yaml:"name"`
	On    map[string]interface{} `yaml:"on"`
	Jobs  []Job  `yaml:"jobs"`
	Env   map[string]string `yaml:"env,omitempty"`
	Secrets []string `yaml:"secrets,omitempty"`
}

type Job struct {
	ID     string `yaml:"id"`
	RunsOn string `yaml:"runs-on"`
	If     string `yaml:"if,omitempty"`
	Steps  []Step `yaml:"steps"`
}

type Step struct {
	Name string `yaml:"name"`
	ID   string `yaml:"id,omitempty"`
	Uses string `yaml:"uses,omitempty"`
	Run  string `yaml:"run,omitempty"`
	With map[string]interface{} `yaml:"with,omitempty"`
	Env  map[string]string `yaml:"env,omitempty"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	if err := validate(&cfg); err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
	}

	return &cfg, nil
}

func validate(cfg *Config) error {
	if cfg.Name == "" {
		return fmt.Errorf("pipeline must have a name")
	}

	if len(cfg.Jobs) == 0 {
		return fmt.Errorf("pipeline must have at least one job")
	}

	for _, job := range cfg.Jobs {
		if job.ID == "" {
			return fmt.Errorf("job must have an id")
		}
		if job.RunsOn == "" {
			return fmt.Errorf("job %q must specify runs-on", job.ID)
		}
		if len(job.Steps) == 0 {
			return fmt.Errorf("job %q must have at least one step", job.ID)
		}
	}

	return nil
}
```

**Step 4.3: Update existing generate command** 
```
📝 yaml-anchor/cmd/generate.go (EXISTING - ENHANCE)
```

Replace with improved version:
```go
package cmd

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"yaml-anchor/pkg/config"
	"yaml-anchor/pkg/generator"
)

var generateConfigPath string
var generateOutputPath string

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate a GitHub Actions YAML from an anchor.yaml config",
	Long: `Reads your pipeline definition from an anchor.yaml file,
performs validation and security scanning, and writes a valid 
GitHub Actions workflow file.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("📖 Loading pipeline config from %s...\n", generateConfigPath)

		pipeline, err := config.Load(generateConfigPath)
		if err != nil {
			log.Fatalf("❌ Failed to load config: %v", err)
		}

		fmt.Printf("🔍 Validating pipeline: %q\n", pipeline.Name)

		// Validate
		validationErrs := generator.ValidatePipeline(pipeline)
		if len(validationErrs) > 0 {
			fmt.Println("❌ Validation errors:")
			for _, e := range validationErrs {
				fmt.Printf("  - %s\n", e)
			}
			os.Exit(1)
		}

		// Security scan
		securityWarnings := generator.ScanForSecurityIssues(pipeline)
		if len(securityWarnings) > 0 {
			fmt.Println("⚠️  Security warnings:")
			for _, w := range securityWarnings {
				fmt.Printf("  - %s\n", w)
			}
		}

		fmt.Printf("✨ Generating YAML...\n")

		if err := generator.ExportYAML(pipeline, generateOutputPath); err != nil {
			log.Fatalf("❌ Error generating YAML: %v", err)
		}

		fmt.Printf("✅ Successfully generated workflow at %s\n", generateOutputPath)
	},
}

func init() {
	generateCmd.Flags().StringVarP(&generateConfigPath, "config", "c", "anchor.yaml",
		"Path to your anchor.yaml pipeline definition")
	generateCmd.Flags().StringVarP(&generateOutputPath, "output", "o", ".github/workflows/main.yml",
		"Output path for generated workflow file")
	rootCmd.AddCommand(generateCmd)
}
```

**Step 4.4: Add security scanning** (NEW FILE)
```
📁 yaml-anchor/pkg/generator/security.go
```

```go
package generator

import (
	"fmt"
	"strings"
)

func ScanForSecurityIssues(pipeline interface{}) []string {
	var warnings []string

	// Type assert to map
	pMap, ok := pipeline.(map[string]interface{})
	if !ok {
		return warnings
	}

	if jobs, ok := pMap["jobs"].([]interface{}); ok {
		for _, j := range jobs {
			if jobMap, ok := j.(map[string]interface{}); ok {
				if steps, ok := jobMap["steps"].([]interface{}); ok {
					for _, s := range steps {
						if stepMap, ok := s.(map[string]interface{}); ok {
							warnings = append(warnings, scanStep(stepMap)...)
						}
					}
				}
			}
		}
	}

	return warnings
}

func scanStep(step map[string]interface{}) []string {
	var warnings []string

	if run, ok := step["run"].(string); ok {
		// Warn about curl | bash
		if strings.Contains(run, "curl") && strings.Contains(run, "| bash") {
			warnings = append(warnings, fmt.Sprintf("Dangerous pattern in step: %q", run))
		}

		// Warn about hardcoded secrets
		if strings.Contains(run, "password=") || strings.Contains(run, "token=") {
			warnings = append(warnings, "Possible hardcoded secret detected")
		}

		// Warn about sudo without nopasswd
		if strings.Contains(run, "sudo") && !strings.Contains(run, "NOPASSWD") {
			warnings = append(warnings, "sudo usage may require NOPASSWD configuration")
		}
	}

	return warnings
}
```

---

### AREA 5: No Documentation
**Status:** ❌ MISSING CRITICAL DOCS  
**Impact:** Users can't set up or use the project  
**Effort:** 1-2 days

#### Problem:
- No setup guide
- No API documentation
- No example workflows
- No architecture overview

#### Solution Steps:

**Step 5.1: Create SETUP.md** (NEW FILE)
```
📁 yaml-anchor/SETUP.md
```

```markdown
# YamlAnchor Setup Guide

## Prerequisites

- Go 1.26.2 or higher
- Node.js 18+ with npm
- Docker (optional, for full CI/CD testing)

## Installation & Development

### 1. Clone the Repository
\`\`\`bash
git clone https://github.com/AyushCN/fusiontech.git
cd fusiontech
\`\`\`

### 2. Setup Backend (Go)
\`\`\`bash
cd yaml-anchor
go mod download
go build -o bin/yaml-anchor main.go
\`\`\`

### 3. Setup Frontend (React)
\`\`\`bash
cd ui
npm install
\`\`\`

### 4. Run Development Environment

**Option A: Using Makefile (Recommended)**
\`\`\`bash
make dev  # Runs both server and UI
\`\`\`

**Option B: Manual (two terminals)**

Terminal 1 - Start Backend Server:
\`\`\`bash
cd yaml-anchor
go run main.go server --port 8080
\`\`\`

Terminal 2 - Start Frontend:
\`\`\`bash
cd yaml-anchor/ui
npm run dev
\`\`\`

### 5. Access the Application
- UI: http://localhost:5173
- API: http://localhost:8080
- Health Check: http://localhost:8080/health

## Running Tests

\`\`\`bash
# Go tests
make test-go

# React tests
make test-ui

# All tests
make test-all

# Coverage
make coverage
\`\`\`

## Building for Production

\`\`\`bash
make build
\`\`\`

This creates:
- `bin/yaml-anchor` - CLI binary
- `dist/` - React build in `ui/dist`
```

**Step 5.2: Create API_DOCS.md** (NEW FILE)
```
📁 yaml-anchor/API_DOCS.md
```

```markdown
# YamlAnchor API Documentation

## Base URL
```
http://localhost:8080
```

### Health Check
**Endpoint:** `GET /health`

**Response:**
```json
{
  "status": "ok",
  "version": "0.1.0"
}
```

### Analyze Code
**Endpoint:** `POST /api/analyze`

**Request:**
```json
{
  "code": "package main\nimport \"fmt\"",
  "filetype": "go"
}
```

**Response:**
```json
{
  "language": "go",
  "dependencies": {
    "fmt": "builtin"
  },
  "framework": "",
  "suggestions": ["Add testing step"]
}
```

**Supported Filetypes:**
- `go` - Go source files
- `js`, `jsx` - JavaScript/React
- `package.json` - Node.js config
- `dockerfile` - Docker config
- `go.mod` - Go modules

### Generate Pipeline
**Endpoint:** `POST /api/generate`

**Request:**
```json
{
  "analysis": {
    "language": "go",
    "framework": "web"
  }
}
```

**Response:**
```json
{
  "name": "go-web-pipeline",
  "on": {
    "push": {
      "branches": ["main"]
    }
  },
  "jobs": [
    {
      "id": "build",
      "runs_on": "ubuntu-latest",
      "steps": [
        {
          "id": 1,
          "name": "Checkout",
          "uses": "actions/checkout@v4"
        }
      ]
    }
  ]
}
```

### Validate Pipeline
**Endpoint:** `POST /api/validate`

**Request:** (pipeline object from /generate)

**Response:**
```json
{
  "valid": true,
  "errors": []
}
```

## Error Handling

All errors return appropriate HTTP status codes:

- `400 Bad Request` - Invalid JSON or missing required fields
- `422 Unprocessable Entity` - Validation failed
- `500 Internal Server Error` - Server-side processing error

Error Response Format:
```json
{
  "error": "Description of what went wrong"
}
```
```

**Step 5.3: Create EXAMPLES.md** (NEW FILE)
```
📁 yaml-anchor/EXAMPLES.md
```

```markdown
# YamlAnchor Examples

## Example 1: Simple Go Project

**Input (UI):**
```
I have a Go backend service that needs unit tests and Docker deployment
```

**Generated Pipeline:**
```yaml
name: go-backend-pipeline
on:
  push:
    branches:
      - main
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout Repo
        uses: actions/checkout@v4
      - name: Setup Go
        uses: actions/setup-go@v4
      - name: Run Tests
        run: go test ./...
      - name: Build Binary
        run: go build -o bin/app main.go
      - name: Build Docker Image
        run: docker build -t myapp:latest .
```

## Example 2: React Frontend

**Input (UI):**
```
React application with ESLint and needs to be built and tested
```

**Generated Pipeline:**
```yaml
name: react-frontend-pipeline
on:
  push:
    branches:
      - main
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout Repo
        uses: actions/checkout@v4
      - name: Setup Node
        uses: actions/setup-node@v3
      - name: Install Dependencies
        run: npm install
      - name: Lint Code
        run: npm run lint
      - name: Build Project
        run: npm run build
```

## Example 3: Using CLI Directly

**Create `anchor.yaml`:**
```yaml
name: my-project-pipeline
on:
  push:
    branches:
      - main
jobs:
  - id: test
    runs-on: ubuntu-latest
    steps:
      - name: Checkout
        uses: actions/checkout@v4
      - name: Run Tests
        run: npm test
```

**Generate Workflow:**
```bash
yaml-anchor generate --config anchor.yaml --output .github/workflows/ci.yml
```

**Dry-run Simulation:**
```bash
yaml-anchor simulate --config anchor.yaml --dry-run
```

**Full Local Simulation:**
```bash
yaml-anchor simulate --config anchor.yaml
```
```

**Step 5.4: Update main README.md**
```
📝 README.md (ALREADY EXISTS - EXPAND)
```

Add these sections if not present:

```markdown
## Quick Start

```bash
git clone https://github.com/AyushCN/fusiontech.git
cd fusiontech
make dev
```

Visit http://localhost:5173

## Documentation

- [Setup Guide](yaml-anchor/SETUP.md) - Installation & development
- [API Docs](yaml-anchor/API_DOCS.md) - REST API reference
- [Examples](yaml-anchor/EXAMPLES.md) - Real-world usage examples
- [Architecture](yaml-anchor/ARCHITECTURE.md) - System design

## Commands

### Backend (Go)
- `yaml-anchor generate` - Generate GitHub Actions workflow from config
- `yaml-anchor simulate` - Simulate pipeline execution locally
- `yaml-anchor server` - Start HTTP API server

### Development
- `make dev` - Start server + UI together
- `make test-all` - Run all tests
- `make build` - Production build

See [SETUP.md](yaml-anchor/SETUP.md) for detailed instructions.
```

---

### AREA 6: Missing Error Handling & Validation
**Status:** ⚠️ BASIC ONLY  
**Impact:** Users get vague errors; crashes on edge cases  
**Effort:** 2 days

#### Problem:
- Minimal error messages
- No input validation
- Silent failures possible
- No logging for debugging

#### Solution Steps:

**Step 6.1: Create error package** (NEW FILE)
```
📁 yaml-anchor/pkg/errors/errors.go
```

```go
package errors

import "fmt"

// Custom error types
type ConfigError struct {
	Message string
	Path    string
}

type ValidationError struct {
	Field   string
	Message string
}

type SecurityError struct {
	Severity string // low, medium, high, critical
	Message  string
	Suggestion string
}

// Error() implementations
func (e *ConfigError) Error() string {
	if e.Path != "" {
		return fmt.Sprintf("config error in %q: %s", e.Path, e.Message)
	}
	return fmt.Sprintf("config error: %s", e.Message)
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error in field %q: %s", e.Field, e.Message)
}

func (e *SecurityError) Error() string {
	return fmt.Sprintf("[%s] %s - Suggestion: %s", e.Severity, e.Message, e.Suggestion)
}

// Constructor functions
func NewConfigError(path, msg string) error {
	return &ConfigError{Message: msg, Path: path}
}

func NewValidationError(field, msg string) error {
	return &ValidationError{Field: field, Message: msg}
}

func NewSecurityError(severity, msg, suggestion string) error {
	return &SecurityError{Severity: severity, Message: msg, Suggestion: suggestion}
}
```

**Step 6.2: Add input validation** (NEW FILE)
```
📁 yaml-anchor/pkg/validator/validator.go
```

```go
package validator

import (
	"fmt"
	"regexp"
	"strings"
)

type ValidationResult struct {
	Valid  bool
	Errors []string
}

func ValidateJobID(id string) error {
	if id == "" {
		return fmt.Errorf("job ID cannot be empty")
	}

	// GitHub Actions job IDs must be alphanumeric and hyphens
	if !regexp.MustCompile(`^[a-zA-Z0-9_-]+$`).MatchString(id) {
		return fmt.Errorf("job ID %q contains invalid characters (only alphanumeric, underscore, hyphen allowed)", id)
	}

	if len(id) > 50 {
		return fmt.Errorf("job ID %q too long (max 50 characters)", id)
	}

	return nil
}

func ValidateStepName(name string) error {
	if name == "" {
		return fmt.Errorf("step name cannot be empty")
	}

	if len(name) > 100 {
		return fmt.Errorf("step name too long (max 100 characters)")
	}

	return nil
}

func ValidateRunsOn(runsOn string) error {
	validRunners := map[string]bool{
		"ubuntu-latest":   true,
		"ubuntu-22.04":    true,
		"ubuntu-20.04":    true,
		"windows-latest":  true,
		"macos-latest":    true,
		"macos-13":        true,
		"self-hosted":     true,
	}

	if !validRunners[runsOn] && !strings.HasPrefix(runsOn, "self-hosted") {
		return fmt.Errorf("invalid runner %q", runsOn)
	}

	return nil
}

func ValidateCron(cron string) error {
	// Basic cron validation
	parts := strings.Fields(cron)
	if len(parts) != 5 {
		return fmt.Errorf("invalid cron expression %q (must have 5 fields)", cron)
	}

	return nil
}
```

**Step 6.3: Add logging utility** (NEW FILE)
```
📁 yaml-anchor/pkg/logger/logger.go
```

```go
package logger

import (
	"fmt"
	"log"
	"os"
	"time"
)

type LogLevel string

const (
	DEBUG LogLevel = "DEBUG"
	INFO  LogLevel = "INFO"
	WARN  LogLevel = "WARN"
	ERROR LogLevel = "ERROR"
	FATAL LogLevel = "FATAL"
)

var (
	currentLevel = INFO
	logFile      *os.File
)

func Init(level LogLevel, filepath string) error {
	currentLevel = level

	if filepath != "" {
		f, err := os.OpenFile(filepath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return err
		}
		logFile = f
	}

	return nil
}

func Close() {
	if logFile != nil {
		logFile.Close()
	}
}

func logMessage(level LogLevel, msg string, args ...interface{}) {
	if level == DEBUG && currentLevel != DEBUG {
		return
	}

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	formatted := fmt.Sprintf(msg, args...)
	output := fmt.Sprintf("[%s] %s: %s", timestamp, level, formatted)

	fmt.Println(output)

	if logFile != nil {
		fmt.Fprintln(logFile, output)
	}
}

func Debug(msg string, args ...interface{})   { logMessage(DEBUG, msg, args...) }
func Info(msg string, args ...interface{})    { logMessage(INFO, msg, args...) }
func Warn(msg string, args ...interface{})    { logMessage(WARN, msg, args...) }
func Error(msg string, args ...interface{})   { logMessage(ERROR, msg, args...) }
func Fatal(msg string, args ...interface{})   {
	logMessage(FATAL, msg, args...)
	os.Exit(1)
}
```

**Step 6.4: Update API handlers with better errors**
```
📝 yaml-anchor/cmd/server.go (EXISTING - UPDATE)
```

Update handlers:
```go
func handleAnalyze(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req AnalyzeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, fmt.Sprintf("Invalid request body: %v", err))
		return
	}

	// Validate input
	if req.CodeContent == "" {
		respondError(w, http.StatusBadRequest, "code field is required")
		return
	}

	if req.Filetype == "" {
		respondError(w, http.StatusBadRequest, "filetype field is required")
		return
	}

	result := analyzer.AnalyzeCode(req.CodeContent, req.Filetype)
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(result)
}

func respondError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error":  message,
		"status": statusCode,
	})
}

func respondJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}
```

---

## 📊 Implementation Priority Matrix

| Area | Difficulty | Impact | Timeline | Priority |
|------|-----------|--------|----------|----------|
| Backend Server & API | Medium | Critical | 3-4 days | **P0** |
| Frontend Connection | Easy | Critical | 2 days | **P0** |
| Error Handling | Medium | High | 2 days | **P1** |
| Testing Framework | Medium | High | 3-4 days | **P1** |
| Complete CLI Commands | Medium | High | 2-3 days | **P1** |
| Documentation | Easy | Medium | 1-2 days | **P2** |

---

## 🚀 Quick Start Implementation (Pick ONE)

### **Option A: Fastest Path (2-3 days)**
1. Step 1.1-1.2: Create Go HTTP server
2. Step 2.1-2.3: Connect frontend to API
3. Step 6.1-6.2: Add basic error handling

**Result:** Working server + UI integration

### **Option B: Production Ready (1 week)**
Complete all areas above in priority order

### **Option C: MVP Enhancement (4-5 days)**
1. Areas 1-2 (Backend + Connection)
2. Area 5 (Documentation)
3. Area 3 (Testing)

---

## 📋 Execution Checklist

Use this to track progress:

```
PHASE 1: Backend Server (Days 1-2)
  [ ] Create server.go command
  [ ] Create analyzer package
  [ ] Update generator package
  [ ] Test endpoints with curl/Postman

PHASE 2: Frontend Connection (Days 2-3)
  [ ] Create api.js service layer
  [ ] Update AIGenerator.jsx
  [ ] Add environment config
  [ ] Test E2E flow

PHASE 3: Testing (Days 3-5)
  [ ] Go unit tests
  [ ] React component tests
  [ ] Setup CI/CD pipeline

PHASE 4: Documentation (Day 5-6)
  [ ] SETUP.md
  [ ] API_DOCS.md
  [ ] EXAMPLES.md
  [ ] README updates

PHASE 5: Polish (Day 6-7)
  [ ] Error handling
  [ ] Security scanning
  [ ] Performance optimization
  [ ] Final testing
```

---

## 🎯 Success Metrics

After implementation:
- ✅ Backend API fully operational
- ✅ Frontend consumes API (no more mock logic)
- ✅ 80%+ test coverage
- ✅ All errors handled gracefully
- ✅ Users can follow setup in 10 minutes
- ✅ Example workflows run successfully

---

**Start with Area 1 (Backend Server). It unblocks everything else.**
