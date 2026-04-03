# Architecture

The CLI is organized as thin Cobra command handlers over shared config, HTTP, formatting, and TUI layers.

## Executable and packaging model

The product is one `kagi` binary with agent-facing skill wrappers that all delegate to the same command tree.

The executable is built from `cmd/kagi/`, and user-facing capabilities are exposed as subcommands under [[cli#CLI Commands]]. The `skills/` folders package installation guidance and agent prompts around that shared binary rather than introducing alternate runtimes or duplicate backends.

## Repository layout

The source tree separates entrypoint wiring, shared infrastructure, command handlers, interactive UI code, and supporting assets into stable layers.

- `cmd/kagi/` owns process startup.
- `internal/api/` holds HTTP clients, auth resolution, and shared API error handling.
- `internal/config/` loads YAML defaults and saved credentials.
- `internal/commands/` defines the Cobra command tree.
- `internal/tui/` renders the Bubble Tea interface that reuses the CLI.
- `internal/output/` owns shared format parsing and JSON writers.
- `internal/version/` reports build metadata.
- `data/` stores embedded assets such as news filter presets.
- `skills/` contains agent-facing wrappers and setup docs.
- `testdata/` holds reusable fixtures and golden-style assets when needed.

## Technology choices

The project favors a small Go-native stack tuned for CLI ergonomics rather than custom runtime layers.

Fang and Cobra provide command assembly, Bubble Tea and related Charm libraries power the TUI, YAML parsing uses `gopkg.in/yaml.v3`, and readable extraction comes from `codeberg.org/readeck/go-readability/v2`. This keeps the executable self-contained while still supporting both agent automation and human terminal workflows.

## Process entry and command assembly

Startup wires version metadata, config loading, shared flags, and an optional interactive TUI around one root command.

`cmd/kagi/main.go` executes the tree from [[internal/commands/root.go#NewRootCmd]]. The root command loads config in `PersistentPreRunE`, registers shared flags such as `--format`, and launches the Bubble Tea interface when no explicit subcommand is given in an interactive terminal.

## Configuration and credential precedence

Configuration is optional, but when present it gives the CLI stable defaults and fallback credentials.

[[internal/config/config.go#Load]] prefers `./.kagi.yaml` over `~/.config/kagi/config.yaml`. [[internal/api/auth.go#ResolveAPIKey]] and [[internal/api/auth.go#ResolveSessionToken]] prefer environment variables over file settings so CI and ephemeral shells can override saved values safely.

## HTTP clients and safety boundaries

Network code separates trusted API traffic from untrusted page fetching to reduce SSRF and local-network exposure.

[[internal/api/client.go#NewHTTPClient]] clones the default transport for normal API calls. [[internal/api/client.go#NewSafeContentClient]], [[internal/api/client.go#ValidateRemoteFetchURL]], and [[internal/api/client.go#IsBlockedIP]] block private or local targets before `search --content` and `search content` fetch remote pages.

## Output contract

Commands favor machine-readable output first so both shell users and the TUI can reuse the same behavior.

[[internal/output/format.go#ParseFormat]] validates the shared format flag. JSON and compact JSON are the stable interchange layer, while human-oriented commands may also emit pretty text, markdown, or csv when that command supports it.

## Interactive terminal model

The no-argument interactive path is a state machine over menu selection, form input, loading, results, detail, and error recovery.

[[internal/tui/menu.go#MenuCommands]] defines the interactive command surface, and [[internal/tui/keys.go#DefaultKeyMap]] defines the default shortcut model. The major states are menu, input, loading, results, detail, and error.

## Interactive TUI reuse

The TUI is a presentation layer over the CLI, not a second API client implementation.

[[internal/tui/app.go#RunWithOptions]] owns navigation state and rendering. [[internal/tui/executor.go#NewExecutor]] shells back into the current `kagi` binary with `--no-tui --format json`, so interactive flows reuse the same command handlers, auth rules, and JSON contracts as non-interactive execution.
