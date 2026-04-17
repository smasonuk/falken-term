# Architecture

This document maps the main runtime flow in `falken-term` so contributors can find the right package quickly when they need to change startup behavior, add screens, or wire new Falken session events into the terminal UI.

## High-Level Design

`falken-term` is a thin UI application on top of `falken-core`:

- `falken-core` owns session lifecycle, tool execution, diff generation, diff application, sandbox startup, and permissions persistence.
- `falken-term` owns screen flow, keyboard handling, modal interaction, and rendering Falken events in a terminal UI.

At startup, `main.go` creates:

1. Runtime paths from `falken.NewPaths("", "")`
2. A permissions config loaded from `.falken.yaml`
3. An OpenAI-compatible client using the `PK` environment variable
4. A `falken.Session` configured with a system prompt and a UI bridge
5. A Bubble Tea program rooted at `internal/ui/tui.Model`

## Main Runtime Flow

The app flow is route-based. `internal/ui/app` defines the route and modal state, and `internal/ui/tui` turns that state into screen models and Bubble Tea commands.

The route sequence is:

1. `RouteStartupGitPreflight`
2. `RouteStartupInitWizard`
3. `RouteStartupPluginApproval`
4. `RouteStartupPermissionOverview`
5. `RouteBooting`
6. `RouteChat`
7. `RouteReviewDiff`

The initial route is chosen by `app.InitialRoute` and `app.NextStartupRoute`, using three inputs:

- Whether the current git worktree is dirty
- Whether caches have already been configured
- Whether there are unapproved external plugins or a pending permission overview

## Package Map

### `main.go`

Top-level composition only. It parses flags, loads config, creates the session, and launches the TUI.

Important details:

- `--debug` enables logging to `.falken/debug.log`
- `PK` must be set or startup fails
- The current model name is hard-coded to `gpt-5.2`

### `internal/bootstrap`

Bootstrap code that should stay small and explicit:

- `client.go` builds the OpenAI client and injects the Portkey base URL and provider header
- `prompt.go` packages the system prompt string used when sessions are created

This package is a good fit for startup-only concerns, but not general UI logic.

### `internal/ui/app`

Pure app/workflow state with no rendering:

- `routes.go` defines route and modal enums
- `session.go` defines `RunnerConfig`, `AppSession`, and runtime state for the active agent
- `messages.go` defines the Bubble Tea message types used across the app
- `update.go` contains workflow helpers such as startup routing, pending plugin filtering, and state mutation helpers

If a change is about application flow rather than visual layout, this is usually the right layer.

### `internal/ui/app/services`

Side-effect wrappers that return Bubble Tea commands:

- `sandbox.go` starts the Falken session and sandbox
- `agent.go` starts a run and waits for streamed session events
- `diff.go` generates and applies diffs through the session
- `git.go` checks whether the repo is dirty and exposes stash and WIP commit actions
- `config.go` persists the permissions config

This layer keeps the TUI models from calling session methods directly in most cases.

### `internal/ui/tui`

The root Bubble Tea model that coordinates all screens and modals.

Key responsibilities:

- Build child screen models
- Keep route and modal state synchronized with those screens
- Fan workflow messages into the active screen
- Trigger startup commands when route changes require work
- Listen for permission and plan requests through the bridge

The central files are:

- `model.go` for root model construction and startup initialization
- `update.go` for workflow transitions and message routing
- `view.go` for selecting the active route or modal view
- `runtime_bridge.go` for the bridge between `falken.Session` callbacks and Bubble Tea messages

### `internal/ui/tui/screens`

One package per primary screen:

- `chat` for prompt entry, logs, slash commands, runtime output, and todos
- `gitpreflight` for dirty-worktree protection before the sandbox boots
- `initwizard` for first-run cache and allowlist setup
- `pluginapproval` for AOT approval of external plugins
- `permissionoverview` for summarizing the effective security posture
- `diffreview` for browsing and applying submitted diffs

