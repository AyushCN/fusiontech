package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"yaml-anchor/pkg/config"
	"yaml-anchor/pkg/simulator"
)

var execCmd = &cobra.Command{
	Use:   "exec [job_name]",
	Short: "Drop into an interactive shell inside the simulated runner",
	Long: `Starts a Dagger container matching the environment for the specified job
and attaches your terminal for interactive debugging.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		jobName := args[0]
		configFile, _ := cmd.Flags().GetString("config")

		pipeline, err := config.Load(configFile)
		if err != nil {
			fmt.Printf("Error loading config: %v\n", err)
			os.Exit(1)
		}

		ctx := context.Background()
		err = simulator.RunInteractive(ctx, pipeline, jobName)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	execCmd.Flags().StringP("config", "c", "anchor.yaml", "Path to anchor.yaml config file")
	rootCmd.AddCommand(execCmd)
}
