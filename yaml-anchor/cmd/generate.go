package cmd

import (
	"os"
	"strings"

	"github.com/spf13/cobra"
	"yaml-anchor/pkg/config"
	"yaml-anchor/pkg/exporter/gitlab"
	"yaml-anchor/pkg/generator"
	"yaml-anchor/pkg/logger"
)

var (
	generateOutputPath string
	generateDryRun     bool
	generatePlatform   string
)

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate CI workflow from anchor.yaml",
	Long: `Reads your pipeline definition from an anchor.yaml file,
validates it, scans for secrets, and writes a CI workflow file.

Supported platforms (--platform flag):
  github    → .github/workflows/main.yml   (default)
  gitlab    → .gitlab-ci.yml`,
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

		platform := strings.ToLower(strings.TrimSpace(generatePlatform))

		switch platform {
		case "github", "":
			logger.Info("✨ Generating GitHub Actions workflow...")
			outPath := generateOutputPath
			if outPath == "" {
				outPath = ".github/workflows/main.yml"
			}
			if err := generator.ExportYAML(pipeline, outPath); err != nil {
				logger.Error("❌ Error generating YAML: %v", err)
				os.Exit(1)
			}
			logger.Info("✅ Successfully generated GitHub Actions workflow at %s", outPath)

		case "gitlab":
			logger.Info("✨ Generating GitLab CI configuration...")
			e := gitlab.New()
			outPath := generateOutputPath
			if outPath == "" {
				outPath = e.DefaultOutputPath()
			}
			data, err := e.Export(pipeline)
			if err != nil {
				logger.Error("❌ Error generating GitLab CI YAML: %v", err)
				os.Exit(1)
			}
			if err := os.MkdirAll(dirOf(outPath), 0755); err != nil {
				logger.Error("❌ Could not create output directory: %v", err)
				os.Exit(1)
			}
			if err := os.WriteFile(outPath, data, 0644); err != nil {
				logger.Error("❌ Could not write %s: %v", outPath, err)
				os.Exit(1)
			}
			logger.Info("✅ Successfully generated GitLab CI config at %s", outPath)

		default:
			logger.Error("❌ Unknown platform %q. Supported: github, gitlab", platform)
			os.Exit(1)
		}
	},
}

func dirOf(path string) string {
	i := strings.LastIndex(path, "/")
	if i < 0 {
		return "."
	}
	return path[:i]
}

func init() {
	generateCmd.Flags().StringVarP(&generateOutputPath, "output", "o", "",
		"Output path for generated workflow file (default depends on --platform)")
	generateCmd.Flags().BoolVar(&generateDryRun, "dry-run", false,
		"Validate pipeline without writing any files to disk")
	generateCmd.Flags().StringVar(&generatePlatform, "platform", "github",
		"CI platform to target: github (default), gitlab")

	rootCmd.AddCommand(generateCmd)
}
