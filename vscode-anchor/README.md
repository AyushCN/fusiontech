# YamlAnchor VS Code Extension

Adds first-class support for `anchor.yaml` pipelines directly in VS Code.

## Features

- **▶ Run Locally** — CodeLens "⚓ Run Locally" buttons appear above every job definition in `anchor.yaml`. Click to instantly run that job via Dagger in the integrated terminal.
- **🐚 Open Shell** — CodeLens "🐚 Open Shell" button drops you into an interactive shell inside the configured Dagger container for that job.
- **🔐 Scan for Secrets** — Right-click context menu on `anchor.yaml` to run `anchor scan .`.
- **⚙️ Generate Actions YAML** — Command palette: generate `.github/workflows/main.yml` without leaving VS Code.

## Requirements

- The `anchor` CLI must be installed and available in your `PATH`.
- Docker must be running locally (for `anchor local` and `anchor exec`).

## Configuration

| Setting | Default | Description |
|---|---|---|
| `yamlanchor.anchorPath` | `anchor` | Path to the `anchor` binary |
| `yamlanchor.configFile` | `anchor.yaml` | Path to your config file |

## Installation (Development)

```bash
cd vscode-anchor
npm install
npm run compile
# Press F5 in VS Code to open a new Extension Development Host window
```
