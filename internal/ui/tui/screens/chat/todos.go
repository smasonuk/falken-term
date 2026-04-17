package chat

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/smasonuk/falken-term/internal/ui/tui/shared"

	"github.com/charmbracelet/lipgloss"
)

type todoStatus string

const (
	todoPending    todoStatus = "pending"
	todoInProgress todoStatus = "in_progress"
	todoCompleted  todoStatus = "completed"
)

type todoItem struct {
	ID       string     `json:"id"`
	Content  string     `json:"content"`
	Status   todoStatus `json:"status"`
	Priority string     `json:"priority,omitempty"`
}

func formatTodosForUI(data []byte) string {
	var allTodos []todoItem
	if err := json.Unmarshal(data, &allTodos); err != nil {
		return "Error reading todos: " + err.Error()
	}

	if len(allTodos) == 0 {
		return "No todos yet."
	}

	var sb strings.Builder
	for _, t := range allTodos {
		var statusPrefix string
		var currentStyle lipgloss.Style

		switch t.Status {
		case todoPending:
			statusPrefix = "[ ]"
			currentStyle = shared.TaskPendingStyle
		case todoInProgress:
			statusPrefix = "[▶]"
			currentStyle = shared.TaskInProgressStyle
		case todoCompleted:
			statusPrefix = "[✓]"
			currentStyle = shared.TaskCompletedStyle
		default:
			statusPrefix = "[?]"
			currentStyle = shared.TaskPendingStyle
		}

		line := fmt.Sprintf("%s %s: %s", statusPrefix, t.ID, t.Content)
		if t.Priority != "" {
			line += fmt.Sprintf(" (Pri: %s)", t.Priority)
		}
		sb.WriteString(currentStyle.Render(line) + "\n")
	}

	return sb.String()
}
