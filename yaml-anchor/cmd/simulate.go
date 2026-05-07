package cmd

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"yaml-anchor/pkg/config"
)

var simulateConfigPath string
var dryRun bool

var simulateCmd = &cobra.Command{
	Use:   "simulate",
	Short: "Simulate pipeline execution locally",
	Long: `Reads your anchor.yaml configuration and simulates the entire
pipeline execution locally. Use --dry-run to preview without executing.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("📖 Loading pipeline from %s...\n", simulateConfigPath)

		pipeline, err := config.Load(simulateConfigPath)
		if err != nil {
			log.Fatalf("❌ Failed to load config: %v", err)
		}

		fmt.Printf("🚀 Simulating pipeline: %q...\n", pipeline.Name)

		if dryRun {
			fmt.Println("[DRY RUN] Would execute:")
			for jobID, job := range pipeline.Jobs {
				fmt.Printf("\n  Job: %s\n", jobID)
				fmt.Printf("    Runs on: %s\n", job.RunsOn)
				fmt.Println("    Steps:")
				for _, step := range job.Steps {
					fmt.Printf("      - %s\n", step.Name)
					if step.Uses != "" {
						fmt.Printf("        uses: %s\n", step.Uses)
					}
					if step.Run != "" {
						fmt.Printf("        run: %s\n", step.Run)
					}
				}
			}
			fmt.Println("\n✅ Dry-run preview complete")
			return
		}

	fmt.Println("💡 Full simulation requires Docker (coming soon)")
	},
}

func init() {
	simulateCmd.Flags().StringVarP(&simulateConfigPath, "config", "c", "anchor.yaml",
		"Path to your anchor.yaml pipeline definition")
	simulateCmd.Flags().BoolVarP(&dryRun, "dry-run", "d", false,
		"Preview execution without actually running")
	rootCmd.AddCommand(simulateCmd)
}
