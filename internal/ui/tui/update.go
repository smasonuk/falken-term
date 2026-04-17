package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/smasonuk/falken-core/pkg/falken"
	"github.com/smasonuk/falken-term/internal/ui/app"
	"github.com/smasonuk/falken-term/internal/ui/app/services"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.chat, _ = m.chat.Update(msg)
		m.initWizard, _ = m.initWizard.Update(msg)
		m.planApproval, _ = m.planApproval.Update(msg)
		m.diffReview, _ = m.diffReview.Update(msg)
		m.gitPreflight, _ = m.gitPreflight.Update(msg)
		m.pluginApproval, _ = m.pluginApproval.Update(msg)
		m.permissionOverview, _ = m.permissionOverview.Update(msg)
		m.permissionPrompt, _ = m.permissionPrompt.Update(msg)
		m.toolDetails, _ = m.toolDetails.Update(msg)
		return m, nil

	case app.NavigateToRouteMsg:
		return m.handleNavigate(msg.Route)

	case app.ShowModalMsg:
		if msg.Modal == app.ModalToolDetails && !m.toolDetails.HasContent() {
			return m, nil
		}
		m.workflow = m.workflow.ShowModal(msg.Modal)
		return m, nil

	case app.HideModalMsg:
		m.workflow = m.workflow.HideModal()
		return m, nil

	case app.StartChatRunMsg:
		handle, startCmd := services.StartAgentRun(m.workflow.Session.RunnerConfig.Session, m.bridge.EventChannel(), msg.Prompt)
		m.workflow.Session.Agent.IsActive = true
		m.workflow.Session.Agent.EventChan = handle.EventChan
		m.workflow.Session.Agent.CancelFunc = handle.Cancel
		m.chat = m.chat.BeginRun(msg.DisplayPrompt)
		return m, tea.Batch(startCmd, m.chat.SpinnerTickCmd())

	case app.PermissionRequestedMsg:
		m.workflow = m.workflow.WithPermissionRequest(msg.Request)
		m.workflow.Session.CurrentPermResp = msg.Response
		return m.syncScreens(), nil

	case app.PermissionResponseSelectedMsg:
		if m.workflow.Session.CurrentPermResp != nil {
			m.workflow.Session.CurrentPermResp <- msg.Response
		}
		m.workflow = m.workflow.ClearPermissionRequest()
		return m.syncScreens(), m.waitPermissionRequest()

	case app.PlanApprovalRequestedMsg:
		m.workflow = m.workflow.WithPlanApprovalRequest(msg.Request)
		m.workflow.Session.CurrentPlanResp = msg.Response
		return m.syncScreens(), nil

	case app.PlanApprovalResponseMsg:
		if m.workflow.Session.CurrentPlanResp != nil {
			m.workflow.Session.CurrentPlanResp <- falken.PlanApprovalResponse{
				Approved: msg.Approved,
				Feedback: msg.Feedback,
			}
		}
		eventChan := m.workflow.Session.Agent.EventChan
		m.workflow = m.workflow.ClearPlanApprovalRequest()
		if eventChan != nil {
			return m.syncScreens(), tea.Batch(services.WaitForAgentEvent(eventChan), m.waitPlanRequest())
		}
		return m.syncScreens(), m.waitPlanRequest()

	case app.GitPreflightActionMsg:
		switch msg.Action {
		case services.GitPreflightActionIgnore:
			return m.handleNavigate(app.NextStartupRoute(m.workflow.Session.RunnerConfig.PermConfig, len(m.workflow.Session.PendingPlugins) > 0))
		}
		return m, nil

	case app.GitPreflightFinishedMsg:
		m.gitPreflight, _ = m.gitPreflight.Update(msg)
		if msg.Err != nil {
			return m, nil
		}
		return m.handleNavigate(app.NextStartupRoute(m.workflow.Session.RunnerConfig.PermConfig, len(m.workflow.Session.PendingPlugins) > 0))

	case app.InitWizardSubmittedMsg:
		m.workflow = m.workflow.ApplyInitWizardSubmission(msg, RecommendedCaches)
		return m, tea.Batch(
			services.SavePermissionsConfigCmd(m.workflow.Session.RunnerConfig.PermConfig),
			func() tea.Msg {
				return app.NavigateToRouteMsg{
					Route: app.NextStartupRoute(m.workflow.Session.RunnerConfig.PermConfig, len(m.workflow.Session.PendingPlugins) > 0),
				}
			},
		)

	case app.PluginApprovalDecisionMsg:
		m.workflow = m.workflow.ApplyPluginApprovalDecision(msg.Approved)
		nextRoute := app.NextStartupRoute(m.workflow.Session.RunnerConfig.PermConfig, len(m.workflow.Session.PendingPlugins) > 0)
		if len(m.workflow.Session.PendingPlugins) > 0 {
			nextRoute = app.RouteStartupPluginApproval
		}
		return m.syncScreens(), tea.Batch(
			services.SavePermissionsConfigCmd(m.workflow.Session.RunnerConfig.PermConfig),
			func() tea.Msg { return app.NavigateToRouteMsg{Route: nextRoute} },
		)

	case app.PermissionOverviewDecisionMsg:
		if msg.DontShowAgain {
			m.workflow.Session.RunnerConfig.PermConfig.ShowPermissionOverview = false
			return m, tea.Batch(
				services.SavePermissionsConfigCmd(m.workflow.Session.RunnerConfig.PermConfig),
				func() tea.Msg { return app.NavigateToRouteMsg{Route: app.RouteBooting} },
			)
		}
		return m.handleNavigate(app.RouteBooting)

	case app.SandboxBootFinishedMsg:
		if msg.Err != nil {
			m.workflow.Session.BootError = msg.Err
			return m, nil
		}

		m.workflow.Session.BootError = nil
		m.workflow = m.workflow.Navigate(app.RouteChat)
		m.chat, cmd = m.chat.SetPromptState()
		return m, tea.Batch(m.chat.Init(), cmd)

	case app.DiffGeneratedMsg:
		m.diffReview, cmd = m.diffReview.Update(msg)
		if msg.Err == nil {
			m.workflow.Session.LastGeneratedDiff = msg.Diff
			m.diffReview.SetDiff(msg.Diff)
		}
		return m, cmd

	case app.DiscardDiffRequestedMsg:
		return m.handleNavigate(app.RouteChat)

	case app.DiffApplyFinishedMsg:
		m.diffReview, cmd = m.diffReview.Update(msg)
		if !msg.Applied {
			m.chat.AppendSystemError(fmt.Sprintf("Could not apply reviewed changes to the host workspace: %v", msg.Err))
			return m, cmd
		}
		if msg.Partial {
			warning := "Applied reviewed changes to the host workspace, but skipped blocked files."
			if len(msg.SkippedFiles) > 0 {
				warning = fmt.Sprintf("%s Skipped: %s", warning, strings.Join(msg.SkippedFiles, ", "))
			}
			m.chat.AppendSystemWarning(warning)
			return m.handleNavigate(app.RouteChat)
		}
		m.chat.AppendSystemSuccess("Successfully applied changes to the host workspace.")
		return m.handleNavigate(app.RouteChat)

	case app.PermissionsConfigSavedMsg:
		if msg.Err != nil {
			m.chat.AppendSystemError(fmt.Sprintf("Failed to save permissions config: %v", msg.Err))
		}
		return m, nil
	}

	if sessionMsg, ok := msg.(app.SessionEventMsg); ok {
		m.workflow = m.workflow.ApplySessionEvent(sessionMsg.Event)
		m = m.syncScreens()
	}

	if m.workflow.Modal != app.ModalNone {
		if _, ok := msg.(tea.KeyMsg); ok {
			return m.updateActiveModal(msg)
		}
	}

	return m.updateActiveRoute(msg)
}

