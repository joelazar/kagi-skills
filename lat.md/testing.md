# Testing Strategy

Tests protect command behavior, API parsing, TUI state transitions, and embedded data integrity.

## Package-level unit coverage

Most packages keep tests beside the code they exercise so small contracts can evolve independently.

`internal/api/*_test.go` checks error parsing and remote-fetch safety rules. `internal/commands/*_test.go` covers command validation and output shaping, while `internal/config`, `internal/output`, `internal/version`, and `data` verify their narrower package contracts.

## Unit-test techniques

Low-level behavior is usually verified with table-driven inputs, mock responses, and focused parser assertions.

HTTP-heavy code generally uses `httptest` servers and fixture-like payloads instead of real network calls. Formatter and command tests prefer stable serialized output so regressions show up as clear textual diffs.

## Interactive TUI state tests

The Bubble Tea interface is verified as a state machine instead of relying only on manual terminal testing.

`internal/tui/app_test.go` and `internal/tui/executor_test.go` drive key presses, command-result messages, layout modes, and executor flows so menu, input, results, detail, and error transitions stay predictable.

## Integration tests

Integration tests exercise the root command against a mock HTTP server instead of the live Kagi service.

`internal/commands/integration_test.go` runs behind the `integration` build tag and overrides endpoint environment variables to test search, FastGPT, summarize, enrich, translate, config, completion, and version flows without real network dependencies.