Each screen owns its own `Model`, `Update`, and `View` behavior.

### `internal/ui/tui/modals`

Transient overlays shown while the current route stays active:

- `permissionprompt` for file, shell, and network approvals
- `planapproval` for approving or rejecting an implementation plan
- `tooldetails` for inspecting the last tool name, arguments, and result

## Event And Interaction Bridge

`falken.Session` reports back to the UI through `tui.SessionBridge`.

The bridge exposes:

- An event channel for streaming `falken.Event` values
- A permission request channel that blocks until the UI responds
- A plan request channel that blocks until the UI responds

This lets the session remain UI-agnostic while `falken-term` decides how to present interactive requests.

The `tui.Model` bootstraps two long-running Bubble Tea commands:

- `waitPermissionRequest()`
- `waitPlanRequest()`

Those commands block on the bridge channels and convert interactions into app messages that open the appropriate modal.

## Chat Screen Internals

The `chat` package is the most stateful part of the UI.

It switches between three states:

- `StatePrompt` for the initial multi-line request box
- `StateRunning` while the agent is active
- `StateDone` after a run finishes and the single-line input is available

The layout is computed dynamically:

- Left pane: logs and assistant output
- Optional stream pane: live command chunks while shell output is streaming
- Right pane: todo state loaded from the session todos file
- Bottom bar: input and contextual help

Session events update the screen incrementally:

- thoughts append internal progress indicators
- assistant text appends visible output
- tool calls and results update the "last tool" modal content
- command chunks fill the live stream viewport
- work submission sends the app to diff review
- run completion or failure moves the screen to `StateDone`

Slash commands are implemented inside `chat/commands.go`, not globally at the root model.

## Startup Safety Model

The startup experience is intentionally defensive:

- Dirty git state is detected before sandbox boot
- Initial cache mounts and recommended network domains are captured early
- External plugins require approval before they are used
- The permission overview can remind users what the current allow and block posture looks like

This means startup behavior is split between route-selection helpers in `internal/ui/app` and screen-specific UI in `internal/ui/tui/screens`.

## Diff Review Model

The diff review screen is a boundary between sandbox work and host workspace mutation.

The flow is:

1. Agent signals `EventTypeWorkSubmitted`
2. Root TUI navigates to `RouteReviewDiff`
3. `services.GenerateDiffCmd` asks the session for a diff
4. `diffreview.Model` parses file names from the unified diff
5. User applies or discards the submission
6. `services.ApplyDiffCmd` asks `falken-core` to apply the diff back to the host workspace

Partial applies are surfaced back in chat as warnings with skipped file names.

## Recommended Editing Entry Points

If you need to make a targeted change, start here:

- New startup step: `internal/ui/app/update.go`, `internal/ui/tui/model.go`, and the relevant `screens/*` package
- New modal: `internal/ui/app/routes.go`, `internal/ui/tui/update.go`, and a new package under `internal/ui/tui/modals`
- New session event rendering: `internal/ui/tui/screens/chat/update.go`
- New slash command: `internal/ui/tui/screens/chat/commands.go`
- Different permission UX: `internal/ui/tui/modals/permissionprompt/model.go`
- Different diff flow: `internal/ui/app/messages.go`, `internal/ui/app/services/diff.go`, and `internal/ui/tui/screens/diffreview`

## Testing Strategy

The existing tests lean toward workflow correctness rather than snapshot rendering:

- `internal/ui/app/update_test.go` verifies route selection and workflow helpers
- `internal/ui/tui/main_view_test.go` verifies route and modal coordination
- `internal/ui/tui/screens/gitpreflight/model_test.go` verifies injected git actions and state transitions
- `internal/ui/app/services/diff_test.go` verifies diff-apply result mapping

If you add new startup branches or workflow transitions, add tests beside the relevant model or service instead of relying only on manual terminal testing.
