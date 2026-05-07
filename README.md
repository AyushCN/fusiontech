# YamlAnchor ⚓ — The Debugger for CI Pipelines

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=for-the-badge&logo=go)](https://go.dev/)
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

### 6. 🎨 YamlAnchor Studio (Web UI)
A premium, glassmorphic React/Vite web application for visual pipeline management:
- **Simulated AI Generator**: Describe your stack and generate a config instantly.
- **Real-time Preview**: Synchronized YAML output with syntax highlighting.
- **Visual Flowchart**: SVG-based graph showing job dependencies with **active fault detection**.

### 7. 📊 Telemetry & Insights
Every local run generates a **Telemetry Report**:
- Actual local execution time vs. Estimated remote CI time.
- Calculation of total CI minutes saved.
- Metrics to justify local testing before pushing.

---

## 🛠️ Getting Started

### Prerequisites
- **Go 1.21+**
- **Docker Desktop** (running locally)
- **Node.js 18+** (for YamlAnchor Studio)

### Installation

```bash
# Clone the repository
git clone https://github.com/ayushcn/fusiontech.git
cd fusiontech/yaml-anchor

# Build the CLI
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
| `anchor generate` | Validates IR, scans for secrets, and exports `.github/workflows/main.yml`. |
| `anchor local` | Starts the **Pulse Dashboard** and executes the pipeline in Dagger. |
| `anchor clean` | Prunes dangling containers and clears the Dagger/Docker cache. |

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
├── cmd/                # CLI implementation (Cobra)
│   ├── generate.go     # YAML Export + Secret Scanning
│   ├── local.go        # Dagger Engine + TUI integration
│   └── clean.go        # Resource management
├── pkg/
│   ├── schema/         # Type-Safe Pipeline IR (Job, Step, Needs)
│   ├── config/         # YAML Loader + Blueprint Expansion
│   ├── simulator/      # Dagger Engine + Action Shims + Telemetry
│   ├── tui/            # Bubbletea Pulse Dashboard
│   └── scanner/        # Automated Secret Detection
└── ui/                 # YamlAnchor Studio (React + Vite)
    ├── src/components/ # AIGenerator, VisualGraph, YAMLPreview
    └── src/App.jsx     # Glassmorphic Layout
```

---

## 🗺️ Roadmap
- [x] Type-Safe Go IR
- [x] Dagger Local Execution
- [x] Bubbletea TUI Dashboard
- [x] Action Shims & Blueprints
- [x] Secret Scanner
- [x] YamlAnchor Studio (Web UI)
- [ ] VS Code Extension for live IR validation
- [ ] Support for GitLab CI and Bitbucket Pipelines

---

Developed with ❤️ for the DevOps community.
**Stop Pushing. Start Anchoring.** ⚓
