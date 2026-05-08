package schema

import (
	"fmt"
	"regexp"
	"strings"
)

// Pipeline represents a complete CI/CD workflow
type Pipeline struct {
	Name        string                 `yaml:"name" json:"name"`
	On          map[string]interface{} `yaml:"on" json:"on"`
	Env         map[string]string      `yaml:"env,omitempty" json:"env,omitempty"`
	Concurrency interface{}            `yaml:"concurrency,omitempty" json:"concurrency,omitempty"`
	Defaults    *Defaults              `yaml:"defaults,omitempty" json:"defaults,omitempty"`
	Jobs        map[string]*Job        `yaml:"jobs" json:"jobs"`
}

// Job represents a workflow job
type Job struct {
	Name      string `yaml:"name,omitempty" json:"name,omitempty"`
	Blueprint string `yaml:"blueprint,omitempty" json:"blueprint,omitempty"`
	// RunsOn accepts a string, a list of strings, or a {group, labels} map.
	// Use RunsOnLabels() to get the resolved list of runner labels.
	RunsOn         interface{}            `yaml:"runs-on" json:"runs_on"`
	Environment    string                 `yaml:"environment,omitempty" json:"environment,omitempty"`
	Concurrency    interface{}            `yaml:"concurrency,omitempty" json:"concurrency,omitempty"`
	Outputs        map[string]interface{} `yaml:"outputs,omitempty" json:"outputs,omitempty"`
	Env            map[string]string      `yaml:"env,omitempty" json:"env,omitempty"`
	Defaults       *Defaults              `yaml:"defaults,omitempty" json:"defaults,omitempty"`
	If             string                 `yaml:"if,omitempty" json:"if,omitempty"`
	Steps          []*Step                `yaml:"steps" json:"steps"`
	Strategy       *Strategy              `yaml:"strategy,omitempty" json:"strategy,omitempty"`
	Needs          []string               `yaml:"needs,omitempty" json:"needs,omitempty"`
	Container      *Container             `yaml:"container,omitempty" json:"container,omitempty"`
	Services       map[string]*Service    `yaml:"services,omitempty" json:"services,omitempty"`
	TimeoutMinutes int                    `yaml:"timeout-minutes,omitempty" json:"timeout_minutes,omitempty"`
}

// Step represents a single step in a job
type Step struct {
	Id               string                 `yaml:"id,omitempty" json:"id,omitempty"`
	Name             string                 `yaml:"name,omitempty" json:"name,omitempty"`
	Uses             string                 `yaml:"uses,omitempty" json:"uses,omitempty"`
	Run              string                 `yaml:"run,omitempty" json:"run,omitempty"`
	Shell            string                 `yaml:"shell,omitempty" json:"shell,omitempty"`
	With             map[string]interface{} `yaml:"with,omitempty" json:"with,omitempty"`
	Env              map[string]string      `yaml:"env,omitempty" json:"env,omitempty"`
	If               string                 `yaml:"if,omitempty" json:"if,omitempty"`
	Timeout          int                    `yaml:"timeout-minutes,omitempty" json:"timeout_minutes,omitempty"`
	ContinueOnError  bool                   `yaml:"continue-on-error,omitempty" json:"continue_on_error,omitempty"`
	WorkingDirectory string                 `yaml:"working-directory,omitempty" json:"working_directory,omitempty"`
}

// Strategy defines matrix and other execution strategies
type Strategy struct {
	Matrix      interface{} `yaml:"matrix,omitempty" json:"matrix,omitempty"`
	FailFast    bool        `yaml:"fail-fast,omitempty" json:"fail_fast,omitempty"`
	MaxParallel int         `yaml:"max-parallel,omitempty" json:"max_parallel,omitempty"`
}

// Container specifies a Docker container for the job
type Container struct {
	Image   string            `yaml:"image" json:"image"`
	Creds   *Credentials      `yaml:"credentials,omitempty" json:"credentials,omitempty"`
	Env     map[string]string `yaml:"env,omitempty" json:"env,omitempty"`
	Ports   []int             `yaml:"ports,omitempty" json:"ports,omitempty"`
	Volumes []string          `yaml:"volumes,omitempty" json:"volumes,omitempty"`
	Options string            `yaml:"options,omitempty" json:"options,omitempty"`
}

