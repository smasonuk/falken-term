package services

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/smasonuk/falken-core/pkg/falken"
	"github.com/smasonuk/falken-term/internal/ui/app"
)

func GenerateDiffCmd(session *falken.Session) tea.Cmd {
	return func() tea.Msg {
		if session == nil {
			return app.DiffGeneratedMsg{}
		}
		diff, err := session.GenerateDiff()
		return app.DiffGeneratedMsg{Diff: diff, Err: err}
	}
}

func ApplyDiffCmd(session *falken.Session, diff string) tea.Cmd {
	return func() tea.Msg {
		if session == nil {
			return app.DiffApplyFinishedMsg{Applied: false, Err: nil}
		}
		result, err := session.ApplyDiff(diff)
		return diffApplyFinishedMsg(result, err)
	}
}

func diffApplyFinishedMsg(result falken.DiffApplyResult, err error) app.DiffApplyFinishedMsg {
	return app.DiffApplyFinishedMsg{
		Applied:      err == nil,
		Partial:      result.Partial,
		SkippedFiles: result.SkippedFiles,
		Err:          err,
	}
}
