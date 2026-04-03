# CLI Commands

The CLI exposes a unified surface for search, synthesis, feeds, and local setup workflows.

## Root behavior

The root command decides whether the user gets a direct subcommand run or the interactive terminal UI.

Running `kagi` with no subcommand launches the TUI when stdout is a terminal. `--no-tui` disables that fallback, and shared flags such as `--format`, `--alt-screen`, and `--compact` shape cross-command behavior.

## Interactive terminal mode

The interactive UI is the operator-facing way to browse the same command surface without typing full subcommands.

[[internal/tui/menu.go#MenuCommands]] exposes search, synthesis, feed, and setup workflows through a menu-driven interface. The implementation details of state transitions and CLI reuse are documented in [[architecture#Interactive terminal model]] and [[architecture#Interactive TUI reuse]].

### Keybindings

The default key map favors lightweight navigation and browsing shortcuts over shell-style flags.

[[internal/tui/keys.go#DefaultKeyMap]] supports arrows plus `j/k` for movement, `h/l` for back and detail navigation, `o` to open URLs, `y` to copy URLs, `/` to filter lists, `?` to toggle help, and `esc`, `backspace`, `q`, or `ctrl+c` for back or quit depending on context.

## Search and retrieval

These commands fetch source material from the web and can optionally extract readable page content.

`search` calls Kagi Search, `search content` extracts readable text from one URL, and `batch` runs multiple searches with concurrency and rate controls. The page-fetching paths rely on the safety rules described in [[architecture#HTTP clients and safety boundaries]].

## Synthesis and question answering

These commands turn search results or page content into direct answers.

`fastgpt` returns live-search-grounded answers through the paid API. `summarize` accepts either a URL or raw text, while `quick`, `askpage`, and `assistant` use subscriber features for shorter answers, page-specific questions, and threaded conversations.

## Discovery feeds and enrichment

These commands surface content collections instead of a general web-search result page.

`enrich web` and `enrich news` query Kagi's independent-web indexes, while `news` and `smallweb` expose curated feeds aimed at discovery rather than arbitrary search.

## Local operator workflows

These commands help users configure, inspect, and integrate the binary into their shell environment.

`auth check` validates saved credentials, `config init` and `config path` manage the config file location, `completion` prints shell completions, `balance` shows the last cached API balance, and `version` reports build metadata.
