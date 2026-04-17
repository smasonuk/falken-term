package ui

import (
	"fmt"

	"github.com/smasonuk/falken-core/pkg/falken"
	"github.com/smasonuk/falken-term/internal/ui/app"
	"github.com/smasonuk/falken-term/internal/ui/tui"

	tea "github.com/charmbracelet/bubbletea"
)

type Bridge = tui.SessionBridge

func NewBridge() *Bridge {
	return tui.NewSessionBridge()
}

type Config struct {
	Session    *falken.Session
	PermConfig *falken.PermissionsConfig
	Bridge     *Bridge
}

func Run(cfg Config) error {
	if cfg.Session == nil {
		return fmt.Errorf("session is required")
	}
	if cfg.Bridge == nil {
		return fmt.Errorf("bridge is required")
	}

	runnerConfig := app.RunnerConfig{
		Session:    cfg.Session,
		PermConfig: cfg.PermConfig,
		Tools:      cfg.Session.ToolInfos(),
		Plugins:    cfg.Session.PluginInfos(),
	}

	p := tea.NewProgram(tui.NewModel(runnerConfig, cfg.Bridge), tea.WithAltScreen())
	_, err := p.Run()
	return err
}
