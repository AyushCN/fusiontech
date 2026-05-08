package schema

import (
	"testing"
)

// --- Pipeline validation ---

func TestPipeline_Validate_Valid(t *testing.T) {
	p := &Pipeline{
		Name: "Valid Pipeline",
		Jobs: map[string]*Job{
			"build": {
				RunsOn: "ubuntu-latest",
				Steps:  []*Step{{Name: "Checkout", Uses: "actions/checkout@v4"}},
			},
		},
	}
	if err := p.Validate(); err != nil {
		t.Errorf("Validate() unexpected error: %v", err)
	}
}

func TestPipeline_Validate_MissingName(t *testing.T) {
	p := &Pipeline{
		Jobs: map[string]*Job{
			"build": {RunsOn: "ubuntu-latest", Steps: []*Step{{Run: "echo hi"}}},
		},
	}
	if err := p.Validate(); err == nil {
		t.Error("Expected error for missing pipeline name")
	}
}

func TestPipeline_Validate_NoJobs(t *testing.T) {
	p := &Pipeline{Name: "Empty", Jobs: map[string]*Job{}}
	if err := p.Validate(); err == nil {
		t.Error("Expected error for empty jobs map")
	}
}

// --- Job validation ---

func TestJob_Validate_MissingRunsOn(t *testing.T) {
	j := &Job{Steps: []*Step{{Run: "echo hi"}}}
	if err := j.Validate("build"); err == nil {
		t.Error("Expected error for missing runs-on")
	}
}

func TestJob_Validate_NoSteps(t *testing.T) {
	j := &Job{RunsOn: "ubuntu-latest", Steps: []*Step{}}
	if err := j.Validate("build"); err == nil {
		t.Error("Expected error for empty steps")
	}
}

// --- RunsOnLabels ---

func TestRunsOnLabels_String(t *testing.T) {
	j := &Job{RunsOn: "ubuntu-latest"}
	labels := j.RunsOnLabels()
	if len(labels) != 1 || labels[0] != "ubuntu-latest" {
		t.Errorf("Expected ['ubuntu-latest'], got %v", labels)
	}
}

func TestRunsOnLabels_Array(t *testing.T) {
	j := &Job{RunsOn: []interface{}{"self-hosted", "linux", "arm64"}}
	labels := j.RunsOnLabels()
	if len(labels) != 3 || labels[0] != "self-hosted" || labels[1] != "linux" || labels[2] != "arm64" {
		t.Errorf("Expected ['self-hosted','linux','arm64'], got %v", labels)
	}
}

func TestRunsOnLabels_Map_GroupAndLabels(t *testing.T) {
	j := &Job{RunsOn: map[string]interface{}{
		"group":  "my-runners",
		"labels": []interface{}{"large", "gpu"},
	}}
	labels := j.RunsOnLabels()
	if len(labels) != 3 {
		t.Errorf("Expected 3 labels (group + 2), got %v", labels)
	}
}

func TestRunsOnLabels_Map_LabelString(t *testing.T) {
	j := &Job{RunsOn: map[string]interface{}{
		"labels": "large",
	}}
	labels := j.RunsOnLabels()
	if len(labels) != 1 || labels[0] != "large" {
		t.Errorf("Expected ['large'], got %v", labels)
	}
}

func TestRunsOnLabels_Empty(t *testing.T) {
	j := &Job{RunsOn: ""}
	labels := j.RunsOnLabels()
	if len(labels) != 0 {
		t.Errorf("Expected empty labels for empty RunsOn, got %v", labels)
	}
}

// --- Step type classification ---

func TestStep_Type_Run(t *testing.T) {
	s := &Step{Run: "echo hello"}
	if s.Type() != StepTypeRun {
		t.Errorf("Expected StepTypeRun, got %s", s.Type())
	}
}

func TestStep_Type_RemoteAction(t *testing.T) {
	s := &Step{Uses: "actions/checkout@v4"}
	if s.Type() != StepTypeUsesActionRemote {
		t.Errorf("Expected StepTypeUsesActionRemote, got %s", s.Type())
	}
}

func TestStep_Type_DockerURL(t *testing.T) {
	s := &Step{Uses: "docker://alpine:3.18"}
	if s.Type() != StepTypeUsesDockerURL {
		t.Errorf("Expected StepTypeUsesDockerURL, got %s", s.Type())
	}
}

func TestStep_Type_LocalAction(t *testing.T) {
	s := &Step{Uses: "./my-local-action"}
	if s.Type() != StepTypeUsesActionLocal {
		t.Errorf("Expected StepTypeUsesActionLocal, got %s", s.Type())
	}
}

