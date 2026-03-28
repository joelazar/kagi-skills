Unified `kagi` CLI. Single Go module at `github.com/joelazar/kagi`.

```
cmd/kagi/main.go           # Entrypoint
internal/api/              # HTTP client, auth, errors, balance cache
internal/config/           # YAML config loading
internal/commands/         # Cobra commands (search, fastgpt, summarize, enrich, balance)
internal/output/           # Output formatting (JSON, compact, pretty, markdown, CSV)
internal/version/          # Build-time version injection
kagi-search/SKILL.md       # Skill definitions (reference `kagi <subcommand>`)
kagi-fastgpt/SKILL.md
kagi-summarizer/SKILL.md
kagi-enrich/SKILL.md
```

Tasks are defined in `.mise.toml` and run via `mise run`:

| Task               | Description                                     |
| ------------------ | ----------------------------------------------- |
| `mise run build`   | Build `bin/kagi` binary with version injection  |
| `mise run lint`    | Run golangci-lint on all packages               |
| `mise run test`    | Run `go test ./...`                             |
| `mise run fmt`     | Run `gofumpt -w -l .`                           |
| `mise run clean`   | Remove built binaries                           |
| `mise run install` | `go install` the binary                         |
