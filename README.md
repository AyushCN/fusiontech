<div align="center">

# ⚓ YamlAnchor

### CI/CD Pipelines as Type-Safe Code

**Stop pushing blind. Start anchoring.**

[![CI](https://github.com/AyushCN/fusiontech/actions/workflows/ci.yml/badge.svg)](https://github.com/AyushCN/fusiontech/actions/workflows/ci.yml)
[![Release](https://github.com/AyushCN/fusiontech/actions/workflows/release.yml/badge.svg)](https://github.com/AyushCN/fusiontech/releases)
[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-green)](LICENSE)

</div>

---

> **The problem**: CI/CD pipelines are written in YAML — a format with no types, weak local feedback, and late failure signals. Errors are often discovered only after a remote push. YamlAnchor gives you a safer preflight path.

YamlAnchor loads an `anchor.yaml` pipeline into a typed Go model, validates dependencies and step structure, scans for secrets, can simulate common shell steps locally with Docker/Dagger, and exports CI configuration. No more "push → wait → fail → cry."

---

## ✨ Why YamlAnchor?

| Traditional CI/CD | YamlAnchor |
|:---|:---|
| ❌ Errors found after a remote push | ✅ Caught earlier with typed validation |
| ❌ No circular dependency detection | ✅ Compile-time DAG cycle detection |
| ❌ Boilerplate steps repeated everywhere | ✅ **Blueprints** auto-expand common stacks |
| ❌ Local CI feedback is hard | ✅ **Dagger simulation** runs common steps locally |
| ❌ Blind log scrolling | ✅ **Pulse Dashboard** (Bubbletea TUI) |
| ❌ Hardcoded secrets committed accidentally | ✅ **Secret Scanner** blocks export on `HIGH`/`CRITICAL` |
| ❌ CI config is hard to inspect visually | ✅ **Studio UI** generates, previews, and graphs pipelines |

---

## 🚀 Features

### ⚡ Pulse Dashboard (TUI)
A live Bubbletea terminal dashboard replaces raw log scrolling. Real-time job progress, color-coded status, and concurrent step tracking via Go channels.

### 🏗️ Blueprints
Write less boilerplate. A single `blueprint: "go-app"` line expands into a fully-configured checkout → setup → build → test pipeline. Supported: `go-app`, `node-app`.

### 🛡️ Action Shims & Local Simulation
YamlAnchor simulates common workflow behavior inside Dagger containers:
- `actions/checkout` → mounts your local workspace
- `actions/setup-go` → uses a Go-capable base image
- `actions/setup-node` → uses a Node-capable base image
- Custom shims via `.anchor/plugins/<owner>/<action>.sh`

Local simulation is intentionally approximate: it is a preflight sanity check, not a perfect replacement for GitHub-hosted runners, macOS/Windows runners, marketplace action internals, cloud permissions, artifacts, or deployment environments.

### 🔗 DAG Validation
Builds a mathematical Directed Acyclic Graph of your `needs:` dependencies. Detects circular dependencies and invalid references at load-time, before a single container starts.

### 🔐 Secret Scanner
Multi-pattern + Shannon entropy secret detection. Blocks YAML export on `HIGH`/`CRITICAL` findings. Prints warnings for `MEDIUM`/`LOW`. Supports:
- AWS Access Keys (`AKIA...`)
- GitHub Tokens (`ghp_`, `ghs_`, etc.)
- Slack Tokens (`xoxb-`, `xoxp-`)
- Azure JWTs
- SSH Private Keys
- Bearer Tokens
- Password assignments

### 🔢 Multi-Dimensional Matrix
The `strategy.matrix` block supports any number of dimensions. Keys are sorted for deterministic output. All combinations are expanded into named sub-runs with env var injection.

### 📊 Telemetry & Cost Insights
Every local run generates a telemetry report with actual vs. estimated remote CI time and dollar savings based on GitHub runner pricing ($0.008/minute).

### 🎨 YamlAnchor Studio
A React/Vite workbench for building and inspecting pipelines:
- **Pipeline Intake** — describe your stack or paste config/code
- **Keyless generation** — calls `anchor server` at `/api/generate`
- **Local LLM optional** — uses Ollama on your machine when available
- **Offline fallback** — deterministic generator when no model is running
- **anchor.yaml Preview** — synchronized YAML with copy/download actions
- **Flow Trace** — SVG job/step graph with inline fault hints

No API key is required. The generator first tries a local Ollama model (`llama3.2` by default) and falls back to an in-process generator for Go, Node/React, Python, Docker, deploy, and mixed-stack prompts.

---

## 🛠️ Installation

### Option A — Binary (recommended)

Download the latest release for your platform:

```bash
# Linux (amd64)
curl -L https://github.com/AyushCN/fusiontech/releases/latest/download/anchor_linux_amd64.tar.gz | tar xz
sudo mv anchor /usr/local/bin/

# macOS (Apple Silicon)
curl -L https://github.com/AyushCN/fusiontech/releases/latest/download/anchor_darwin_arm64.tar.gz | tar xz
sudo mv anchor /usr/local/bin/
```

### Option B — Build from source

```bash
git clone https://github.com/AyushCN/fusiontech.git
cd fusiontech/yaml-anchor
go build -o anchor main.go
sudo mv anchor /usr/local/bin/
```

### Prerequisites

| Tool | Required For |
|:---|:---|
| Go 1.21+ | CLI + API server |
| Docker Desktop | `anchor local` (Dagger simulation) |
| Node.js 18+ | YamlAnchor Studio (`ui/`) |

---

## 📦 Quick Start

**1. Scaffold a pipeline for your project:**
```bash
anchor init
```
Auto-detects your stack and writes `anchor.yaml`. It scans project context such as `package.json`, `go.mod`, `Dockerfile`, `requirements.txt`, `pyproject.toml`, existing `.github/workflows`, and a lightweight project tree. Mixed projects such as Go + React + Docker produce multiple jobs.

**2. Validate without writing:**
```bash
anchor generate --dry-run
```

**3. Generate the GitHub Actions workflow:**
```bash
anchor generate
# → .github/workflows/main.yml
```

**4. Simulate locally (requires Docker):**
```bash
anchor local
```

**5. Scan for secrets:**
```bash
anchor scan ./
```

**6. Run YamlAnchor Studio:**
```bash
# Terminal 1: backend API
cd yaml-anchor
go run main.go server

# Terminal 2: React UI
cd yaml-anchor/ui
npm install
npm run dev
```

Open the Vite URL printed in the terminal, usually `http://localhost:5173/`. The UI proxies `/api` and `/health` to the Go backend on port `8080`.

**Optional: use a local LLM with no API key**
```bash
# Install/run Ollama separately, then pull a model:
ollama pull llama3.2

# Start YamlAnchor normally:
cd yaml-anchor
go run main.go server
```

YamlAnchor will call `http://localhost:11434/api/generate` automatically. You can change the model or disable local LLM attempts:

```bash
YAML_ANCHOR_MODEL=mistral go run main.go server
YAML_ANCHOR_LLM=off go run main.go server
```

**Generate with project context over the API**
```bash
curl -X POST http://localhost:8080/api/generate \
  -H 'Content-Type: application/json' \
  -d '{
    "prompt": "Generate CI for this repo",
    "project_tree": ["go.mod", "ui/package.json", "Dockerfile"],
    "context_files": {
      "go.mod": "module example.com/app\n\ngo 1.22\n",
      "ui/package.json": "{\"scripts\":{\"lint\":\"eslint .\",\"test\":\"vitest\",\"build\":\"vite build\"},\"dependencies\":{\"react\":\"latest\"}}",
      "Dockerfile": "FROM node:20"
    },
    "existing_ci": {
      ".github/workflows/ci.yml": "name: old-ci"
    }
  }'
```

---

## 📋 CLI Reference

| Command | Description |
|:---|:---|
| `anchor init` | Auto-detect stack, scaffold `anchor.yaml` |
| `anchor generate` | Validate + scan + export `.github/workflows/main.yml` |
| `anchor generate --dry-run` | Validate only — no files written |
| `anchor local` | Run pipeline in Dagger with live Pulse TUI |
| `anchor exec <job>` | Drop into an interactive shell inside the runner container |
| `anchor scan <path>` | Scan files for hardcoded secrets |
| `anchor server` | Start the REST API server (for Studio UI) |
| `anchor version` | Print version, commit, Go version, OS/arch |
| `anchor clean` | Prune Dagger/Docker cache |

**Global flags** available on every command:

```
-c, --config string   Path to anchor.yaml  (default: anchor.yaml)
-v, --verbose         Enable debug-level structured logging
```

---

## 📄 anchor.yaml Reference

```yaml
name: "My App"

jobs:
  build:
    blueprint: "go-app"   # auto-expands to checkout + setup-go + build + test

  lint:
    runs-on: ubuntu-latest
    steps:
      - name: Lint
        run: golangci-lint run ./...

  deploy:
    runs-on: ubuntu-latest
    needs: [build, lint]  # DAG: only runs if both pass
    if: github.ref == 'refs/heads/main'
    steps:
      - name: Deploy
        run: ./scripts/deploy.sh
    env:
      STAGE: production
```

**Matrix strategy:**
```yaml
jobs:
  test:
    runs-on: ubuntu-latest
    strategy:
      matrix:
        go-version: ["1.20", "1.21", "1.22"]
        os: ["ubuntu-latest", "macos-latest"]
    steps:
      - run: go test ./...
```
Produces 6 expanded jobs: `test (1.20, ubuntu-latest)`, `test (1.20, macos-latest)`, etc.

---

## 🏗️ Project Structure

```
fusiontech/
├── yaml-anchor/
│   ├── cmd/                    # CLI commands (Cobra)
│   │   ├── root.go             # Global --config, --verbose flags
│   │   ├── generate.go         # Export + secret blocking + --dry-run
│   │   ├── local.go            # Dagger + TUI integration
│   │   ├── scan.go             # Standalone scanner + git hooks
│   │   ├── server.go           # REST API server
│   │   ├── init.go             # Stack-aware scaffolding
│   │   ├── exec.go             # Interactive debug shell
│   │   ├── version.go          # Build-time version info
│   │   └── clean.go            # Cache management
│   ├── pkg/
│   │   ├── schema/             # Type-safe Pipeline IR + DAG validation
│   │   ├── config/             # YAML loader + multi-dim matrix expansion
│   │   ├── blueprints/         # Blueprint → job step expansion
│   │   ├── detector/           # Stack detection (go.mod, package.json)
│   │   ├── analyzer/           # Code analysis for Studio AI
│   │   ├── generator/          # YAML export + secret scanning
│   │   ├── scanner/            # Multi-pattern + entropy secret scanner
│   │   ├── debugger/           # Pattern-based error analysis
│   │   ├── simulator/          # Dagger engine + action shims
│   │   ├── tui/                # Bubbletea Pulse Dashboard
│   │   ├── errors/             # ConfigError, ValidationError, SecurityError
│   │   ├── validator/          # Job ID, runner, cron, step validation
│   │   └── logger/             # Structured leveled logger (color + file)
│   ├── ui/                     # YamlAnchor Studio (React + Vite)
│   ├── examples/               # Real anchor.yaml configs
│   └── anchor.yaml             # This project's own pipeline
├── vscode-anchor/              # VS Code extension scaffold
├── .github/
│   ├── workflows/ci.yml        # Test + lint on every push/PR
│   └── workflows/release.yml   # GoReleaser multi-platform builds on tags
├── Makefile                    # make build | test-go | coverage | lint
├── .goreleaser.yaml            # Multi-platform release config
├── CONTRIBUTING.md             # Contribution guide
├── SETUP.md                    # Full installation guide
└── API_DOCS.md                 # REST API reference
```

---

## 🧪 Development

```bash
# From the repository root, go.work points at ./yaml-anchor
go test ./yaml-anchor/...

# Run all tests
make test-go

# Run with race detector
cd yaml-anchor
go test -race ./pkg/...

# Build binary
make build

# Show coverage
make coverage

# Lint
make lint

# UI checks
cd yaml-anchor/ui
npm run lint
npm run build
```

See [`CONTRIBUTING.md`](CONTRIBUTING.md) for the full contribution guide.

---

## 🗺️ Roadmap

- [x] Type-safe Go IR (`pkg/schema`)
- [x] Dagger local simulation (`pkg/simulator`)
- [x] Bubbletea TUI Dashboard (`pkg/tui`)
- [x] Blueprints — `go-app`, `node-app`
- [x] Action Shims — checkout, setup-go, setup-node
- [x] Multi-dimensional matrix expansion
- [x] Compile-time DAG cycle detection
- [x] Secret scanner — 7 pattern types + Shannon entropy
- [x] Smart scaffolding (`anchor init`)
- [x] Interactive debug shell (`anchor exec`)
- [x] Action plugin system (`.anchor/plugins/`)
- [x] Telemetry & cost dashboard
- [x] Pattern-based debugger (`pkg/debugger`)
- [x] YamlAnchor Studio (React/Vite)
- [x] REST API server
- [x] VS Code extension scaffold
- [x] GoReleaser multi-platform binary releases
- [x] GitHub Actions CI pipeline
- [x] GitLab CI export
- [ ] Bitbucket Pipelines export
- [ ] LLM-powered intelligent fix suggestions

---

## 📄 License

MIT © [AyushCN](https://github.com/AyushCN)

---

<div align="center">

**Built for developers who are tired of pushing blind.**

⚓ *Stop Pushing. Start Anchoring.*

</div>
