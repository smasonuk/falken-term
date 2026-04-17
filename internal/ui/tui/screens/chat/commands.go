package chat

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/smasonuk/falken-term/internal/ui/app"
	"github.com/smasonuk/falken-term/internal/ui/tui/shared"
)

type SlashCommandChat struct {
	Names       []string
	Description string
	Handler     func(m Model, args string) (Model, tea.Cmd)
}

func (m Model) slashCommands() []SlashCommandChat {
	return []SlashCommandChat{
		{
			Names:       []string{"/exit", "/quit"},
			Description: "Exit the application",
			Handler: func(m Model, args string) (Model, tea.Cmd) {
				if m.runtime != nil && m.runtime.CancelFunc != nil {
					m.runtime.CancelFunc()
				}
				return m, tea.Quit
			},
		},
		{
			Names:       []string{"/new"},
			Description: "Start a new session and clear persisted conversation state",
			Handler: func(m Model, args string) (Model, tea.Cmd) {
				m.state = StateDone
				if m.runtime != nil {
					if m.runtime.CancelFunc != nil {
						m.runtime.CancelFunc()
					}
					m.runtime.IsActive = false
					m.runtime.Logs = ""
					m.runtime.TodoState = "No todos yet."
					m.runtime.TransientStatus = ""
					m.runtime.HiddenToolUses = 0
					m.runtime.LiveCommandOutput = ""
					m.runtime.CancelFunc = nil
					m.runtime.EventChan = nil
				}
				if m.session != nil {
					if err := m.session.ResetConversationState(); err != nil {
						if m.runtime != nil {
							m.runtime.Logs = shared.ErrorStyle.Render(fmt.Sprintf("Failed to fully reset session state: %v", err))
						}
					}
				} else if m.todosPath != "" {
					_ = os.Remove(m.todosPath)
				}
				m.textinput.SetValue("")
				m.viewport.SetContent("Agent logs will appear here...")
				m.textarea.Focus()
				m.textarea.SetValue("")
				return m, nil
			},
		},
		{
			Names:       []string{"/plan"},
			Description: "Force the agent to enter plan mode before executing",
			Handler: func(m Model, args string) (Model, tea.Cmd) {
				if m.session != nil {
					m.session.ForcePlanMode(true)
					_ = os.WriteFile(filepath.Join(m.session.Paths().WorkspaceDir, ".agent_plan.md"), []byte("# Implementation Plan\n\n"), 0644)
				}
				prompt := fmt.Sprintf("System: The user has placed you directly into Plan Mode for this request. Read the codebase, write your plan to `.agent_plan.md`, and call the `exit_plan_mode` tool. You must do this before taking any action.\n\nUser Request: %s", args)
				return m, func() tea.Msg {
					return app.StartChatRunMsg{
						Prompt:        prompt,
						DisplayPrompt: "/plan " + args,
					}
				}
			},
		},
		{
			Names:       []string{"/push"},
			Description: "Submit current sandbox changes for review and application to host",
			Handler: func(m Model, args string) (Model, tea.Cmd) {
				m.runtime.TransientStatus = ""
				m.runtime.HiddenToolUses = 0
				if m.runtime.CancelFunc != nil {
					m.runtime.CancelFunc()
					m.runtime.CancelFunc = nil
				}
				m.runtime.Logs += "\n" + shared.SuccessStyle.Render("User initiated /push. Entering diff review.")
				m.viewport.SetContent(m.runtime.Logs)
				m.viewport.GotoBottom()
				m.textinput.SetValue("")
				m.textarea.SetValue("")
				return m, func() tea.Msg {
					return app.NavigateToRouteMsg{Route: app.RouteReviewDiff}
				}
			},
		},
		{
			Names:       []string{"/help"},
			Description: "Show this help message",
			Handler: func(m Model, args string) (Model, tea.Cmd) {
				var helpText strings.Builder
				helpText.WriteString("\n" + shared.SystemStyle.Render("Available Commands:") + "\n")
				for _, cmd := range m.slashCommands() {
					line := fmt.Sprintf("  %-14s - %s", strings.Join(cmd.Names, ", "), cmd.Description)
					helpText.WriteString(shared.SystemStyle.Render(line) + "\n")
				}
				if m.runtime.Logs == "" {
					m.runtime.Logs = strings.TrimSpace(helpText.String())
				} else {
					m.runtime.Logs += helpText.String()
				}
				m.viewport.SetContent(m.runtime.Logs)
				m.viewport.GotoBottom()
				if m.state == StateDone {
					m.state = StateDone
					m.textinput.Focus()
					m.textarea.Blur()
				}
				m.textinput.SetValue("")
				m.textarea.SetValue("")
				return m, nil
			},
		},
	}
}

func (m Model) runSlashCommand(input string) (Model, tea.Cmd) {
	parts := strings.SplitN(strings.TrimSpace(input), " ", 2)
	cmdName := strings.ToLower(parts[0])
	args := ""
	if len(parts) > 1 {
		args = parts[1]
	}

	for _, cmd := range m.slashCommands() {
		for _, name := range cmd.Names {
			if name == cmdName {
				return cmd.Handler(m, args)
			}
		}
	}

	errLog := "\n" + shared.ErrorStyle.Render(fmt.Sprintf("Unknown command: %s. Type /help for a list of commands.", cmdName))
	if m.runtime.Logs == "" {
		m.runtime.Logs = strings.TrimPrefix(errLog, "\n")
	} else {
		m.runtime.Logs += errLog
	}
	m.viewport.SetContent(m.runtime.Logs)
	m.viewport.GotoBottom()
	// if m.state == StatePrompt {
	// 	m.state = StateDone
	// 	m.textinput.Focus()
	// 	m.textarea.Blur()
	// }
	m.textinput.SetValue("")
	m.textarea.SetValue("")
	return m, nil
}
