package improver

import (
	"context"
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
	"yaml-anchor/pkg/aigen"
	"yaml-anchor/pkg/schema"
)

// ImproveYAML rewrites an anchor.yaml candidate using the failure context.
// It is keyless: aigen may use local Ollama when available, otherwise it uses
// the built-in offline generator.
func ImproveYAML(ctx context.Context, currentYAML, failureReason string) (string, string, error) {
	if repaired, ok := repairYAML(currentYAML, failureReason); ok {
		return repaired, "repair", nil
	}

	prompt := buildImprovePrompt(currentYAML, failureReason)
	pipeline, source, err := aigen.Generate(ctx, prompt, "yaml")
	if err != nil {
		return "", source, err
	}

	data, err := yaml.Marshal(pipeline)
	if err != nil {
		return "", source, fmt.Errorf("failed to serialize improved pipeline: %w", err)
	}
	return string(data), source, nil
}

func buildImprovePrompt(currentYAML, failureReason string) string {
	var b strings.Builder
	b.WriteString("Improve this YamlAnchor anchor.yaml so it validates and runs locally.\n")
	b.WriteString("Preserve the user's intent, but fix missing jobs, missing steps, bad dependencies, wrong commands, or missing setup steps.\n\n")
	if strings.TrimSpace(failureReason) != "" {
		b.WriteString("Failure reason:\n")
		b.WriteString(failureReason)
		b.WriteString("\n\n")
	}
	b.WriteString("Current anchor.yaml:\n")
	b.WriteString(currentYAML)
	return b.String()
}

func repairYAML(currentYAML, failureReason string) (string, bool) {
	var pipeline schema.Pipeline
	if err := yaml.Unmarshal([]byte(currentYAML), &pipeline); err != nil {
		return "", false
	}

	changed := repairPipeline(&pipeline, failureReason)
	if !changed {
		return "", false
	}
	if err := pipeline.Validate(); err != nil {
		return "", false
	}

	data, err := yaml.Marshal(&pipeline)
	if err != nil {
		return "", false
	}
	return string(data), true
}

func repairPipeline(pipeline *schema.Pipeline, failureReason string) bool {
	changed := false
	if strings.TrimSpace(pipeline.Name) == "" {
		pipeline.Name = "Improved Pipeline"
		changed = true
	}
	if pipeline.On == nil {
		pipeline.On = map[string]interface{}{
			"push": map[string]interface{}{"branches": []string{"main"}},
		}
		changed = true
	}
	if len(pipeline.Jobs) == 0 {
		pipeline.Jobs = map[string]*schema.Job{
			"build": fallbackJob(failureReason),
		}
		return true
	}

	for jobID, job := range pipeline.Jobs {
		if job == nil {
			pipeline.Jobs[jobID] = fallbackJob(failureReason)
			changed = true
			continue
		}
		if len(job.RunsOnLabels()) == 0 {
			job.RunsOn = "ubuntu-latest"
			changed = true
		}
		if len(job.Steps) == 0 {
			job.Steps = fallbackJob(failureReason).Steps
			changed = true
		}
		if filtered, ok := removeMissingNeeds(job.Needs, pipeline.Jobs); ok {
			job.Needs = filtered
			changed = true
		}
		if repairSteps(job, failureReason) {
			changed = true
		}
	}

	return changed
}

func removeMissingNeeds(needs []string, jobs map[string]*schema.Job) ([]string, bool) {
	if len(needs) == 0 {
		return needs, false
	}
	filtered := make([]string, 0, len(needs))
	changed := false
	for _, need := range needs {
		if _, ok := jobs[need]; ok {
			filtered = append(filtered, need)
			continue
		}
		changed = true
	}
	return filtered, changed
}

func repairSteps(job *schema.Job, failureReason string) bool {
	changed := false
	if needsCheckout(job) && !hasUses(job, "actions/checkout") {
		job.Steps = prependStep(job.Steps, &schema.Step{Name: "Checkout", Uses: "actions/checkout@v4"})
		changed = true
	}
	if usesGo(job) && !hasUses(job, "actions/setup-go") {
		job.Steps = insertAfterCheckout(job.Steps, &schema.Step{Name: "Setup Go", Uses: "actions/setup-go@v4"})
		changed = true
	}
	if usesNode(job) && !hasUses(job, "actions/setup-node") {
		job.Steps = insertAfterCheckout(job.Steps, &schema.Step{Name: "Setup Node", Uses: "actions/setup-node@v3"})
		changed = true
	}
	if usesPython(job) && !hasUses(job, "actions/setup-python") {
		job.Steps = insertAfterCheckout(job.Steps, &schema.Step{Name: "Setup Python", Uses: "actions/setup-python@v4"})
		changed = true
	}
	if repairNPMCommands(job, failureReason) {
		changed = true
	}
	return changed
}

