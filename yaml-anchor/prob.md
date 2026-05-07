# Challenges & Problem Report: YamlAnchor ⚓

Building a tool that bridges the gap between Go code and containerized CI/CD execution presented several high-level technical challenges.

## 1. The "Push and Pray" Replacement (Dagger Integration)
*   **The Challenge:** Connecting a local CLI to a container engine (Dagger) required managing the Dagger engine lifecycle. 
*   **The Problem:** We initially faced version mismatches where the `go.mod` required a newer Go version than the default Dagger images provided.
*   **The Solution:** Implemented a dynamic `resolveImage` helper that detects project requirements and automatically pulls the correct `golang:1.26` or `node` images.

## 2. Real-time TUI State Synchronization
*   **The Challenge:** The Pulse Dashboard (Bubbletea) runs on a different loop than the Dagger execution engine.
*   **The Problem:** Streaming logs from a running container into a stylized UI without blocking the main execution thread or causing "flicker."
*   **The Solution:** Built a custom `UpdateMsg` channel architecture. The simulation engine emits granular events (Job started, Step log, Success/Error) which the TUI consumes asynchronously to update the UI state.

## 3. Security vs. Automation (Secret Scanning)
*   **The Challenge:** Automating YAML generation is dangerous if users hardcode credentials.
*   **The Problem:** Standard regex scanning can be slow or produce false positives.
*   **The Solution:** Integrated a mandatory pre-generation scan in `pkg/scanner/secrets.go`. By blocking the `generate` command entirely when a secret is found, we forced a "Security by Default" workflow.

## 4. Environment Variable Injection
*   **The Challenge:** GitHub Actions allow per-step environment variables, but Docker containers are often ephemeral.
*   **The Problem:** Mapping complex Go `map[string]string` structures into the Dagger `WithEnvVariable` chain.
*   **The Solution:** Enhanced the `Step` schema and the engine logic to recursively apply environment variables before each `exec` call.

## 5. The Blueprint Paradox
*   **The Challenge:** Users want simplicity (blueprints) but also control (manual steps).
*   **The Problem:** Creating a system that can "expand" a single line like `blueprint: go-app` into a full suite of commands.
*   **The Solution:** Implemented a blueprint expansion layer in the simulator engine that detects the template and prepends the standard CI steps (Download, Build, Test) before any custom user steps are executed.