func TestStep_Type_LocalReusableWorkflow(t *testing.T) {
	s := &Step{Uses: "./.github/workflows/deploy.yml"}
	if s.Type() != StepTypeReusableWorkflowLocal {
		t.Errorf("Expected StepTypeReusableWorkflowLocal, got %s", s.Type())
	}
}

func TestStep_Type_RemoteReusableWorkflow(t *testing.T) {
	s := &Step{Uses: "org/repo/.github/workflows/deploy.yml@main"}
	if s.Type() != StepTypeReusableWorkflowRemote {
		t.Errorf("Expected StepTypeReusableWorkflowRemote, got %s", s.Type())
	}
}

func TestStep_Type_Invalid_NeitherRunNorUses(t *testing.T) {
	s := &Step{Name: "only-name"}
	if s.Type() != StepTypeInvalid {
		t.Errorf("Expected StepTypeInvalid, got %s", s.Type())
	}
}

func TestStep_Type_Invalid_BothRunAndUses(t *testing.T) {
	s := &Step{Run: "echo hi", Uses: "actions/checkout@v4"}
	if s.Type() != StepTypeInvalid {
		t.Errorf("Expected StepTypeInvalid, got %s", s.Type())
	}
}

// --- ShellCommand ---

func TestShellCommand_Default_Bash(t *testing.T) {
	s := &Step{Run: "echo hi"}
	if s.ShellCommand() != "bash --noprofile --norc -eo pipefail {0}" {
		t.Errorf("Unexpected shell command: %s", s.ShellCommand())
	}
}

func TestShellCommand_Python(t *testing.T) {
	s := &Step{Run: "print('hi')", Shell: "python"}
	if s.ShellCommand() != "python {0}" {
		t.Errorf("Unexpected shell command: %s", s.ShellCommand())
	}
}

func TestShellCommand_Sh(t *testing.T) {
	s := &Step{Shell: "sh"}
	if s.ShellCommand() != "sh -e {0}" {
		t.Errorf("Unexpected shell command: %s", s.ShellCommand())
	}
}

func TestShellCommand_Custom(t *testing.T) {
	s := &Step{Shell: "fish --no-config {0}"}
	if s.ShellCommand() != "fish --no-config {0}" {
		t.Errorf("Unexpected shell command: %s", s.ShellCommand())
	}
}

// --- Step Validate ---

func TestStep_Validate_Valid_Run(t *testing.T) {
	s := &Step{Name: "build", Run: "go build ./..."}
	if err := s.Validate("build", 0); err != nil {
		t.Errorf("Validate() unexpected error: %v", err)
	}
}

func TestStep_Validate_Valid_Uses(t *testing.T) {
	s := &Step{Uses: "actions/checkout@v4"}
	if err := s.Validate("build", 0); err != nil {
		t.Errorf("Validate() unexpected error: %v", err)
	}
}

func TestStep_Validate_BothRunAndUses(t *testing.T) {
	s := &Step{Run: "echo hi", Uses: "actions/checkout@v4"}
	if err := s.Validate("build", 0); err == nil {
		t.Error("Expected error when both run and uses are set")
	}
}

func TestStep_Validate_NeitherRunNorUses(t *testing.T) {
	s := &Step{}
	if err := s.Validate("build", 0); err == nil {
		t.Error("Expected error when neither run nor uses is set")
	}
}

// --- DAG validation ---

func TestValidateDAG_CircularDependency(t *testing.T) {
	jobs := map[string]*Job{
		"a": {RunsOn: "ubuntu-latest", Needs: []string{"b"}, Steps: []*Step{{Run: "echo a"}}},
		"b": {RunsOn: "ubuntu-latest", Needs: []string{"a"}, Steps: []*Step{{Run: "echo b"}}},
	}
	p := &Pipeline{Name: "Circular", Jobs: jobs}
	if err := p.Validate(); err == nil {
		t.Error("Expected circular dependency error")
	}
}

func TestValidateDAG_LinearDependency(t *testing.T) {
	jobs := map[string]*Job{
		"build": {RunsOn: "ubuntu-latest", Steps: []*Step{{Run: "go build ./..."}}},
		"test":  {RunsOn: "ubuntu-latest", Needs: []string{"build"}, Steps: []*Step{{Run: "go test ./..."}}},
	}
	p := &Pipeline{Name: "Linear", Jobs: jobs}
	if err := p.Validate(); err != nil {
		t.Errorf("Expected valid DAG, got error: %v", err)
	}
}

func TestValidateDAG_MissingDependency(t *testing.T) {
	jobs := map[string]*Job{
		"deploy": {RunsOn: "ubuntu-latest", Needs: []string{"build"}, Steps: []*Step{{Run: "echo deploy"}}},
	}
	p := &Pipeline{Name: "Missing dependency", Jobs: jobs}
	if err := p.Validate(); err == nil {
		t.Error("Expected missing dependency error")
	}
}
