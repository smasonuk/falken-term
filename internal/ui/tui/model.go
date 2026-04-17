package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/smasonuk/falken-core/pkg/falken"
	"github.com/smasonuk/falken-term/internal/ui/app"
	"github.com/smasonuk/falken-term/internal/ui/app/services"
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

type Model struct {
	workflow app.Model
	bridge   *SessionBridge
	width    int
	height   int

	gitPreflight       gitpreflight.Model
	initWizard         initwizard.Model
	pluginApproval     pluginapproval.Model
	permissionOverview permissionoverview.Model
	permissionPrompt   permissionprompt.Model
	planApproval       planapproval.Model
	toolDetails        tooldetails.Model
	diffReview         diffreview.Model
	chat               chat.Model
}

var RecommendedCaches = map[string]falken.CacheConfig{
	"go-mod":   {ContainerPath: "/caches/go-mod", Env: []string{"GOMODCACHE=/caches/go-mod"}},
	"go-build": {ContainerPath: "/caches/go-build", Env: []string{"GOCACHE=/caches/go-build"}},
	"npm":      {ContainerPath: "/caches/npm", Env: []string{"npm_config_cache=/caches/npm"}},
	"pip":      {ContainerPath: "/caches/pip", Env: []string{"PIP_CACHE_DIR=/caches/pip"}},
}

var RecommendedDomains = map[string][]string{
	"go-mod":   {"*.golang.org", "github.com", "proxy.golang.org", "sum.golang.org"},
	"go-build": {},
	"npm":      {"registry.npmjs.org"},
	"pip":      {"pypi.org", "files.pythonhosted.org"},
}

func NewModel(config app.RunnerConfig, bridge *SessionBridge) Model {
	dirty, dirtyFiles, err := services.CheckDirtyWorktree()
	if err != nil {
		dirty = false
		dirtyFiles = ""
	}

	initwizard.RecommendedDomains = RecommendedDomains

	workflow := app.NewWorkflow(config, dirty)
	model := Model{
		workflow: workflow,
		bridge:   bridge,
		gitPreflight: gitpreflight.NewModel(dirtyFiles, gitpreflight.Actions{
			Stash:     services.GitStashCmd,
			CommitWIP: services.GitCommitWIPCmd,
		}),
		initWizard:         initwizard.NewModel(),
		pluginApproval:     pluginapproval.NewModel(),
		permissionOverview: permissionoverview.NewModel(config.PermConfig, config.Tools, config.Plugins),
		permissionPrompt:   permissionprompt.NewModel(),
		planApproval:       planapproval.NewModel(),
		toolDetails:        tooldetails.NewModel(),
		diffReview: diffreview.NewModel(diffreview.Actions{
			Apply: func(diff string) tea.Cmd {
				return services.ApplyDiffCmd(config.Session, diff)
			},
		}),
		chat: chat.NewModel(config.Session, config.Session.Paths().TodosPath(), workflow.Session.Agent),
	}

	return model.syncScreens()
}

func (m Model) Init() tea.Cmd {
	cmds := []tea.Cmd{
		m.chat.Init(),
		m.initWizard.Init(),
		m.planApproval.Init(),
		m.diffReview.Init(),
		m.gitPreflight.Init(),
		m.pluginApproval.Init(),
		m.permissionOverview.Init(),
		m.permissionPrompt.Init(),
		m.toolDetails.Init(),
		m.waitPermissionRequest(),
		m.waitPlanRequest(),
	}

	if m.workflow.Route == app.RouteBooting {
		cmds = append(cmds, services.BootSandboxCmd(m.workflow.Session.RunnerConfig.Session))
	}

	return tea.Batch(cmds...)
}
