package simulator

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"dagger.io/dagger"
	"yaml-anchor/pkg/debugger"
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

	startTime := time.Now()
	var stepsExecuted int
	var jobsSimulated int

	for jobName, job := range pipeline.Jobs {
		jobsSimulated++
		send(UpdateMsg{JobName: jobName, Status: "running", Step: "Initializing Container"})

		image := resolveImage(job)
		
		// Blueprint expansion
		if job.Blueprint == "go-app" {
			job.Steps = append([]*schema.Step{
				{Name: "Go Dependencies", Run: "go mod download"},
				{Name: "Go Build", Run: "go build ./..."},
				{Name: "Go Test", Run: "go test -v ./..."},
			}, job.Steps...)
		}

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

			// Handle 'uses' steps with Action Shims
			if step.Uses != "" {
				if strings.Contains(step.Uses, "actions/checkout") {
					send(UpdateMsg{
						JobName: jobName, Step: stepName, Status: "success",
						LogLine: "✓ Action Shimmed: Local directory successfully mounted.",
					})
					stepsExecuted++
					continue
				} else if strings.Contains(step.Uses, "actions/setup-go") || strings.Contains(step.Uses, "actions/setup-node") {
					send(UpdateMsg{
						JobName: jobName, Step: stepName, Status: "success",
						LogLine: "✓ Action Shimmed: Environment satisfied by base image.",
					})
					stepsExecuted++
					continue
				} else {
					// Check for local plugin shim
					pluginPath := filepath.Join(cwd, ".anchor", "plugins", step.Uses+".sh")
					if _, err := os.Stat(pluginPath); err == nil {
						send(UpdateMsg{
							JobName: jobName, Step: stepName, Status: "running",
							LogLine: fmt.Sprintf("⚙️ Executing custom plugin shim: %s", step.Uses),
						})

						// Mount the script and run it
						container = container.WithFile("/tmp/shim.sh", client.Host().File(pluginPath))
						container = container.WithExec([]string{"sh", "/tmp/shim.sh"})

						_, err = container.Sync(ctx)
						if err != nil {
							send(UpdateMsg{JobName: jobName, Step: stepName, Status: "error", Error: err})
							return
						}

						send(UpdateMsg{
							JobName: jobName, Step: stepName, Status: "success",
							LogLine: "✓ Custom plugin shim executed successfully.",
						})
						stepsExecuted++
						continue
					}
				}

				send(UpdateMsg{
					JobName: jobName, Step: stepName, Status: "skipped",
					LogLine: fmt.Sprintf("Skipping unsupported action '%s'", step.Uses),
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
				suggestions := debugger.Analyze(stepName, step.Run, err.Error())
				debugOutput := debugger.Format(suggestions)
				send(UpdateMsg{JobName: jobName, Step: stepName, Status: "error", Error: err,
					LogLine: debugOutput})
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
			stepsExecuted++
		}

		send(UpdateMsg{JobName: jobName, Status: "success", Step: "All steps completed"})
	}

	// Calculate and report Telemetry Metrics
	duration := time.Since(startTime)
	// Rough heuristic: CI takes ~3x longer due to queuing, image pulls over network, and overhead
	estimatedSaved := duration * 3
	
	// Financial Cost Calculation
	// GitHub Ubuntu runners are $0.008/minute
	minutesSaved := estimatedSaved.Minutes()
	costSaved := minutesSaved * 0.008
	
	metricsLog := fmt.Sprintf("==== TELEMETRY REPORT ====\nJobs Simulated: %d\nSteps Executed: %d\nLocal Execution Time: %s\nEstimated CI Time Saved: %s\n💸 Estimated Cost Saved: $%.4f\n==========================", 
		jobsSimulated, stepsExecuted, duration.Round(time.Second), estimatedSaved.Round(time.Second), costSaved)
	
	send(UpdateMsg{
		JobName: "System", Step: "Telemetry", Status: "success",
		LogLine: metricsLog,
	})
}

// resolveImage maps GitHub Actions runner names to real Docker image names.
func resolveImage(job *schema.Job) string {
	switch job.RunsOn {
	case "ubuntu-latest", "ubuntu-22.04":
		// Check if any step uses Go — prefer a Go image for better compatibility
		for _, step := range job.Steps {
			if strings.Contains(step.Run, "go ") || strings.Contains(step.Run, "go\t") || strings.Contains(step.Uses, "setup-go") {
				return "golang:1.26"
			}
			if strings.Contains(step.Run, "npm ") || strings.Contains(step.Uses, "setup-node") {
				return "node:18"
			}
		}
		return "ubuntu:22.04"
	case "ubuntu-20.04":
		return "ubuntu:20.04"
	// Cross-Platform Shims: macOS runners map to a Linux base image.
	// True macOS simulation is not possible in Docker; we use ubuntu as an approximation.
	case "macos-latest", "macos-13", "macos-12":
		fmt.Println("⚠️  Cross-Platform Shim: macOS runner detected. Using ubuntu:22.04 as a Linux approximation.")
		fmt.Println("   Note: macOS-specific tooling (e.g., Homebrew, Xcode) will not be available.")
		return "ubuntu:22.04"
	// Cross-Platform Shims: Windows runners map to a Linux base image.
	// True Windows container simulation requires Windows hosts.
	case "windows-latest", "windows-2022", "windows-2019":
		fmt.Println("⚠️  Cross-Platform Shim: Windows runner detected. Using ubuntu:22.04 as a Linux approximation.")
		fmt.Println("   Note: Windows-specific tooling (e.g., PowerShell Core, MSVC) will not be available.")
		return "ubuntu:22.04"
	default:
		return job.RunsOn
	}
}

// resolveStepName returns a display name for a step.
func resolveStepName(step *schema.Step) string {
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