func (m Model) updateActiveRoute(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch m.workflow.Route {
	case app.RouteStartupGitPreflight:
		m.gitPreflight, cmd = m.gitPreflight.Update(msg)
	case app.RouteStartupInitWizard:
		m.initWizard, cmd = m.initWizard.Update(msg)
	case app.RouteStartupPluginApproval:
		m.pluginApproval, cmd = m.pluginApproval.Update(msg)
	case app.RouteStartupPermissionOverview:
		m.permissionOverview, cmd = m.permissionOverview.Update(msg)
	case app.RouteReviewDiff:
		m.diffReview, cmd = m.diffReview.Update(msg)
	case app.RouteChat:
		m.chat, cmd = m.chat.Update(msg)
	case app.RouteBooting:
		if keyMsg, ok := msg.(tea.KeyMsg); ok && m.workflow.Session.BootError != nil {
			if keyMsg.String() == "q" || keyMsg.String() == "ctrl+c" {
				return m, tea.Quit
			}
		}
	}
	return m, cmd
}

func (m Model) updateActiveModal(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch m.workflow.Modal {
	case app.ModalPermissionPrompt:
		m.permissionPrompt, cmd = m.permissionPrompt.Update(msg)
	case app.ModalPlanApproval:
		m.planApproval, cmd = m.planApproval.Update(msg)
	case app.ModalToolDetails:
		m.toolDetails, cmd = m.toolDetails.Update(msg)
	}
	return m, cmd
}

func (m Model) handleNavigate(route app.Route) (tea.Model, tea.Cmd) {
	m.workflow = m.workflow.Navigate(route)
	m = m.syncScreens()

	switch route {
	case app.RouteBooting:
		return m, services.BootSandboxCmd(m.workflow.Session.RunnerConfig.Session)
	case app.RouteReviewDiff:
		var cmd tea.Cmd
		m.diffReview, cmd = m.diffReview.BeginDiffGeneration()
		return m, tea.Batch(cmd, services.GenerateDiffCmd(m.workflow.Session.RunnerConfig.Session))
	case app.RouteChat:
		return m, nil
	default:
		return m, nil
	}
}

func (m Model) syncScreens() Model {
	m.pluginApproval = m.pluginApproval.SetPlugin(m.currentPlugin(), len(m.workflow.Session.PendingPlugins))
	m.permissionPrompt = m.permissionPrompt.SetRequest(m.workflow.Session.CurrentPermReq)
	m.planApproval = m.planApproval.SetRequest(m.workflow.Session.CurrentPlanReq)
	m.toolDetails = m.toolDetails.SetContent(
		m.workflow.Session.Agent.LastToolName,
		m.workflow.Session.Agent.LastToolArgs,
		m.workflow.Session.Agent.LastToolResult,
	)
	m.chat = m.chat.SetRuntime(m.workflow.Session.Agent)
	return m
}

func (m Model) currentPlugin() *falken.PluginInfo {
	if len(m.workflow.Session.PendingPlugins) == 0 {
		return nil
	}
	plugin := m.workflow.Session.PendingPlugins[0]
	return &plugin
}

func (m Model) waitPermissionRequest() tea.Cmd {
	if m.bridge == nil {
		return nil
	}
	return func() tea.Msg {
		req := <-m.bridge.NextPermissionRequest()
		return app.PermissionRequestedMsg{Request: req.Request, Response: req.Response}
	}
}

func (m Model) waitPlanRequest() tea.Cmd {
	if m.bridge == nil {
		return nil
	}
	return func() tea.Msg {
		req := <-m.bridge.NextPlanRequest()
		return app.PlanApprovalRequestedMsg{Request: req.Request, Response: req.Response}
	}
}
