package diffreview

import (
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/smasonuk/falken-term/internal/ui/app"
	"github.com/smasonuk/falken-term/internal/ui/tui/shared"
)

type Actions struct {
	Apply func(diff string) tea.Cmd
}

type Model struct {
	diffFiles        []string
	selectedDiffFile int
	fullDiff         string
	viewport         viewport.Model
	spinner          spinner.Model
	actions          Actions
	width, height    int
	isGenerating     bool
	isApplying       bool
	lastError        string
}

func NewModel(actions Actions) Model {
	vp := viewport.New(80, 20)
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	return Model{
		viewport: vp,
		spinner:  s,
		actions:  actions,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case app.DiffGeneratedMsg:
		m.isGenerating = false
		if msg.Err != nil {
			m.lastError = msg.Err.Error()
		}
		return m, nil

	case app.DiffApplyFinishedMsg:
		m.isApplying = false
		if !msg.Applied && msg.Err != nil {
			m.lastError = msg.Err.Error()
		} else {
			m.lastError = ""
		}
		return m, nil

	case spinner.TickMsg:
		if !m.isGenerating && !m.isApplying {
			return m, nil
		}

		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case tea.KeyMsg:
		if m.isGenerating || m.isApplying {
			return m, nil
		}

		switch msg.String() {
		case "up", "k":
			if m.selectedDiffFile > 0 {
				m.selectedDiffFile--
				m.updateDiffViewport()
			}
		case "down", "j":
			if m.selectedDiffFile < len(m.diffFiles)-1 {
				m.selectedDiffFile++
				m.updateDiffViewport()
			}
		case "y", "Y", "enter":
			if m.fullDiff == "" || m.actions.Apply == nil {
				return m, nil
			}

			m.lastError = ""
			m.isApplying = true
			return m, tea.Batch(m.actions.Apply(m.fullDiff), m.spinner.Tick)
		case "n", "N", "esc":
			return m, func() tea.Msg { return app.DiscardDiffRequestedMsg{} }
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	}

	var vpCmd tea.Cmd
	m.viewport, vpCmd = m.viewport.Update(msg)
	return m, vpCmd
}

func (m Model) View() string {
	leftPaneStyle := lipgloss.NewStyle().
		Width(m.width * 1 / 4).
		Height(m.height - 5).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62"))

	rightPaneStyle := lipgloss.NewStyle().
		Width(m.width*3/4 - 4).
		Height(m.height - 5).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62"))

	var fileList strings.Builder
	fileList.WriteString(shared.HeaderStyle.Render(" FILES ") + "\n\n")
	for i, file := range m.diffFiles {
		style := lipgloss.NewStyle()
		if i == m.selectedDiffFile {
			style = style.Background(lipgloss.Color("62")).Foreground(lipgloss.Color("230"))
		}
		fileList.WriteString(style.Render(file) + "\n")
	}

	m.viewport.Width = m.width*3/4 - 6
	m.viewport.Height = m.height - 7

	leftPane := leftPaneStyle.Render(fileList.String())
	rightPane := rightPaneStyle.Render(m.viewport.View())

	topSection := lipgloss.JoinHorizontal(lipgloss.Top, leftPane, rightPane)

	footer := shared.HelpStyle.Render(" [j/k] Navigate  [y] Apply Changes  [n/esc] Discard & Exit")
	if m.isGenerating {
		footer = shared.HelpStyle.Render(" " + m.spinner.View() + " Generating diff for review... Please wait. Input is temporarily disabled.")
	} else if m.isApplying {
		footer = shared.HelpStyle.Render(" " + m.spinner.View() + " Applying changes to host workspace... Please wait. Input is temporarily disabled.")
	}

	if m.lastError != "" {
		errorBanner := lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")). // Red
			Bold(true).
			Padding(0, 1).
			Render("Error: " + m.lastError)
		footer = lipgloss.JoinVertical(lipgloss.Center, errorBanner, footer)
	}

	return lipgloss.JoinVertical(lipgloss.Center,
		shared.HeaderStyle.Render("REVIEW AGENT SUBMISSION"),
		topSection,
		footer)
}

func (m *Model) updateDiffViewport() {
	if len(m.diffFiles) == 0 {
		m.viewport.SetContent("No changes detected.")
		return
	}

	selectedFile := m.diffFiles[m.selectedDiffFile]

	lines := strings.Split(m.fullDiff, "\n")
	var fileDiff []string
	found := false
	for _, line := range lines {
		if strings.HasPrefix(line, "diff --git") {
			if strings.Contains(line, " b/"+selectedFile) {
				found = true
				fileDiff = append(fileDiff, line)
			} else if found {
				break
			}
		} else if found {
			fileDiff = append(fileDiff, line)
		}
	}

	m.viewport.SetContent(strings.Join(fileDiff, "\n"))
	m.viewport.GotoTop()
}

func (m *Model) SetDiff(diffStr string) {
	m.isGenerating = false
	m.fullDiff = diffStr
	m.diffFiles = parseDiffFiles(diffStr)
	m.selectedDiffFile = 0
	m.lastError = ""
	m.updateDiffViewport()
}

func (m Model) FullDiff() string {
	return m.fullDiff
}

func (m Model) BeginDiffGeneration() (Model, tea.Cmd) {
	m.isGenerating = true
	m.isApplying = false
	m.lastError = ""
	m.fullDiff = ""
	m.diffFiles = nil
	m.selectedDiffFile = 0
	m.viewport.SetContent("Generating diff...")
	m.viewport.GotoTop()
	return m, m.spinner.Tick
}

func parseDiffFiles(diffStr string) []string {
	var files []string
	lines := strings.Split(diffStr, "\n")
	for _, line := range lines {
		if !strings.HasPrefix(line, "diff --git") {
			continue
		}
		parts := strings.Split(line, " ")
		if len(parts) < 4 {
			continue
		}
		files = append(files, strings.TrimPrefix(parts[3], "b/"))
	}
	return files
}
