package shared

import "github.com/charmbracelet/lipgloss"

var (
	HeaderStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("205")).
			Bold(true)

	HelpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))

	BorderStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("62"))

	ThoughtStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("63"))
	SystemStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	ErrorStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true)
	SuccessStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
	WarningStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Bold(true)
	PromptStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("39")).Bold(true)
	ToolNameStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("208")).Bold(true)

	TaskPendingStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	TaskInProgressStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Bold(true)
	TaskCompletedStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Strikethrough(true)
	TaskBlockedStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
)
