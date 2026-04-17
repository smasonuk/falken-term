package chat

import (
	"fmt"
	"os"
	"strings"

	"github.com/smasonuk/falken-core/pkg/falken"
	"github.com/smasonuk/falken-term/internal/ui/app"
	"github.com/smasonuk/falken-term/internal/ui/app/services"
	"github.com/smasonuk/falken-term/internal/ui/tui/shared"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var (
		tiCmd tea.Cmd
		vpCmd tea.Cmd
		taCmd tea.Cmd
		spCmd tea.Cmd
	)

	switch msg := msg.(type) {
	case spinner.TickMsg:
		if m.runtime != nil && m.runtime.IsActive {
			m.spinner, spCmd = m.spinner.Update(msg)
			return m, spCmd
		}
		return m, nil

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			if m.state == StateRunning {
				if m.runtime != nil && m.runtime.CancelFunc != nil {
					m.runtime.CancelFunc()
					m.AppendSystemError("[System: Agent interrupted]")
				}
				return m, nil
			}
			return m, tea.Quit

		case tea.KeyEsc:
			return m, tea.Quit

		case tea.KeyCtrlI:
			if m.state == StateRunning || m.state == StateDone {
				return m, func() tea.Msg { return app.ShowModalMsg{Modal: app.ModalToolDetails} }
			}
			return m, nil

		case tea.KeyEnter:
			if m.state == StatePrompt {
				prompt := strings.TrimSpace(m.textarea.Value())
				return m.processInput(prompt)
			}
			if m.state == StateDone {
				prompt := strings.TrimSpace(m.textinput.Value())
				return m.processInput(prompt)
			}
		}

	case app.SessionEventMsg:
		return m.handleSessionEvent(msg.Event)

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.textarea.SetWidth(m.width - 4)
		m.textinput.Width = m.width - 2
	}

	if m.state == StatePrompt {
		m.textarea, taCmd = m.textarea.Update(msg)
	} else {
		if m.state == StateDone && !m.textinput.Focused() {
			m.textinput.Focus()
		}
		m.viewport, vpCmd = m.viewport.Update(msg)
		m.textinput, tiCmd = m.textinput.Update(msg)
	}

	return m, tea.Batch(taCmd, vpCmd, tiCmd)
}

func (m Model) handleSessionEvent(event falken.Event) (Model, tea.Cmd) {
	switch event.Type {
	case falken.EventTypeThought:
		m.appendThought(event.Thought.Text)
		return m, services.WaitForAgentEvent(m.runtime.EventChan)

	case falken.EventTypeAssistantText:
		m.finishHiddenReadSummary()
		m.runtime.Logs += event.AssistantText.Text
		m.viewport.SetContent(m.runtime.Logs)
		m.viewport.GotoBottom()
		return m, services.WaitForAgentEvent(m.runtime.EventChan)

	case falken.EventTypeToolCall:
		m.recordToolCall(*event.ToolCall)
		if isReadOnlyTool(event.ToolCall.Name) {
			return m, services.WaitForAgentEvent(m.runtime.EventChan)
		}
		return m, tea.Batch(services.WaitForAgentEvent(m.runtime.EventChan), m.spinner.Tick)

	case falken.EventTypeToolResult:
		m.recordToolResult(*event.ToolResult)
		if data, err := os.ReadFile(m.todosPath); err == nil {
			m.runtime.TodoState = formatTodosForUI(data)
		}
		m.runtime.LiveCommandOutput = ""
		m.streamViewport.SetContent("")
		return m, services.WaitForAgentEvent(m.runtime.EventChan)

	case falken.EventTypeCommandChunk:
		m.runtime.LiveCommandOutput += event.CommandChunk.Chunk
		if len(m.runtime.LiveCommandOutput) > 10000 {
			m.runtime.LiveCommandOutput = m.runtime.LiveCommandOutput[len(m.runtime.LiveCommandOutput)-10000:]
		}
		m.streamViewport.SetContent(m.runtime.LiveCommandOutput)
		m.streamViewport.GotoBottom()
		return m, services.WaitForAgentEvent(m.runtime.EventChan)

	case falken.EventTypeWorkSubmitted:
		m.finishSubmission()
		return m, func() tea.Msg { return app.NavigateToRouteMsg{Route: app.RouteReviewDiff} }

	case falken.EventTypeRunCompleted:
		m.finishAgentDone(nil)
		return m, nil

	case falken.EventTypeRunFailed:
		m.finishAgentDone(event.RunFailed.Error)
		return m, nil
	}

	return m, nil
}

func (m Model) processInput(input string) (Model, tea.Cmd) {
	input = strings.TrimSpace(input)
	if input == "" {
		return m, nil
	}

	if strings.HasPrefix(input, "/") {
		return m.runSlashCommand(input)
	}

	return m, func() tea.Msg {
		return app.StartChatRunMsg{
			Prompt:        input,
			DisplayPrompt: input,
		}
	}
}

func (m *Model) appendUserPrompt(displayPrompt string) {
	userLog := "\n" + shared.PromptStyle.Render("> "+displayPrompt)
	if m.runtime.Logs == "" {
		m.runtime.Logs = strings.TrimPrefix(userLog, "\n")
	} else {
		m.runtime.Logs += userLog
	}
	m.viewport.SetContent(m.runtime.Logs)
	m.viewport.GotoBottom()
}

func (m *Model) finishSubmission() {
	m.runtime.IsActive = false
	m.runtime.TransientStatus = ""
	m.runtime.HiddenToolUses = 0
	m.runtime.CancelFunc = nil
	m.state = StateDone
	m.runtime.Logs += "\n" + shared.SuccessStyle.Render("Agent submitted changes for review.")
	m.viewport.SetContent(m.runtime.Logs)
	m.viewport.GotoBottom()
}

func (m *Model) finishAgentDone(err error) {
	m.finishHiddenReadSummary()
	m.state = StateDone
	m.runtime.IsActive = false
	m.runtime.CancelFunc = nil
	if err != nil && err.Error() != "context canceled" {
		m.runtime.Logs += "\n" + shared.ErrorStyle.Render(fmt.Sprintf("Agent finished with error: %v", err))
	} else {
		m.runtime.Logs += "\n" + shared.SuccessStyle.Render("Agent session complete.")
	}
	m.viewport.SetContent(m.runtime.Logs)
	m.viewport.GotoBottom()
	m.textinput.Focus()
}
