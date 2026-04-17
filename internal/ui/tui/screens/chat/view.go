package chat

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

type chatLayout struct {
	leftWidth    int
	rightWidth   int
	totalHeight  int
	mainHeight   int
	streamHeight int
	showStream   bool
}

func (m Model) View() string {
	if m.state == StatePrompt {
		return m.renderPromptView()
	}
	return m.renderChatView()
}

func (m Model) renderPromptView() string {
	m.help.Width = max(0, m.width-2)

	return lipgloss.Place(
		m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		lipgloss.JoinVertical(
			lipgloss.Left,
			"What would you like me to do?",
			"",
			m.textarea.View(),
			"",
			m.help.View(m.keyMap.forPrompt()),
		),
	)
}

func (m Model) renderChatView() string {
	layout := m.computeLayout()
	m = m.applyLayout(layout)

	leftPane := m.renderLogsPane(layout)
	if layout.showStream {
		leftPane = lipgloss.JoinVertical(lipgloss.Left, leftPane, m.renderStreamPane(layout))
	}
	rightPane := m.renderTodoPane(layout)
	topSection := lipgloss.JoinHorizontal(lipgloss.Top, leftPane, rightPane)
	bottomSection := m.renderInputBar()

	return lipgloss.JoinVertical(lipgloss.Left, topSection, bottomSection)
}

func (m Model) computeLayout() chatLayout {
	layout := chatLayout{
		leftWidth:   m.width * 4 / 5,
		rightWidth:  (m.width * 1 / 5) - 2,
		totalHeight: m.height - 3,
		mainHeight:  m.height - 3,
	}

	if m.runtime != nil && m.runtime.LiveCommandOutput != "" {
		layout.showStream = true
		layout.streamHeight = 10
		layout.mainHeight -= layout.streamHeight
	}

	if layout.rightWidth < 0 {
		layout.rightWidth = 0
	}
	if layout.totalHeight < 0 {
		layout.totalHeight = 0
	}
	if layout.mainHeight < 0 {
		layout.mainHeight = 0
	}

	return layout
}

func (m Model) applyLayout(layout chatLayout) Model {
	m.viewport.Width = layout.leftWidth
	m.viewport.Height = max(0, layout.mainHeight-2)

	m.streamViewport.Width = max(0, layout.leftWidth-4)
	m.streamViewport.Height = max(0, layout.streamHeight-2)
	if m.runtime != nil {
		m.streamViewport.SetContent(m.runtime.LiveCommandOutput)
	}

	m.help.Width = max(0, m.width-2)

	return m
}

func (m Model) renderLogsPane(layout chatLayout) string {
	leftPaneStyle := lipgloss.NewStyle().
		Width(layout.leftWidth).
		Height(layout.mainHeight).
		Border(lipgloss.RoundedBorder())

	m.viewport.SetContent(m.logsContent())
	return leftPaneStyle.Render(m.viewport.View())
}

func (m Model) renderStreamPane(layout chatLayout) string {
	streamBoxStyle := lipgloss.NewStyle().
		Width(layout.leftWidth).
		Height(layout.streamHeight).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("214"))

	return streamBoxStyle.Render(m.streamViewport.View())
}

func (m Model) renderTodoPane(layout chatLayout) string {
	rightPaneStyle := lipgloss.NewStyle().
		Width(layout.rightWidth).
		Height(layout.totalHeight).
		Border(lipgloss.RoundedBorder())

	todoState := "No todos yet."
	if m.runtime != nil && m.runtime.TodoState != "" {
		todoState = m.runtime.TodoState
	}

	return rightPaneStyle.Render(todoState)
}

func (m Model) renderInputBar() string {
	var content strings.Builder
	content.WriteString(m.textinput.View())

	helpView := m.help.View(m.keyMap.forState(m.state))
	if helpView != "" {
		content.WriteString("\n")
		content.WriteString(helpView)
	}

	return lipgloss.NewStyle().
		Width(m.width).
		Padding(0, 1).
		Render(content.String())
}

func (m Model) logsContent() string {
	if m.runtime == nil {
		return ""
	}

	viewContent := m.runtime.Logs
	if m.runtime.TransientStatus != "" {
		transientStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Italic(true)
		viewContent += "\n\n" + transientStyle.Render(m.runtime.TransientStatus)
	}
	if m.runtime.IsActive {
		viewContent += "\n\n" + m.spinner.View() + " Agent is thinking..."
	}

	return viewContent
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
