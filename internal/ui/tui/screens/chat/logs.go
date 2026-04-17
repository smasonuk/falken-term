package chat

import (
	"encoding/json"
	"fmt"

	"github.com/smasonuk/falken-core/pkg/falken"
	"github.com/smasonuk/falken-term/internal/ui/tui/shared"

	"github.com/charmbracelet/lipgloss"
)

func (m *Model) appendThought(text string) {
	m.runtime.Logs += "\n" + shared.ThoughtStyle.Render("THOUGHT: "+text)
	m.viewport.SetContent(m.runtime.Logs)
	m.viewport.GotoBottom()
}

func (m *Model) recordToolCall(msg falken.ToolCallEvent) {
	argBytes, _ := json.MarshalIndent(msg.Args, "", "  ")
	if isReadOnlyTool(msg.Name) {
		m.runtime.HiddenToolUses++
		if m.runtime.HiddenToolUses <= 3 {
			summary := formatToolSummary(msg.Name, msg.Args)
			logLine := fmt.Sprintf("\n%s %s", shared.SystemStyle.Render("Inspecting:"), shared.ToolNameStyle.Render(msg.Name))
			logLine += fmt.Sprintf("\n%s\n", shared.SystemStyle.Render("   ↳ "+summary))
			m.runtime.Logs += logLine
			m.viewport.SetContent(m.runtime.Logs)
			m.viewport.GotoBottom()
		} else {
			m.runtime.TransientStatus = fmt.Sprintf("Agent exploring deeply... (%d consecutive read operations)", m.runtime.HiddenToolUses)
		}
		_ = argBytes
		return
	}

	m.finishHiddenReadSummary()
	summary := formatToolSummary(msg.Name, msg.Args)
	logLine := fmt.Sprintf("\n%s %s", shared.SystemStyle.Render("Executing Tool:"), shared.ToolNameStyle.Render(msg.Name))
	logLine += fmt.Sprintf("\n%s\n", shared.SystemStyle.Render("   ↳ "+summary))
	m.runtime.Logs += logLine
	m.viewport.SetContent(m.runtime.Logs)
	m.viewport.GotoBottom()
	m.runtime.LiveCommandOutput = ""
	m.streamViewport.SetContent("")
}

func (m *Model) recordToolResult(msg falken.ToolResultEvent) {
	if msg.Name == "edit_file" || msg.Name == "write_file" {
		if resStr, ok := msg.Result["result"].(string); ok {
			diffStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
			m.runtime.Logs += "\n" + diffStyle.Render(resStr) + "\n"
			m.viewport.SetContent(m.runtime.Logs)
			m.viewport.GotoBottom()
		}
	}
}

func (m *Model) finishHiddenReadSummary() {
	if m.runtime.HiddenToolUses > 0 {
		if m.runtime.HiddenToolUses > 3 {
			hiddenCount := m.runtime.HiddenToolUses - 3
			m.runtime.Logs += "\n" + shared.SystemStyle.Render(fmt.Sprintf("Finished deep exploration (%d additional hidden operations)", hiddenCount))
		}
		m.runtime.HiddenToolUses = 0
		m.runtime.TransientStatus = ""
	}
}

func formatToolSummary(name string, args map[string]any) string {
	switch name {
	case "read_file", "list_directory", "ast_get_symbols":
		if path, ok := args["Path"].(string); ok {
			return fmt.Sprintf("Path: %s", path)
		}
		if path, ok := args["path"].(string); ok {
			return fmt.Sprintf("Path: %s", path)
		}
	case "write_file", "edit_file":
		if path, ok := args["Path"].(string); ok {
			return fmt.Sprintf("File: %s (Content truncated)", path)
		}
		if path, ok := args["path"].(string); ok {
			return fmt.Sprintf("File: %s (Content truncated)", path)
		}
	case "execute_command":
		if cmd, ok := args["Command"].(string); ok {
			return fmt.Sprintf("Cmd: %s", cmd)
		}
		if cmd, ok := args["command"].(string); ok {
			return fmt.Sprintf("Cmd: %s", cmd)
		}
	}

	keys := make([]string, 0, len(args))
	for k := range args {
		keys = append(keys, k)
	}
	return fmt.Sprintf("Args: %v", keys)
}

func isReadOnlyTool(name string) bool {
	switch name {
	case "read_file", "read_files", "glob", "grep", "search_tools", "fetch_url", "TaskList", "TaskGet":
		return true
	default:
		return false
	}
}

func (m *Model) AppendSystemError(text string) {
	m.runtime.Logs += "\n" + shared.ErrorStyle.Render(text)
	m.viewport.SetContent(m.runtime.Logs)
	m.viewport.GotoBottom()
}

func (m *Model) AppendSystemSuccess(text string) {
	m.runtime.Logs += "\n" + shared.SuccessStyle.Render(text)
	m.viewport.SetContent(m.runtime.Logs)
	m.viewport.GotoBottom()
}

func (m *Model) AppendSystemWarning(text string) {
	m.runtime.Logs += "\n" + shared.WarningStyle.Render(text)
	m.viewport.SetContent(m.runtime.Logs)
	m.viewport.GotoBottom()
}
