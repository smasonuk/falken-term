package services

import (
	"context"

	"github.com/smasonuk/falken-core/pkg/falken"
	"github.com/smasonuk/falken-term/internal/ui/app"

	tea "github.com/charmbracelet/bubbletea"
)

type AgentRunHandle struct {
	EventChan chan falken.Event
	Cancel    context.CancelFunc
}

func StartAgentRun(session *falken.Session, eventChan chan falken.Event, prompt string) (*AgentRunHandle, tea.Cmd) {
	ctx, cancel := context.WithCancel(context.Background())

	handle := &AgentRunHandle{
		EventChan: eventChan,
		Cancel:    cancel,
	}

	cmd := func() tea.Msg {
		go func() {
			_ = session.Run(ctx, prompt)
		}()
		return app.SessionEventMsg{Event: <-eventChan}
	}

	return handle, cmd
}

func WaitForAgentEvent(eventChan chan falken.Event) tea.Cmd {
	return func() tea.Msg {
		return app.SessionEventMsg{Event: <-eventChan}
	}
}
