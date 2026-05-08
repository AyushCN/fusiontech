package cmd

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/spf13/cobra"
	"yaml-anchor/pkg/analyzer"
	"yaml-anchor/pkg/generator"
	"yaml-anchor/pkg/schema"
)

var (
	port string
)

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start YamlAnchor HTTP API server",
	Long: `Starts a local HTTP server that provides REST API endpoints
for pipeline generation and analysis.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("🚀 Starting YamlAnchor API Server on port %s...\n", port)
		fmt.Println("📍 Available endpoints:")
		fmt.Println("   GET  /health         - Health check")
		fmt.Println("   POST /api/analyze    - Analyze code")
		fmt.Println("   POST /api/generate   - Generate pipeline")
		fmt.Println("   POST /api/validate   - Validate config")
		fmt.Println("")

		// Register routes
		http.HandleFunc("/", corsMiddleware(handleDocs))
		http.HandleFunc("/health", corsMiddleware(handleHealth))
		http.HandleFunc("/api/analyze", corsMiddleware(handleAnalyze))
		http.HandleFunc("/api/generate", corsMiddleware(handleGenerate))
		http.HandleFunc("/api/validate", corsMiddleware(handleValidate))

		if err := http.ListenAndServe(":"+port, nil); err != nil {
			log.Fatalf("❌ Server error: %v", err)
		}
	},
}

func init() {
	serverCmd.Flags().StringVarP(&port, "port", "p", "8080", "Port to listen on")
	rootCmd.AddCommand(serverCmd)
}

// Request/Response types
type AnalyzeRequest struct {
	Code     string `json:"code"`
	FileType string `json:"file_type"`
}

type GenerateRequest struct {
	Code     string `json:"code"`
	FileType string `json:"file_type"`
}

type ValidateRequest struct {
	Pipeline *schema.Pipeline `json:"pipeline"`
}

type ErrorResponse struct {
	Error  string `json:"error"`
	Status int    `json:"status"`
}

type ValidationResponse struct {
	Valid  bool     `json:"valid"`
	Errors []string `json:"errors"`
}

// CORS middleware
func corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next(w, r)
	}
}

// Handlers
func handleDocs(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	j := json.NewEncoder(w)
	j.Encode(map[string]interface{}{
		"name": "YamlAnchor API",
		"version": "0.1.0",
		"endpoints": []map[string]string{
			{"path": "/health", "method": "GET"},
			{"path": "/api/analyze", "method": "POST"},
			{"path": "/api/generate", "method": "POST"},
			{"path": "/api/validate", "method": "POST"},
		},
	})
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "ok",
		"version": "0.1.0",
	})
}

func handleAnalyze(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req AnalyzeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Code == "" {
		respondError(w, http.StatusBadRequest, "code field is required")
		return
	}

	if req.FileType == "" {
		respondError(w, http.StatusBadRequest, "file_type field is required")
		return
	}

	result := analyzer.AnalyzeCode(req.Code, req.FileType)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func handleGenerate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req GenerateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Code == "" {
		respondError(w, http.StatusBadRequest, "code field is required")
		return
	}

	// Analyze and generate
	analysis := analyzer.AnalyzeCode(req.Code, req.FileType)

	// Create basic pipeline
	pipeline := &schema.Pipeline{
		Name: "Generated Pipeline",
		On: map[string]interface{}{
			"push": map[string]interface{}{
				"branches": []string{"main"},
			},
		},
		Jobs: make(map[string]*schema.Job),
	}

	// Add jobs based on analysis
	if analysis.Framework != "" {
		job := createJobForFramework(analysis.Framework)
		pipeline.Jobs["build"] = job
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(pipeline)
}

func handleValidate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req ValidateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	var errStrings []string
	if req.Pipeline != nil {
		valErrs := generator.ValidatePipeline(req.Pipeline)
		for _, e := range valErrs {
			errStrings = append(errStrings, e.Error())
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ValidationResponse{
		Valid:  len(errStrings) == 0,
		Errors: errStrings,
	})
}

func createJobForFramework(framework string) *schema.Job {
	var steps []*schema.Step

	switch framework {
	case "go":
		steps = []*schema.Step{
			{Name: "Checkout", Uses: "actions/checkout@v4"},
			{Name: "Setup Go", Uses: "actions/setup-go@v4"},
			{Name: "Run Tests", Run: "go test ./..."},
			{Name: "Build", Run: "go build -o bin/app main.go"},
		}
	case "nodejs", "react":
		steps = []*schema.Step{
			{Name: "Checkout", Uses: "actions/checkout@v4"},
			{Name: "Setup Node", Uses: "actions/setup-node@v3"},
			{Name: "Install", Run: "npm ci"},
			{Name: "Lint", Run: "npm run lint"},
			{Name: "Build", Run: "npm run build"},
		}
	case "python":
		steps = []*schema.Step{
			{Name: "Checkout", Uses: "actions/checkout@v4"},
			{Name: "Setup Python", Uses: "actions/setup-python@v4"},
			{Name: "Install Dependencies", Run: "pip install -r requirements.txt"},
			{Name: "Run Tests", Run: "pytest"},
		}
	default:
		steps = []*schema.Step{
			{Name: "Checkout", Uses: "actions/checkout@v4"},
			{Name: "Run", Run: "echo 'Building...'"},
		}
	}

	return &schema.Job{
		Name:     strings.Title(framework) + " Build",
		RunsOn:   "ubuntu-latest",
		Steps:    steps,
	}
}

func respondError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(ErrorResponse{
		Error:  message,
		Status: statusCode,
	})
}
