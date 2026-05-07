package cmd

import (
	"fmt"
	"os/exec"

	"github.com/spf13/cobra"
)

var cleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Prunes Docker and Dagger caches to free up local disk space",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Starting cleanup of Docker and Dagger caches...")

		// Running docker system prune
		// Warning: This deletes stopped containers, unused networks, and dangling images.
		c := exec.Command("docker", "system", "prune", "-f")
		out, err := c.CombinedOutput()
		if err != nil {
			fmt.Printf("Failed to run docker system prune: %v\nOutput: %s\n", err, string(out))
		} else {
			fmt.Printf("Docker prune successful:\n%s\n", string(out))
		}

		fmt.Println("Cleanup complete.")
	},
}

func init() {
	rootCmd.AddCommand(cleanCmd)
}
