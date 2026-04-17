package bootstrap

import (
	"io"
	"log"
	"testing"
)

func TestSystemPromptIsPackaged(t *testing.T) {
	if SystemPrompt == "" {
		t.Fatalf("expected packaged system prompt to be non-empty")
	}
}

func TestNewOpenAIClientBuilds(t *testing.T) {
	client := NewOpenAIClient("test-key", log.New(io.Discard, "", 0))
	if client == nil {
		t.Fatalf("expected client to be created")
	}
}
