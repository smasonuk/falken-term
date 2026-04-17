package services

import (
	"fmt"
	"os/exec"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/smasonuk/falken-term/internal/ui/app"
)

const (
	GitPreflightActionStash     = "stash"
	GitPreflightActionCommitWIP = "commit_wip"
	GitPreflightActionIgnore    = "ignore"
)

func CheckDirtyWorktree() (bool, string, error) {
	if err := exec.Command("git", "rev-parse", "--is-inside-work-tree").Run(); err != nil {
		return false, "", nil
	}

	out, err := exec.Command("git", "status", "--porcelain").Output()
	if err != nil {
		return false, "", err
	}

	status := strings.TrimSpace(string(out))
	return status != "", status, nil
}

func GitStashCmd() tea.Cmd {
	return gitActionCmd(GitPreflightActionStash, func() error {
		return exec.Command("git", "stash").Run()
	})
}

func GitCommitWIPCmd() tea.Cmd {
	return gitActionCmd(GitPreflightActionCommitWIP, func() error {
		if err := exec.Command("git", "add", ".").Run(); err != nil {
			return err
		}
		return exec.Command("git", "commit", "-m", "WIP: Pre-Falken agent run").Run()
	})
}

func gitActionCmd(action string, run func() error) tea.Cmd {
	return func() tea.Msg {
		err := run()
		if err != nil {
			err = fmt.Errorf("git %s failed: %w", action, err)
		}
		return app.GitPreflightFinishedMsg{Action: action, Err: err}
	}
}
