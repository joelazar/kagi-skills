---
name: kagi-search
description: Fast web search and content extraction via Kagi Search API. Uses a Go backend for quick startup and supports JSON output.
---

# Kagi Search

Fast web search and content extraction using the official Kagi Search API.

This skill uses the unified `kagi` CLI binary for fast startup and no runtime dependencies.

## Setup

Requires a Kagi account with API access enabled.

1. Create an account at https://kagi.com/signup
2. Navigate to Settings -> Advanced -> API portal: https://kagi.com/settings/api
3. Generate an API Token
4. Add funds to your API balance at: https://kagi.com/settings/billing_api
5. Add to your shell profile (`~/.profile` or `~/.zprofile` for zsh):
   ```bash
   export KAGI_API_KEY="your-api-key-here"
   ```
6. Install the binary — see [Installation](#installation) below

## Installation

### Option A — Install from source (requires Go 1.26+)

```bash
cd {baseDir}/../.. && go install -ldflags "-X github.com/joelazar/kagi/internal/version.Version=$(git describe --tags --always)" ./cmd/kagi
```

### Option B — Build locally

```bash
cd {baseDir}/../.. && mise run build
# Binary at {baseDir}/../../bin/kagi
```

## Pricing

The Kagi Search API is priced at $25 for 1000 queries (2.5 cents per search).

## Search

```bash
kagi search "query"                                       # Basic search (10 results)
kagi search "query" -n 20                                 # More results (max 100)
kagi search "query" --content                             # Include extracted page content
kagi search "query" --format json                         # JSON output
kagi search "query" --show-balance                        # Show API balance for this call
kagi search "query" -n 5 --content --format json          # Combined options
```

### Search options

- `-n, --num <num>` - Number of results (default: 10, max: 100)
- `--content` - Fetch and include page content for each result
- `--format <fmt>` - Output format: json (default), compact, pretty, markdown, csv
- `--show-balance` - Print API balance to stderr for this call
- `--timeout <sec>` - HTTP timeout in seconds (default: 15)
- `--max-content-chars <num>` - Max chars per fetched result content (default: 5000)

## Extract Page Content

```bash
kagi search content https://example.com/article
kagi search content https://example.com/article --format json
```

### Content options

- `--format <fmt>` - Output format (default: json)
- `--timeout <sec>` - HTTP timeout in seconds (default: 20)
- `--max-chars <num>` - Max chars to output (default: 20000)

## API Balance

Balance is not printed by default. You can either:

- add `--show-balance` to `search`
- run the dedicated command:

```bash
kagi balance
kagi balance --format json
```

## Output

### Default (JSON)

`kagi search` returns JSON by default:

- `query`
- `meta` (includes API metadata like `ms`, `api_balance` when provided)
- `results[]` with `title`, `link`, `snippet`, optional `published`, optional `content`
- `related_searches[]`

### Pretty text (`--format pretty`)

Prints readable text blocks with labeled fields.

## When to Use

- Searching for documentation or API references
- Looking up facts or current information
- Fetching content from specific URLs
- Any task requiring web search without interactive browsing

## Notes

- Search results inherit your Kagi account settings (personalized results, blocked/promoted sites)
- Results may include related search suggestions
- Content extraction uses `codeberg.org/readeck/go-readability/v2` (Readability v2)
