package analyzer

import (
	"regexp"
	"strings"
)

// AnalysisResult contains code analysis results
type AnalysisResult struct {
	Language string            `json:"language"`
	Framework string           `json:"framework"`
	Dependencies map[string]string `json:"dependencies"`
	Suggestions []string        `json:"suggestions"`
}

// AnalyzeCode analyzes code content and returns analysis result
func AnalyzeCode(content, fileType string) AnalysisResult {
	switch fileType {
	case "go":
		return analyzeGo(content)
	case "js", "jsx", "ts", "tsx":
		return analyzeJavaScript(content)
	case "python", "py":
		return analyzePython(content)
	case "package.json":
		return analyzePackageJSON(content)
	case "dockerfile", "Dockerfile":
		return analyzeDockerfile(content)
	case "go.mod":
		return analyzeGoMod(content)
	default:
		return AnalysisResult{
			Language: "unknown",
			Dependencies: make(map[string]string),
		}
	}
}

func analyzeGo(content string) AnalysisResult {
	result := AnalysisResult{
		Language: "go",
		Dependencies: make(map[string]string),
	}

	// Detect framework
	if strings.Contains(content, "github.com/gin-gonic/gin") {
		result.Framework = "gin"
		result.Suggestions = append(result.Suggestions, "go build, go test, docker build suggested")
	}
	if strings.Contains(content, "github.com/gorilla/mux") {
		result.Framework = "gorilla"
		result.Suggestions = append(result.Suggestions, "HTTP routing detected")
	}

	// Extract imports
	importRegex := regexp.MustCompile(`import\s*\([\s\S]*?\)`)
	imports := importRegex.FindString(content)
	if imports != "" {
		importLines := strings.Split(imports, "\n")
		for _, line := range importLines {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "\"github.com") {
				lib := strings.Trim(line, "\"")
				result.Dependencies[lib] = "remote"
			}
		}
	}

	if len(result.Dependencies) == 0 {
		result.Suggestions = append(result.Suggestions, "go test ./...")
		result.Suggestions = append(result.Suggestions, "go build -o bin/app main.go")
	}

	return result
}

func analyzeJavaScript(content string) AnalysisResult {
	result := AnalysisResult{
		Language: "javascript",
		Dependencies: make(map[string]string),
	}

	// Detect frameworks
	if strings.Contains(content, "from 'react'") || strings.Contains(content, 'from "react"') {
		result.Framework = "react"
		result.Suggestions = append(result.Suggestions, "npm install")
		result.Suggestions = append(result.Suggestions, "npm run lint")
		result.Suggestions = append(result.Suggestions, "npm run build")
	}

	if strings.Contains(content, "from 'vue'") || strings.Contains(content, 'from "vue"') {
		result.Framework = "vue"
		result.Suggestions = append(result.Suggestions, "npm ci && npm run build")
	}

	if strings.Contains(content, "next/") {
		result.Framework = "nextjs"
		result.Suggestions = append(result.Suggestions, "npm install && npm run build")
	}

	if len(result.Suggestions) == 0 {
		result.Suggestions = append(result.Suggestions, "npm install")
		result.Suggestions = append(result.Suggestions, "npm test")
	}

	return result
}

func analyzePython(content string) AnalysisResult {
	result := AnalysisResult{
		Language: "python",
		Dependencies: make(map[string]string),
	}

	if strings.Contains(content, "import django") {
		result.Framework = "django"
		result.Suggestions = append(result.Suggestions, "pip install -r requirements.txt")
		result.Suggestions = append(result.Suggestions, "python manage.py test")
	}

	if strings.Contains(content, "import flask") || strings.Contains(content, "from flask") {
		result.Framework = "flask"
		result.Suggestions = append(result.Suggestions, "pip install -r requirements.txt")
		result.Suggestions = append(result.Suggestions, "pytest")
	}

	if len(result.Suggestions) == 0 {
		result.Suggestions = append(result.Suggestions, "python -m pytest")
	}

	return result
}

func analyzePackageJSON(content string) AnalysisResult {
	result := AnalysisResult{
		Language: "nodejs",
		Dependencies: make(map[string]string),
	}

	if strings.Contains(content, "\"react\"") {
		result.Framework = "react"
	}

	if strings.Contains(content, "\"test\"") {
		result.Suggestions = append(result.Suggestions, "npm test")
	}

	if strings.Contains(content, "\"build\"") {
		result.Suggestions = append(result.Suggestions, "npm run build")
	}

	if strings.Contains(content, "\"dev\"") {
		result.Suggestions = append(result.Suggestions, "npm run dev")
	}

	result.Suggestions = append(result.Suggestions, "npm ci")

	return result
}

func analyzeDockerfile(content string) AnalysisResult {
	result := AnalysisResult{
		Language: "dockerfile",
		Dependencies: make(map[string]string),
	}

	if strings.Contains(strings.ToLower(content), "from golang") {
		result.Framework = "go"
	}
	if strings.Contains(strings.ToLower(content), "from node") {
		result.Framework = "nodejs"
	}
	if strings.Contains(strings.ToLower(content), "from python") {
		result.Framework = "python"
	}

	result.Suggestions = append(result.Suggestions, "docker build -t myapp:latest .")
	result.Suggestions = append(result.Suggestions, "docker push")

	return result
}

func analyzeGoMod(content string) AnalysisResult {
	result := AnalysisResult{
		Language: "go",
		Dependencies: make(map[string]string),
	}

	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "require") {
			result.Suggestions = append(result.Suggestions, "go mod tidy")
			result.Suggestions = append(result.Suggestions, "go build")
			result.Suggestions = append(result.Suggestions, "go test ./...")
		}
	}

	return result
}
