package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "anchor",
	Short: "YamlAnchor - CI/CD Pipeline Debugger",
	Long: `YamlAnchor treats CI/CD pipelines as type-safe code.

Instead of push → wait → fail → cry, validate and simulate locally.

Features:
  • Type-safe pipeline definitions
  • Local execution with Dagger
  • Real-time monitoring with Pulse Dashboard
  • Automatic secret scanning
  • Code analysis and suggestions

Usage:
  anchor generate -c anchor.yaml        Generate GitHub Actions workflow
  anchor simulate -c anchor.yaml        Simulate pipeline locally
  anchor server -p 8080                 Start REST API server
  anchor clean                          Clean up cache

Examples:
  # Create your pipeline
  echo 'name: my-pipeline
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - run: echo Done!' > anchor.yaml

  # Generate workflow
  anchor generate -c anchor.yaml

  # Start API server
  anchor server -p 8080

  # Test locally
  anchor simulate -c anchor.yaml --dry-run
`,
	Version: "0.1.0",
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.CompletionOptions.DisableDefaultCmd = true
}
