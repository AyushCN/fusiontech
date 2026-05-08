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

	// Sort jobNames to ensure deterministic processing order
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

			var keys []string
			for k := range matrixRaw {
				keys = append(keys, k)
			}
			sort.Strings(keys)

			var values [][]string
			for _, k := range keys {
				vRaw := matrixRaw[k]
				vArr, ok := vRaw.([]interface{})
				if !ok {
					// Fallback to treat non-array as single item array if needed, but per schema it should be array
					values = append(values, []string{fmt.Sprintf("%v", vRaw)})
					continue
				}
				var strVals []string
				for _, val := range vArr {
					strVals = append(strVals, fmt.Sprintf("%v", val))
				}
				values = append(values, strVals)
			}

			// Cartesian product
			combinations := cartesianProduct(values)

			for _, combo := range combinations {
				var suffixParts []string
				for _, val := range combo {
					suffixParts = append(suffixParts, val)
				}
				suffix := strings.Join(suffixParts, ", ")
				newJobName := fmt.Sprintf("%s (%s)", jobName, suffix)

				newJob := deepCopyJob(job)
				newJob.Strategy = nil
				if newJob.Env == nil {
					newJob.Env = make(map[string]string)
				}
				for i, k := range keys {
					newJob.Env[k] = combo[i]
				}
				expanded[newJobName] = newJob
			}
			continue
		}

		expanded[jobName] = job
	}
	return expanded
}

func cartesianProduct(arrays [][]string) [][]string {
	if len(arrays) == 0 {
		return [][]string{}
	}
	if len(arrays) == 1 {
		var res [][]string
		for _, v := range arrays[0] {
			res = append(res, []string{v})
		}
		return res
	}

	var res [][]string
	rest := cartesianProduct(arrays[1:])
	for _, v := range arrays[0] {
		for _, r := range rest {
			combo := append([]string{v}, r...)
			res = append(res, combo)
		}
	}
	return res
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
