package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"yaml-anchor/pkg/logger"
)

var (
	globalConfigPath string
	globalVerbose    bool
)

var rootCmd = &cobra.Command{
	Use:   "anchor",
	Short: "YamlAnchor - CI/CD Pipeline as Code",
	Long: `YamlAnchor treats CI/CD pipelines as type-safe code.

Instead of push → wait → fail → cry, validate, simulate, and auto-fix locally.

Features:
  • Type-safe pipeline definitions
  • Local execution with Dagger Simulation (pkg/simulator)
  • Real-time monitoring with Bubbletea TUI (pkg/tui)
  • Keyless YAML improvement loop with local/offline generation
  • Automatic secret scanning
  • Code analysis and suggestions

Usage:
  anchor generate -c anchor.yaml        Generate GitHub Actions workflow
  anchor improve -c anchor.yaml         Validate, run, and auto-fix YAML
  anchor simulate -c anchor.yaml        Simulate pipeline locally
  anchor server -p 8080                 Start REST API server
  anchor scan ./                        Scan for secrets
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

  # Improve until it validates and runs
  anchor improve -c anchor.yaml --max-iterations 5

  # Start API server
  anchor server -p 8080

  # Test locally
  anchor simulate -c anchor.yaml --dry-run
`,
	Version: "0.1.0",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		level := logger.LevelInfo
		if globalVerbose {
			level = logger.LevelDebug
		}
		logger.Init(level, "")
	},
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

	// Global flags
	rootCmd.PersistentFlags().StringVarP(&globalConfigPath, "config", "c", "anchor.yaml", "Path to anchor.yaml pipeline definition")
	rootCmd.PersistentFlags().BoolVarP(&globalVerbose, "verbose", "v", false, "Enable verbose/debug logging")
}
