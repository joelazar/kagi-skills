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

A `Makefile` at the repo root provides common tasks:

| Target       | Description                                     |
| ------------ | ----------------------------------------------- |
| `make build` | Build `bin/kagi` binary with version injection  |
| `make lint`  | Run golangci-lint on all packages               |
| `make test`  | Run `go test ./...`                             |
| `make fmt`   | Run `gofumpt -w -l .`                           |
| `make clean` | Remove built binaries                           |
| `make install` | `go install` the binary                       |
