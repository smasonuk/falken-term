package app

import (
	"encoding/json"

	"github.com/smasonuk/falken-core/pkg/falken"
)

func NewWorkflow(config RunnerConfig, dirty bool) Model {
	session := AppSession{
		RunnerConfig:   config,
		PendingPlugins: PendingPlugins(config.Plugins, config.PermConfig),
		Agent:          &AgentSessionState{},
	}

	return Model{
		Route:   InitialRoute(config.PermConfig, dirty, len(session.PendingPlugins) > 0),
		Modal:   ModalNone,
		Session: session,
	}
}

func PendingPlugins(plugins []falken.PluginInfo, cfg *falken.PermissionsConfig) []falken.PluginInfo {
	pending := make([]falken.PluginInfo, 0, len(plugins))
	for _, hook := range plugins {
		if hook.Internal {
			continue
		}
		if cfg == nil || !cfg.ApprovedPlugins[hook.Name] {
			pending = append(pending, hook)
		}
	}
	return pending
}

func InitialRoute(cfg *falken.PermissionsConfig, dirty bool, hasPendingPlugins bool) Route {
	if dirty {
		return RouteStartupGitPreflight
	}
	return NextStartupRoute(cfg, hasPendingPlugins)
}

func NextStartupRoute(cfg *falken.PermissionsConfig, hasPendingPlugins bool) Route {
	if cfg != nil && len(cfg.Caches) == 0 {
		return RouteStartupInitWizard
	}
	if hasPendingPlugins {
		return RouteStartupPluginApproval
	}
	if cfg != nil && cfg.ShowPermissionOverview {
		return RouteStartupPermissionOverview
	}
	return RouteBooting
}

func (m Model) Navigate(route Route) Model {
	m.Route = route
	return m
}

func (m Model) ShowModal(modal Modal) Model {
	m.Modal = modal
	return m
}

func (m Model) HideModal() Model {
	m.Modal = ModalNone
	return m
}

func (m Model) WithPermissionRequest(req falken.PermissionRequest) Model {
	m.Session.CurrentPermReq = &req
	m.Modal = ModalPermissionPrompt
	return m
}

func (m Model) ClearPermissionRequest() Model {
	m.Session.CurrentPermReq = nil
	m.Session.CurrentPermResp = nil
	if m.Modal == ModalPermissionPrompt {
		m.Modal = ModalNone
	}
	return m
}

func (m Model) WithPlanApprovalRequest(req falken.PlanApprovalRequest) Model {
	m.Session.CurrentPlanReq = &req
	m.Modal = ModalPlanApproval
	return m
}

func (m Model) ClearPlanApprovalRequest() Model {
	m.Session.CurrentPlanReq = nil
	m.Session.CurrentPlanResp = nil
	if m.Modal == ModalPlanApproval {
		m.Modal = ModalNone
	}
	return m
}

func (m Model) ApplyInitWizardSubmission(msg InitWizardSubmittedMsg, recommendedCaches map[string]falken.CacheConfig) Model {
	cfg := m.Session.RunnerConfig.PermConfig
	if cfg.Caches == nil {
		cfg.Caches = make(map[string]falken.CacheConfig)
	}

	for _, cache := range msg.SelectedCaches {
		if cacheCfg, ok := recommendedCaches[cache]; ok {
			cfg.Caches[cache] = cacheCfg
		}
	}
	for _, domain := range msg.AllowedDomains {
		cfg.AddGlobalAllowedURL(domain)
	}

	return m
}

func (m Model) ApplyPluginApprovalDecision(approved bool) Model {
	if len(m.Session.PendingPlugins) == 0 {
		return m
	}
	cfg := m.Session.RunnerConfig.PermConfig
	plugin := m.Session.PendingPlugins[0]
	cfg.ApprovedPlugins[plugin.Name] = approved
	m.Session.PendingPlugins = m.Session.PendingPlugins[1:]
	return m
}

func (m Model) ApplySessionEvent(event falken.Event) Model {
	if m.Session.Agent == nil {
		return m
	}

	if event.Type == falken.EventTypeToolCall && event.ToolCall != nil {
		argBytes, _ := json.MarshalIndent(event.ToolCall.Args, "", "  ")
		m.Session.Agent.LastToolName = event.ToolCall.Name
		m.Session.Agent.LastToolArgs = string(argBytes)
		m.Session.Agent.LastToolResult = "Running..."
	}
	if event.Type == falken.EventTypeToolResult && event.ToolResult != nil {
		resBytes, _ := json.MarshalIndent(event.ToolResult.Result, "", "  ")
		m.Session.Agent.LastToolResult = string(resBytes)
	}
	if event.Type == falken.EventTypeRunCompleted || event.Type == falken.EventTypeRunFailed {
		m.Session.Agent.IsActive = false
		m.Session.Agent.CancelFunc = nil
	}
	return m
}
