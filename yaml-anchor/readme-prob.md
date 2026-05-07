# Project Crutches & Improvement Roadmap: YamlAnchor ⚓

This document outlines the internal "hacks" (crutches), technical difficulties, and the strategic roadmap for improving YamlAnchor.

## 🛠️ Current Crutches (The Hacks)

Every ambitious project has shortcuts. Here are the "crutches" currently holding YamlAnchor together:

1.  **Hardcoded Image Resolver**: 
    - *The Hack*: In `pkg/simulator/engine.go`, we map `ubuntu-latest` to `golang:1.26` or `node:18` based on simple string checks for "go" or "npm".
    - *The Risk*: If a project uses both or a different version, the simulation might fail or use an incorrect environment.
    - *Improvement*: Implement a real version parser for `go.mod` and `package.json`.

2.  **Naïve Action Shimming**:
    - *The Hack*: Common actions like `actions/checkout` are "shimmed" by simply doing nothing (since we mount the directory anyway). Other actions are silently skipped.
    - *The Risk*: A pipeline that relies on a complex third-party action (e.g., S3 upload) will fail silently in the simulation.
    - *Improvement*: Create a plugin-based shim architecture.

3.  **Heuristic-Based Telemetry**:
    - *The Hack*: CI time savings are calculated as `local_duration * 3`.
    - *The Risk*: This is a guestimate. Actual savings depend on runner availability and network speed.
    - *Improvement*: Integrate with GitHub APIs to fetch historical run data for accurate comparisons.

4.  **No Artifact/Cache Persistence**:
    - *The Hack*: Every `anchor local` run starts with a clean slate (mostly).
    - *The Risk*: Slow execution for large projects that rely heavily on `actions/cache`.
    - *Improvement*: Implement persistent Dagger volumes for common build caches.

## 🌋 Technical Difficulties

The "Hard Problems" we faced during development:

*   **The TUI Synchronization**: Keeping the **Bubbletea** loop responsive while **Dagger** blocks on heavy container operations required a complex channel-based event system.
*   **Context Management**: Properly propagating `context.Context` to ensure that stopping the TUI (Ctrl+C) actually kills the underlying Docker containers immediately.
*   **Path Mapping**: Aligning host paths with container paths (`/src`) consistently across different OS environments.

## 🚀 Improvement Roadmap

### Phase 1: Robustness (Short Term)
- [x] **Real Version Detection**: Parse project manifest files to select the exact Docker tag.
- [x] **Enhanced Secret Scanner**: Move beyond regex to entropy-based detection.
- [x] **`anchor exec`**: Add a command to drop into an interactive shell inside the simulated runner.

### Phase 2: Features (Medium Term)
- [x] **Action Plugin System**: Drop a `.sh` script in `.anchor/plugins/my-org/action.sh` and YamlAnchor will mount and execute it instead of skipping the action.
- [x] **Matrix Support**: Jobs with a `strategy.matrix` are automatically expanded into separate named sub-jobs in the Pulse Dashboard.
- [x] **Cost Dashboard**: The Telemetry Report now calculates actual dollar amounts saved based on GitHub runner pricing ($0.008/min).

### Phase 3: Intelligence (Long Term)
- [ ] **LLM Debugger**: When a step fails, feed the logs to an LLM to suggest the exact YAML fix.
- [ ] **VS Code Extension**: Visual "Play" buttons next to jobs in your `anchor.yaml`.
- [ ] **Cross-Platform Shims**: Use lightweight VMs to simulate Windows/macOS runners.

---

*Found a new crutch? Open an issue and let's turn it into a feature.* ⚓
