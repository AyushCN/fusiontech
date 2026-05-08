package cmd

import (
	"fmt"
	"os"

	"yaml-anchor/pkg/blueprints"
	"yaml-anchor/pkg/detector"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/sergi/go-diff/diffmatchpatch"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new anchor.yaml file for your project",
	Long: `Scans the current directory to detect your tech stack (Go, Node, etc.) 
and generates a recommended anchor.yaml configuration.`,
	Run: func(cmd *cobra.Command, args []string) {
		existingYAML := ""
		if data, err := os.ReadFile("anchor.yaml"); err == nil {
			existingYAML = string(data)
		}

		profile, err := detector.Detect(".")
		if err != nil {
			fmt.Printf("Detection failed: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("⚓ Detected stack: %s\n", profile.Stack)
		if profile.HasExistingCI {
			fmt.Println("⚠️  Note: Existing GitHub Actions detected in .github/workflows/")
		}

		suggestedYAML := blueprints.MapToYAML(profile)
		isOverwrite := existingYAML != ""

		if isOverwrite {
			fmt.Println("\n⚠️  anchor.yaml already exists. Here is the diff:")
			dmp := diffmatchpatch.New()
			diffs := dmp.DiffMain(existingYAML, suggestedYAML, false)
			fmt.Println(dmp.DiffPrettyText(diffs))
		} else {
			fmt.Println("\nProposed anchor.yaml:")
			fmt.Println("---")
			fmt.Println(suggestedYAML)
			fmt.Println("---")
		}

		p := tea.NewProgram(initialModel(suggestedYAML, isOverwrite))
		if _, err := p.Run(); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
	},
}

type initModel struct {
	yaml        string
	choice      string
	quitting    bool
	isOverwrite bool
}

func initialModel(yaml string, isOverwrite bool) initModel {
	return initModel{yaml: yaml, isOverwrite: isOverwrite}
}

func (m initModel) Init() tea.Cmd {
	return nil
}

func (m initModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "y", "Y", "enter":
			m.choice = "yes"
			m.quitting = true
			return m, tea.Quit
		case "n", "N", "q", "ctrl+c":
			m.choice = "no"
			m.quitting = true
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m initModel) View() string {
	if m.quitting {
		if m.choice == "yes" {
			err := os.WriteFile("anchor.yaml", []byte(m.yaml), 0644)
			if err != nil {
				return fmt.Sprintf("\n❌ Error writing file: %v\n", err)
			}
			return "\n✅ anchor.yaml generated successfully!\n"
		}
		return "\n❌ Initialization cancelled.\n"
	}
	if m.isOverwrite {
		return "\nOverwrite anchor.yaml? [Y/n] "
	}
	return "\nWrite anchor.yaml? [Y/n] "
}

func init() {
	rootCmd.AddCommand(initCmd)
}
