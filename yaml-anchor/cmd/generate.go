package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
	"yaml-anchor/pkg/config"
	"yaml-anchor/pkg/generator"
)

var generateConfigPath string
var generateOutputPath string

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate GitHub Actions workflow from anchor.yaml",
	Long: `Reads your pipeline definition from an anchor.yaml file,
validates it, scans for secrets, and writes a GitHub Actions
workflow file to .github/workflows/main.yml.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("📖 Loading pipeline from %s...\n", generateConfigPath)

		pipeline, err := config.Load(generateConfigPath)
		if err != nil {
			log.Fatalf("❌ Failed to load config: %v", err)
		}

		fmt.Printf("🔍 Validating pipeline: %q...\n", pipeline.Name)

		// Validate
		validationErrors := generator.ValidatePipeline(pipeline)
		if len(validationErrors) > 0 {
			fmt.Println("❌ Validation errors:")
			for _, e := range validationErrors {
				fmt.Printf("   - %s\n", e)
			}
			os.Exit(1)
		}

		fmt.Printf("✨ Generating YAML...\n")

		if err := generator.ExportYAML(pipeline, generateOutputPath); err != nil {
			log.Fatalf("❌ Error generating YAML: %v", err)
		}

		fmt.Printf("✅ Successfully generated workflow at %s\n", generateOutputPath)
	},
}

func init() {
	generateCmd.Flags().StringVarP(&generateConfigPath, "config", "c", "anchor.yaml",
		"Path to your anchor.yaml pipeline definition")
	generateCmd.Flags().StringVarP(&generateOutputPath, "output", "o", ".github/workflows/main.yml",
		"Output path for generated workflow file")
	rootCmd.AddCommand(generateCmd)
}
