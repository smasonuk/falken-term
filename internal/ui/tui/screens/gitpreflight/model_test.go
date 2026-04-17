package gitpreflight

import (
	"errors"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/smasonuk/falken-term/internal/ui/app"
)

func TestGitPreflightStartsInjectedStashAction(t *testing.T) {
	called := 0
	model := NewModel("", Actions{
		Stash: func() tea.Cmd {
			called++
			return func() tea.Msg {
				return app.GitPreflightFinishedMsg{Action: "stash"}
			}
		},
	})

	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})

	if !updated.isRunning {
		t.Fatalf("expected model to enter running state")
	}
	if updated.runningAction != "stash" {
		t.Fatalf("expected running action stash, got %q", updated.runningAction)
	}
	if cmd == nil {
		t.Fatalf("expected stash command")
	}

	msg := cmd()
	if called != 1 {
		t.Fatalf("expected stash action to be started once, got %d", called)
	}
	if finished, ok := msg.(app.GitPreflightFinishedMsg); !ok || finished.Action != "stash" {
		t.Fatalf("expected stash finished message, got %#v", msg)
	}
}

func TestGitPreflightStartsInjectedCommitAction(t *testing.T) {
	called := 0
	model := NewModel("", Actions{
		CommitWIP: func() tea.Cmd {
			called++
			return func() tea.Msg {
				return app.GitPreflightFinishedMsg{Action: "commit_wip"}
			}
		},
	})

	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}})

	if !updated.isRunning {
		t.Fatalf("expected model to enter running state")
	}
	if updated.runningAction != "commit_wip" {
		t.Fatalf("expected running action commit_wip, got %q", updated.runningAction)
	}
	if cmd == nil {
		t.Fatalf("expected commit command")
	}

	msg := cmd()
	if called != 1 {
		t.Fatalf("expected commit action to be started once, got %d", called)
	}
	if finished, ok := msg.(app.GitPreflightFinishedMsg); !ok || finished.Action != "commit_wip" {
		t.Fatalf("expected commit finished message, got %#v", msg)
	}
}

func TestGitPreflightIgnoresAdditionalActionsWhileRunning(t *testing.T) {
	called := 0
	model := NewModel("", Actions{
		Stash: func() tea.Cmd {
			called++
			return func() tea.Msg { return app.GitPreflightFinishedMsg{Action: "stash"} }
		},
		CommitWIP: func() tea.Cmd {
			called++
			return func() tea.Msg { return app.GitPreflightFinishedMsg{Action: "commit_wip"} }
		},
	})

	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	if cmd == nil {
		t.Fatalf("expected first command to be returned")
	}
	if called != 1 {
		t.Fatalf("expected first action to start once, got %d", called)
	}

	updated, cmd = updated.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}})
	if cmd != nil {
		t.Fatalf("expected no command while action is already running")
	}
	if called != 1 {
		t.Fatalf("expected no additional action while running, got %d calls", called)
	}
}

func TestGitPreflightFinishedUpdatesLocalErrorState(t *testing.T) {
	model := NewModel("", Actions{})
	model.isRunning = true
	model.runningAction = "stash"

	failed, _ := model.Update(app.GitPreflightFinishedMsg{Action: "stash", Err: errors.New("boom")})
	if failed.isRunning {
		t.Fatalf("expected running state to clear after completion")
	}
	if failed.lastError != "boom" {
		t.Fatalf("expected error to be stored, got %q", failed.lastError)
	}

	cleared, _ := failed.Update(app.GitPreflightFinishedMsg{Action: "stash"})
	if cleared.lastError != "" {
		t.Fatalf("expected error to clear on success, got %q", cleared.lastError)
	}
}
