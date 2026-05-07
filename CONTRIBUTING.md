# Contributing to YamlAnchor

Thanks for your interest in contributing! This guide will get you set up quickly.

## Project Structure

```
fusiontech/
├── yaml-anchor/          # Go CLI + API server
│   ├── cmd/              # Cobra CLI commands
│   ├── pkg/              # Core packages
│   │   ├── analyzer/     # Code-to-pipeline analysis
│   │   ├── config/       # YAML loading + matrix expansion
│   │   ├── debugger/     # Error pattern analysis
│   │   ├── detector/     # Stack detection
│   │   ├── errors/       # Typed error system
│   │   ├── generator/    # YAML export
│   │   ├── logger/       # Structured leveled logger
│   │   ├── scanner/      # Secret + entropy scanner
│   │   ├── schema/       # IR types + DAG validation
│   │   ├── simulator/    # Dagger local execution engine
│   │   ├── tui/          # Bubbletea dashboard
│   │   └── validator/    # Input validation rules
│   ├── ui/               # React/Vite Studio web app
│   └── examples/         # Real anchor.yaml examples
├── vscode-anchor/        # VS Code extension scaffold
├── Makefile              # Dev shortcuts
└── .goreleaser.yaml      # Multi-platform release config
```

## Getting Started

```bash
git clone https://github.com/AyushCN/fusiontech.git
cd fusiontech
make build        # Build the anchor binary
make test-go      # Run all Go tests
```

## Development Workflow

1. **Fork** the repository and create a branch from `main`.
2. **Write code** — new packages go in `yaml-anchor/pkg/`.
3. **Write tests** — every new package must have a `*_test.go` file.
4. **Ensure all tests pass**: `make test-go`
5. **Open a pull request** against `main`.

## Package Guidelines

| Rule | Why |
|:---|:---|
| One package per concern | Keeps testability high |
| No circular imports | Use `pkg/schema` as the shared IR boundary |
| `pkg/simulator` and `pkg/tui` are Docker-dependent | Don't import them from core logic packages |
| Use `pkg/errors` for all custom error types | Consistent error handling across CLI |
| Use `pkg/logger` for all output, not `fmt.Println` | Enables log level filtering |

## Adding a New CLI Command

1. Create `yaml-anchor/cmd/<command>.go`
2. Register it with `rootCmd.AddCommand(myCmd)` in an `init()` function
3. Add it to the commands table in `SETUP.md`

## Adding a New `pkg/` Package

1. Create `yaml-anchor/pkg/<name>/<name>.go`
2. Create `yaml-anchor/pkg/<name>/<name>_test.go` with at least 3 tests
3. Run `go test ./pkg/<name>/...` to verify

## Action Plugin Shims

Place shell scripts at `.anchor/plugins/<owner>/<action>.sh` in your project to override how the simulator handles a specific `uses:` step. The script runs inside the Dagger container with the repo mounted at `/src`.

## Commit Message Format

We follow a simple prefix convention:

| Prefix | When to use |
|:---|:---|
| `feat:` | New feature or capability |
| `fix:` | Bug fix |
| `test:` | Adding or fixing tests |
| `docs:` | Documentation only |
| `refactor:` | Code cleanup (no behavior change) |
| `chore:` | Build, CI, dependency updates |

## Release Process

Releases are automated via GoReleaser. To trigger a release:

```bash
git tag v0.2.0
git push origin v0.2.0
```

This triggers `.github/workflows/release.yml` which builds Linux/macOS/Windows binaries and attaches them to a GitHub Release.

## Questions?

Open an issue or start a Discussion on GitHub. We're friendly! ⚓
