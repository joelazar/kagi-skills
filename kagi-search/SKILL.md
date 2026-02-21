---
name: kagi-search
description: Fast web search and content extraction via Kagi Search API. Uses a Go backend for quick startup and supports JSON output.
---

# Kagi Search

Fast web search and content extraction using the official Kagi Search API.

This skill uses a Go binary for fast startup and no runtime dependencies. The binary can be downloaded pre-built or compiled from source.

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

### Option A — Download pre-built binary (no Go required)

```bash
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)
case "$ARCH" in
  x86_64)        ARCH="amd64" ;;
  aarch64|arm64) ARCH="arm64" ;;
esac

TAG=$(curl -fsSL "https://api.github.com/repos/joelazar/kagi-skills/releases/latest" | grep '"tag_name"' | cut -d'"' -f4)
BINARY="kagi-search_${TAG}_${OS}_${ARCH}"

mkdir -p {baseDir}/.bin
curl -fsSL "https://github.com/joelazar/kagi-skills/releases/download/${TAG}/${BINARY}" \
  -o {baseDir}/.bin/kagi-search
chmod +x {baseDir}/.bin/kagi-search

# Verify checksum (recommended)
curl -fsSL "https://github.com/joelazar/kagi-skills/releases/download/${TAG}/checksums.txt" | \
  grep "${BINARY}" | sha256sum --check
```

Pre-built binaries are available for Linux and macOS (amd64 + arm64) and Windows (amd64).

### Option B — Build from source (requires Go 1.26+)

```bash
cd {baseDir} && go build -o .bin/kagi-search .
```

Alternatively, just run `{baseDir}/kagi-search.sh` directly — the wrapper auto-builds on first run if Go is available.

## Pricing

The Kagi Search API is priced at $25 for 1000 queries (2.5 cents per search).

## Search

```bash
{baseDir}/kagi-search.sh search "query"                              # Basic search (10 results)
{baseDir}/kagi-search.sh search "query" -n 20                        # More results (max 100)
{baseDir}/kagi-search.sh search "query" --content                    # Include extracted page content
{baseDir}/kagi-search.sh search "query" --json                       # JSON output
{baseDir}/kagi-search.sh search "query" -n 5 --content --json        # Combined options
```

### Search options

- `-n <num>` - Number of results (default: 10, max: 100)
- `--content` - Fetch and include page content for each result
- `--json` - Emit JSON output
- `--timeout <sec>` - HTTP timeout in seconds (default: 15)
- `--max-content-chars <num>` - Max chars per fetched result content (default: 5000)

## Extract Page Content

```bash
{baseDir}/kagi-search.sh content https://example.com/article
{baseDir}/kagi-search.sh content https://example.com/article --json
```

### Content options

- `--json` - Emit JSON output
- `--timeout <sec>` - HTTP timeout in seconds (default: 20)
- `--max-chars <num>` - Max chars to output (default: 20000)

## Output

### Default (text)

`kagi-search search` prints readable text blocks, and `kagi-search content` prints extracted content.

### JSON (`--json`)

`kagi-search search --json` returns:

- `query`
- `meta` (includes API metadata like `ms`, `api_balance` when provided)
- `results[]` with `title`, `link`, `snippet`, optional `published`, optional `content`
- `related_searches[]`

`kagi-search content --json` returns:

- `url`
- `title`
- `content`
- `error` (only when extraction fails)

## When to Use

- Searching for documentation or API references
- Looking up facts or current information
- Fetching content from specific URLs
- Any task requiring web search without interactive browsing

## Notes

- Search results inherit your Kagi account settings (personalized results, blocked/promoted sites)
- Results may include related search suggestions (`t:1` objects)
- Content extraction uses `codeberg.org/readeck/go-readability/v2` (Readability v2)
- The binary lives at `{baseDir}/.bin/kagi-search`; the wrapper rebuilds it automatically when source changes (requires Go)
