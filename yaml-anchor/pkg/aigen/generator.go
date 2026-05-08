package aigen

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"yaml-anchor/pkg/analyzer"
	"yaml-anchor/pkg/schema"
)

const defaultOllamaHost = "http://localhost:11434"
const defaultOllamaModel = "llama3.2"

// Generate creates a YamlAnchor pipeline from a prompt or pasted project data.
// It is keyless: when a local Ollama model is available it asks that model,
// otherwise it falls back to a deterministic offline generator.
func Generate(ctx context.Context, content, fileType string) (*schema.Pipeline, string, error) {
	content = strings.TrimSpace(content)
	if content == "" {
		return nil, "", fmt.Errorf("content is required")
	}
	if fileType == "" {
		fileType = InferFileType(content)
	}

	if localLLMEnabled() {
		pipeline, err := generateWithOllama(ctx, content, fileType)
		if err == nil {
			return pipeline, "ollama", nil
		}
	}

	return generateOffline(content, fileType), "offline", nil
}

func localLLMEnabled() bool {
	return strings.ToLower(strings.TrimSpace(os.Getenv("YAML_ANCHOR_LLM"))) != "off"
}

func generateWithOllama(ctx context.Context, content, fileType string) (*schema.Pipeline, error) {
	host := strings.TrimRight(os.Getenv("OLLAMA_HOST"), "/")
	if host == "" {
		host = defaultOllamaHost
	}
	model := strings.TrimSpace(os.Getenv("YAML_ANCHOR_MODEL"))
	if model == "" {
		model = defaultOllamaModel
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, 8*time.Second)
	defer cancel()

	body := map[string]interface{}{
		"model":  model,
		"stream": false,
		"format": "json",
		"prompt": buildPrompt(content, fileType),
	}
	data, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(timeoutCtx, http.MethodPost, host+"/api/generate", bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return nil, fmt.Errorf("ollama returned %s", res.Status)
	}

	var ollamaRes struct {
		Response string `json:"response"`
	}
	if err := json.NewDecoder(res.Body).Decode(&ollamaRes); err != nil {
		return nil, err
	}

	pipeline, err := decodePipelineJSON(ollamaRes.Response)
	if err != nil {
		return nil, err
	}
	return pipeline, nil
}

func buildPrompt(content, fileType string) string {
	return fmt.Sprintf(`Generate a YamlAnchor pipeline as strict JSON.
Return only one JSON object. Do not wrap it in markdown.

Required schema:
{
  "name": "string",
  "on": { "push": { "branches": ["main"] }, "pull_request": { "branches": ["main"] } },
  "jobs": {
    "job-id": {
      "name": "string",
      "runs_on": "ubuntu-latest",
      "needs": ["optional-job-id"],
      "steps": [
        { "name": "Checkout", "uses": "actions/checkout@v4" },
        { "name": "Run something", "run": "shell command" }
      ]
    }
  }
}

Rules:
- Use YamlAnchor JSON field names, especially "runs_on".
- Every job must have runs_on and at least one step.
- Prefer common CI commands: go test ./..., npm ci, npm run lint, npm test, npm run build, pytest, docker build.
- Include checkout steps where useful.
- Do not include secrets or real credentials.
- Keep the pipeline practical and small.

Input type: %s
Input:
%s`, fileType, content)
}

func decodePipelineJSON(raw string) (*schema.Pipeline, error) {
	raw = strings.TrimSpace(raw)
	raw = strings.TrimPrefix(raw, "```json")
	raw = strings.TrimPrefix(raw, "```")
	raw = strings.TrimSuffix(raw, "```")
	raw = strings.TrimSpace(raw)

	var pipeline schema.Pipeline
	if err := json.Unmarshal([]byte(raw), &pipeline); err != nil {
		return nil, err
	}
	normalizePipeline(&pipeline)
	if err := pipeline.Validate(); err != nil {
		return nil, err
	}
	return &pipeline, nil
}

func generateOffline(content, fileType string) *schema.Pipeline {
	lower := strings.ToLower(content)
	analysis := analyzer.AnalyzeCode(content, fileType)

	pipeline := &schema.Pipeline{
		Name: "Generated Pipeline",
		On: map[string]interface{}{
			"push": map[string]interface{}{
				"branches": []string{"main"},
			},
			"pull_request": map[string]interface{}{
				"branches": []string{"main"},
			},
		},
		Jobs: make(map[string]*schema.Job),
	}

	if mentionsGo(lower, analysis) {
		pipeline.Jobs["backend-test"] = goJob()
	}
	if mentionsNode(lower, analysis) {
		pipeline.Jobs["frontend-build"] = nodeJob()
	}
	if mentionsPython(lower, analysis) {
		pipeline.Jobs["python-test"] = pythonJob()
	}
	if strings.Contains(lower, "docker") || strings.Contains(lower, "image") || strings.Contains(lower, "container") {
		needs := existingJobIDs(pipeline)
		pipeline.Jobs["docker-build"] = dockerJob(needs)
	}
	if strings.Contains(lower, "deploy") || strings.Contains(lower, "release") {
		needs := existingJobIDs(pipeline)
		pipeline.Jobs["deploy"] = deployJob(needs)
	}

	if len(pipeline.Jobs) == 0 {
		framework := analysis.Framework
		if framework == "" {
			framework = analysis.Language
		}
		pipeline.Jobs["build"] = jobForFramework(framework)
	}

	normalizePipeline(pipeline)
	return pipeline
}

