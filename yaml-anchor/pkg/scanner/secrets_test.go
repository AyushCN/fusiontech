package scanner_test

import (
	"testing"

	"yaml-anchor/pkg/scanner"
	"yaml-anchor/pkg/schema"
)

func cleanPipeline() *schema.Pipeline {
	return &schema.Pipeline{
		Name: "Test Pipeline",
		On:   &schema.Triggers{Push: &schema.PushTrigger{Branches: []string{"main"}}},
		Jobs: map[string]schema.Job{
			"build": {
				RunsOn: "ubuntu-latest",
				Steps: []schema.Step{
					{Name: "Run Tests", Run: "go test ./..."},
				},
			},
		},
	}
}

func TestHasSecret_CleanPipeline(t *testing.T) {
	if err := scanner.HasSecret(cleanPipeline()); err != nil {
		t.Errorf("expected no secret in clean pipeline, got: %v", err)
	}
}

func TestHasSecret_AWSKey(t *testing.T) {
	p := cleanPipeline()
	p.Jobs["build"] = schema.Job{
		RunsOn: "ubuntu-latest",
		Steps: []schema.Step{
			{Name: "Deploy", Run: "aws s3 cp . s3://bucket --access-key AKIAIOSFODNN7EXAMPLE"},
		},
	}
	if err := scanner.HasSecret(p); err == nil {
		t.Error("expected AWS key to be detected, got nil")
	}
}

func TestHasSecret_GitHubToken(t *testing.T) {
	p := cleanPipeline()
	p.Jobs["build"] = schema.Job{
		RunsOn: "ubuntu-latest",
		Steps: []schema.Step{
			{Name: "Push", Run: "git push https://ghp_ABCDEFGHIJKLMNOPQRSTUVWXYZ012345678901:x-oauth-basic@github.com/user/repo"},
		},
	}
	if err := scanner.HasSecret(p); err == nil {
		t.Error("expected GitHub token to be detected, got nil")
	}
}

func TestHasSecret_BearerToken(t *testing.T) {
	p := cleanPipeline()
	p.Jobs["build"] = schema.Job{
		RunsOn: "ubuntu-latest",
		Steps: []schema.Step{
			{Name: "API Call", Run: `curl -H "Authorization: Bearer eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9" https://api.example.com`},
		},
	}
	if err := scanner.HasSecret(p); err == nil {
		t.Error("expected Bearer token to be detected, got nil")
	}
}

func TestHasSecret_MultipleJobs(t *testing.T) {
	p := cleanPipeline()
	// Second job has a secret, first is clean
	p.Jobs["deploy"] = schema.Job{
		RunsOn: "ubuntu-latest",
		Steps: []schema.Step{
			{Name: "Upload", Run: "upload --key AKIAIOSFODNN7EXAMPLE --secret mysecret"},
		},
	}
	if err := scanner.HasSecret(p); err == nil {
		t.Error("expected secret to be detected in multi-job pipeline")
	}
}
