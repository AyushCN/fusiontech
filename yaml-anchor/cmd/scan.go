package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"yaml-anchor/pkg/scanner"
)

var (
	scanRecursive bool
	scanEntropy   float64
	scanFormat    string
	installHook   bool
)

var scanCmd = &cobra.Command{
	Use:   "scan [path]",
	Short: "Scan a directory or file for hardcoded secrets and sensitive files",
	Long: `Performs a security audit of your codebase. 
It looks for common secret patterns (AWS, GitHub, Bearer tokens), 
detects high-entropy strings, and identifies sensitive files like .env.`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if installHook {
			err := installPreCommitHook()
			if err != nil {
				fmt.Printf("Failed to install pre-commit hook: %v\n", err)
				os.Exit(1)
			}
			fmt.Println("Pre-commit hook installed successfully! ⚓")
			return
		}

		path := "."
		if len(args) > 0 {
			path = args[0]
		}

		opts := scanner.ScanOptions{
			Recursive:     scanRecursive,
			EntropyLimit:  scanEntropy,
			OutputFormat:  scanFormat,
			IncludeDotEnv: true,
		}

		findings, err := scanner.Scan(path, opts)
		if err != nil {
			fmt.Printf("Scan failed: %v\n", err)
			os.Exit(1)
		}

		fmt.Print(scanner.FormatFindings(findings, scanFormat))

		if len(findings) > 0 && scanFormat == "human" {
			fmt.Printf("\nFound %d potential security issues.\n", len(findings))
			os.Exit(1)
		}
	},
}

func init() {
	scanCmd.Flags().BoolVarP(&scanRecursive, "recursive", "r", true, "Scan subdirectories recursively")
	scanCmd.Flags().Float64VarP(&scanEntropy, "entropy", "e", 4.5, "Entropy threshold (0 to disable)")
	scanCmd.Flags().StringVarP(&scanFormat, "format", "f", "human", "Output format (human, json, github)")
	scanCmd.Flags().BoolVar(&installHook, "install-hook", false, "Install a git pre-commit hook to run anchor scan")
	rootCmd.AddCommand(scanCmd)
}

func installPreCommitHook() error {
	gitDir := ".git"
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		return fmt.Errorf("not a git repository (no .git directory found)")
	}

	hookPath := filepath.Join(gitDir, "hooks", "pre-commit")
	hookContent := `#!/bin/sh
# YamlAnchor Pre-commit Hook
echo "⚓ Running YamlAnchor security scan..."
anchor scan --entropy 4.5 --recursive
if [ $? -ne 0 ]; then
  echo "❌ Security scan failed. Commit aborted."
  exit 1
fi
`

	err := os.WriteFile(hookPath, []byte(hookContent), 0755)
	if err != nil {
		return fmt.Errorf("could not write hook file: %w", err)
	}

	return nil
}
