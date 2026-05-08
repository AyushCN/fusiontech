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

> **The problem**: CI/CD pipelines are written in YAML — a format with no types, no compile-time checks, and no local execution. Errors are discovered only after a remote push. YamlAnchor fixes this.

YamlAnchor defines your pipeline in **Go structs**, validates it at load-time, simulates it locally in **Docker via Dagger**, and exports a valid `.github/workflows/main.yml`. No more "push → wait → fail → cry."

---

## ✨ Why YamlAnchor?

| Traditional CI/CD | YamlAnchor |
|:---|:---|
| ❌ Errors found after a remote push | ✅ Caught at load-time with typed validation |
| ❌ No circular dependency detection | ✅ Compile-time DAG cycle detection |
| ❌ Boilerplate steps repeated everywhere | ✅ **Blueprints** auto-expand common stacks |
| ❌ `uses:` actions skipped locally | ✅ **Action Shims** simulate them in Dagger |
| ❌ Blind log scrolling | ✅ **Pulse Dashboard** (Bubbletea TUI) |
| ❌ Hardcoded secrets committed accidentally | ✅ **Secret Scanner** blocks export on `HIGH`/`CRITICAL` |
| ❌ No financial insight on CI usage | ✅ **Telemetry** reports minutes and cost saved |

---

## 🚀 Features

### ⚡ Pulse Dashboard (TUI)
A live Bubbletea terminal dashboard replaces raw log scrolling. Real-time job progress, color-coded status, and concurrent step tracking via Go channels.

### 🏗️ Blueprints
Write less boilerplate. A single `blueprint: "go-app"` line expands into a fully-configured checkout → setup → build → test pipeline. Supported: `go-app`, `node-app`.

### 🛡️ Action Shims & Local Simulation
YamlAnchor doesn't skip GitHub Actions — it *simulates* them inside Dagger containers:
- `actions/checkout` → mounts your local workspace
- `actions/setup-go` → resolves the correct Go image from `go.mod`
- `actions/setup-node` → resolves the Node image from `package.json`
- Custom shims via `.anchor/plugins/<owner>/<action>.sh`

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
A glassmorphic React/Vite web UI with:
- **AI Generator** — describe your stack, get an `anchor.yaml`
- **Visual Flowchart** — SVG job dependency graph with fault detection
- **YAML Preview** — live synchronized output

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
Auto-detects your stack (Go, Node, Python, Rust) and writes `anchor.yaml`.

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
# Run all tests
make test-go

# Run with race detector
go test -race ./pkg/...

# Build binary
make build

# Show coverage
make coverage

# Lint
make lint
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
- [x] Pattern-based LLM debugger (`pkg/debugger`)
- [x] YamlAnchor Studio (React/Vite)
- [x] REST API server
- [x] VS Code extension scaffold
- [x] GoReleaser multi-platform binary releases
- [x] GitHub Actions CI pipeline
- [ ] GitLab CI export
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
