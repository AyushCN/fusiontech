.PHONY: build test-go coverage lint clean help

## ── Build ──────────────────────────────────────────────────────────────────

build: ## Build the anchor CLI binary
	cd yaml-anchor && go build -o anchor main.go
	@echo "✅  Binary: yaml-anchor/anchor"

build-ui: ## Build the Studio web UI for production
	cd yaml-anchor/ui && npm run build
	@echo "✅  UI dist: yaml-anchor/ui/dist"

## ── Development ────────────────────────────────────────────────────────────

run-server: ## Start the Go API server (port 8080)
	cd yaml-anchor && go run main.go server

run-ui: ## Start the Vite dev server (port 5173)
	cd yaml-anchor/ui && npm run dev

## ── Testing ────────────────────────────────────────────────────────────────

test-go: ## Run all Go package tests
	cd yaml-anchor && go test ./pkg/... -v -race 2>&1

coverage: ## Generate Go test coverage report
	cd yaml-anchor && go test ./pkg/... -coverprofile=coverage.out
	cd yaml-anchor && go tool cover -func=coverage.out
	@echo "💡 Run 'go tool cover -html=yaml-anchor/coverage.out' for HTML view"

## ── Code Quality ───────────────────────────────────────────────────────────

fmt: ## Format all Go source files
	cd yaml-anchor && gofmt -w ./...

vet: ## Run go vet
	cd yaml-anchor && go vet ./...

## ── Cleanup ────────────────────────────────────────────────────────────────

clean: ## Remove build artifacts
	rm -f yaml-anchor/anchor yaml-anchor/coverage.out
	@echo "🧹 Cleaned"

## ── Help ───────────────────────────────────────────────────────────────────

help: ## Show this help message
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-16s\033[0m %s\n", $$1, $$2}'

.DEFAULT_GOAL := help
