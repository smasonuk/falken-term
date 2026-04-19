# falken-term

`falken-term` is a terminal user interface for running a Falken coding session from the command line. It wraps a `falken-core` session in a Bubble Tea application, walks the user through startup safety checks, and then provides a chat-style agent interface with permission prompts, plan approval, live command output, todo tracking, and diff review before changes are applied back to the host workspace.

## What It Does

- Starts a `falken-core` session backed by a local sibling checkout of `../falken-core`.
- Loads project permissions from `.falken.yaml`.
- Boots the sandbox environment before entering chat.
- Streams agent events into a terminal UI with logs, live command output, and todo state.
- Intercepts permission requests and plan approvals with modal prompts.
- Reviews generated diffs before applying them to the host workspace.

## Requirements

- Go `1.25`
- A sibling checkout of `falken-core` at `../falken-core`
- Docker or whatever sandbox runtime `falken-core` expects in your environment
- A Portkey API key exported as `PK`
- A project-level `.falken.yaml` permissions file in the workspace you launch from

`go.mod` uses a local `replace` directive:

```go
replace github.com/smasonuk/falken-core => ../falken-core
```

That means this project will not build correctly unless `falken-core` is present beside `falken-term`.

## Build And Run

Build the binary:

```bash
./build.sh
```

Or with `make`:

```bash
make all
```

Run the app with debug logging enabled:

```bash
PK=your-portkey-key ./run.sh
```

Run the binary directly:

```bash
PK=your-portkey-key ./falken-term --debug
```

The `--debug` flag writes detailed request and response logs to `.falken/debug.log` inside the resolved Falken state directory.

## Startup Flow

When the app launches, it builds a `falken.Session`, creates a UI bridge, and moves through the following route flow:

1. Git preflight if the current worktree is dirty.
2. First-run cache and network setup if no caches are configured.
3. External plugin approval for unapproved plugins.
4. Permission overview if configured to show on startup.
5. Sandbox boot.
6. Chat screen.

This flow is computed in `internal/ui/app/update.go` and rendered through the screen models in `internal/ui/tui/screens`.

## Chat Experience

The main chat screen has four pieces:

- A prompt composer for the initial request
- A scrolling log pane for thoughts, tool activity, and assistant output
- A live command stream pane that appears while shell output is streaming
- A todo sidebar sourced from the session todo file

While the agent is active, the UI can open modal prompts for:

- Permission requests
- Plan approval
- Inspection of the last tool call and result

After the agent submits work, the UI switches to a diff review screen where the user can inspect changed files and decide whether to apply the patch back to the host workspace.

## Slash Commands

The chat screen supports these built-in commands:

- `/help` shows available commands.
- `/new` resets the current conversation state and clears persisted session context.
- `/plan <request>` forces the next run into plan mode and initializes the runtime plan.
- `/push` skips to diff review for the current sandbox state.
- `/exit` or `/quit` closes the application.

Keyboard shortcuts also vary by state:

- `Enter` submits the current prompt.
- `Ctrl+C` interrupts a running agent, or quits when idle.
- `Ctrl+I` opens the tool details modal after a run has started.
- `Esc` quits from the chat screen and closes the tool details modal.

## Project Layout

```text
.
├── main.go                     # Entrypoint: flags, config, session, TUI startup
├── internal/bootstrap         # OpenAI client setup and packaged system prompt
├── internal/ui/app            # Route state, messages, and app-level session state
├── internal/ui/app/services   # Commands that call session, git, sandbox, and config APIs
├── internal/ui/tui            # Bubble Tea root model and workflow orchestration
└── internal/ui/tui/screens    # Screen-specific models such as chat and diff review
```

## Testing

Run the Go test suite:

```bash
go test ./...
```

The current test coverage is focused on:

- Startup route selection
- Modal and workflow transitions
- Git preflight behavior
- Diff apply result handling

## Notes For Contributors

- The app assumes `falken-core` owns session execution, diff generation, diff application, and permissions persistence.
- `falken-term` is mainly responsible for UI state, event routing, and user interaction decisions.
- If you are changing UI flow or adding a new screen, start with `internal/ui/tui/model.go` and `internal/ui/tui/update.go`.
- If you are changing startup sequencing, start with `internal/ui/app/update.go`.
- If you are changing chat behavior, start with `internal/ui/tui/screens/chat`.

See [ARCHITECTURE.md](ARCHITECTURE.md) for a contributor-oriented walkthrough of the runtime flow and package boundaries.
