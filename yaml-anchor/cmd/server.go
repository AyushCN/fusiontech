package cmd

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/spf13/cobra"
	"yaml-anchor/pkg/aigen"
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
	Code         string            `json:"code"`
	FileType     string            `json:"file_type"`
	Prompt       string            `json:"prompt"`
	ProjectTree  []string          `json:"project_tree"`
	ContextFiles map[string]string `json:"context_files"`
	ExistingCI   map[string]string `json:"existing_ci"`
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
		"name":    "YamlAnchor API",
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
		"status":  "ok",
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

	content := strings.TrimSpace(req.Code)
	if content == "" {
		content = strings.TrimSpace(req.Prompt)
	}
	content = appendProjectContext(content, req)
	if content == "" {
		respondError(w, http.StatusBadRequest, "code or prompt field is required")
		return
	}
	fileType := strings.TrimSpace(req.FileType)
	if fileType == "" {
		fileType = aigen.InferFileType(content)
	}

	pipeline, source, err := aigen.Generate(r.Context(), content, fileType)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-YamlAnchor-Generator", source)
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

func appendProjectContext(content string, req GenerateRequest) string {
	var b strings.Builder
	if strings.TrimSpace(content) != "" {
		b.WriteString(strings.TrimSpace(content))
		b.WriteString("\n\n")
	}
	if len(req.ProjectTree) > 0 {
		b.WriteString("Project tree:\n")
		for _, path := range req.ProjectTree {
			b.WriteString("- ")
			b.WriteString(path)
			b.WriteString("\n")
		}
		b.WriteString("\n")
	}
	if len(req.ContextFiles) > 0 {
		b.WriteString("Context files:\n")
		for path, data := range req.ContextFiles {
			b.WriteString("### ")
			b.WriteString(path)
			b.WriteString("\n")
			b.WriteString(data)
			b.WriteString("\n\n")
		}
	}
	if len(req.ExistingCI) > 0 {
		b.WriteString("Existing CI workflows:\n")
		for path, data := range req.ExistingCI {
			b.WriteString("### ")
			b.WriteString(path)
			b.WriteString("\n")
			b.WriteString(data)
			b.WriteString("\n\n")
		}
	}
	return strings.TrimSpace(b.String())
}

func respondError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(ErrorResponse{
		Error:  message,
		Status: statusCode,
	})
}
