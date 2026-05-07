# YamlAnchor Setup Guide

## Prerequisites

- **Go** 1.21 or higher
- **Node.js** 18+ with npm
- **Docker** — required for `anchor local` and `anchor exec` (uses Dagger)
- **Git** — for `anchor scan --install-hook`

## Quick Start

```bash
git clone https://github.com/AyushCN/fusiontech.git
cd fusiontech/yaml-anchor
go build -o anchor main.go
./anchor --help
```

## Installation

### 1. Build the CLI

```bash
cd yaml-anchor
go build -o anchor main.go

# Optionally move to your PATH
sudo mv anchor /usr/local/bin/
```

### 2. Setup the Web UI (Optional)

```bash
cd yaml-anchor/ui
npm install
```

### 3. Initialize a Project

Navigate to any project directory and run:

```bash
anchor init          # Auto-detects your stack and scaffolds anchor.yaml
anchor scan .        # Scan for hardcoded secrets
anchor generate      # Emit .github/workflows/main.yml
anchor local         # Simulate the pipeline locally in Docker
```

## Running the Full Stack (CLI + UI + API)

**Option A: Single command (recommended)**
```bash
cd yaml-anchor/ui
npm run dev:full     # Starts Go API server on :8080 + Vite UI on :5173
```

**Option B: Two terminals**

Terminal 1 — API Server:
```bash
cd yaml-anchor
go run main.go server --port 8080
```

Terminal 2 — Web UI:
```bash
cd yaml-anchor/ui
npm run dev
```

Access:
- **Studio UI**: http://localhost:5173
- **API Server**: http://localhost:8080
- **Health Check**: http://localhost:8080/health

## CLI Commands Reference

| Command | Description |
|:---|:---|
| `anchor init` | Detect stack, scaffold `anchor.yaml` |
| `anchor scan [path]` | Scan for hardcoded secrets |
| `anchor scan --install-hook` | Install pre-commit git hook |
| `anchor generate` | Validate and export `.github/workflows/main.yml` |
| `anchor local` | Run pipeline locally via Dagger (launches TUI) |
| `anchor exec <job>` | Drop into interactive shell inside the runner |
| `anchor server` | Start the HTTP API server for the Studio |
| `anchor clean` | Prune Dagger/Docker caches |

## Running Tests

```bash
# All Go tests
make test-go

# With coverage report
make coverage

# Build binary + UI
make build
```

## Docker Notes

YamlAnchor uses [Dagger](https://dagger.io/) for local simulation. Dagger requires Docker Desktop or a Docker daemon. The first run will pull base images (`ubuntu:22.04`, `golang:1.26`, `node:18`) which may take a minute.

## Environment Variables (UI)

Create `yaml-anchor/ui/.env` from the example:
```bash
cp yaml-anchor/ui/.env.example yaml-anchor/ui/.env
```

| Variable | Default | Description |
|:---|:---|:---|
| `VITE_API_URL` | `http://localhost:8080` | Go backend API URL |
