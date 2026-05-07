package cmd

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

// Build-time variables — injected by GoReleaser via -ldflags
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("⚓ YamlAnchor\n")
		fmt.Printf("  Version : %s\n", version)
		fmt.Printf("  Commit  : %s\n", commit)
		fmt.Printf("  Built   : %s\n", date)
		fmt.Printf("  Go      : %s\n", runtime.Version())
		fmt.Printf("  OS/Arch : %s/%s\n", runtime.GOOS, runtime.GOARCH)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
