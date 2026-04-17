package gitpreflight

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/smasonuk/falken-term/internal/ui/app"
	"github.com/smasonuk/falken-term/internal/ui/tui/shared"
)

type Model struct {
	dirtyFiles    string
	actions       Actions
	lastError     string
	isRunning     bool
	runningAction string
	width         int
	height        int
}

type Actions struct {
	Stash     func() tea.Cmd
	CommitWIP func() tea.Cmd
}

func NewModel(dirtyFiles string, actions Actions) Model {
	return Model{
		dirtyFiles: dirtyFiles,
		actions:    actions,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.isRunning {
			return m, nil
		}

		switch msg.String() {
		case "s":
			if m.actions.Stash == nil {
				return m, nil
			}
			m.isRunning = true
			m.runningAction = "stash"
			return m, m.actions.Stash()
		case "c":
			if m.actions.CommitWIP == nil {
				return m, nil
			}
			m.isRunning = true
			m.runningAction = "commit_wip"
			return m, m.actions.CommitWIP()
		case "i":
			return m, func() tea.Msg { return app.GitPreflightActionMsg{Action: "ignore"} }
		case "q", "esc", "ctrl+c":
			return m, tea.Quit
		}
	case app.GitPreflightFinishedMsg:
		m.isRunning = false
		m.runningAction = ""
		return m.SetError(msg.Err), nil
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}
	return m, nil
}

func (m Model) View() string {
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("208")). // Orange
		Padding(1, 3).
		Align(lipgloss.Left)

	content := fmt.Sprintf(
		"%s UNCOMMITTED CHANGES DETECTED\n\n"+
			"The following files have changes:\n%s\n\n"+
			"To prevent accidental data loss, please secure your work:\n"+
			"  [s] Stash changes\n"+
			"  [c] Commit changes (WIP: Pre-Falken)\n"+
			"  [i] Ignore and continue\n"+
			"  [q] Quit",
		shared.ErrorStyle.Render("⚠️"), m.dirtyFiles)

	if m.isRunning {
		content += "\n\n" + shared.HelpStyle.Render(fmt.Sprintf("git %s in progress. Input is temporarily disabled.", m.runningAction))
	}

	if m.lastError != "" {
		content += "\n\n" + shared.ErrorStyle.Render("Last git action failed: "+m.lastError)
	}

	return lipgloss.Place(
		m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		boxStyle.Render(content),
	)
}

func (m Model) SetError(err error) Model {
	if err == nil {
		m.lastError = ""
		return m
	}
	m.lastError = err.Error()
	return m
}

func (m Model) LastError() string {
	return m.lastError
}
