package schema

import (
	"testing"
)

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

func TestJob_Validate_MissingRunsOn(t *testing.T) {
	j := &Job{Steps: []*Step{{Run: "echo hi"}}}
	if err := j.Validate("build"); err == nil {
		t.Error("Expected error for missing runs-on")
	}
}

func TestJob_Validate_InvalidRunner(t *testing.T) {
	j := &Job{RunsOn: "solaris-latest", Steps: []*Step{{Run: "echo hi"}}}
	if err := j.Validate("build"); err == nil {
		t.Error("Expected error for invalid runner")
	}
}

func TestJob_Validate_NoSteps(t *testing.T) {
	j := &Job{RunsOn: "ubuntu-latest", Steps: []*Step{}}
	if err := j.Validate("build"); err == nil {
		t.Error("Expected error for empty steps")
	}
}

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
