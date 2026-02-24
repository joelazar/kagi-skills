Multi-module monorepo. Each skill lives in its own directory with its own `go.mod`:

```
kagi-search/      # Search API — Teclis + content extraction
kagi-fastgpt/     # FastGPT API — AI-synthesized answers
kagi-summarizer/  # Universal Summarizer API
kagi-enrich/      # Enrichment API — Teclis (web) + TinyGem (news)
```

A `Makefile` at the repo root provides common tasks:

| Target       | Description                                     |
| ------------ | ----------------------------------------------- |
| `make build` | Build all four binaries into their `.bin/` dirs |
| `make lint`  | Run golangci-lint on all modules                |
| `make test`  | Run `go test ./...` on all modules              |
| `make fmt`   | Run `gofumpt -w -l .`                           |
| `make clean` | Remove built binaries                           |
