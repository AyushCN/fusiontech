# YamlAnchor ⚓ — The Debugger for CI Pipelines

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=for-the-badge&logo=go)](https://go.dev/)
[![CI](https://github.com/AyushCN/fusiontech/actions/workflows/ci.yml/badge.svg)](https://github.com/AyushCN/fusiontech/actions/workflows/ci.yml)
[![Release](https://github.com/AyushCN/fusiontech/actions/workflows/release.yml/badge.svg)](https://github.com/AyushCN/fusiontech/releases)
[![Dagger Engine](https://img.shields.io/badge/Engine-Dagger-FF6C37?style=for-the-badge&logo=docker)](https://dagger.io/)
[![TUI](https://img.shields.io/badge/UI-Bubbletea-ED5282?style=for-the-badge)](https://github.com/charmbracelet/bubbletea)
[![Web UI](https://img.shields.io/badge/Studio-Vite%20%2B%20React-646CFF?style=for-the-badge&logo=vite)](https://vitejs.dev/)

**YamlAnchor** is the ultimate developer tool that treats CI/CD pipelines as **Type-Safe Code** rather than Indentation-Sensitive Text. Stop the "Push → Wait → Fail → Cry" loop. YamlAnchor turns invisible remote execution into **Visual Local Execution**, allowing you to pause, inspect, and fix pipeline failures instantly.

---

## ✨ Why YamlAnchor?

| Traditional YAML | YamlAnchor |
|:--- |:--- |
| ❌ Errors discovered after push | ✅ Errors caught at compile-time |
| ❌ No dependency validation | ✅ Built-in DAG cycle detection |
| ❌ Manual action configuration | ✅ **Blueprints** auto-expand complex steps |
| ❌ `uses:` steps skipped locally | ✅ **Action Shims** simulate actions in Dagger |
| ❌ Blind log scrolling | ✅ **Pulse Dashboard** (TUI) for real-time status |
| ❌ No feedback on savings | ✅ **Telemetry** reports CI minutes/cost saved |
| ❌ Config-only workflow | ✅ **YamlAnchor Studio** for visual design |

---

## 🚀 Core Features

### 1. ⚡ Visual Execution (Pulse Dashboard)
Powered by **Bubbletea**, YamlAnchor provides a live, interactive TUI. Instead of scrolling terminal text, you see:
- Real-time job progress with spinners.
- Color-coded success/failure status.
- Concurrent step tracking via Go channels.
- Localized log feeds for the active step.

### 2. 🏗️ High-Level Blueprints
Stop writing boilerplate. Use **Blueprints** to define common stacks.
- `blueprint: "go-app"`: Automatically adds checkout, setup-go, build, and test steps.
- `blueprint: "node-app"`: Handles npm installation and testing with environment detection.

### 3. 🛡️ Action Shims & Local Simulation
YamlAnchor doesn't just "skip" GitHub Actions. It uses **Action Shims** to intelligently simulate them inside Dagger:
- `actions/checkout`: Automatically mounts your local workspace.
- `actions/setup-go`: Detects `go.mod` and resolves the correct Docker image dynamically.
- Support for step-level `env:` variables and dynamic environment injection.

### 4. 🔗 Intelligent DAG Validation
The engine builds a mathematical **Directed Acyclic Graph (DAG)** of your pipeline.
- Catch circular dependencies *before* starting a single container.
- Guarantee execution order with the `needs:` keyword.
- Validate job hierarchies at load-time.

### 5. 🔐 Security-First Generation
The `anchor generate` command includes a built-in **Secret Scanner**.
- Scans IR for hardcoded AWS keys, GitHub tokens, and Bearer tokens.
- Blocks YAML generation if leaks are detected, preventing accidental pushes to remote.

### 6. 🧩 Action Plugin System
Don't let unsupported third-party actions silently fail. Write a custom shim:
- Drop a shell script at `.anchor/plugins/<owner>/<action>.sh` in your project.
- YamlAnchor will automatically find, mount, and execute it inside the Dagger container.
- Teams can codify their proprietary CI logic without modifying YamlAnchor itself.

### 7. 🔢 Matrix Build Support
Simulate parallel matrix strategies locally.
- Add a `strategy.matrix` block to any job in `anchor.yaml`.
- YamlAnchor automatically expands the job into named sub-runs (e.g. `test (1.21)`, `test (1.22)`).
- Each expanded job gets the matrix variable injected as an environment variable.

### 8. 🎨 YamlAnchor Studio (Web UI)
A premium, glassmorphic React/Vite web application for visual pipeline management:
- **Simulated AI Generator**: Describe your stack and generate a config instantly.
- **Real-time Preview**: Synchronized YAML output with syntax highlighting.
- **Visual Flowchart**: SVG-based graph showing job dependencies with **active fault detection**.

### 8. 📊 Telemetry & Cost Insights
Every local run generates a **Telemetry Report**:
- Actual local execution time vs. Estimated remote CI time.
- Calculation of total CI minutes saved.
- 💸 **Financial Cost Saved**: Dollar amount saved based on GitHub runner pricing ($0.008/minute).

---

## 🛠️ Getting Started

### Prerequisites
- **Go 1.21+**
- **Docker Desktop** (running locally)
- **Node.js 18+** (for YamlAnchor Studio)

### Installation

#### Option A — Download Binary (recommended)
```bash
# Linux amd64
curl -L https://github.com/AyushCN/fusiontech/releases/latest/download/anchor_linux_amd64.tar.gz | tar xz
sudo mv anchor /usr/local/bin/
```

#### Option B — Build from source
```bash
git clone https://github.com/ayushcn/fusiontech.git
cd fusiontech/yaml-anchor
go build -o anchor main.go
sudo mv anchor /usr/local/bin/
```

---

## 📦 Usage & Commands

### Define Your Pipeline (`anchor.yaml`)
```yaml
name: "Production Pipeline"

jobs:
  build-backend:
    blueprint: "go-app"
  
  deploy:
    runs-on: "ubuntu-latest"
    needs: [build-backend]
    steps:
      - name: "Deploy to Production"
        run: "echo 'Deploying to cloud...'"
        env:
          STAGE: "prod"
```

### CLI Commands

| Command | Description |
|:--- |:--- |
| `anchor init` | Auto-detect your stack and scaffold a smart `anchor.yaml`. |
| `anchor scan` | Standalone secret scanner (entropy + regex) with git hook support. |
| `anchor generate` | Validate, scan for secrets, and export `.github/workflows/main.yml`. |
| `anchor generate --dry-run` | Validate only — no files written to disk. |
| `anchor local` | Start the **Pulse Dashboard** and execute the pipeline in Dagger. |
| `anchor exec <job>` | Drop into an interactive shell inside a Dagger container for the job. |
| `anchor server` | Start the REST API server for YamlAnchor Studio. |
| `anchor version` | Print version, commit hash, Go version, OS/arch. |
| `anchor clean` | Prune dangling containers and clear Dagger/Docker cache. |

**Global flags** (work with every command):
```
-c, --config string   Path to anchor.yaml (default: anchor.yaml)
-v, --verbose         Enable debug-level structured logging
```

### Running the Studio (Web UI)
```bash
cd ui
npm install
npm run dev
```

---

## 📁 Project Architecture

```
yaml-anchor/
├── cmd/                # CLI commands (Cobra)
│   ├── generate.go     # YAML export — --dry-run, --verbose, secret blocking
│   ├── local.go        # Dagger + Bubbletea TUI integration
│   ├── server.go       # REST API (/health, /api/analyze, /api/generate, /api/validate)
│   ├── scan.go         # Standalone secret scanner with git hook support
│   ├── version.go      # Build-time version info (injected by GoReleaser)
│   └── root.go         # Global --config, --verbose flags
├── pkg/
│   ├── schema/         # Type-safe Pipeline IR + DAG validation
│   ├── config/         # YAML loader + multi-dimensional matrix expansion
│   ├── blueprints/     # Blueprint → job step expansion
│   ├── detector/       # Auto stack detection from go.mod / package.json
│   ├── analyzer/       # Code analysis for Studio AI generator
│   ├── simulator/      # Dagger engine + action shims + telemetry
│   ├── tui/            # Bubbletea Pulse Dashboard
│   ├── scanner/        # Secret scanner (AWS/GitHub/Slack/Azure/SSH regex + entropy)
│   ├── debugger/       # Pattern-based error analysis + fix suggestions
│   ├── errors/         # Typed errors: ConfigError, ValidationError, SecurityError
│   ├── validator/      # Input validation: job IDs, runners, cron, step names
│   └── logger/         # Structured leveled logger with color + file output
├── ui/                 # YamlAnchor Studio (React + Vite)
│   ├── src/components/ # AIGenerator, VisualGraph, YAMLPreview
│   └── src/App.jsx     # Glassmorphic layout
├── examples/           # Real anchor.yaml examples (Go, Node, Python, Matrix, Full-stack)
├── vscode-anchor/      # VS Code extension scaffold
├── .github/
│   ├── workflows/ci.yml      # CI: go test -race on every push/PR
│   └── workflows/release.yml # Release: GoReleaser builds multi-platform binaries on tags
├── Makefile            # make build | test-go | coverage | lint
├── .goreleaser.yaml    # Multi-platform release config
├── CONTRIBUTING.md     # Contribution guide
├── SETUP.md            # Installation & usage guide
└── API_DOCS.md         # REST API reference
```

---

## 🗺️ Roadmap
- [x] Type-Safe Go IR
- [x] Dagger Local Execution
- [x] Bubbletea TUI Dashboard
- [x] Action Shims & Blueprints
- [x] Secret Scanner (`anchor scan`)
- [x] Smart Scaffolding (`anchor init`)
- [x] Interactive Debug Shell (`anchor exec`)
- [x] Matrix Build Support
- [x] Action Plugin System (`.anchor/plugins/`)
- [x] Financial Cost Dashboard
- [x] Pattern-based LLM Debugger
- [x] VS Code Extension Scaffold
- [x] Cross-Platform Runner Shims (macOS/Windows)

---

Developed with ❤️ for the DevOps community.
**Stop Pushing. Start Anchoring.** ⚓
