package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"yaml-anchor/pkg/config"
	"yaml-anchor/pkg/generator"
	"yaml-anchor/pkg/logger"
)

var (
	generateOutputPath string
	generateDryRun     bool
)

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate GitHub Actions workflow from anchor.yaml",
	Long: `Reads your pipeline definition from an anchor.yaml file,
validates it, scans for secrets, and writes a GitHub Actions
workflow file to .github/workflows/main.yml.`,
	Run: func(cmd *cobra.Command, args []string) {
		logger.Info("📖 Loading pipeline from %s...", globalConfigPath)

		pipeline, err := config.Load(globalConfigPath)
		if err != nil {
			logger.Error("❌ Failed to load config: %v\nSuggestion: Check if the file exists and is valid YAML.", err)
			os.Exit(1)
		}

		logger.Info("🔍 Validating pipeline: %q...", pipeline.Name)

		// Validate
		validationErrors := generator.ValidatePipeline(pipeline)
		if len(validationErrors) > 0 {
			logger.Error("❌ Validation errors found in pipeline")
			for _, e := range validationErrors {
				logger.Error("   - %s", e.Error())
			}
			os.Exit(1)
		}

		logger.Debug("Validation passed.")

		if generateDryRun {
			logger.Info("✅ Dry-run successful! Pipeline is valid and free of CRITICAL secrets.")
			logger.Info("Skipping YAML generation because --dry-run is enabled.")
			return
		}

		logger.Info("✨ Generating YAML...")

		if err := generator.ExportYAML(pipeline, generateOutputPath); err != nil {
			logger.Error("❌ Error generating YAML: %v", err)
			os.Exit(1)
		}

		logger.Info("✅ Successfully generated workflow at %s", generateOutputPath)
	},
}

func init() {
	generateCmd.Flags().StringVarP(&generateOutputPath, "output", "o", ".github/workflows/main.yml",
		"Output path for generated workflow file")
	generateCmd.Flags().BoolVar(&generateDryRun, "dry-run", false,
		"Validate pipeline without writing any files to disk")
	
	rootCmd.AddCommand(generateCmd)
}
