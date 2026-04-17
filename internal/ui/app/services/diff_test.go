package services

import (
	"errors"

	"testing"

	"github.com/smasonuk/falken-core/pkg/falken"
)

func TestDiffApplyFinishedMsgTreatsNilErrorAsSuccess(t *testing.T) {
	msg := diffApplyFinishedMsg(falken.DiffApplyResult{}, nil)
	if !msg.Applied {
		t.Fatalf("expected apply to be marked successful when no error is returned")
	}
	if msg.Err != nil {
		t.Fatalf("expected nil error, got %v", msg.Err)
	}
}

func TestDiffApplyFinishedMsgTreatsErrorAsFailure(t *testing.T) {
	expectedErr := errors.New("apply failed")

	msg := diffApplyFinishedMsg(falken.DiffApplyResult{}, expectedErr)
	if msg.Applied {
		t.Fatalf("expected apply to be marked unsuccessful when an error is returned")
	}
	if !errors.Is(msg.Err, expectedErr) {
		t.Fatalf("expected error %v, got %v", expectedErr, msg.Err)
	}
}

func TestDiffApplyFinishedMsgIncludesPartialResult(t *testing.T) {
	msg := diffApplyFinishedMsg(falken.DiffApplyResult{
		Partial:      true,
		SkippedFiles: []string{"secret.txt"},
	}, nil)
	if !msg.Applied {
		t.Fatalf("expected partial apply to still be treated as applied")
	}
	if !msg.Partial {
		t.Fatal("expected partial flag to be preserved")
	}
	if len(msg.SkippedFiles) != 1 || msg.SkippedFiles[0] != "secret.txt" {
		t.Fatalf("unexpected skipped files: %#v", msg.SkippedFiles)
	}
}
