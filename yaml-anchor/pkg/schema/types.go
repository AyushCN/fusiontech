package schema

// Pipeline represents a GitHub Actions workflow.
type Pipeline struct {
	Name string         `yaml:"name"`
	On   *Triggers      `yaml:"on"`
	Jobs map[string]Job `yaml:"jobs"`
}

// Triggers represents the events that trigger the workflow.
type Triggers struct {
	Push        *PushTrigger        `yaml:"push,omitempty"`
	PullRequest *PullRequestTrigger `yaml:"pull_request,omitempty"`
}

type PushTrigger struct {
	Branches []string `yaml:"branches,omitempty"`
}

type PullRequestTrigger struct {
	Branches []string `yaml:"branches,omitempty"`
}

// Job represents a single job in the workflow.
type Job struct {
	RunsOn string `yaml:"runs-on"`
	Steps  []Step `yaml:"steps"`
}

// Step represents a single step within a job.
type Step struct {
	Name string            `yaml:"name,omitempty"`
	Uses string            `yaml:"uses,omitempty"`
	Run  string            `yaml:"run,omitempty"`
	Env  map[string]string `yaml:"env,omitempty"`
}
