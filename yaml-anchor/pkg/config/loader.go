package config

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
	"yaml-anchor/pkg/errors"
	"yaml-anchor/pkg/schema"
)

// Load reads a YAML configuration file and returns a validated Pipeline
func Load(filepath string) (*schema.Pipeline, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, errors.NewConfigError(filepath, fmt.Sprintf("failed to read file: %v", err))
	}

	pipeline, err := ParseYAML(string(data))
	if err != nil {
		return nil, errors.NewConfigError(filepath, err.Error())
	}
	return pipeline, nil
}

// ParseYAML parses YAML string and returns a validated Pipeline
func ParseYAML(content string) (*schema.Pipeline, error) {
	var pipeline schema.Pipeline

	if err := yaml.Unmarshal([]byte(content), &pipeline); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	pipeline.Jobs = expandMatrix(pipeline.Jobs)

	if err := pipeline.Validate(); err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
	}

	return &pipeline, nil
}

// Write saves a pipeline to a YAML file
func Write(pipeline *schema.Pipeline, filepath string) error {
	if err := pipeline.Validate(); err != nil {
		return err
	}

	data, err := yaml.Marshal(pipeline)
	if err != nil {
		return fmt.Errorf("failed to marshal to YAML: %w", err)
	}

	if err := os.WriteFile(filepath, data, 0644); err != nil {
		return fmt.Errorf("failed to write file %q: %w", filepath, err)
	}

	return nil
}

func expandMatrix(jobs map[string]*schema.Job) map[string]*schema.Job {
	expanded := make(map[string]*schema.Job)

	var jobNames []string
	for jobName := range jobs {
		jobNames = append(jobNames, jobName)
	}
	sort.Strings(jobNames)

	for _, jobName := range jobNames {
		job := jobs[jobName]

		if job.Strategy != nil && job.Strategy.Matrix != nil {
			matrixRaw, ok := job.Strategy.Matrix.(map[string]interface{})
			if !ok {
				expanded[jobName] = job
				continue
			}

			matrixes, err := getMatrixes(matrixRaw)
			if err != nil {
				// Fall back to unexpanded job on parse error
				expanded[jobName] = job
				continue
			}

			for _, combo := range matrixes {
				// Sort keys for a deterministic job name suffix
				var keys []string
				for k := range combo {
					keys = append(keys, k)
				}
				sort.Strings(keys)

				var suffixParts []string
				for _, k := range keys {
					suffixParts = append(suffixParts, fmt.Sprintf("%v", combo[k]))
				}
				newJobName := fmt.Sprintf("%s (%s)", jobName, strings.Join(suffixParts, ", "))

				newJob := deepCopyJob(job)
				newJob.Strategy = nil
				if newJob.Env == nil {
					newJob.Env = make(map[string]string)
				}
				for k, v := range combo {
					newJob.Env[k] = fmt.Sprintf("%v", v)
				}
				expanded[newJobName] = newJob
			}
			continue
		}

		expanded[jobName] = job
	}
	return expanded
}

// getMatrixes returns all matrix combinations with include/exclude applied.
// This is a direct port of GitHub Actions semantics from nektos/act.
func getMatrixes(matrixRaw map[string]interface{}) ([]map[string]interface{}, error) {
	includes := extractMatrixList(matrixRaw["include"])
	excludes := extractMatrixList(matrixRaw["exclude"])

	dims := make(map[string][]interface{})
	var dimKeys []string
	for k, v := range matrixRaw {
		if k == "include" || k == "exclude" {
			continue
		}
		if vArr, ok := v.([]interface{}); ok {
			dims[k] = vArr
		} else {
			dims[k] = []interface{}{v}
		}
		dimKeys = append(dimKeys, k)
	}
	sort.Strings(dimKeys)

	// Validate excludes — all keys must be known dimension keys
	for _, excl := range excludes {
		for k := range excl {
			if _, ok := dims[k]; !ok {
				return nil, fmt.Errorf("matrix exclude key %q does not match any key in the matrix", k)
			}
		}
	}

	// Compute full Cartesian product of dimension values
	product := matrixCartesian(dims, dimKeys)

	// Apply excludes
	var afterExcludes []map[string]interface{}
PRODUCT:
	for _, combo := range product {
		for _, excl := range excludes {
			if matrixKeysMatch(combo, excl) {
				continue PRODUCT
			}
		}
		afterExcludes = append(afterExcludes, combo)
	}

	// Apply includes:
	// - If an include matches existing combos on their dimension keys, merge new keys in.
	// - If no match, append as a new standalone combo.
	var extraIncludes []map[string]interface{}
	for _, incl := range includes {
		matched := false
		for _, combo := range afterExcludes {
			if matrixKeysMatch2(combo, incl, dims) {
				matched = true
				for k, v := range incl {
					combo[k] = v
				}
			}
		}
		if !matched {
			extraIncludes = append(extraIncludes, incl)
		}
	}
	afterExcludes = append(afterExcludes, extraIncludes...)

	if len(afterExcludes) == 0 {
		afterExcludes = append(afterExcludes, make(map[string]interface{}))
	}
	return afterExcludes, nil
}

func extractMatrixList(raw interface{}) []map[string]interface{} {
	if raw == nil {
		return nil
	}
	var result []map[string]interface{}
	switch t := raw.(type) {
	case []interface{}:
		for _, item := range t {
			if m, ok := item.(map[string]interface{}); ok {
				result = append(result, m)
			}
		}
	case map[string]interface{}:
		result = append(result, t)
	}
	return result
}

