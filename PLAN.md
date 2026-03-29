# Kagi CLI — Consolidation & Rewrite Plan

## Overview

Rewrite the multi-binary `kagi-skills` repo into a single `kagi` CLI binary with all Kagi API features, an interactive TUI mode, and a solid Go toolchain foundation.

- **Module**: `github.com/joelazar/kagi`
- **Binary**: `kagi`
- **Go version**: 1.26
- **License**: MIT (unchanged)

---

## Architecture

### Project Layout

```
.
├── cmd/
│   └── kagi/
│       └── main.go              # Entrypoint
├── internal/
│   ├── api/                     # HTTP client, auth, shared request/response handling
│   │   ├── client.go            # Shared HTTP client with timeouts, retries
│   │   ├── auth.go              # API token + session token resolution
│   │   └── errors.go            # Typed API error handling
│   ├── config/
│   │   └── config.go            # YAML config file (~/.config/kagi/config.yaml)
│   ├── commands/                # One file per command (cobra commands)
│   │   ├── search.go
│   │   ├── fastgpt.go
│   │   ├── summarize.go
│   │   ├── enrich.go
│   │   ├── quick.go
│   │   ├── assistant.go
│   │   ├── askpage.go
│   │   ├── translate.go
│   │   ├── news.go
│   │   ├── smallweb.go
│   │   ├── batch.go
│   │   └── root.go              # Root command, global flags, TUI launcher
│   ├── tui/                     # Bubble Tea TUI components
│   │   ├── app.go               # Main TUI app model
│   │   ├── menu.go              # Command picker menu
│   │   ├── results.go           # Result list with arrow/vim navigation
│   │   ├── detail.go            # Detail view for a selected result
│   │   ├── input.go             # Query input with huh forms
│   │   ├── styles.go            # Lip Gloss theme/styles
│   │   └── keys.go              # Keybindings (arrows + vim hjkl)
│   ├── output/                  # Output formatting
│   │   ├── format.go            # Format dispatcher (json, pretty, markdown, csv, compact)
│   │   ├── json.go
│   │   ├── pretty.go            # Glamour markdown + Lip Gloss styling
│   │   ├── markdown.go
│   │   ├── csv.go
│   │   └── compact.go
│   └── version/
│       └── version.go           # Build-time version injection
├── data/
│   └── news-filter-presets.json # Embedded news filter preset definitions
├── testdata/                    # Golden files, fixtures for tests
├── go.mod
├── go.sum
├── Makefile
├── .golangci.yml
├── .mise.toml
├── .goreleaser.yml              # Milestone 5
├── .github/
│   └── workflows/
│       ├── ci.yml
│       └── release.yml          # Milestone 5
├── skills/
│   ├── kagi-search/SKILL.md     # Updated skill definitions → call `kagi search`
│   ├── kagi-fastgpt/SKILL.md
│   ├── kagi-summarizer/SKILL.md
│   ├── kagi-enrich/SKILL.md
│   ├── kagi-assistant/SKILL.md  # New
│   ├── kagi-translate/SKILL.md  # New
│   ├── kagi-news/SKILL.md       # New
│   ├── kagi-askpage/SKILL.md    # New
│   ├── kagi-quick/SKILL.md      # New
│   └── kagi-smallweb/SKILL.md   # New
├── PLAN.md
├── CHANGELOG.md
├── README.md
└── LICENSE
```

### Key Dependencies

| Package | Purpose |
|---------|---------|
| `github.com/spf13/cobra` | CLI framework with subcommands |
| `github.com/charmbracelet/bubbletea` | TUI framework |
| `github.com/charmbracelet/lipgloss` | Terminal styling |
| `github.com/charmbracelet/glamour` | Markdown rendering in terminal |
| `github.com/charmbracelet/huh` | Interactive forms/prompts |
| `github.com/charmbracelet/bubbles` | TUI components (list, textinput, viewport, etc.) |
| `gopkg.in/yaml.v3` | Config file parsing |
| `codeberg.org/readeck/go-readability/v2` | Content extraction (carried over from kagi-search) |

### Auth Strategy

Priority order for credential resolution:

1. **Environment variables**: `KAGI_API_KEY` (paid API), `KAGI_SESSION_TOKEN` (subscriber features)
2. **Config file**: `~/.config/kagi/config.yaml`
3. **Local config**: `./.kagi.yaml` (project-level override)

```yaml
# ~/.config/kagi/config.yaml
api_key: "your-api-key"
session_token: "https://kagi.com/search?token=..."
defaults:
  format: pretty
  search:
    region: us
```

