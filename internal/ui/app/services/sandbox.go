package services

import (
	"context"

	"github.com/smasonuk/falken-core/pkg/falken"
	"github.com/smasonuk/falken-term/internal/ui/app"

	tea "github.com/charmbracelet/bubbletea"
)

func BootSandboxCmd(session *falken.Session) tea.Cmd {
	return func() tea.Msg {
		if session == nil {
			return app.SandboxBootFinishedMsg{Err: nil}
		}
		return app.SandboxBootFinishedMsg{Err: session.Start(context.Background())}
	}
}
