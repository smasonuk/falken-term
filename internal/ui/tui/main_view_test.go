package tui

import (
	"strings"
	"testing"

	"github.com/smasonuk/falken-core/pkg/falken"
	"github.com/smasonuk/falken-term/internal/ui/app"
	"github.com/smasonuk/falken-term/internal/ui/tui/modals/permissionprompt"
	"github.com/smasonuk/falken-term/internal/ui/tui/modals/planapproval"
	"github.com/smasonuk/falken-term/internal/ui/tui/modals/tooldetails"
	"github.com/smasonuk/falken-term/internal/ui/tui/screens/chat"
	"github.com/smasonuk/falken-term/internal/ui/tui/screens/diffreview"
	"github.com/smasonuk/falken-term/internal/ui/tui/screens/gitpreflight"
	"github.com/smasonuk/falken-term/internal/ui/tui/screens/initwizard"
	"github.com/smasonuk/falken-term/internal/ui/tui/screens/permissionoverview"
	"github.com/smasonuk/falken-term/internal/ui/tui/screens/pluginapproval"
)

func newTestModel() Model {
	cfg := app.RunnerConfig{
		PermConfig: &falken.PermissionsConfig{
			Caches:                 map[string]falken.CacheConfig{"go-mod": {ContainerPath: "/caches/go-mod"}},
			ShowPermissionOverview: false,
			ApprovedPlugins:        map[string]bool{},
		},
	}

	initwizard.RecommendedDomains = RecommendedDomains

	m := Model{
		workflow: app.Model{
			Route: app.RouteChat,
			Modal: app.ModalNone,
			Session: app.AppSession{
				RunnerConfig: cfg,
				Agent:        &app.AgentSessionState{},
			},
		},
		chat:               chat.NewModel(nil, "", &app.AgentSessionState{}),
		gitPreflight:       gitpreflight.NewModel("", gitpreflight.Actions{}),
		initWizard:         initwizard.NewModel(),
		pluginApproval:     pluginapproval.NewModel(),
		permissionOverview: permissionoverview.NewModel(cfg.PermConfig, nil, nil),
		permissionPrompt:   permissionprompt.NewModel(),
		planApproval:       planapproval.NewModel(),
		toolDetails:        tooldetails.NewModel(),
		diffReview:         diffreview.NewModel(diffreview.Actions{}),
	}
	return m.syncScreens()
}

func TestPermissionRequestOpensModalWithoutChangingRoute(t *testing.T) {
	m := newTestModel()
	req := falken.PermissionRequest{
		Kind:       "file",
		Target:     ".env",
		AccessType: "read",
	}

	updatedAny, _ := m.Update(app.PermissionRequestedMsg{Request: req, Response: make(chan falken.PermissionResponse, 1)})
	updated := updatedAny.(Model)

	if updated.workflow.Route != app.RouteChat {
		t.Fatalf("expected route to remain chat, got %v", updated.workflow.Route)
	}
	if updated.workflow.Modal != app.ModalPermissionPrompt {
		t.Fatalf("expected permission modal, got %v", updated.workflow.Modal)
	}
}

func TestPlanApprovalModalKeepsChatState(t *testing.T) {
	m := newTestModel()
	m.chat = m.chat.BeginRun("running")
	m.workflow.Session.Agent.EventChan = make(chan falken.Event)
	req := falken.PlanApprovalRequest{Plan: "test plan"}

	updatedAny, _ := m.Update(app.PlanApprovalRequestedMsg{Request: req, Response: make(chan falken.PlanApprovalResponse, 1)})
	updated := updatedAny.(Model)

	if updated.workflow.Route != app.RouteChat {
		t.Fatalf("expected route to remain chat, got %v", updated.workflow.Route)
	}
	if updated.workflow.Modal != app.ModalPlanApproval {
		t.Fatalf("expected plan modal, got %v", updated.workflow.Modal)
	}
	if updated.chat.CurrentState() != chat.StateRunning {
		t.Fatalf("expected chat state to remain running, got %v", updated.chat.CurrentState())
	}
}

func TestDiffApplySuccessReturnsToChat(t *testing.T) {
	m := newTestModel()
	m.workflow.Route = app.RouteReviewDiff
	m.chat, _ = m.chat.SetDoneState()

	updatedAny, _ := m.Update(app.DiffApplyFinishedMsg{Applied: true})
	updated := updatedAny.(Model)

	if updated.workflow.Route != app.RouteChat {
		t.Fatalf("expected route to return to chat, got %v", updated.workflow.Route)
	}
}

func TestDiffApplyPartialReturnsToChatWithWarning(t *testing.T) {
	m := newTestModel()
	m.workflow.Route = app.RouteReviewDiff
	m.chat, _ = m.chat.SetDoneState()

	updatedAny, _ := m.Update(app.DiffApplyFinishedMsg{
		Applied:      true,
		Partial:      true,
		SkippedFiles: []string{"secret.txt"},
	})
	updated := updatedAny.(Model)

	if updated.workflow.Route != app.RouteChat {
		t.Fatalf("expected route to return to chat, got %v", updated.workflow.Route)
	}
	if updated.workflow.Session.Agent == nil || !strings.Contains(updated.workflow.Session.Agent.Logs, "secret.txt") {
		t.Fatalf("expected warning log to mention skipped file, got %q", updated.workflow.Session.Agent.Logs)
	}
}

func TestSyncScreensUsesPendingPlugin(t *testing.T) {
	m := newTestModel()
	m.workflow.Session.PendingPlugins = []falken.PluginInfo{{Name: "example"}}

	m = m.syncScreens()
	if plugin := m.pluginApproval.Plugin(); plugin == nil || plugin.Name != "example" {
		t.Fatalf("expected pending plugin to be loaded into screen")
	}
}

func TestGitPreflightFinishedSuccessNavigatesToNextStartupRoute(t *testing.T) {
	m := newTestModel()
	m.workflow.Route = app.RouteStartupGitPreflight
	m.workflow.Session.RunnerConfig.PermConfig.Caches = map[string]falken.CacheConfig{"go-mod": {ContainerPath: "/caches/go-mod"}}
	m.workflow.Session.RunnerConfig.PermConfig.ShowPermissionOverview = false

	updatedAny, _ := m.Update(app.GitPreflightFinishedMsg{Action: "stash"})
	updated := updatedAny.(Model)

	if updated.workflow.Route != app.RouteBooting {
		t.Fatalf("expected successful preflight to continue to booting, got %v", updated.workflow.Route)
	}
	if updated.gitPreflight.LastError() != "" {
		t.Fatalf("expected git preflight error to be cleared, got %q", updated.gitPreflight.LastError())
	}
}

func TestGitPreflightIgnoreNavigatesThroughParentWorkflow(t *testing.T) {
	m := newTestModel()
	m.workflow.Route = app.RouteStartupGitPreflight
	m.workflow.Session.RunnerConfig.PermConfig.Caches = map[string]falken.CacheConfig{"go-mod": {ContainerPath: "/caches/go-mod"}}
	m.workflow.Session.RunnerConfig.PermConfig.ShowPermissionOverview = false

	updatedAny, _ := m.Update(app.GitPreflightActionMsg{Action: "ignore"})
	updated := updatedAny.(Model)

	if updated.workflow.Route != app.RouteBooting {
		t.Fatalf("expected ignore action to continue to booting, got %v", updated.workflow.Route)
	}
}
