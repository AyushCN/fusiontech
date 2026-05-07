package simulator

import (
	"context"
	"fmt"
	"os"
	"strings"

	"dagger.io/dagger"
	"yaml-anchor/pkg/schema"
)

// UpdateMsg represents a message from the simulation engine to the TUI.
type UpdateMsg struct {
	JobName string
	Step    string
	Status  string // "running", "success", "error", "skipped"
	LogLine string // optional
	Error   error  // optional
}

// RunLocal executes a pipeline locally using Dagger and Docker, emitting
// structured UpdateMsg events to the provided channel for the TUI to render.
func RunLocal(ctx context.Context, pipeline *schema.Pipeline, updates chan<- UpdateMsg) {
	defer close(updates)

	send := func(msg UpdateMsg) {
		select {
		case updates <- msg:
		case <-ctx.Done():
		}
	}

	// Initialize Dagger client
	client, err := dagger.Connect(ctx, dagger.WithLogOutput(os.Stderr))
	if err != nil {
		send(UpdateMsg{Status: "error", Error: fmt.Errorf("failed to connect to Dagger: %w", err)})
		return
	}
	defer client.Close()

	cwd, err := os.Getwd()
	if err != nil {
		send(UpdateMsg{Status: "error", Error: fmt.Errorf("failed to get working directory: %w", err)})
		return
	}
	hostDir := client.Host().Directory(cwd)

	for jobName, job := range pipeline.Jobs {
		send(UpdateMsg{JobName: jobName, Status: "running", Step: "Initializing Container"})

		image := resolveImage(job)
		container := client.Container().
			From(image).
			WithMountedDirectory("/src", hostDir).
			WithWorkdir("/src")

		send(UpdateMsg{JobName: jobName, Status: "running",
			Step:    "Container Ready",
			LogLine: fmt.Sprintf("Using image: %s", image),
		})

		for _, step := range job.Steps {
			stepName := resolveStepName(step)

			// Handle 'uses' steps — not executable locally, skip gracefully
			if step.Uses != "" {
				send(UpdateMsg{
					JobName: jobName, Step: stepName, Status: "skipped",
					LogLine: fmt.Sprintf("Skipping action '%s' (not supported in local mode)", step.Uses),
				})
				continue
			}

			if step.Run == "" {
				continue
			}

			// Apply any step-level env vars
			for k, v := range step.Env {
				container = container.WithEnvVariable(k, v)
			}

			send(UpdateMsg{
				JobName: jobName, Step: stepName, Status: "running",
				LogLine: fmt.Sprintf("$ %s", step.Run),
			})

			container = container.WithExec([]string{"sh", "-c", step.Run})

			// Sync executes the step and captures stdout
			_, err = container.Sync(ctx)
			if err != nil {
				send(UpdateMsg{JobName: jobName, Step: stepName, Status: "error", Error: err})
				return
			}

			// Capture and stream real stdout line by line
			stdout, err := container.Stdout(ctx)
			if err == nil && strings.TrimSpace(stdout) != "" {
				for _, line := range strings.Split(strings.TrimRight(stdout, "\n"), "\n") {
					if line != "" {
						send(UpdateMsg{JobName: jobName, Step: stepName, Status: "running", LogLine: line})
					}
				}
			}

			send(UpdateMsg{JobName: jobName, Step: stepName, Status: "success", LogLine: "✓ Step completed"})
		}

		send(UpdateMsg{JobName: jobName, Status: "success", Step: "All steps completed"})
	}
}

// resolveImage maps GitHub Actions runner names to real Docker image names.
func resolveImage(job schema.Job) string {
	switch job.RunsOn {
	case "ubuntu-latest", "ubuntu-22.04":
		// Check if any step uses Go — prefer a Go image for better compatibility
		for _, step := range job.Steps {
			if strings.Contains(step.Run, "go ") || strings.Contains(step.Run, "go\t") {
				return "golang:1.21"
			}
		}
		return "ubuntu:22.04"
	case "ubuntu-20.04":
		return "ubuntu:20.04"
	default:
		return job.RunsOn
	}
}

// resolveStepName returns a display name for a step.
func resolveStepName(step schema.Step) string {
	if step.Name != "" {
		return step.Name
	}
	if step.Uses != "" {
		return step.Uses
	}
	if step.Run != "" {
		// Trim to first line for display
		lines := strings.SplitN(step.Run, "\n", 2)
		return "$ " + lines[0]
	}
	return "unnamed step"
}
