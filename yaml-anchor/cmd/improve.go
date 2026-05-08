package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"yaml-anchor/pkg/config"
	"yaml-anchor/pkg/improver"
	"yaml-anchor/pkg/simulator"
)

var (
	improveConfigPath    string
	improveMaxIterations int
	improveSkipRun       bool
	improveBackup        bool
)

var improveCmd = &cobra.Command{
	Use:   "improve",
	Short: "Iteratively fix anchor.yaml until it validates and runs",
	Long: `Validates anchor.yaml, runs it locally, and rewrites the YAML when
validation or execution fails. Uses a local Ollama model when available and
falls back to YamlAnchor's built-in offline generator. No API key required.`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()

		for iteration := 1; iteration <= improveMaxIterations; iteration++ {
			fmt.Printf("🔄 Improve iteration %d/%d\n", iteration, improveMaxIterations)

			currentYAML, err := os.ReadFile(improveConfigPath)
			if err != nil {
				fmt.Printf("❌ Could not read %s: %v\n", improveConfigPath, err)
				os.Exit(1)
			}

			pipeline, err := config.ParseYAML(string(currentYAML))
			if err != nil {
				fmt.Printf("❌ Validation failed: %v\n", err)
				if err := rewriteImprovedYAML(ctx, string(currentYAML), err.Error()); err != nil {
					fmt.Printf("❌ Could not improve YAML: %v\n", err)
					os.Exit(1)
				}
				continue
			}

			if improveSkipRun {
				fmt.Println("✅ YAML validates. Skipping local execution because --skip-run is enabled.")
				return
			}

			result := simulator.Run(ctx, pipeline)
			if result.ExitCode == 0 {
				fmt.Printf("✅ Pipeline passed on iteration %d\n", iteration)
				return
			}

			reason := strings.TrimSpace(result.Logs)
			if result.Error != nil {
				reason += "\n" + result.Error.Error()
			}
			fmt.Println("❌ Local simulation failed; improving anchor.yaml...")
			if err := rewriteImprovedYAML(ctx, string(currentYAML), reason); err != nil {
				fmt.Printf("❌ Could not improve YAML: %v\n", err)
				os.Exit(1)
			}
		}

		fmt.Printf("❌ Reached max iterations (%d). Manual review needed.\n", improveMaxIterations)
		os.Exit(1)
	},
}

func rewriteImprovedYAML(ctx context.Context, currentYAML, reason string) error {
	improved, source, err := improver.ImproveYAML(ctx, currentYAML, reason)
	if err != nil {
		return err
	}
	if _, err := config.ParseYAML(improved); err != nil {
		return fmt.Errorf("improved YAML is still invalid: %w", err)
	}
	if improveBackup && improved != currentYAML {
		backupPath := improveConfigPath + ".bak"
		if err := os.WriteFile(backupPath, []byte(currentYAML), 0644); err != nil {
			return fmt.Errorf("failed to write backup %s: %w", backupPath, err)
		}
	}
	if err := os.WriteFile(improveConfigPath, []byte(improved), 0644); err != nil {
		return err
	}
	fmt.Printf("🛠️  Rewrote %s using %s generator\n", improveConfigPath, source)
	return nil
}

func init() {
	improveCmd.Flags().StringVarP(&improveConfigPath, "config", "c", "anchor.yaml", "Path to anchor.yaml pipeline definition")
	improveCmd.Flags().IntVar(&improveMaxIterations, "max-iterations", 5, "Maximum validation/run/fix attempts")
	improveCmd.Flags().BoolVar(&improveSkipRun, "skip-run", false, "Only validate and rewrite YAML without running Docker/Dagger simulation")
	improveCmd.Flags().BoolVar(&improveBackup, "backup", true, "Write a .bak copy before replacing the YAML file")
	rootCmd.AddCommand(improveCmd)
}