// matrixCartesian computes the Cartesian product of dims in deterministic key order.
func matrixCartesian(dims map[string][]interface{}, keys []string) []map[string]interface{} {
	if len(keys) == 0 {
		return []map[string]interface{}{{}}
	}
	first := keys[0]
	rest := matrixCartesian(dims, keys[1:])
	var result []map[string]interface{}
	for _, v := range dims[first] {
		for _, r := range rest {
			combo := make(map[string]interface{})
			for k, val := range r {
				combo[k] = val
			}
			combo[first] = v
			result = append(result, combo)
		}
	}
	return result
}

// matrixKeysMatch returns true if all keys in b that also appear in a have equal string values.
func matrixKeysMatch(a, b map[string]interface{}) bool {
	for k, bv := range b {
		if av, ok := a[k]; ok {
			if fmt.Sprintf("%v", av) != fmt.Sprintf("%v", bv) {
				return false
			}
		}
	}
	return true
}

// matrixKeysMatch2 checks that b's dimension-keyed values match a.
func matrixKeysMatch2(a, b map[string]interface{}, dims map[string][]interface{}) bool {
	for k, bv := range b {
		if _, isDim := dims[k]; isDim {
			if av, ok := a[k]; ok {
				if fmt.Sprintf("%v", av) != fmt.Sprintf("%v", bv) {
					return false
				}
			}
		}
	}
	return true
}



func deepCopyJob(job *schema.Job) *schema.Job {
	newJob := &schema.Job{
		Name:           job.Name,
		Blueprint:      job.Blueprint,
		RunsOn:         job.RunsOn,
		Environment:    job.Environment,
		Concurrency:    job.Concurrency,
		If:             job.If,
		TimeoutMinutes: job.TimeoutMinutes,
	}

	if job.Needs != nil {
		newJob.Needs = make([]string, len(job.Needs))
		copy(newJob.Needs, job.Needs)
	}

	if job.Outputs != nil {
		newJob.Outputs = make(map[string]interface{})
		for k, v := range job.Outputs {
			newJob.Outputs[k] = v // Shallow copy for interface{}
		}
	}

	if job.Env != nil {
		newJob.Env = make(map[string]string)
		for k, v := range job.Env {
			newJob.Env[k] = v
		}
	}

	if job.Defaults != nil {
		newJob.Defaults = &schema.Defaults{}
		if job.Defaults.Run != nil {
			newJob.Defaults.Run = &schema.RunDefaults{
				Shell:            job.Defaults.Run.Shell,
				WorkingDirectory: job.Defaults.Run.WorkingDirectory,
			}
		}
	}

	if job.Steps != nil {
		newJob.Steps = make([]*schema.Step, len(job.Steps))
		for i, step := range job.Steps {
			newStep := &schema.Step{
				Id:               step.Id,
				Name:             step.Name,
				Uses:             step.Uses,
				Run:              step.Run,
				Shell:            step.Shell,
				If:               step.If,
				Timeout:          step.Timeout,
				ContinueOnError:  step.ContinueOnError,
				WorkingDirectory: step.WorkingDirectory,
			}

			if step.With != nil {
				newStep.With = make(map[string]interface{})
				for k, v := range step.With {
					newStep.With[k] = v
				}
			}

			if step.Env != nil {
				newStep.Env = make(map[string]string)
				for k, v := range step.Env {
					newStep.Env[k] = v
				}
			}

			newJob.Steps[i] = newStep
		}
	}

	if job.Strategy != nil {
		newJob.Strategy = &schema.Strategy{
			Matrix:      job.Strategy.Matrix, // Shallow copy for interface{}
			FailFast:    job.Strategy.FailFast,
			MaxParallel: job.Strategy.MaxParallel,
		}
	}

	if job.Container != nil {
		newJob.Container = &schema.Container{
			Image:   job.Container.Image,
			Options: job.Container.Options,
		}
		if job.Container.Creds != nil {
			newJob.Container.Creds = &schema.Credentials{
				Username: job.Container.Creds.Username,
				Password: job.Container.Creds.Password,
			}
		}
		if job.Container.Env != nil {
			newJob.Container.Env = make(map[string]string)
			for k, v := range job.Container.Env {
				newJob.Container.Env[k] = v
			}
		}
		if job.Container.Ports != nil {
			newJob.Container.Ports = make([]int, len(job.Container.Ports))
			copy(newJob.Container.Ports, job.Container.Ports)
		}
		if job.Container.Volumes != nil {
			newJob.Container.Volumes = make([]string, len(job.Container.Volumes))
			copy(newJob.Container.Volumes, job.Container.Volumes)
		}
	}

	if job.Services != nil {
		newJob.Services = make(map[string]*schema.Service)
		for k, svc := range job.Services {
			newSvc := &schema.Service{
				Image:   svc.Image,
				Options: svc.Options,
			}
			if svc.Creds != nil {
				newSvc.Creds = &schema.Credentials{
					Username: svc.Creds.Username,
					Password: svc.Creds.Password,
				}
			}
			if svc.Env != nil {
				newSvc.Env = make(map[string]string)
				for envK, envV := range svc.Env {
					newSvc.Env[envK] = envV
				}
			}
			if svc.Ports != nil {
				newSvc.Ports = make([]int, len(svc.Ports))
				copy(newSvc.Ports, svc.Ports)
			}
			newJob.Services[k] = newSvc
		}
	}

	return newJob
}
