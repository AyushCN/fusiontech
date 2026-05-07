# YamlAnchor ⚓

> **Treat CI/CD pipelines as type-safe code, not indentation-sensitive text.**

YamlAnchor eliminates "YAML Hell" in CI/CD workflows. Define your pipeline in a structured `anchor.yaml`, let YamlAnchor validate it with compile-time guarantees (DAG cycle detection, dependency graph checks), generate GitHub Actions YAML, and simulate execution locally in real Docker containers — all before pushing to the cloud.

---

## ✨ What Makes This Different

| Traditional YAML | YamlAnchor |
|---|---|
| Errors discovered after push | Errors caught at load-time |
| No dependency validation | DAG cycle detection built-in |
| Manual action configuration | High-level **Blueprints** auto-expand steps |
| `uses:` steps skipped locally | **Action Shims** simulate common actions |
| No feedback on savings | **Telemetry** reports CI minutes saved |
| Config-only workflow | Web UI for visual pipeline design |

---

## 🚀 Getting Started

### Prerequisites
- [Go 1.21+](https://go.dev/)
- [Docker](https://www.docker.com/) (running locally)

### Installation

```bash
git clone https://github.com/ayushcn/fusiontech.git
cd fusiontech/yaml-anchor
go build -o anchor main.go
sudo mv anchor /usr/local/bin/
```

---

## 🛠️ Defining Your Pipeline

### Option 1: Manual Configuration

```yaml
# anchor.yaml
name: "CI Pipeline"

on:
  push:
    branches: [main]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - name: "Checkout Code"
        uses: "actions/checkout@v4"
      - name: "Run Tests"
        run: "go test ./..."
      - name: "Build Binary"
        run: "go build -o ./bin/app ./..."
        env:
          CGO_ENABLED: "0"
          GOOS: "linux"
```

### Option 2: Blueprints (Recommended)

Blueprints let you define a pipeline without knowing the underlying GitHub Action syntax. YamlAnchor expands them into the correct steps automatically.

```yaml
name: "My Go App"

jobs:
  build:
    blueprint: "go-app"   # Expands into: checkout → setup-go → build → test
```

**Available Blueprints:**

| Blueprint | Expands Into |
|---|---|
| `go-app` | checkout → setup-go@v4 → `go build` → `go test` |
| `node-app` | checkout → setup-node@v3 → `npm ci` → `npm test` |

### Option 3: YamlAnchor Studio (Web UI)

```bash
cd ui && npm install && npm run dev
# Open http://localhost:5173
```

A glassmorphic React web app with:
- **Simulated AI Generator** — describe your stack, generate a pipeline instantly
- **Real-time YAML preview** with syntax highlighting
- **Visual pipeline graph** — SVG flowchart of jobs and steps with live **fault detection**

---

## 🔗 Job Dependencies (DAG)

YamlAnchor performs compile-time DAG validation. Circular dependencies are caught before any container is started.

```yaml
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - name: "Build"
        run: "go build ./..."

  deploy:
    runs-on: ubuntu-latest
    needs: [build]          # 'deploy' waits for 'build' to succeed
    steps:
      - name: "Deploy"
        run: "echo deploying..."
```

```bash
# Circular dependency example — caught immediately:
$ anchor local -c bad-pipeline.yaml
Failed to load config: circular dependency detected involving job: job-a
```

---

## 📦 Commands

### `anchor generate` — Generate GitHub Actions YAML

```bash
anchor generate --config anchor.yaml
```

- Validates pipeline structure and DAG dependencies
- **Scans for hardcoded secrets** — blocks generation if found
- Writes `.github/workflows/main.yml`

### `anchor local` — Run Locally with the Pulse TUI

```bash
anchor local --config anchor.yaml
```

- Blueprints are expanded before execution
- Runner names mapped to real images (`ubuntu-latest` → `ubuntu:22.04`, `setup-go` → `golang:1.21`)
- **Action Shims** for common `uses:` steps (no more silent skips):
  - `actions/checkout` → shimmed (local directory is already mounted)
  - `actions/setup-go` / `actions/setup-node` → shimmed via base image selection
- Real stdout streamed to the **Pulse Dashboard** TUI
- **Telemetry Report** printed after execution:
  ```
  ==== TELEMETRY REPORT ====
  Jobs Simulated: 1
  Steps Executed: 3
  Local Execution Time: 12s
  Estimated CI Time Saved: 36s
  ==========================
  ```

### `anchor clean` — Free Docker/Dagger Cache

```bash
anchor clean
```

---

## 🔐 Secret Scanner

Blocks `anchor generate` if any step contains:

| Pattern | Example |
|---|---|
| AWS Access Key | `AKIA[0-9A-Z]{16}` |
| GitHub Token | `ghp_...`, `gho_...` |
| Bearer Token | `Bearer <token>` |

---

## 📁 Project Structure

```
yaml-anchor/
├── anchor.yaml              # Example pipeline definition
├── test-blueprint.yaml      # Blueprint feature demo
├── test-circle.yaml         # DAG cycle detection demo
├── main.go                  # CLI entry point
├── cmd/
│   ├── root.go
│   ├── generate.go          # anchor generate
│   ├── local.go             # anchor local (with error reporting)
│   └── clean.go             # anchor clean
├── pkg/
│   ├── config/              # Loader, Blueprint expansion, DAG validation
│   ├── schema/              # Pipeline IR: Pipeline, Job (Needs/Blueprint), Step
│   ├── generator/           # YAML export
│   ├── scanner/             # Secret detection
│   ├── simulator/           # Dagger engine with Action Shims + Telemetry
│   └── tui/                 # Bubbletea Pulse dashboard
└── ui/                      # YamlAnchor Studio (React + Vite)
    └── src/
        ├── App.jsx           # 3-panel layout
        ├── components/
        │   ├── AIGenerator.jsx   # Simulated AI pipeline generator
        │   └── VisualGraph.jsx   # SVG flowchart + fault detection
        └── index.css         # Dark terminal aesthetic
```
