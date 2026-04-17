package app

import "github.com/smasonuk/falken-core/pkg/falken"

type NavigateToRouteMsg struct {
	Route Route
}

type ShowModalMsg struct {
	Modal Modal
}

type HideModalMsg struct{}

type SandboxBootStartedMsg struct{}

type SandboxBootFinishedMsg struct {
	Err error
}

type PlanApprovalRequestedMsg struct {
	Request  falken.PlanApprovalRequest
	Response chan falken.PlanApprovalResponse
}

type PermissionRequestedMsg struct {
	Request  falken.PermissionRequest
	Response chan falken.PermissionResponse
}

type DiffGeneratedMsg struct {
	Diff string
	Err  error
}

type DiscardDiffRequestedMsg struct{}

type DiffApplyFinishedMsg struct {
	Applied      bool
	Partial      bool
	SkippedFiles []string
	Err          error
}

type GitPreflightActionMsg struct {
	Action string
}

type GitPreflightFinishedMsg struct {
	Action string
	Err    error
}

type InitWizardSubmittedMsg struct {
	SelectedCaches []string
	AllowedDomains []string
}

type PermissionsConfigSavedMsg struct {
	Err error
}

type StartChatRunMsg struct {
	Prompt        string
	DisplayPrompt string
}

type PluginApprovalDecisionMsg struct {
	Approved bool
}

type PermissionOverviewDecisionMsg struct {
	DontShowAgain bool
}

type PermissionResponseSelectedMsg struct {
	Response falken.PermissionResponse
}

type SessionEventMsg struct {
	Event falken.Event
}

type PlanApprovalResponseMsg struct {
	Approved bool
	Feedback string
}
