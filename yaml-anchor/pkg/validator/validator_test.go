package validator

import "testing"

func TestValidateJobID_Valid(t *testing.T) {
	cases := []string{"build", "test-job", "deploy_prod", "Job123"}
	for _, id := range cases {
		if err := ValidateJobID(id); err != nil {
			t.Errorf("ValidateJobID(%q) unexpected error: %v", id, err)
		}
	}
}

func TestValidateJobID_Invalid(t *testing.T) {
	cases := []string{"", "bad job", "has spaces", "has@symbol", string(make([]byte, 51))}
	for _, id := range cases {
		if err := ValidateJobID(id); err == nil {
			t.Errorf("ValidateJobID(%q) expected error, got nil", id)
		}
	}
}

func TestValidateStepName_Valid(t *testing.T) {
	if err := ValidateStepName("Run Tests"); err != nil {
		t.Errorf("ValidateStepName() unexpected error: %v", err)
	}
}

func TestValidateStepName_Empty(t *testing.T) {
	if err := ValidateStepName(""); err == nil {
		t.Error("ValidateStepName() expected error for empty name")
	}
}

func TestValidateRunsOn_Valid(t *testing.T) {
	valid := []string{"ubuntu-latest", "ubuntu-22.04", "windows-latest", "macos-latest", "self-hosted"}
	for _, r := range valid {
		if err := ValidateRunsOn(r); err != nil {
			t.Errorf("ValidateRunsOn(%q) unexpected error: %v", r, err)
		}
	}
}

func TestValidateRunsOn_Invalid(t *testing.T) {
	if err := ValidateRunsOn("solaris-latest"); err == nil {
		t.Error("ValidateRunsOn() expected error for unknown runner")
	}
}

func TestValidateCron_Valid(t *testing.T) {
	if err := ValidateCron("0 0 * * *"); err != nil {
		t.Errorf("ValidateCron() unexpected error: %v", err)
	}
}

func TestValidateCron_Invalid(t *testing.T) {
	cases := []string{"* * *", "0 0 0 0 0 0", "", "not-a-cron"}
	for _, c := range cases {
		if err := ValidateCron(c); err == nil {
			t.Errorf("ValidateCron(%q) expected error, got nil", c)
		}
	}
}

func TestValidatePipelineName(t *testing.T) {
	if err := ValidatePipelineName("My CI Pipeline"); err != nil {
		t.Errorf("ValidatePipelineName() unexpected error: %v", err)
	}
	if err := ValidatePipelineName(""); err == nil {
		t.Error("ValidatePipelineName() expected error for empty name")
	}
}
