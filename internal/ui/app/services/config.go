package services

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/smasonuk/falken-core/pkg/falken"
	"github.com/smasonuk/falken-term/internal/ui/app"
)

func SavePermissionsConfig(cfg *falken.PermissionsConfig) error {
	return falken.SavePermissionsConfig(cfg)
}

func SavePermissionsConfigCmd(cfg *falken.PermissionsConfig) tea.Cmd {
	return func() tea.Msg {
		return app.PermissionsConfigSavedMsg{Err: SavePermissionsConfig(cfg)}
	}
}