Commands that need the paid API (`search`, `fastgpt`, `summarize`, `enrich`) use `api_key`.
Commands that need subscriber access (`assistant`, `translate`, `askpage`, `quick`, `news`, `smallweb`, `summarize --subscriber`) use `session_token`.

### Output Formats

All commands support `--format` flag:

| Format | Description |
|--------|-------------|
| `json` | Default. Structured JSON for scripts/agents |
| `pretty` | Colorized terminal output with Lip Gloss styling |
| `compact` | Minified JSON |
| `markdown` | Markdown for documentation |
| `csv` | Tabular output (where applicable) |

### TUI Mode

Running `kagi` with no arguments (or `kagi --interactive`) launches the TUI:

1. **Command picker**: Menu to select a command (search, fastgpt, etc.)
2. **Query input**: `huh` form for the selected command's parameters
3. **Results list**: Navigable with arrow keys and vim bindings (j/k up/down, gg/G top/bottom, / to filter)
4. **Detail view**: Enter on a result shows full details, `o` opens URL in browser, `y` copies URL
5. **Back navigation**: Esc/q goes back, Ctrl+C exits

### Keybindings (TUI)

| Key | Action |
|-----|--------|
| `j` / `↓` | Move down |
| `k` / `↑` | Move up |
| `h` / `←` | Go back |
| `l` / `→` / `Enter` | Select / view detail |
| `g g` | Jump to top |
| `G` | Jump to bottom |
| `/` | Filter/search results |
| `o` | Open URL in browser |
| `y` | Copy URL to clipboard |
| `q` / `Esc` | Back / quit |
| `?` | Show help |

---

## Milestones

### Milestone 1 — Foundation & Core Commands

**Goal**: Single binary with the 4 existing commands, shared infrastructure, proper tooling.

#### Tasks

- [ ] Initialize single `go.mod` at `github.com/joelazar/kagi`
- [ ] Set up `.mise.toml` with Go 1.26, golangci-lint, gofumpt, goreleaser
- [ ] Set up `.golangci.yml` (carry over + adapt existing config)
- [ ] Create `internal/api/client.go` — shared HTTP client with timeouts
- [ ] Create `internal/api/auth.go` — credential resolution (env vars + config file)
- [ ] Create `internal/api/errors.go` — typed error handling
- [ ] Create `internal/config/config.go` — YAML config loading
- [ ] Create `internal/version/version.go` — build-time version
- [ ] Create `internal/output/` — format dispatcher + JSON, pretty, compact, markdown, CSV formatters
- [ ] Implement `cmd/kagi/main.go` + `internal/commands/root.go` (cobra root)
- [ ] Port `kagi-search` → `internal/commands/search.go` with advanced filters (lens, region, time, date, order, verbatim, personalization)
- [ ] Port `kagi-fastgpt` → `internal/commands/fastgpt.go`
- [ ] Port `kagi-summarizer` → `internal/commands/summarize.go` (including subscriber web mode)
- [ ] Port `kagi-enrich` → `internal/commands/enrich.go`
- [ ] Write unit tests for all shared infrastructure (client, auth, config, output, errors)
- [ ] Write unit tests for each command (mock HTTP responses, golden file comparisons)
- [ ] Set up `Makefile` (build, test, lint, fmt, clean)
- [ ] Set up `.github/workflows/ci.yml`
- [ ] Update existing SKILL.md files to call `kagi search`, `kagi fastgpt`, etc.
- [ ] Remove old per-module `go.mod` files and individual `main.go` binaries

### Milestone 2 — New API Commands

**Goal**: Add all remaining Kagi API commands.

#### Tasks

- [ ] Implement `kagi quick` — Quick Answer via subscriber session
- [ ] Implement `kagi assistant` — Chat with Kagi Assistant
- [ ] Implement `kagi assistant thread list/get/delete/export` — Thread management
- [ ] Implement `kagi askpage` — Ask questions about a URL
- [ ] Implement `kagi translate` — Full translation with all options (formality, gender, context, alternatives, word insights, suggestions, alignments)
- [ ] Implement `kagi news` — News feed with categories, chaos index, content filters, filter presets
- [ ] Implement `kagi smallweb` — Small Web feed
- [ ] Implement `kagi batch` — Parallel searches with configurable concurrency + rate limiting
- [ ] Embed `data/news-filter-presets.json` via `go:embed`
- [ ] Write unit tests for each new command (red/green)
- [ ] Create SKILL.md files for new commands: `kagi-assistant/`, `kagi-translate/`, `kagi-news/`, `kagi-askpage/`, `kagi-quick/`, `kagi-smallweb/`

