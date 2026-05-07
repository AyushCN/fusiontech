# YamlAnchor ⚓

YamlAnchor is a developer tool that treats CI/CD pipelines as **Type-Safe Code** rather than Indentation-Sensitive Text. Define your workflow in a simple `anchor.yaml`, let YamlAnchor validate it and generate a mathematically perfect GitHub Actions YAML — then test it completely locally using real Docker containers, before ever pushing.

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

Copy the included `anchor.yaml` into your project root and edit it:

```yaml
# anchor.yaml
name: "CI Pipeline"

on:
  push:
    branches:
      - main
  pull_request:
    branches:
      - main

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

**Schema reference:**

| Field | Required | Description |
|---|---|---|
| `name` | ✅ | Pipeline display name |
| `on.push.branches` | optional | Branches to trigger on push |
| `on.pull_request.branches` | optional | Branches to trigger on PR |
| `jobs.<id>.runs-on` | ✅ | Runner image (`ubuntu-latest`, etc.) |
| `jobs.<id>.steps[].name` | optional | Step display name |
| `jobs.<id>.steps[].uses` | * | GitHub Action to use |
| `jobs.<id>.steps[].run` | * | Shell command to run |
| `jobs.<id>.steps[].env` | optional | Environment variables for this step |

*Each step must have either `run` or `uses`.*

---

## 📦 Commands

### `anchor generate` — Generate GitHub Actions YAML

```bash
anchor generate --config anchor.yaml
```

- Loads your `anchor.yaml` and validates it structurally
- **Scans for hardcoded secrets** (AWS keys, GitHub tokens, Bearer tokens) — blocks generation if found
- Writes a valid workflow to `.github/workflows/main.yml`

```
Loading pipeline config from anchor.yaml...
Generating YAML for pipeline: "CI Pipeline"
✓ Successfully generated workflow at .github/workflows/main.yml
```

### `anchor local` — Run Locally with the Pulse Dashboard

```bash
anchor local --config anchor.yaml
```

- Connects to your local Docker daemon via Dagger
- Maps runner names to real images (`ubuntu-latest` → `ubuntu:22.04`, detects Go projects and uses `golang:1.21`)
- Applies per-step `env:` variables inside the container
- Streams **real stdout output** from each command to the live TUI
- `uses:` steps are gracefully skipped with a notice

### `anchor clean` — Free Up Docker/Dagger Cache

```bash
anchor clean
```

Runs `docker system prune -f` to remove dangling containers, unused layers, and build caches.

---

## 🔐 Secret Scanner

YamlAnchor automatically detects and **blocks** generation if any step contains:

| Pattern | Example |
|---|---|
| AWS Access Key | `AKIA[0-9A-Z]{16}` |
| GitHub Token | `ghp_...`, `gho_...`, etc. |
| Bearer Token | `Bearer <token>` |

---

## 🧪 Running Tests

```bash
go test ./...
```

---

## 📁 Project Structure

```
yaml-anchor/
├── anchor.yaml              # Example pipeline definition (edit this)
├── main.go                  # CLI entry point
├── cmd/
│   ├── root.go              # Cobra root command
│   ├── generate.go          # anchor generate
│   ├── local.go             # anchor local
│   └── clean.go             # anchor clean
└── pkg/
    ├── config/              # anchor.yaml loader & validator
    ├── schema/              # Pipeline IR type definitions
    ├── generator/           # YAML export logic
    ├── scanner/             # Hardcoded secret detection
    ├── simulator/           # Dagger-based local execution engine
    └── tui/                 # Bubbletea Pulse dashboard
```
