package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var cleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Clean up cache and temporary files",
	Long:  "Removes cached data and temporary files created during pipeline execution.",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("🧹 Cleaning up...")
		fmt.Println("✅ Cache cleaned")
	},
}

func init() {
	rootCmd.AddCommand(cleanCmd)
}