### Milestone 3 — TUI Mode

**Goal**: Interactive terminal UI for command selection and result browsing.

#### Tasks

- [ ] Create `internal/tui/app.go` — Main Bubble Tea model with state machine (menu → input → results → detail)
- [ ] Create `internal/tui/menu.go` — Command picker using `bubbles/list`
- [ ] Create `internal/tui/input.go` — Query input forms using `huh`
- [ ] Create `internal/tui/results.go` — Result list with vim keybindings using `bubbles/list`
- [ ] Create `internal/tui/detail.go` — Detail view with glamour markdown rendering using `bubbles/viewport`
- [ ] Create `internal/tui/keys.go` — Keybinding definitions (arrows + vim)
- [ ] Create `internal/tui/styles.go` — Lip Gloss theme (respect terminal color scheme)
- [ ] Wire TUI launch from `kagi` (no args) and `kagi --interactive`
- [ ] Implement browser open (`o` key) and clipboard copy (`y` key)
- [ ] Write tests for TUI state transitions
- [ ] Add `--no-tui` flag to force non-interactive mode in pipes

### Milestone 4 — Polish & Documentation

**Goal**: Production-ready CLI with full documentation.

#### Tasks

- [ ] Write comprehensive README.md with install instructions, usage examples, screenshots
- [ ] Add shell completion generation (`kagi completion bash/zsh/fish`)
- [ ] Add `kagi config init` command for guided config setup
- [ ] Add `kagi auth check` command to validate credentials
- [ ] Add `kagi version` command
- [ ] Update CHANGELOG.md
- [ ] Add man page generation (cobra-doc)
- [ ] End-to-end integration tests against mock server
- [ ] Performance profiling and optimization

### Milestone 5 — Release & Distribution

**Goal**: Automated releases via goreleaser and homebrew tap.

#### Tasks

- [ ] Set up `.goreleaser.yml` (cross-compile linux/darwin/windows, amd64/arm64)
- [ ] Set up homebrew tap repo (`joelazar/homebrew-tap`)
- [ ] Set up `.github/workflows/release.yml` with goreleaser
- [ ] Tag `v2.0.0` (breaking change from multi-binary to single binary)
- [ ] Update README with homebrew install instructions
- [ ] Archive/deprecate old release artifacts

---

## Testing Strategy

**Red/Green approach throughout all milestones.**

### Unit Tests

- Every package gets `_test.go` files
- Mock HTTP responses using `httptest.NewServer`
- Golden file comparisons for output formatters (`testdata/*.golden`)
- Table-driven tests for input validation, error mapping, auth resolution

### Integration Tests

- Build tag `//go:build integration` for tests that hit real APIs
- Require `KAGI_API_KEY` / `KAGI_SESSION_TOKEN` env vars
- Run separately: `go test -tags integration ./...`
- Not part of CI (needs credentials)

### TUI Tests

- Test state machine transitions (menu → input → results → detail → back)
- Use `bubbletea`'s test helpers for simulating key presses
- Snapshot tests for rendered views

---

## Tooling Setup

### `.mise.toml`

```toml
[tools]
go = "1.26"
golangci-lint = "latest"
gofumpt = "latest"
goreleaser = "latest"
```

### Makefile Targets

| Target | Description |
|--------|-------------|
| `make build` | `go build -o bin/kagi ./cmd/kagi` |
| `make test` | `go test ./...` |
| `make test-integration` | `go test -tags integration ./...` |
| `make lint` | `golangci-lint run` |
| `make fmt` | `gofumpt -w -l .` |
| `make clean` | `rm -rf bin/` |
| `make install` | `go install ./cmd/kagi` |

### CI Pipeline (`.github/workflows/ci.yml`)

1. Format check (gofumpt)
2. Lint (golangci-lint)
3. Build
4. Test (unit only)

---

## Migration Notes

- Old binaries (`kagi-search`, `kagi-fastgpt`, `kagi-summarizer`, `kagi-enrich`) are removed
- SKILL.md files updated: binary calls change from `kagi-search ...` to `kagi search ...`
- Env var `KAGI_API_KEY` stays the same (backward compatible)
- New env var `KAGI_SESSION_TOKEN` for subscriber features
- Config file is optional (env vars always work)
- Version bumps to `v2.0.0` (semver major for breaking changes)
