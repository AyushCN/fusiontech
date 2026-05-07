package cmd

import (
	"context"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"yaml-anchor/pkg/config"
	"yaml-anchor/pkg/simulator"
	"yaml-anchor/pkg/tui"
)

var localConfigPath string

var localCmd = &cobra.Command{
	Use:   "local",
	Short: "Run the pipeline locally using Dagger with the Pulse TUI",
	Long: `Reads your pipeline definition from an anchor.yaml file,
spins up isolated Docker containers via Dagger, and streams
live execution logs to the Pulse interactive dashboard.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Loading pipeline config from %s...\n", localConfigPath)

		pipeline, err := config.Load(localConfigPath)
		if err != nil {
			fmt.Printf("Failed to load config: %v\n", err)
			os.Exit(1)
		}

		ctx := context.Background()
		updates := make(chan simulator.UpdateMsg, 64)

		// Run simulation in background
		go simulator.RunLocal(ctx, pipeline, updates)

		// Start Bubbletea TUI
		m := tui.NewDashboard(updates)
		p := tea.NewProgram(m, tea.WithAltScreen())
		finalM, err := p.Run()
		if err != nil {
			fmt.Printf("Error running dashboard: %v\n", err)
			os.Exit(1)
		}

		finalModel := finalM.(tui.DashboardModel)
		if finalModel.Err != nil {
			fmt.Printf("\n[YAML-ANCHOR ERROR] Simulation failed:\n%v\n", finalModel.Err)
			os.Exit(1)
		}
	},
}

func init() {
	localCmd.Flags().StringVarP(&localConfigPath, "config", "c", "anchor.yaml",
		"Path to your anchor.yaml pipeline definition")
	rootCmd.AddCommand(localCmd)
}