// Service represents a service container
type Service struct {
	Image   string            `yaml:"image" json:"image"`
	Creds   *Credentials      `yaml:"credentials,omitempty" json:"credentials,omitempty"`
	Env     map[string]string `yaml:"env,omitempty" json:"env,omitempty"`
	Ports   []int             `yaml:"ports,omitempty" json:"ports,omitempty"`
	Options string            `yaml:"options,omitempty" json:"options,omitempty"`
}

// Credentials for container/service authentication
type Credentials struct {
	Username string `yaml:"username" json:"username"`
	Password string `yaml:"password" json:"password"`
}

// Defaults specifies default settings
type Defaults struct {
	Run *RunDefaults `yaml:"run,omitempty" json:"run,omitempty"`
}

// RunDefaults specifies defaults for run steps
type RunDefaults struct {
	Shell            string `yaml:"shell,omitempty" json:"shell,omitempty"`
	WorkingDirectory string `yaml:"working-directory,omitempty" json:"working_directory,omitempty"`
}

// Validate checks if the pipeline is valid
func (p *Pipeline) Validate() error {
	if p.Name == "" {
		return fmt.Errorf("pipeline must have a name")
	}

	if len(p.Jobs) == 0 {
		return fmt.Errorf("pipeline must have at least one job")
	}

	for jobID, job := range p.Jobs {
		if err := job.Validate(jobID); err != nil {
			return err
		}
	}

	// Check for circular dependencies
	if err := validateDAG(p.Jobs); err != nil {
		return err
	}

	return nil
}

// Validate checks if a job is valid
func (j *Job) Validate(jobID string) error {
	labels := j.RunsOnLabels()
	if len(labels) == 0 {
		return fmt.Errorf("job %q must specify runs-on", jobID)
	}

	if len(j.Steps) == 0 {
		return fmt.Errorf("job %q must have at least one step", jobID)
	}

	for idx, step := range j.Steps {
		if err := step.Validate(jobID, idx); err != nil {
			return err
		}
	}

	return nil
}

// RunsOnLabels returns the list of runner labels from the runs-on field.
// Handles all three GitHub Actions forms:
//   - string:  "ubuntu-latest"
//   - array:   ["self-hosted", "linux"]
//   - map:     {group: mygroup, labels: [large]}
func (j *Job) RunsOnLabels() []string {
	switch v := j.RunsOn.(type) {
	case string:
		if v == "" {
			return nil
		}
		return []string{v}
	case []interface{}:
		var labels []string
		for _, item := range v {
			if s, ok := item.(string); ok {
				labels = append(labels, s)
			}
		}
		return labels
	case map[string]interface{}:
		var labels []string
		if group, ok := v["group"].(string); ok && group != "" {
			labels = append(labels, group)
		}
		if rawLabels, ok := v["labels"]; ok {
			switch l := rawLabels.(type) {
			case string:
				labels = append(labels, l)
			case []interface{}:
				for _, item := range l {
					if s, ok := item.(string); ok {
						labels = append(labels, s)
					}
				}
			}
		}
		return labels
	}
	return nil
}

// --- Step type classification (from nektos/act) ---

// StepType describes the kind of step.
type StepType int

const (
	StepTypeRun                    StepType = iota // `run:` shell command
	StepTypeUsesDockerURL                          // `uses: docker://...`
	StepTypeUsesActionLocal                        // `uses: ./local-action`
	StepTypeUsesActionRemote                       // `uses: owner/repo@ref`
	StepTypeReusableWorkflowLocal                  // `uses: ./.github/workflows/foo.yml`
	StepTypeReusableWorkflowRemote                 // `uses: org/repo/.github/workflows/foo.yml@ref`
	StepTypeInvalid                                // missing or conflicting fields
)

func (t StepType) String() string {
	switch t {
	case StepTypeRun:
		return "run"
	case StepTypeUsesDockerURL:
		return "docker"
	case StepTypeUsesActionLocal:
		return "local-action"
	case StepTypeUsesActionRemote:
		return "remote-action"
	case StepTypeReusableWorkflowLocal:
		return "local-reusable-workflow"
	case StepTypeReusableWorkflowRemote:
		return "remote-reusable-workflow"
	case StepTypeInvalid:
		return "invalid"
	}
	return "unknown"
}

