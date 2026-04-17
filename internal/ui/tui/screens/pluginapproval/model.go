package pluginapproval

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/smasonuk/falken-core/pkg/falken"
	"github.com/smasonuk/falken-term/internal/ui/app"
	"github.com/smasonuk/falken-term/internal/ui/tui/shared"
)

type Model struct {
	plugin        *falken.PluginInfo
	remaining     int
	width, height int
}

func NewModel() Model {
	return Model{}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "y", "Y":
			return m, func() tea.Msg { return app.PluginApprovalDecisionMsg{Approved: true} }
		case "n", "N":
			return m, func() tea.Msg { return app.PluginApprovalDecisionMsg{Approved: false} }
		case "q", "esc", "ctrl+c":
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}
	return m, nil
}

func (m Model) View() string {
	if m.plugin == nil {
		return "No plugin awaiting approval."
	}

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("196")).
		Padding(1, 3).
		Align(lipgloss.Left)

	var perms strings.Builder
	if len(m.plugin.NetworkTargets) > 0 {
		perms.WriteString("\n- Network:")
		for _, n := range m.plugin.NetworkTargets {
			perms.WriteString(" " + n)
		}
	}
	if len(m.plugin.ShellCommands) > 0 {
		perms.WriteString("\n- Shell (Sandbox Only): " + strings.Join(m.plugin.ShellCommands, ", "))
	}
	if len(m.plugin.FilePermissions) > 0 {
		perms.WriteString("\n- Files:")
		for _, f := range m.plugin.FilePermissions {
			perms.WriteString(" " + f)
		}
	}

	remainingText := ""
	if m.remaining > 1 {
		remainingText = fmt.Sprintf("\n\n%d plugin approvals remaining.", m.remaining)
	}

	content := fmt.Sprintf(
		"NEW EXTERNAL PLUGIN DETECTED: %s\n"+
			"Description: %s\n\n"+
			"Requested AOT Permissions:%s%s\n\n"+
			"Do you want to allow this plugin to load?\n"+
			"[y] Approve   [n] Deny & Disable",
		shared.ToolNameStyle.Render(m.plugin.Name),
		m.plugin.Description,
		perms.String(),
		remainingText,
	)

	return lipgloss.Place(
		m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		boxStyle.Render(content),
	)
}

func (m Model) SetPlugin(plugin *falken.PluginInfo, remaining int) Model {
	m.plugin = plugin
	m.remaining = remaining
	return m
}

func (m Model) Plugin() *falken.PluginInfo {
	return m.plugin
}
