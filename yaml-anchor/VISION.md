# YamlAnchor: CI DevTools & The Debugger for CI Pipelines

## The Problem
Current CI workflows are defined by a miserable emotional experience: **push → wait → read logs → cry.**
Nobody cares about successful pipelines. But when things break, caches fail, secrets mismatch, or flaky builds happen, developers waste hours blindly guessing what went wrong in a remote, invisible environment.

## The Pivot
YamlAnchor is no longer just "local execution" or "a YAML generator." That's a feature, not a product.

**YamlAnchor is "CI DevTools".** It is the definitive debugger for CI pipelines. 

We turn invisible execution into **VISUAL execution**, making debugging feel interactive. 
* "Pause → Inspect → Replay → Fix instantly."

## Core Pillars of the New Vision

### 1. Visual Execution (Reveal Hidden Systems)
* Live dependency graphs.
* Animated execution flow.
* Failed-step highlighting.
* Environment inspection (seeing exact env vars at the time of failure).
* Execution snapshots.

### 2. Focus on the Failure Experience
The product must shine when things break.
* Flaky build root-cause analysis.
* Cache miss visualizations.
* Secret injection tracing.

### 3. Time-Travel Debugging
Because Dagger handles execution as a DAG of container states, deterministic replay is possible.
* Replay failed pipeline states instantly without re-running earlier steps.
* Inspect exact file system diffs between successful and failed runs.
* Pause execution on failure and drop into an interactive debug shell.

### 4. Visually Addictive Demos
We move away from terminal logs and YAML edits.
Demos will showcase a Chrome DevTools-like interface (YamlAnchor Studio) where developers visually inspect and fix pipelines in real-time.