// Type classifies the step, mirroring nektos/act's step type resolution.
func (s *Step) Type() StepType {
	if s.Run == "" && s.Uses == "" {
		return StepTypeInvalid
	}
	if s.Run != "" && s.Uses != "" {
		return StepTypeInvalid // cannot have both
	}
	if s.Run != "" {
		return StepTypeRun
	}
	// uses-based classification
	if strings.HasPrefix(s.Uses, "docker://") {
		return StepTypeUsesDockerURL
	}
	isLocalWorkflow := strings.HasPrefix(s.Uses, "./.github/workflows") &&
		(strings.HasSuffix(s.Uses, ".yml") || strings.HasSuffix(s.Uses, ".yaml"))
	if isLocalWorkflow {
		return StepTypeReusableWorkflowLocal
	}
	isRemoteWorkflow := !strings.HasPrefix(s.Uses, "./") &&
		strings.Contains(s.Uses, ".github/workflows") &&
		(strings.Contains(s.Uses, ".yml@") || strings.Contains(s.Uses, ".yaml@"))
	if isRemoteWorkflow {
		return StepTypeReusableWorkflowRemote
	}
	if strings.HasPrefix(s.Uses, "./") {
		return StepTypeUsesActionLocal
	}
	return StepTypeUsesActionRemote
}

// ShellCommand returns the shell invocation string for this step's shell setting.
// Mirrors GitHub Actions runner behavior and nektos/act's shell mapping.
func (s *Step) ShellCommand() string {
	switch s.Shell {
	case "", "bash":
		return "bash --noprofile --norc -eo pipefail {0}"
	case "sh":
		return "sh -e {0}"
	case "pwsh":
		return "pwsh -command . '{0}'"
	case "powershell":
		return "powershell -command . '{0}'"
	case "python":
		return "python {0}"
	case "cmd":
		return `cmd /D /E:ON /V:OFF /S /C "CALL "{0}""`
	default:
		// Custom shell — use as-is, caller substitutes {0} with script path
		return s.Shell
	}
}

// String returns a human-readable name for the step.
func (s *Step) String() string {
	if s.Name != "" {
		return s.Name
	}
	if s.Uses != "" {
		return s.Uses
	}
	if s.Run != "" {
		lines := strings.SplitN(s.Run, "\n", 2)
		if len(lines[0]) > 60 {
			return lines[0][:60] + "..."
		}
		return lines[0]
	}
	return s.Id
}

// Validate checks if a step is valid
func (s *Step) Validate(jobID string, stepIdx int) error {
	if s.Name == "" && s.Uses == "" && s.Run == "" {
		return fmt.Errorf("job %q step %d must have name, uses, or run", jobID, stepIdx)
	}

	if s.Uses != "" && s.Run != "" {
		return fmt.Errorf("job %q step %d cannot have both 'uses' and 'run'", jobID, stepIdx)
	}

	return nil
}

func isValidRunner(runner string) bool {
	validRunners := []string{
		"ubuntu-latest",
		"ubuntu-22.04",
		"ubuntu-20.04",
		"windows-latest",
		"windows-2022",
		"windows-2019",
		"macos-latest",
		"macos-13",
		"macos-12",
		"self-hosted",
	}

	for _, valid := range validRunners {
		if runner == valid {
			return true
		}
	}

	// Allow custom self-hosted runners
	return regexp.MustCompile(`^self-hosted-[a-z0-9-]+$`).MatchString(runner)
}

func validateDAG(jobs map[string]*Job) error {
	visited := make(map[string]bool)
	rec := make(map[string]bool)

	var visit func(jobID string) error
	visit = func(jobID string) error {
		job, ok := jobs[jobID]
		if !ok {
			return fmt.Errorf("job %q depends on missing job %q", jobID, jobID)
		}
		if rec[jobID] {
			return fmt.Errorf("circular dependency detected in job %q", jobID)
		}

		if visited[jobID] {
			return nil
		}

		rec[jobID] = true

		for _, need := range job.Needs {
			if _, ok := jobs[need]; !ok {
				return fmt.Errorf("job %q depends on missing job %q", jobID, need)
			}
			if err := visit(need); err != nil {
				return err
			}
		}

		rec[jobID] = false
		visited[jobID] = true

		return nil
	}

	for jobID := range jobs {
		if !visited[jobID] {
			if err := visit(jobID); err != nil {
				return err
			}
		}
	}

	return nil
}
