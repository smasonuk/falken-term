package tooldetails

import (
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/smasonuk/falken-term/internal/ui/app"
	"github.com/smasonuk/falken-term/internal/ui/tui/shared"
)

type Model struct {
	name          string
	args          string
	result        string
	viewport      viewport.Model
	width, height int
}

func NewModel() Model {
	return Model{viewport: viewport.New(80, 20)}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlI, tea.KeyEsc:
			return m, func() tea.Msg { return app.HideModalMsg{} }
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.viewport.Width = m.width - 8
		m.viewport.Height = m.height - 8
	}

	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

func (m Model) View() string {
	modalStyle := lipgloss.NewStyle().
		Width(m.width - 4).
		Height(m.height - 4).
		Border(lipgloss.DoubleBorder()).
		BorderForeground(lipgloss.Color("62")).
		Padding(1)

	title := lipgloss.NewStyle().
		Background(lipgloss.Color("62")).
		Foreground(lipgloss.Color("230")).
		Render(" INSPECT LAST TOOL (Press Ctrl+I or Esc to close) ")

	return lipgloss.Place(
		m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		lipgloss.JoinVertical(lipgloss.Center, title, modalStyle.Render(m.viewport.View())),
	)
}

func (m Model) SetContent(name, args, result string) Model {
	m.name = name
	m.args = args
	m.result = result
	content := lipgloss.JoinVertical(lipgloss.Left,
		shared.ToolNameStyle.Render("Tool: "+name),
		"\n--- ARGUMENTS ---",
		args,
		"\n--- RESULT ---",
		result,
	)
	m.viewport.SetContent(strings.TrimSpace(content))
	m.viewport.GotoTop()
	return m
}

func (m Model) HasContent() bool {
	return strings.TrimSpace(m.name) != ""
}
