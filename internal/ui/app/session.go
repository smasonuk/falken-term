package app

import (
	"context"

	"github.com/smasonuk/falken-core/pkg/falken"
)

type RunnerConfig struct {
	Session    *falken.Session
	PermConfig *falken.PermissionsConfig
	Tools      []falken.ToolInfo
	Plugins    []falken.PluginInfo
}

type AppSession struct {
	RunnerConfig      RunnerConfig
	BootError         error
	CurrentPermReq    *falken.PermissionRequest
	CurrentPermResp   chan falken.PermissionResponse
	CurrentPlanReq    *falken.PlanApprovalRequest
	CurrentPlanResp   chan falken.PlanApprovalResponse
	PendingPlugins    []falken.PluginInfo
	LastGeneratedDiff string
	Agent             *AgentSessionState
}

type AgentSessionState struct {
	IsActive          bool
	EventChan         chan falken.Event
	LastToolName      string
	LastToolArgs      string
	LastToolResult    string
	Logs              string
	TodoState         string
	LiveCommandOutput string
	TransientStatus   string
	HiddenToolUses    int
	CancelFunc        context.CancelFunc
}
