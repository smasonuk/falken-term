package planapproval

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/smasonuk/falken-core/pkg/falken"
	"github.com/smasonuk/falken-term/internal/ui/app"
	"github.com/smasonuk/falken-term/internal/ui/tui/shared"
)

type Model struct {
	currentPlanReq falken.PlanApprovalRequest
	viewport       viewport.Model
	textinput      textinput.Model
	width, height  int
}

func NewModel() Model {
	ti := textinput.New()
	ti.Placeholder = "Enter feedback..."
	ti.Width = 80

	vp := viewport.New(80, 20)

	return Model{
		textinput: ti,
		viewport:  vp,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var (
		tiCmd tea.Cmd
		vpCmd tea.Cmd
	)

	switch msg := msg.(type) {
	case falken.PlanApprovalRequest:
		m.currentPlanReq = msg
		m.textinput.Focus()
		m.textinput.SetValue("")
		m.viewport.SetContent(msg.Plan)
		m.viewport.GotoTop()
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "y", "Y":
			return m, func() tea.Msg { return app.PlanApprovalResponseMsg{Approved: true} }
		case "enter":
			feedback := m.textinput.Value()
			if feedback == "" {
				return m, nil
			}
			m.textinput.SetValue("")
			return m, func() tea.Msg {
				return app.PlanApprovalResponseMsg{Approved: false, Feedback: feedback}
			}
		case "esc":
			return m, func() tea.Msg { return app.PlanApprovalResponseMsg{Approved: false} }
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.viewport.Width = m.width - 10
		m.viewport.Height = m.height - 15
		m.textinput.Width = m.width - 2
	}

	m.textinput, tiCmd = m.textinput.Update(msg)
	m.viewport, vpCmd = m.viewport.Update(msg)
	return m, tea.Batch(tiCmd, vpCmd)
}

func (m Model) View() string {
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		Padding(1, 3).
		Align(lipgloss.Left)

	content := fmt.Sprintf(
		"%s PROPOSED IMPLEMENTATION PLAN:\n\n%s\n\n"+
			"Do you approve this plan?\n"+
			"  [y] Approve and start coding\n"+
			"  Or type feedback below + [Enter] to reject\n\n"+
			"> %s",
		shared.HeaderStyle.Render("📋"),
		m.viewport.View(),
		m.textinput.View(),
	)

	return lipgloss.Place(
		m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		boxStyle.Render(content),
	)
}

func (m Model) SetRequest(req *falken.PlanApprovalRequest) Model {
	if req == nil {
		m.currentPlanReq = falken.PlanApprovalRequest{}
		m.textinput.SetValue("")
		m.viewport.SetContent("")
		return m
	}
	m.currentPlanReq = *req
	m.textinput.Focus()
	m.textinput.SetValue("")
	m.viewport.SetContent(req.Plan)
	m.viewport.GotoTop()
	return m
}
