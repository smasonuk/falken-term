package app

import (
	"testing"

	"github.com/smasonuk/falken-core/pkg/falken"
)

func TestInitialRoute(t *testing.T) {
	cfg := &falken.PermissionsConfig{
		Caches:                 map[string]falken.CacheConfig{},
		ShowPermissionOverview: true,
		ApprovedPlugins:        map[string]bool{},
	}

	if route := InitialRoute(cfg, true, false); route != RouteStartupGitPreflight {
		t.Fatalf("expected git preflight route, got %v", route)
	}

	if route := InitialRoute(cfg, false, false); route != RouteStartupInitWizard {
		t.Fatalf("expected init wizard route, got %v", route)
	}

	cfg.Caches["go-mod"] = falken.CacheConfig{ContainerPath: "/caches/go-mod"}
	if route := InitialRoute(cfg, false, true); route != RouteStartupPluginApproval {
		t.Fatalf("expected plugin approval route, got %v", route)
	}

	if route := InitialRoute(cfg, false, false); route != RouteStartupPermissionOverview {
		t.Fatalf("expected permission overview route, got %v", route)
	}

	cfg.ShowPermissionOverview = false
	if route := InitialRoute(cfg, false, false); route != RouteBooting {
		t.Fatalf("expected booting route, got %v", route)
	}
}

func TestPendingPluginsFiltersInternalAndApproved(t *testing.T) {
	cfg := &falken.PermissionsConfig{
		ApprovedPlugins: map[string]bool{"approved": true},
	}

	plugins := []falken.PluginInfo{
		{Name: "internal", Internal: true},
		{Name: "approved"},
		{Name: "pending"},
	}

	pending := PendingPlugins(plugins, cfg)
	if len(pending) != 1 || pending[0].Name != "pending" {
		t.Fatalf("unexpected pending plugins: %+v", pending)
	}
}

func TestModalHelpersPreserveRoute(t *testing.T) {
	model := Model{Route: RouteChat, Modal: ModalNone}
	req := falken.PermissionRequest{Kind: "file", Target: ".env"}

	model = model.WithPermissionRequest(req)
	if model.Route != RouteChat {
		t.Fatalf("expected route to stay on chat, got %v", model.Route)
	}
	if model.Modal != ModalPermissionPrompt {
		t.Fatalf("expected permission modal, got %v", model.Modal)
	}

	model = model.ClearPermissionRequest()
	if model.Modal != ModalNone {
		t.Fatalf("expected modal to clear, got %v", model.Modal)
	}
}
