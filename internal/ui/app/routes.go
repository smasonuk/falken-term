package app

type Route int

const (
	RouteStartupGitPreflight Route = iota
	RouteStartupInitWizard
	RouteStartupPluginApproval
	RouteStartupPermissionOverview
	RouteBooting
	RouteChat
	RouteReviewDiff
)

type Modal int

const (
	ModalNone Modal = iota
	ModalPermissionPrompt
	ModalPlanApproval
	ModalToolDetails
)
