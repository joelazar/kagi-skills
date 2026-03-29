# Interactive Mode

This project's Bubble Tea interface is intentionally **not** a 1:1 flag mirror of the raw CLI.
It aims for strong command coverage while keeping the interaction model compact and practical.

## Command Coverage

Interactive mode currently exposes these workflows:

- `search`
- `search content`
- `fastgpt`
- `summarize`
- `enrich`
- `quick`
- `translate`
- `news`
- `smallweb`
- `askpage`
- `assistant`
- `assistant threads`
- `assistant delete thread`
- `balance`
- `auth`
- `config`
- `completion`
- `version`

## Option Parity Philosophy

Interactive mode exposes options that materially change the task outcome:

- `search`: query, result count, optional content fetching
- `summarize`: URL vs text input, engine, summary type, output language
- `enrich`: index choice (`web` vs `news`), result count
- `translate`: source language, target language, formality
- `news`: category, language, item count
- `smallweb`: item count
- `assistant`: prompt plus optional thread continuation
- `assistant threads`: browsing plus thread-detail drill-down

These remain intentionally CLI-only because they are either low-level, presentation-only, or a poor TUI fit:

- `--format`
- `--show-balance`
- `--timeout`
- `--no-cache`
- `--no-refs`
- raw shell-completion script output

## Intentional Gaps

### `batch`

`batch` remains **CLI-only**.

Reason: it is fundamentally a script-friendly, concurrency-focused command whose value comes from raw argument lists and machine-readable output more than from an interactive picker.

### `completion`

The TUI explains how to install completions, but actual completion script generation still belongs to the CLI:

```bash
kagi completion bash
kagi completion zsh
kagi completion fish
kagi completion powershell
```

### `config`

The TUI can inspect the config path and write merged updates, but it does **not** currently provide a destructive "clear this field" flow.
Blank inputs preserve the saved values. Full file editing is still an external-editor / CLI task.

### `auth`

The TUI validates the API key and reports whether a session token is configured.
It does not try to turn every subscriber-token workflow into a separate auth probe.
Instead, live subscriber commands (`assistant`, `quick`, `askpage`, `translate`) remain the real end-to-end validation path.

## Utility Command Stories

Every top-level command now has an interactive-mode story:

- `auth` → status screen with API-key validation and session-token presence
- `config` → inspect path and update saved defaults/secrets
- `version` → build-info screen
- `completion` → setup instructions for shell completions
- `batch` → documented CLI-only workflow

## Styling Notes

The CLI help output and Bubble Tea interface share a Kagi-inspired palette tuned for terminal contrast:

- primary purple: `#6F5AF2` / `#8B7CFF`
- supporting teal: `#1AAE9F` / `#39D0BF`
- warm gold accents: `#B7791F` / `#F6C453`
- panel/background surfaces: `#F7F4FF` / `#1D1830`

The palette is adaptive for light and dark terminals and is applied to:

- Cobra/Fang help output
- TUI titles
- selected items
- URL rendering
- status/help footer text
- borders and error states
