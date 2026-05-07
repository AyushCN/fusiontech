package scanner

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Severity represents the importance of a finding.
type Severity string

const (
	SeverityHigh   Severity = "HIGH"
	SeverityMedium Severity = "MEDIUM"
	SeverityLow    Severity = "LOW"
)

// Finding represents a single security issue detected.
type Finding struct {
	File        string   `json:"file"`
	Line        int      `json:"line"`
	Pattern     string   `json:"pattern"`
	Severity    Severity `json:"severity"`
	Preview     string   `json:"preview"`
	Description string   `json:"description"`
}

// ScanOptions configures the scanner's behavior.
type ScanOptions struct {
	Recursive     bool
	EntropyLimit  float64
	OutputFormat  string // "human", "json", "github"
	IncludeDotEnv bool
}

var defaultPatterns = map[string]*regexp.Regexp{
	"AWS Access Key": regexp.MustCompile(`AKIA[0-9A-Z]{16}`),
	"GitHub Token":   regexp.MustCompile(`(ghp|gho|ghu|ghs|ghr)_[A-Za-z0-9_]{36}`),
	"Bearer Token":   regexp.MustCompile(`Bearer [a-zA-Z0-9\-\._~+/]+=*`),
}

// Scan crawls the given path and looks for secrets and sensitive files.
func Scan(root string, opts ScanOptions) ([]Finding, error) {
	var findings []Finding

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			if !opts.Recursive && path != root {
				return filepath.SkipDir
			}
			// Skip .git
			if info.Name() == ".git" {
				return filepath.SkipDir
			}
			return nil
		}

		// Check for .env files
		if opts.IncludeDotEnv && (info.Name() == ".env" || strings.HasSuffix(info.Name(), ".env")) {
			findings = append(findings, Finding{
				File:        path,
				Pattern:     "Sensitive File",
				Severity:    SeverityHigh,
				Description: "Detected .env file which may contain secrets. Ensure this is not committed.",
			})
		}

		// Only scan text-like files for contents
		ext := filepath.Ext(path)
		if isScanable(ext) {
			fileFindings, err := scanFile(path, opts)
			if err == nil {
				findings = append(findings, fileFindings...)
			}
		}

		return nil
	})

	return findings, err
}

func isScanable(ext string) bool {
	scanable := map[string]bool{
		".go":   true, ".js":   true, ".ts":   true, ".py": true,
		".yaml": true, ".yml":  true, ".json": true, ".md": true,
		".env":  true, ".txt":  true, ".sh":   true,
	}
	return scanable[strings.ToLower(ext)]
}

func scanFile(path string, opts ScanOptions) ([]Finding, error) {
	var findings []Finding
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(data), "\n")
	for i, line := range lines {
		// 1. Regex Patterns
		for name, pattern := range defaultPatterns {
			if pattern.MatchString(line) {
				findings = append(findings, Finding{
					File:     path,
					Line:     i + 1,
					Pattern:  name,
					Severity: SeverityHigh,
					Preview:  redact(line, pattern),
				})
			}
		}

		// 2. Entropy Detection
		if opts.EntropyLimit > 0 {
			// Extract words/strings that look like potential tokens
			words := regexp.MustCompile(`[a-zA-Z0-9\-\._~+/]{8,}`).FindAllString(line, -1)
			for _, word := range words {
				e := shannonEntropy(word)
				if e > opts.EntropyLimit {
					// Skip common non-secret high entropy strings if needed
					findings = append(findings, Finding{
						File:        path,
						Line:        i + 1,
						Pattern:     "High Entropy String",
						Severity:    SeverityMedium,
						Description: fmt.Sprintf("Entropy: %.2f", e),
						Preview:     redactPlain(line, word),
					})
				}
			}
		}
	}

	return findings, nil
}

func shannonEntropy(data string) float64 {
	if data == "" {
		return 0
	}
	charCounts := make(map[rune]int)
	for _, r := range data {
		charCounts[r]++
	}
	var entropy float64
	length := float64(len(data))
	for _, count := range charCounts {
		p := float64(count) / length
		entropy -= p * math.Log2(p)
	}
	return entropy
}

func redact(line string, pattern *regexp.Regexp) string {
	return pattern.ReplaceAllString(line, "[REDACTED]")
}

func redactPlain(line, word string) string {
	return strings.ReplaceAll(line, word, "[REDACTED]")
}

// FormatFindings formats the findings according to the requested format.
func FormatFindings(findings []Finding, format string) string {
	switch format {
	case "json":
		data, _ := json.MarshalIndent(findings, "", "  ")
		return string(data)
	case "github":
		var sb strings.Builder
		for _, f := range findings {
			// GitHub Action annotation format: ::error file={name},line={line},col={col}::{message}
			sb.WriteString(fmt.Sprintf("::error file=%s,line=%d::[YamlAnchor] %s detected: %s\n", f.File, f.Line, f.Pattern, f.Description))
		}
		return sb.String()
	default:
		var sb strings.Builder
		if len(findings) == 0 {
			return "No security issues found. ✅"
		}
		for _, f := range findings {
			color := "\033[31m" // Red
			if f.Severity == SeverityMedium {
				color = "\033[33m" // Yellow
			}
			sb.WriteString(fmt.Sprintf("%s[%s]\033[0m %s:%d - %s\n", color, f.Severity, f.File, f.Line, f.Pattern))
			if f.Preview != "" {
				sb.WriteString(fmt.Sprintf("   > %s\n", strings.TrimSpace(f.Preview)))
			}
		}
		return sb.String()
	}
}
