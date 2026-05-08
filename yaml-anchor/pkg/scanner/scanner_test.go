package scanner_test

import (
	"os"
	"path/filepath"
	"testing"

	"yaml-anchor/pkg/scanner"
)

func TestScan_SlackToken(t *testing.T) {
	f, err := os.CreateTemp("", "slack-*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	// Build token at runtime so no literal token appears in source
	slackToken := "xo" + "xb-" + "000000000000-" + "ZZZZZZZZZZZZZZZZZZZZabcde"
	f.WriteString("token: " + slackToken + "\n")
	f.Close()

	findings, err := scanner.Scan(f.Name(), scanner.ScanOptions{})
	if err != nil {
		t.Fatalf("Scan() error: %v", err)
	}
	if len(findings) == 0 {
		t.Error("Expected Slack token to be detected, got no findings")
	}
}

func TestScan_SSHPrivateKey(t *testing.T) {
	f, err := os.CreateTemp("", "ssh-*.sh")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	f.WriteString("-----BEGIN RSA PRIVATE KEY-----\nMIIE...\n-----END RSA PRIVATE KEY-----\n")
	f.Close()

	findings, err := scanner.Scan(f.Name(), scanner.ScanOptions{})
	if err != nil {
		t.Fatalf("Scan() error: %v", err)
	}
	if len(findings) == 0 {
		t.Error("Expected SSH private key to be detected, got no findings")
	}
}

func TestScan_PasswordAssignment(t *testing.T) {
	f, err := os.CreateTemp("", "config-*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	f.WriteString("db_password = mysecretpassword123\n")
	f.Close()

	findings, err := scanner.Scan(f.Name(), scanner.ScanOptions{})
	if err != nil {
		t.Fatalf("Scan() error: %v", err)
	}
	if len(findings) == 0 {
		t.Error("Expected password assignment to be detected, got no findings")
	}
}

func TestScan_AzureJWT(t *testing.T) {
	f, err := os.CreateTemp("", "azure-*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	// Realistic-looking JWT-shaped token
	f.WriteString("token: eyJhbGciOiJSUzI1NiJ9.eyJzdWIiOiJ1c2VyQGV4YW1wbGUuY29tIn0.SflKxwRJSMeKKF2QT4fwpMeJf\n")
	f.Close()

	findings, err := scanner.Scan(f.Name(), scanner.ScanOptions{})
	if err != nil {
		t.Fatalf("Scan() error: %v", err)
	}
	if len(findings) == 0 {
		t.Error("Expected Azure/JWT token to be detected, got no findings")
	}
}

func TestScan_CleanFile_NoFindings(t *testing.T) {
	f, err := os.CreateTemp("", "clean-*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	f.WriteString("name: My Pipeline\njobs:\n  build:\n    runs-on: ubuntu-latest\n")
	f.Close()

	findings, err := scanner.Scan(f.Name(), scanner.ScanOptions{})
	if err != nil {
		t.Fatalf("Scan() error: %v", err)
	}
	if len(findings) != 0 {
		t.Errorf("Expected no findings for clean file, got %d: %+v", len(findings), findings)
	}
}

func TestScan_EnvFile_FlagsAsSensitive(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".env")
	if err := os.WriteFile(path, []byte("SECRET_KEY=abc123\n"), 0644); err != nil {
		t.Fatal(err)
	}

	findings, err := scanner.Scan(path, scanner.ScanOptions{IncludeDotEnv: true})
	if err != nil {
		t.Fatalf("Scan() error: %v", err)
	}
	if len(findings) == 0 {
		t.Error("Expected .env file to be flagged as sensitive")
	}
}
