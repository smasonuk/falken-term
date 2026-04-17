package main

import (
	"context"
	"flag"
	"io"
	"log"
	"os"

	"github.com/smasonuk/falken-term/internal/bootstrap"
	"github.com/smasonuk/falken-term/internal/ui"

	"github.com/smasonuk/falken-core/pkg/falken"
)

func main() {
	debugFlag := flag.Bool("debug", false, "Enable detailed debug logging to .falken/debug.log")
	flag.Parse()

	paths, err := falken.NewPaths("", "")
	if err != nil {
		log.Fatalf("Failed to resolve runtime paths: %v", err)
	}

	permConfig, err := falken.LoadPermissionsConfigFromPath(paths.WorkspaceDir + "/.falken.yaml")
	if err != nil {
		log.Fatalf("Error loading permissions config: %v", err)
	}

	var debugLog *log.Logger
	if *debugFlag {
		if err := os.MkdirAll(paths.StateDir, 0755); err != nil {
			log.Fatalf("Failed to create state directory: %v", err)
		}
		logFile, err := os.OpenFile(paths.DebugLogPath(), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			log.Fatalf("Failed to open debug log file: %v", err)
		}
		defer logFile.Close()
		debugLog = log.New(logFile, "[DEBUG] ", log.Ldate|log.Ltime|log.Lmicroseconds)
		log.Printf("Debug logging enabled. Writing to %s", paths.DebugLogPath())
	} else {
		debugLog = log.New(io.Discard, "", 0)
	}

	modelName := "gpt-5.2"

	portkeyAPIKey := os.Getenv("PK")
	if portkeyAPIKey == "" {
		log.Fatal("Error: PK environment variable is not set")
	}
	client := bootstrap.NewOpenAIClient(portkeyAPIKey, debugLog)

	bridge := ui.NewBridge()
	session, err := falken.NewSession(falken.Config{
		Client:             client,
		ModelName:          modelName,
		WorkspaceDir:       paths.WorkspaceDir,
		StateDir:           paths.StateDir,
		SystemPrompt:       bootstrap.SystemPrompt,
		Logger:             debugLog,
		PermissionsConfig:  permConfig,
		InteractionHandler: bridge,
		EventHandler:       bridge,
		Debug:              *debugFlag,
	})
	if err != nil {
		log.Fatalf("Failed to initialize session: %v", err)
	}
	defer session.Close(context.Background())

	if err := ui.Run(ui.Config{Session: session, PermConfig: permConfig, Bridge: bridge}); err != nil {
		log.Fatalf("Error running TUI: %v", err)
	}
}
