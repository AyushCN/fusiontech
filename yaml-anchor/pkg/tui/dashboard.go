package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"yaml-anchor/pkg/simulator"
)

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#7D56F4")).
			Padding(0, 1).
			MarginBottom(1)

	jobStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#04B575"))

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF0000"))

	logStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#666666")).
			MarginLeft(2)
)

type DashboardModel struct {
	spinner  spinner.Model
	updates  <-chan simulator.UpdateMsg
	
	currentJob  string
	currentStep string
	status      string
	logs        []string
	err         error
	done        bool
}

type updateMsg simulator.UpdateMsg

func waitForUpdate(sub <-chan simulator.UpdateMsg) tea.Cmd {
	return func() tea.Msg {
		msg, ok := <-sub
		if !ok {
			return updateMsg{Status: "done"}
		}
		return updateMsg(msg)
	}
}

func NewDashboard(updates <-chan simulator.UpdateMsg) DashboardModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	return DashboardModel{
		spinner: s,
		updates: updates,
		logs:    make([]string, 0),
	}
}

func (m DashboardModel) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		waitForUpdate(m.updates),
	)
}

func (m DashboardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "q" || msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	case updateMsg:
		if msg.Status == "done" {
			m.done = true
			return m, tea.Quit
		}
		
		if msg.JobName != "" {
			m.currentJob = msg.JobName
		}
		if msg.Step != "" {
			m.currentStep = msg.Step
		}
		m.status = msg.Status
		
		if msg.Error != nil {
			m.err = msg.Error
			m.done = true
			return m, tea.Quit
		}
		
		if msg.LogLine != "" {
			m.logs = append(m.logs, msg.LogLine)
			// keep only last 5 logs
			if len(m.logs) > 5 {
				m.logs = m.logs[1:]
			}
		}
		return m, waitForUpdate(m.updates)

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m DashboardModel) View() string {
	if m.err != nil {
		return fmt.Sprintf("\n%s\n%s\n\n", 
			titleStyle.Render("YamlAnchor Simulation"), 
			errorStyle.Render(fmt.Sprintf("Pipeline Failed: %v", m.err)))
	}

	if m.done {
		return fmt.Sprintf("\n%s\nPipeline Execution Completed Successfully!\n\n", titleStyle.Render("YamlAnchor Simulation"))
	}

	var b strings.Builder
	b.WriteString("\n")
	b.WriteString(titleStyle.Render("YamlAnchor Pulse Dashboard"))
	b.WriteString("\n\n")

	jobStr := jobStyle.Render(m.currentJob)
	
	if m.status == "running" {
		b.WriteString(fmt.Sprintf("%s Running Job: %s\n", m.spinner.View(), jobStr))
	} else {
		b.WriteString(fmt.Sprintf("✓ Job: %s\n", jobStr))
	}
	
	if m.currentStep != "" {
		b.WriteString(fmt.Sprintf("  ↳ Step: %s\n", m.currentStep))
	}

	if len(m.logs) > 0 {
		b.WriteString("\nLogs:\n")
		for _, l := range m.logs {
			b.WriteString(logStyle.Render(l) + "\n")
		}
	}

	b.WriteString("\n(press q to quit)\n")
	return b.String()
}
