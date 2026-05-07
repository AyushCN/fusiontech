package simulator

import (
	"context"
	"fmt"
	"os"

	"dagger.io/dagger"
	"yaml-anchor/pkg/schema"
)

// RunInteractive executes a Dagger container for the specified job and drops the user into an interactive shell.
func RunInteractive(ctx context.Context, pipeline *schema.Pipeline, jobName string) error {
	job, exists := pipeline.Jobs[jobName]
	if !exists {
		return fmt.Errorf("job '%s' not found in anchor.yaml", jobName)
	}

	fmt.Printf("⚓ Initializing interactive shell for job '%s'...\n", jobName)

	client, err := dagger.Connect(ctx, dagger.WithLogOutput(os.Stderr))
	if err != nil {
		return fmt.Errorf("failed to connect to Dagger: %w", err)
	}
	defer client.Close()

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}
	hostDir := client.Host().Directory(cwd)

	image := resolveImage(job)
	fmt.Printf("🐳 Using image: %s\n", image)

	// Blueprint expansion for env
	if job.Blueprint == "go-app" {
		// Nothing specific to env, but could add setup here
	}

	container := client.Container().
		From(image).
		WithMountedDirectory("/src", hostDir).
		WithWorkdir("/src")

	// Apply any step-level env vars globally for the shell
	for _, step := range job.Steps {
		for k, v := range step.Env {
			container = container.WithEnvVariable(k, v)
		}
	}

	fmt.Println("🚀 Dropping into /bin/sh. Type 'exit' to leave.")

	// Dagger Terminal for interactive execution
	_, err = container.Terminal(dagger.ContainerTerminalOpts{
		Cmd: []string{"/bin/sh"},
	}).Sync(ctx)

	if err != nil {
		return fmt.Errorf("terminal session ended with error: %w", err)
	}

	fmt.Println("👋 Exited interactive shell.")
	return nil
}
