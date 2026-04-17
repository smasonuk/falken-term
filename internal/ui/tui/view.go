package tui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/smasonuk/falken-term/internal/ui/app"
)

func (m Model) View() string {
	if m.width == 0 || m.height == 0 {
		return "Initializing..."
	}

	base := m.routeView()
	if m.workflow.Modal == app.ModalNone {
		return base
	}

	return m.modalView()
}

func (m Model) routeView() string {
	switch m.workflow.Route {
	case app.RouteStartupGitPreflight:
		return m.gitPreflight.View()
	case app.RouteStartupInitWizard:
		return m.initWizard.View()
	case app.RouteStartupPluginApproval:
		return m.pluginApproval.View()
	case app.RouteStartupPermissionOverview:
		return m.permissionOverview.View()
	case app.RouteBooting:
		return m.bootingView()
	case app.RouteReviewDiff:
		return m.diffReview.View()
	case app.RouteChat:
		return m.chat.View()
	default:
		return "Initializing..."
	}
}

func (m Model) modalView() string {
	switch m.workflow.Modal {
	case app.ModalPermissionPrompt:
		return m.permissionPrompt.View()
	case app.ModalPlanApproval:
		return m.planApproval.View()
	case app.ModalToolDetails:
		return m.toolDetails.View()
	default:
		return ""
	}
}

func (m Model) bootingView() string {
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		Padding(1, 3).
		Align(lipgloss.Left)

	content := "Booting Sandbox Environment...\n\n"
	if m.workflow.Session.BootError != nil {
		content = fmt.Sprintf("FAILED TO BOOT SANDBOX:\n\n%v\n\n[q] Quit", m.workflow.Session.BootError)
	} else {
		content += "Starting snapshot and Docker container. Please wait."
	}

	return lipgloss.Place(
		m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		boxStyle.Render(content),
	)
}
