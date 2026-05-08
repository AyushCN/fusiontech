// Package planner builds a stage-based execution plan from a YamlAnchor pipeline.
// Each stage contains jobs that can run in parallel. Jobs are placed in the
// earliest stage where all their dependencies have already been scheduled.
//
// This approach is taken from nektos/act's model/planner.go, adapted for
// YamlAnchor's schema. It allows anchor local to run independent jobs
// concurrently rather than serially.
package planner

import (
	"fmt"
	"sort"

	"yaml-anchor/pkg/schema"
)

// Stage is a group of jobs that can all run in parallel.
type Stage struct {
	// Jobs is an ordered, deterministic list of job IDs in this stage.
	Jobs []string
}

// Plan is the full execution plan for a pipeline: a list of stages to run in series.
type Plan struct {
	Stages []*Stage
}

// BuildPlan creates an execution Plan from a pipeline.
// It returns an error if a cycle is detected (a dependency cannot be resolved).
func BuildPlan(pipeline *schema.Pipeline) (*Plan, error) {
	// Build a dependency map for every job.
	deps := make(map[string][]string)
	for id, job := range pipeline.Jobs {
		deps[id] = job.Needs
	}

	plan := &Plan{}
	scheduled := make(map[string]bool)

	// Iterate until all jobs are placed in stages.
	for len(deps) > 0 {
		stage := &Stage{}

		// Collect jobs whose dependencies are all already scheduled.
		var ready []string
		for id, needs := range deps {
			if allScheduled(needs, scheduled) {
				ready = append(ready, id)
			}
		}

		if len(ready) == 0 {
			// No jobs could be scheduled — there must be an unresolvable dependency.
			var remaining []string
			for id := range deps {
				remaining = append(remaining, id)
			}
			sort.Strings(remaining)
			return nil, fmt.Errorf(
				"unable to build execution plan: jobs %v have unresolvable dependencies (possible cycle or missing job)",
				remaining,
			)
		}

		// Sort for deterministic stage order.
		sort.Strings(ready)
		stage.Jobs = ready

		for _, id := range ready {
			scheduled[id] = true
			delete(deps, id)
		}

		plan.Stages = append(plan.Stages, stage)
	}

	return plan, nil
}

// allScheduled returns true if every job in needs appears in the scheduled map.
func allScheduled(needs []string, scheduled map[string]bool) bool {
	for _, n := range needs {
		if !scheduled[n] {
			return false
		}
	}
	return true
}

// MaxJobNameLen returns the length of the longest job name across all stages.
func (p *Plan) MaxJobNameLen() int {
	max := 0
	for _, stage := range p.Stages {
		for _, id := range stage.Jobs {
			if len(id) > max {
				max = len(id)
			}
		}
	}
	return max
}

// TotalJobs returns the total number of jobs across all stages.
func (p *Plan) TotalJobs() int {
	n := 0
	for _, stage := range p.Stages {
		n += len(stage.Jobs)
	}
	return n
}
