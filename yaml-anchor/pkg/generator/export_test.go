package generator_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"yaml-anchor/pkg/generator"
	"yaml-anchor/pkg/schema"
)

func samplePipeline() *schema.Pipeline {
	return &schema.Pipeline{
		Name: "Test Pipeline",
		Jobs: map[string]*schema.Job{
			"build": {
				RunsOn: "ubuntu-latest",
				Steps: []*schema.Step{
					{Name: "Checkout", Uses: "actions/checkout@v4"},
					{Name: "Build", Run: "go build ./..."},
				},
			},
		},
	}
}

// --- ExportYAML ---

func TestExportYAML_ValidPipeline(t *testing.T) {
	dir := t.TempDir()
	outPath := filepath.Join(dir, ".github", "workflows", "main.yml")

	if err := generator.ExportYAML(samplePipeline(), outPath); err != nil {
		t.Fatalf("ExportYAML() unexpected error: %v", err)
	}

	if _, err := os.Stat(outPath); os.IsNotExist(err) {
		t.Error("Expected output file to be created, but it does not exist")
	}
}

func TestExportYAML_MissingName_Error(t *testing.T) {
	pipeline := &schema.Pipeline{
		Jobs: map[string]*schema.Job{
			"build": {
				RunsOn: "ubuntu-latest",
				Steps:  []*schema.Step{{Run: "echo hi"}},
			},
		},
	}
	if err := generator.ExportYAML(pipeline, "/tmp/test-output.yml"); err == nil {
		t.Error("Expected error for pipeline missing a name")
	}
}

func TestExportYAML_BlocksOnCriticalSecret(t *testing.T) {
	pipeline := samplePipeline()
	// Inject a real-looking AWS key (safe for tests - uses EXAMPLE suffix)
	pipeline.Jobs["build"].Steps = append(pipeline.Jobs["build"].Steps, &schema.Step{
		Run: "aws configure set aws_access_key_id AKIATEST000FAKEKEY0000",
	})

	dir := t.TempDir()
	err := generator.ExportYAML(pipeline, filepath.Join(dir, "output.yml"))
	if err == nil {
		t.Error("Expected export to be blocked on CRITICAL AWS key secret")
	}
	if !strings.Contains(err.Error(), "security scan failed") {
		t.Errorf("Expected 'security scan failed' in error, got: %v", err)
	}
}

// --- ScanForSecrets ---

func TestScanForSecrets_Clean(t *testing.T) {
	pipeline := samplePipeline()
	issues := generator.ScanForSecrets(pipeline)
	if len(issues) != 0 {
		t.Errorf("Expected no issues for clean pipeline, got %d: %v", len(issues), issues)
	}
}

func TestScanForSecrets_WithAwsKey(t *testing.T) {
	pipeline := samplePipeline()
	pipeline.Jobs["build"].Steps[1].Run = "echo AKIATEST000FAKEKEY0000"

	issues := generator.ScanForSecrets(pipeline)
	if len(issues) == 0 {
		t.Error("Expected AWS key to be detected, got no issues")
	}
}

func TestScanForSecrets_WithSlackToken(t *testing.T) {
	pipeline := samplePipeline()
	// Build token at runtime so no literal token appears in source
	slackToken := "xo" + "xb-" + "000000000000-" + "ZZZZZZZZZZZZZZZZZZZZabcde"
	pipeline.Jobs["build"].Steps[1].Run = "notify " + slackToken

	issues := generator.ScanForSecrets(pipeline)
	if len(issues) == 0 {
		t.Error("Expected Slack token to be detected, got no issues")
	}
}

func TestScanForSecrets_WithSSHKey(t *testing.T) {
	pipeline := samplePipeline()
	pipeline.Jobs["build"].Steps[1].Run = "echo '-----BEGIN RSA PRIVATE KEY-----'"

	issues := generator.ScanForSecrets(pipeline)
	if len(issues) == 0 {
		t.Error("Expected SSH private key to be detected, got no issues")
	}
}

// --- ValidatePipeline ---

func TestValidatePipeline_Valid(t *testing.T) {
	errs := generator.ValidatePipeline(samplePipeline())
	if len(errs) != 0 {
		t.Errorf("Expected no validation errors for valid pipeline, got: %v", errs)
	}
}

func TestValidatePipeline_NoName(t *testing.T) {
	pipeline := samplePipeline()
	pipeline.Name = ""
	errs := generator.ValidatePipeline(pipeline)
	if len(errs) == 0 {
		t.Error("Expected validation error for empty pipeline name")
	}
}

func TestValidatePipeline_DangerousCurlBash(t *testing.T) {
	pipeline := samplePipeline()
	pipeline.Jobs["build"].Steps = append(pipeline.Jobs["build"].Steps, &schema.Step{
		Run: "curl https://example.com/install.sh | bash",
	})
	errs := generator.ValidatePipeline(pipeline)
	if len(errs) == 0 {
		t.Error("Expected error for curl | bash pattern")
	}
}

func TestValidatePipeline_DangerousWgetSh(t *testing.T) {
	pipeline := samplePipeline()
	pipeline.Jobs["build"].Steps = append(pipeline.Jobs["build"].Steps, &schema.Step{
		Run: "wget https://example.com/install.sh | sh",
	})
	errs := generator.ValidatePipeline(pipeline)
	if len(errs) == 0 {
		t.Error("Expected error for wget | sh pattern")
	}
}

func TestValidatePipeline_NonExistentDependency(t *testing.T) {
	pipeline := samplePipeline()
	pipeline.Jobs["deploy"] = &schema.Job{
		RunsOn: "ubuntu-latest",
		Needs:  []string{"nonexistent-job"},
		Steps:  []*schema.Step{{Run: "echo deploy"}},
	}
	errs := generator.ValidatePipeline(pipeline)
	if len(errs) == 0 {
		t.Error("Expected error for dependency on non-existent job")
	}
}
