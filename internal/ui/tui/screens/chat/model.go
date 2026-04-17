package chat

import (
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/smasonuk/falken-core/pkg/falken"
	"github.com/smasonuk/falken-term/internal/ui/app"
)

type State int

const (
	// StatePrompt State = iota
	StateRunning State = iota
	StateDone
)

type Model struct {
	state          State
	session        *falken.Session
	runtime        *app.AgentSessionState
	textarea       textarea.Model
	viewport       viewport.Model
	streamViewport viewport.Model
	textinput      textinput.Model
	help           help.Model
	keyMap         keyMap
	spinner        spinner.Model
	todosPath      string
	width, height  int
	resized        bool
	lastChatLayout *chatLayout
}

func NewModel(session *falken.Session, todosPath string, runtime *app.AgentSessionState) Model {
	ta := textarea.New()
	ta.Placeholder = "Enter your task description here... (Enter to submit, /help for commands)"
	ta.Focus()
	ta.SetWidth(80)
	ta.SetHeight(10)

	vp := viewport.New(80, 20)
	vp.SetContent("Agent logs will appear here...")

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	svp := viewport.New(80, 8)
	svp.SetContent("")

	ti := textinput.New()
	ti.Placeholder = "Enter command or message... (/help for commands)"
	ti.CharLimit = 156
	ti.Width = 80

	h := help.New()
	h.ShowAll = false

	return Model{
		state:          StateDone,
		session:        session,
		textarea:       ta,
		viewport:       vp,
		streamViewport: svp,
		spinner:        s,
		textinput:      ti,
		help:           h,
		keyMap:         newKeyMap(),
		todosPath:      todosPath,
		runtime:        runtime,
	}
}

func (m Model) Init() tea.Cmd {
	return textarea.Blink
}

// func (m Model) SetPromptState() (Model, tea.Cmd) {
// 	m.state = StatePrompt
// 	if m.runtime != nil {
// 		m.runtime.IsActive = false
// 		m.runtime.CancelFunc = nil
// 	}
// 	return m, m.textarea.Focus()
// }

func (m Model) SetDoneState() (Model, tea.Cmd) {
	m.state = StateDone
	if m.runtime != nil {
		m.runtime.IsActive = false
	}
	return m, m.textinput.Focus()
}

func (m Model) BeginRun(displayPrompt string) Model {
	m.state = StateRunning
	if m.runtime != nil {
		m.runtime.IsActive = true
	}
	m.textinput.Blur()
	m.textinput.SetValue("")
	m.textarea.Blur()
	m.textarea.SetValue("")
	m.appendUserPrompt(displayPrompt)
	return m
}

func (m Model) SpinnerTickCmd() tea.Cmd {
	return m.spinner.Tick
}

func (m Model) CurrentState() State {
	return m.state
}

func (m Model) SetRuntime(runtime *app.AgentSessionState) Model {
	m.runtime = runtime
	return m
}
