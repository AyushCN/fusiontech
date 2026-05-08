package planner_test

import (
	"testing"

	"yaml-anchor/pkg/planner"
	"yaml-anchor/pkg/schema"
)

func pipeline(jobs map[string]*schema.Job) *schema.Pipeline {
	return &schema.Pipeline{Name: "test", Jobs: jobs}
}

func job(needs ...string) *schema.Job {
	return &schema.Job{
		RunsOn: "ubuntu-latest",
		Needs:  needs,
		Steps:  []*schema.Step{{Run: "echo hi"}},
	}
}

// --- BuildPlan ---

func TestBuildPlan_SingleJob_OneStage(t *testing.T) {
	p, err := planner.BuildPlan(pipeline(map[string]*schema.Job{
		"build": job(),
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(p.Stages) != 1 {
		t.Errorf("expected 1 stage, got %d", len(p.Stages))
	}
	if p.Stages[0].Jobs[0] != "build" {
		t.Errorf("expected 'build' in stage 0, got %v", p.Stages[0].Jobs)
	}
}

func TestBuildPlan_LinearChain_ThreeStages(t *testing.T) {
	// build → test → deploy
	p, err := planner.BuildPlan(pipeline(map[string]*schema.Job{
		"build":  job(),
		"test":   job("build"),
		"deploy": job("test"),
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(p.Stages) != 3 {
		t.Errorf("expected 3 stages, got %d", len(p.Stages))
	}
	if p.Stages[0].Jobs[0] != "build" {
		t.Errorf("stage 0 expected 'build', got %v", p.Stages[0].Jobs)
	}
	if p.Stages[1].Jobs[0] != "test" {
		t.Errorf("stage 1 expected 'test', got %v", p.Stages[1].Jobs)
	}
	if p.Stages[2].Jobs[0] != "deploy" {
		t.Errorf("stage 2 expected 'deploy', got %v", p.Stages[2].Jobs)
	}
}

func TestBuildPlan_FanOut_ParallelJobs(t *testing.T) {
	// build → [test-unit, test-e2e, lint] in parallel
	p, err := planner.BuildPlan(pipeline(map[string]*schema.Job{
		"build":     job(),
		"test-unit": job("build"),
		"test-e2e":  job("build"),
		"lint":      job("build"),
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(p.Stages) != 2 {
		t.Errorf("expected 2 stages, got %d", len(p.Stages))
	}
	if len(p.Stages[1].Jobs) != 3 {
		t.Errorf("expected 3 parallel jobs in stage 1, got %d: %v", len(p.Stages[1].Jobs), p.Stages[1].Jobs)
	}
}

func TestBuildPlan_Diamond_CorrectStages(t *testing.T) {
	// build → [test, lint] → deploy
	p, err := planner.BuildPlan(pipeline(map[string]*schema.Job{
		"build":  job(),
		"test":   job("build"),
		"lint":   job("build"),
		"deploy": job("test", "lint"),
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(p.Stages) != 3 {
		t.Errorf("expected 3 stages, got %d: %+v", len(p.Stages), p.Stages)
	}
	// Stage 0: build
	if len(p.Stages[0].Jobs) != 1 || p.Stages[0].Jobs[0] != "build" {
		t.Errorf("stage 0 expected [build], got %v", p.Stages[0].Jobs)
	}
	// Stage 1: lint + test (sorted)
	if len(p.Stages[1].Jobs) != 2 {
		t.Errorf("stage 1 expected [lint, test], got %v", p.Stages[1].Jobs)
	}
	// Stage 2: deploy
	if len(p.Stages[2].Jobs) != 1 || p.Stages[2].Jobs[0] != "deploy" {
		t.Errorf("stage 2 expected [deploy], got %v", p.Stages[2].Jobs)
	}
}

func TestBuildPlan_Deterministic_SortedJobs(t *testing.T) {
	p, err := planner.BuildPlan(pipeline(map[string]*schema.Job{
		"z-job": job(),
		"a-job": job(),
		"m-job": job(),
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(p.Stages) != 1 {
		t.Errorf("expected 1 stage, got %d", len(p.Stages))
	}
	expected := []string{"a-job", "m-job", "z-job"}
	for i, id := range p.Stages[0].Jobs {
		if id != expected[i] {
			t.Errorf("stage 0 jobs not sorted: expected %v, got %v", expected, p.Stages[0].Jobs)
			break
		}
	}
}

func TestBuildPlan_UnresolvableDep_ReturnsError(t *testing.T) {
	_, err := planner.BuildPlan(pipeline(map[string]*schema.Job{
		"build": job("nonexistent"),
	}))
	if err == nil {
		t.Error("expected error for unresolvable dependency, got nil")
	}
}

func TestBuildPlan_EmptyPipeline_EmptyPlan(t *testing.T) {
	p, err := planner.BuildPlan(pipeline(map[string]*schema.Job{}))
	if err != nil {
		t.Fatalf("unexpected error for empty pipeline: %v", err)
	}
	if len(p.Stages) != 0 {
		t.Errorf("expected 0 stages for empty pipeline, got %d", len(p.Stages))
	}
}

// --- Plan helpers ---

func TestPlan_TotalJobs(t *testing.T) {
	p, _ := planner.BuildPlan(pipeline(map[string]*schema.Job{
		"build":  job(),
		"test":   job("build"),
		"deploy": job("test"),
	}))
	if p.TotalJobs() != 3 {
		t.Errorf("expected TotalJobs=3, got %d", p.TotalJobs())
	}
}

func TestPlan_MaxJobNameLen(t *testing.T) {
	p, _ := planner.BuildPlan(pipeline(map[string]*schema.Job{
		"build":             job(),
		"a-very-long-name":  job("build"),
	}))
	if p.MaxJobNameLen() != len("a-very-long-name") {
		t.Errorf("expected MaxJobNameLen=%d, got %d", len("a-very-long-name"), p.MaxJobNameLen())
	}
}
