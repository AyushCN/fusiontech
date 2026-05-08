package simulator

import (
	"context"
	"strings"

	"yaml-anchor/pkg/schema"
)

// RunResult is machine-readable simulation output for automation loops.
type RunResult struct {
	Logs     string
	ExitCode int
	Error    error
}

// Run executes a pipeline and captures structured simulator output as logs.
func Run(ctx context.Context, pipeline *schema.Pipeline) RunResult {
	updates := make(chan UpdateMsg, 64)
	var logs strings.Builder

	go RunLocal(ctx, pipeline, updates)

	for update := range updates {
		if update.JobName != "" {
			logs.WriteString("[")
			logs.WriteString(update.JobName)
			logs.WriteString("]")
		}
		if update.Step != "" {
			logs.WriteString(" ")
			logs.WriteString(update.Step)
		}
		if update.Status != "" {
			logs.WriteString(" - ")
			logs.WriteString(update.Status)
		}
		if update.LogLine != "" {
			logs.WriteString("\n")
			logs.WriteString(update.LogLine)
		}
		logs.WriteString("\n")
		if update.Error != nil {
			return RunResult{Logs: logs.String(), ExitCode: 1, Error: update.Error}
		}
	}

	return RunResult{Logs: logs.String(), ExitCode: 0}
}
