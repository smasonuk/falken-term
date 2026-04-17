package permissionoverview

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
	cfg           *falken.PermissionsConfig
	tools         []falken.ToolInfo
	plugins       []falken.PluginInfo
	width, height int
}

func NewModel(cfg *falken.PermissionsConfig, tools []falken.ToolInfo, plugins []falken.PluginInfo) Model {
	return Model{cfg: cfg, tools: tools, plugins: plugins}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			return m, func() tea.Msg { return app.PermissionOverviewDecisionMsg{} }
		case "d", "D":
			return m, func() tea.Msg { return app.PermissionOverviewDecisionMsg{DontShowAgain: true} }
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}
	return m, nil
}

func (m Model) View() string {
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		Padding(1, 3).
		Align(lipgloss.Left)

	var sb strings.Builder
	sb.WriteString(shared.HeaderStyle.Render("SECURITY POSTURE OVERVIEW") + "\n\n")

	if m.cfg != nil {
		sb.WriteString(lipgloss.NewStyle().Bold(true).Render("Configured Allow/Block Rules:") + "\n")
		sb.WriteString(fmt.Sprintf("- Blocked URLs: %v\n", m.cfg.GlobalBlockedURLs))
		sb.WriteString(fmt.Sprintf("- Allowed URLs: %v\n", m.cfg.GlobalAllowedURLs))
		sb.WriteString(fmt.Sprintf("- Blocked Files: %v\n", m.cfg.GlobalBlockedFiles))
		sb.WriteString(fmt.Sprintf("- Allowed Files: %v\n", m.cfg.GlobalAllowedFiles))
		sb.WriteString(fmt.Sprintf("- Strict file allowlist: %v\n", m.cfg.StrictFileAllowlist))
		sb.WriteString(fmt.Sprintf("- Blocked Commands: %v\n", m.cfg.GlobalBlockedCommands))
		sb.WriteString(fmt.Sprintf("- Allowed Commands: %v\n", m.cfg.GlobalAllowedCommands))
		sb.WriteString(fmt.Sprintf("- Strict command allowlist: %v\n\n", m.cfg.StrictCommandAllowlist))

		sb.WriteString(lipgloss.NewStyle().Bold(true).Render("Persistent Approvals:") + "\n")
		sb.WriteString(fmt.Sprintf("- URLs: %v\n", m.cfg.PersistentAllowedURLs))
		sb.WriteString(fmt.Sprintf("- Files: %v\n", m.cfg.PersistentAllowedFiles))
		sb.WriteString(fmt.Sprintf("- Commands: %v\n\n", m.cfg.PersistentAllowedCommands))
	}

	sb.WriteString(lipgloss.NewStyle().Bold(true).Render("Loaded Tools (JIT Evaluated):") + "\n")
	for _, t := range m.tools {
		sb.WriteString(fmt.Sprintf("- %s\n", t.Name))
	}
	sb.WriteString("\n")

	sb.WriteString(lipgloss.NewStyle().Bold(true).Render("Loaded Plugins (AOT Approved):") + "\n")
	for _, p := range m.plugins {
		if p.Internal || (m.cfg != nil && m.cfg.ApprovedPlugins[p.Name]) {
			sb.WriteString(fmt.Sprintf("- %s\n", p.Name))
		}
	}

	sb.WriteString("\n" + shared.HelpStyle.Render("[Enter] Start Agent   [d] Start & Don't show this again"))

	return lipgloss.Place(
		m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		boxStyle.Render(sb.String()),
	)
}