func mentionsGo(lower string, analysis analyzer.AnalysisResult) bool {
	return analysis.Language == "go" || strings.Contains(lower, "go ") || strings.Contains(lower, "golang") || strings.Contains(lower, "go.mod")
}

func mentionsNode(lower string, analysis analyzer.AnalysisResult) bool {
	return analysis.Language == "nodejs" || analysis.Language == "javascript" || analysis.Framework == "react" ||
		strings.Contains(lower, "react") || strings.Contains(lower, "node") || strings.Contains(lower, "npm") || strings.Contains(lower, "frontend")
}

func mentionsPython(lower string, analysis analyzer.AnalysisResult) bool {
	return analysis.Language == "python" || strings.Contains(lower, "python") || strings.Contains(lower, "pytest") || strings.Contains(lower, "django") || strings.Contains(lower, "flask")
}

func existingJobIDs(pipeline *schema.Pipeline) []string {
	ids := make([]string, 0, len(pipeline.Jobs))
	for id := range pipeline.Jobs {
		ids = append(ids, id)
	}
	return ids
}

func goJob() *schema.Job {
	return &schema.Job{
		Name:   "Backend Test",
		RunsOn: "ubuntu-latest",
		Steps: []*schema.Step{
			{Name: "Checkout", Uses: "actions/checkout@v4"},
			{Name: "Setup Go", Uses: "actions/setup-go@v4"},
			{Name: "Download Dependencies", Run: "go mod download"},
			{Name: "Run Tests", Run: "go test ./..."},
			{Name: "Build", Run: "go build ./..."},
		},
	}
}

func nodeJob() *schema.Job {
	return &schema.Job{
		Name:   "Frontend Build",
		RunsOn: "ubuntu-latest",
		Steps: []*schema.Step{
			{Name: "Checkout", Uses: "actions/checkout@v4"},
			{Name: "Setup Node", Uses: "actions/setup-node@v3"},
			{Name: "Install Dependencies", Run: "npm ci"},
			{Name: "Lint", Run: "npm run lint --if-present"},
			{Name: "Test", Run: "npm test --if-present"},
			{Name: "Build", Run: "npm run build"},
		},
	}
}

func pythonJob() *schema.Job {
	return &schema.Job{
		Name:   "Python Test",
		RunsOn: "ubuntu-latest",
		Steps: []*schema.Step{
			{Name: "Checkout", Uses: "actions/checkout@v4"},
			{Name: "Setup Python", Uses: "actions/setup-python@v4"},
			{Name: "Install Dependencies", Run: "pip install -r requirements.txt"},
			{Name: "Run Tests", Run: "pytest"},
		},
	}
}

func dockerJob(needs []string) *schema.Job {
	return &schema.Job{
		Name:   "Docker Build",
		RunsOn: "ubuntu-latest",
		Needs:  needs,
		Steps: []*schema.Step{
			{Name: "Checkout", Uses: "actions/checkout@v4"},
			{Name: "Build Image", Run: "docker build -t app:latest ."},
		},
	}
}

func deployJob(needs []string) *schema.Job {
	return &schema.Job{
		Name:   "Deploy",
		RunsOn: "ubuntu-latest",
		Needs:  needs,
		If:     "github.ref == 'refs/heads/main'",
		Steps: []*schema.Step{
			{Name: "Deploy", Run: "echo 'Add deployment command here'"},
		},
	}
}

func jobForFramework(framework string) *schema.Job {
	switch framework {
	case "go":
		return goJob()
	case "nodejs", "javascript", "react":
		return nodeJob()
	case "python":
		return pythonJob()
	default:
		return &schema.Job{
			Name:   "Build",
			RunsOn: "ubuntu-latest",
			Steps: []*schema.Step{
				{Name: "Checkout", Uses: "actions/checkout@v4"},
				{Name: "Build", Run: "echo 'Add project build command here'"},
			},
		}
	}
}

func normalizePipeline(pipeline *schema.Pipeline) {
	if pipeline.Name == "" {
		pipeline.Name = "Generated Pipeline"
	}
	if pipeline.On == nil {
		pipeline.On = map[string]interface{}{
			"push": map[string]interface{}{"branches": []string{"main"}},
		}
	}
	if pipeline.Jobs == nil {
		pipeline.Jobs = make(map[string]*schema.Job)
	}
	for id, job := range pipeline.Jobs {
		if job == nil {
			pipeline.Jobs[id] = jobForFramework("")
			continue
		}
		if len(job.RunsOnLabels()) == 0 {
			job.RunsOn = "ubuntu-latest"
		}
		if len(job.Steps) == 0 {
			job.Steps = []*schema.Step{{Name: "Build", Run: "echo 'Add project build command here'"}}
		}
	}
}

// InferFileType guesses the analyzer input type from pasted text or a prompt.
func InferFileType(content string) string {
	lower := strings.ToLower(content)
	switch {
	case strings.Contains(lower, "package.json") || strings.Contains(lower, "npm ") || strings.Contains(lower, "node"):
		return "package.json"
	case strings.Contains(lower, "dockerfile") || strings.Contains(lower, "docker build"):
		return "dockerfile"
	case strings.Contains(lower, "go.mod"):
		return "go.mod"
	case strings.Contains(lower, "golang") || strings.Contains(lower, "go test"):
		return "go"
	case strings.Contains(lower, "react") || strings.Contains(lower, "jsx"):
		return "jsx"
	case strings.Contains(lower, "python") || strings.Contains(lower, "pytest") || strings.Contains(lower, "django") || strings.Contains(lower, "flask"):
		return "python"
	default:
		return "go"
	}
}
