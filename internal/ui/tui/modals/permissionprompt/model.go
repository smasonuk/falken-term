package permissionprompt

import (
	"fmt"
	"strings"

	"github.com/smasonuk/falken-core/pkg/falken"
	"github.com/smasonuk/falken-term/internal/ui/app"
	"github.com/smasonuk/falken-term/internal/ui/tui/shared"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Model struct {
	request       *falken.PermissionRequest
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
		if m.request == nil {
			return m, nil
		}

		resp, handled := buildPermissionResponse(*m.request, msg.String())
		if handled {
			return m, func() tea.Msg { return app.PermissionResponseSelectedMsg{Response: resp} }
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}

	return m, nil
}

func (m Model) View() string {
	if m.request == nil {
		return ""
	}

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("196")).
		Padding(1, 3).
		Align(lipgloss.Left)

	return lipgloss.Place(
		m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		boxStyle.Render(permissionPromptContent(*m.request)),
	)
}

func (m Model) SetRequest(req *falken.PermissionRequest) Model {
	m.request = req
	return m
}

func buildPermissionResponse(req falken.PermissionRequest, key string) (falken.PermissionResponse, bool) {
	resp := falken.PermissionResponse{}
	handled := true

	switch key {
	case "1":
		resp = falken.PermissionResponse{Allowed: true, Scope: "once", AccessType: "url"}
		if req.Kind != "network" {
			resp.AccessType = req.AccessType
		}
	case "2":
		resp = falken.PermissionResponse{Allowed: true, Scope: "session", AccessType: "url"}
		if req.Kind != "network" {
			resp.AccessType = req.AccessType
		}
	case "3":
		if req.Kind == "shell" {
			resp = falken.PermissionResponse{Allowed: true, Scope: "project", AccessType: "execute"}
		} else if req.Kind == "network" {
			resp = falken.PermissionResponse{Allowed: true, Scope: "project", AccessType: "url"}
		} else {
			resp = falken.PermissionResponse{Allowed: true, Scope: "project", AccessType: "read"}
		}
	case "4":
		if req.Kind == "shell" {
			handled = false
		} else if req.Kind == "network" {
			resp = falken.PermissionResponse{Allowed: true, Scope: "session", AccessType: "domain"}
		} else {
			resp = falken.PermissionResponse{Allowed: true, Scope: "project", AccessType: "read/write"}
		}
	case "5":
		if req.Kind == "network" {
			resp = falken.PermissionResponse{Allowed: true, Scope: "project", AccessType: "domain"}
		} else {
			handled = false
		}
	case "esc", "q", "n", "N":
		resp = falken.PermissionResponse{Allowed: false, Scope: "deny"}
	default:
		handled = false
	}

	return resp, handled
}

func permissionPromptContent(req falken.PermissionRequest) string {
	if req.Kind == "network" {
		domain := req.AccessType
		return fmt.Sprintf(
			"%s Agent requested to access an external network resource:\n\n"+
				"URL: %s\n\n"+
				"Choose an action:\n"+
				"  [1] Allow this exact URL once\n"+
				"  [2] Allow this exact URL for session\n"+
				"  [3] Allow this exact URL permanently (.falken.yaml)\n"+
				"  [4] Allow entire domain (%s) for session\n"+
				"  [5] Allow entire domain (%s) permanently\n"+
				"  [Esc/n] Deny Request",
			shared.ErrorStyle.Render("WARNING"),
			shared.ToolNameStyle.Render(req.Target),
			domain,
			domain,
		)
	}

	if req.Kind == "shell" {
		return fmt.Sprintf(
			"%s Agent requested to %s an unlisted shell command:\n\n"+
				"Command: %s\n\n"+
				"Choose an action:\n"+
				"  [1] Allow this exact command once\n"+
				"  [2] Allow the base command for this session\n"+
				"  [3] Allow the base command permanently (.falken.yaml)\n"+
				"  [Esc/n] Deny Request",
			shared.ErrorStyle.Render("WARNING"),
			strings.ToUpper(req.AccessType),
			shared.ToolNameStyle.Render(req.Target),
		)
	}

	return fmt.Sprintf(
		"%s Agent requested to %s a hidden dotfile:\n\n"+
			"File: %s\n\n"+
			"Choose an action:\n"+
			"  [1] Allow this specific time\n"+
			"  [2] Allow for this session\n"+
			"  [3] Allow in project directory (Read-only, persist config)\n"+
			"  [4] Allow in project directory (Read/Write, persist config)\n"+
			"  [Esc/n] Deny Request",
		shared.ErrorStyle.Render("WARNING"),
		strings.ToUpper(req.AccessType),
		shared.ToolNameStyle.Render(req.Target),
	)
}
