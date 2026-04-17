package initwizard

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/smasonuk/falken-term/internal/ui/app"
	"github.com/smasonuk/falken-term/internal/ui/tui/shared"
)

var RecommendedDomains map[string][]string

type Model struct {
	step           int //= Caches, 1 = Domains
	cursor         int
	options        []string
	selected       map[int]bool
	selectedCaches []string
	textinput      textinput.Model
	width, height  int
}

func NewModel() Model {
	ti := textinput.New()
	ti.Placeholder = "Enter command or message... (/help for commands)"
	ti.CharLimit = 156
	ti.Width = 80

	m := Model{
		options:   []string{"go-mod", "go-build", "npm", "pip"},
		selected:  make(map[int]bool),
		textinput: ti,
	}

	// Auto-detect based on files
	if _, err := os.Stat("go.mod"); err == nil {
		m.selected[0] = true
		m.selected[1] = true
	}
	if _, err := os.Stat("package.json"); err == nil {
		m.selected[2] = true
	}
	if _, err := os.Stat("requirements.txt"); err == nil {
		m.selected[3] = true
	}

	return m
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.step == 0 {
			// Step 1: Select Caches
			switch msg.String() {
			case "up", "k":
				if m.cursor > 0 {
					m.cursor--
				}
			case "down", "j":
				if m.cursor < len(m.options)-1 {
					m.cursor++
				}
			case " ":
				m.selected[m.cursor] = !m.selected[m.cursor]
			case "enter":
				// Gather domains based on selected caches
				domainSet := make(map[string]bool)
				m.selectedCaches = m.selectedCaches[:0]
				for i, opt := range m.options {
					if m.selected[i] {
						m.selectedCaches = append(m.selectedCaches, opt)
						for _, domain := range RecommendedDomains[opt] {
							domainSet[domain] = true
						}
					}
				}
				// Prepare text input for Step 2
				domainList := []string{}
				for d := range domainSet {
					domainList = append(domainList, d)
				}
				sort.Strings(domainList)
				m.textinput.SetValue(strings.Join(domainList, ", "))
				m.textinput.Focus()
				m.step = 1
			case "q", "esc", "ctrl+c":
				return m, tea.Quit
			}
		} else if m.step == 1 {
			// Step 2: Confirm Domains
			switch msg.String() {
			case "enter":
				// Parse domains and let the app layer mutate config/persist.
				domains := strings.Split(m.textinput.Value(), ",")
				allowedDomains := make([]string, 0, len(domains))
				for _, d := range domains {
					trimmed := strings.TrimSpace(d)
					if trimmed != "" {
						allowedDomains = append(allowedDomains, trimmed)
					}
				}
				m.textinput.Blur()
				m.textinput.SetValue("")
				selectedCaches := append([]string(nil), m.selectedCaches...)
				return m, func() tea.Msg {
					return app.InitWizardSubmittedMsg{
						SelectedCaches: selectedCaches,
						AllowedDomains: allowedDomains,
					}
				}
			case "q", "esc", "ctrl+c":
				return m, tea.Quit
			}
			m.textinput, cmd = m.textinput.Update(msg)
			return m, cmd
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.textinput.Width = m.width - 2
	}
	return m, nil
}

func (m Model) View() string {
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		Padding(1, 3).
		Align(lipgloss.Left)

	var content string
	if m.step == 0 {
		var sb strings.Builder
		sb.WriteString(shared.HeaderStyle.Render("FIRST-RUN SETUP: CACHE MOUNTS") + "\n\n")
		sb.WriteString("Select recommended cache mounts for this project:\n\n")

		for i, opt := range m.options {
			cursor := " "
			if m.cursor == i {
				cursor = ">"
			}
			checked := " "
			if m.selected[i] {
				checked = "x"
			}
			sb.WriteString(fmt.Sprintf("%s [%s] %s\n", cursor, checked, opt))
		}

		sb.WriteString("\n" + shared.HelpStyle.Render("[Space] Toggle   [Enter] Next   [q] Quit"))
		content = sb.String()
	} else {
		content = fmt.Sprintf(
			"%s FIRST-RUN SETUP: NETWORK ALLOWLIST\n\n"+
				"Verify or add domains the agent is allowed to access:\n\n"+
				"%s\n\n"+
				"%s",
			shared.HeaderStyle.Render("🌐"),
			m.textinput.View(),
			shared.HelpStyle.Render("[Enter] Save & Finish   [q] Quit"))
	}

	return lipgloss.Place(
		m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		boxStyle.Render(content),
	)
}
