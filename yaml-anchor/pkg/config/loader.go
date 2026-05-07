package config

import (
	"encoding/json"
	"fmt"
	"os"
	
	"gopkg.in/yaml.v3"
	"yaml-anchor/pkg/schema"
)

// Load reads a YAML configuration file and returns a validated Pipeline
func Load(filepath string) (*schema.Pipeline, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %q: %w", filepath, err)
	}

	return ParseYAML(string(data))
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

	for jobName, job := range jobs {
		if job.Strategy != nil && job.Strategy.Matrix != nil {
			matrixRaw, ok := job.Strategy.Matrix.(map[string]interface{})
			if !ok {
				expanded[jobName] = job
				continue
			}

			var keys []string
			var values [][]string

			for k, vRaw := range matrixRaw {
				vArr, ok := vRaw.([]interface{})
				if !ok {
					continue
				}
				keys = append(keys, k)
				var strVals []string
				for _, val := range vArr {
					strVals = append(strVals, fmt.Sprintf("%v", val))
				}
				values = append(values, strVals)
			}

			if len(keys) == 1 {
				k := keys[0]
				for _, val := range values[0] {
					newJobName := fmt.Sprintf("%s (%s)", jobName, val)
					newJob := deepCopyJob(job)
					newJob.Strategy = nil
					if newJob.Env == nil {
						newJob.Env = make(map[string]string)
					}
					newJob.Env[k] = val
					expanded[newJobName] = newJob
				}
				continue
			}
		}

		expanded[jobName] = job
	}
	return expanded
}

func deepCopyJob(job *schema.Job) *schema.Job {
	data, _ := json.Marshal(job)
	var newJob schema.Job
	json.Unmarshal(data, &newJob)
	return &newJob
}
