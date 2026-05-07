package cmd

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"yaml-anchor/pkg/config"
	"yaml-anchor/pkg/generator"
)

var generateConfigPath string

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate a GitHub Actions YAML from an anchor.yaml config",
	Long: `Reads your pipeline definition from an anchor.yaml file,
scans for hardcoded secrets, and writes a valid GitHub Actions
workflow file to .github/workflows/main.yml.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Loading pipeline config from %s...\n", generateConfigPath)

		pipeline, err := config.Load(generateConfigPath)
		if err != nil {
			log.Fatalf("Failed to load config: %v", err)
		}

		fmt.Printf("Generating YAML for pipeline: %q\n", pipeline.Name)

		outputPath := ".github/workflows/main.yml"
		if err := generator.ExportYAML(pipeline, outputPath); err != nil {
			log.Fatalf("Error generating YAML: %v", err)
		}

		fmt.Printf("✓ Successfully generated workflow at %s\n", outputPath)
	},
}

func init() {
	generateCmd.Flags().StringVarP(&generateConfigPath, "config", "c", "anchor.yaml",
		"Path to your anchor.yaml pipeline definition")
	rootCmd.AddCommand(generateCmd)
}
