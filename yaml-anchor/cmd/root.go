package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "anchor",
	Short: "YamlAnchor treats CI/CD pipelines as type-safe code",
	Long: `YamlAnchor is a developer tool that translates Go-based CI/CD logic 
into valid GitHub Actions YAML, with the ability to simulate and 
run those pipelines locally using Dagger.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
