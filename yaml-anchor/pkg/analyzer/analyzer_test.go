package analyzer

import (
	"testing"
)

func TestAnalyzeCode_Go(t *testing.T) {
	code := `package main
import (
	"fmt"
	"github.com/gin-gonic/gin"
)
func main() { fmt.Println("hi") }`

	result := AnalyzeCode(code, "go")
	if result.Language != "go" {
		t.Errorf("Expected language 'go', got %q", result.Language)
	}
	if result.Framework != "gin" {
		t.Errorf("Expected framework 'gin', got %q", result.Framework)
	}
}

func TestAnalyzeCode_JavaScript_React(t *testing.T) {
	code := `import React from 'react'
import { useState } from 'react'`

	result := AnalyzeCode(code, "js")
	if result.Language != "javascript" {
		t.Errorf("Expected language 'javascript', got %q", result.Language)
	}
	if result.Framework != "react" {
		t.Errorf("Expected framework 'react', got %q", result.Framework)
	}
}

func TestAnalyzeCode_JavaScript_NextJS(t *testing.T) {
	code := `import { useRouter } from 'next/router'`
	result := AnalyzeCode(code, "js")
	if result.Framework != "nextjs" {
		t.Errorf("Expected framework 'nextjs', got %q", result.Framework)
	}
}

func TestAnalyzeCode_Python_Django(t *testing.T) {
	code := `import django
from django.db import models`

	result := AnalyzeCode(code, "python")
	if result.Language != "python" {
		t.Errorf("Expected language 'python', got %q", result.Language)
	}
	if result.Framework != "django" {
		t.Errorf("Expected framework 'django', got %q", result.Framework)
	}
}

func TestAnalyzeCode_Dockerfile_Go(t *testing.T) {
	code := `FROM golang:1.21-alpine
COPY . .
RUN go build -o app main.go`

	result := AnalyzeCode(code, "dockerfile")
	if result.Language != "dockerfile" {
		t.Errorf("Expected language 'dockerfile', got %q", result.Language)
	}
	if result.Framework != "go" {
		t.Errorf("Expected framework 'go' for Go Dockerfile, got %q", result.Framework)
	}
}

func TestAnalyzeCode_PackageJSON(t *testing.T) {
	code := `{"name":"app","scripts":{"test":"jest","build":"vite build"},"dependencies":{"react":"^18"}}`
	result := AnalyzeCode(code, "package.json")
	if result.Language != "nodejs" {
		t.Errorf("Expected language 'nodejs', got %q", result.Language)
	}
}

func TestAnalyzeCode_GoMod(t *testing.T) {
	code := `module myapp
go 1.21
require (
	github.com/spf13/cobra v1.7.0
)`
	result := AnalyzeCode(code, "go.mod")
	if result.Language != "go" {
		t.Errorf("Expected language 'go', got %q", result.Language)
	}
	if len(result.Suggestions) == 0 {
		t.Error("Expected suggestions for go.mod with require block")
	}
}

func TestAnalyzeCode_Unknown(t *testing.T) {
	result := AnalyzeCode("random content", "xyz")
	if result.Language != "unknown" {
		t.Errorf("Expected language 'unknown', got %q", result.Language)
	}
}