func repairNPMCommands(job *schema.Job, failureReason string) bool {
	lower := strings.ToLower(failureReason)
	changed := false
	for _, step := range job.Steps {
		run := strings.TrimSpace(step.Run)
		switch {
		case run == "npm ci" && (strings.Contains(lower, "package-lock.json") || strings.Contains(lower, "npm ci can only install")):
			step.Run = "npm install"
			changed = true
		case run == "npm test" && strings.Contains(lower, "missing script") && strings.Contains(lower, "test"):
			step.Run = "npm test --if-present"
			changed = true
		case strings.HasPrefix(run, "npm run ") && !strings.Contains(run, "--if-present"):
			script := strings.TrimSpace(strings.TrimPrefix(run, "npm run "))
			if strings.Contains(lower, "missing script") && strings.Contains(lower, script) {
				step.Run = run + " --if-present"
				changed = true
			}
		}
	}
	return changed
}

func fallbackJob(failureReason string) *schema.Job {
	lower := strings.ToLower(failureReason)
	switch {
	case strings.Contains(lower, "npm") || strings.Contains(lower, "node") || strings.Contains(lower, "react"):
		return &schema.Job{RunsOn: "ubuntu-latest", Steps: []*schema.Step{
			{Name: "Checkout", Uses: "actions/checkout@v4"},
			{Name: "Setup Node", Uses: "actions/setup-node@v3"},
			{Name: "Install Dependencies", Run: "npm ci"},
			{Name: "Test", Run: "npm test --if-present"},
			{Name: "Build", Run: "npm run build --if-present"},
		}}
	case strings.Contains(lower, "pytest") || strings.Contains(lower, "python") || strings.Contains(lower, "pip"):
		return &schema.Job{RunsOn: "ubuntu-latest", Steps: []*schema.Step{
			{Name: "Checkout", Uses: "actions/checkout@v4"},
			{Name: "Setup Python", Uses: "actions/setup-python@v4"},
			{Name: "Install Dependencies", Run: "pip install -r requirements.txt"},
			{Name: "Run Tests", Run: "pytest"},
		}}
	default:
		return &schema.Job{RunsOn: "ubuntu-latest", Steps: []*schema.Step{
			{Name: "Checkout", Uses: "actions/checkout@v4"},
			{Name: "Setup Go", Uses: "actions/setup-go@v4"},
			{Name: "Run Tests", Run: "go test ./..."},
			{Name: "Build", Run: "go build ./..."},
		}}
	}
}

func needsCheckout(job *schema.Job) bool {
	for _, step := range job.Steps {
		if step != nil && step.Run != "" {
			return true
		}
	}
	return false
}

func usesGo(job *schema.Job) bool {
	return stepRunContains(job, "go ")
}

func usesNode(job *schema.Job) bool {
	return stepRunContains(job, "npm ") || stepRunContains(job, "yarn ") || stepRunContains(job, "pnpm ")
}

func usesPython(job *schema.Job) bool {
	return stepRunContains(job, "pytest") || stepRunContains(job, "python ") || stepRunContains(job, "pip ")
}

func stepRunContains(job *schema.Job, needle string) bool {
	for _, step := range job.Steps {
		if step != nil && strings.Contains(strings.ToLower(step.Run), needle) {
			return true
		}
	}
	return false
}

func hasUses(job *schema.Job, needle string) bool {
	for _, step := range job.Steps {
		if step != nil && strings.Contains(strings.ToLower(step.Uses), strings.ToLower(needle)) {
			return true
		}
	}
	return false
}

func prependStep(steps []*schema.Step, step *schema.Step) []*schema.Step {
	return append([]*schema.Step{step}, steps...)
}

func insertAfterCheckout(steps []*schema.Step, step *schema.Step) []*schema.Step {
	insertAt := 0
	for idx, existing := range steps {
		if existing != nil && strings.Contains(strings.ToLower(existing.Uses), "actions/checkout") {
			insertAt = idx + 1
			break
		}
	}
	updated := make([]*schema.Step, 0, len(steps)+1)
	updated = append(updated, steps[:insertAt]...)
	updated = append(updated, step)
	updated = append(updated, steps[insertAt:]...)
	return updated
}
