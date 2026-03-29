---
name: kagi-fastgpt
description: Ask questions and get AI-synthesized answers backed by live web search, via Kagi's FastGPT API. Returns a direct answer with cited references. Use when you need a quick, authoritative answer rather than raw search results.
---

# Kagi FastGPT

Get AI-generated answers with cited web sources using [Kagi's FastGPT API](https://help.kagi.com/kagi/api/fastgpt.html). FastGPT runs a full web search under the hood and synthesizes results into a concise answer — ideal for factual questions, API lookups, and current-events queries.

This skill uses the unified `kagi` CLI binary for fast startup and no runtime dependencies.

## Setup

Requires a Kagi account with API access enabled. Uses the same `KAGI_API_KEY` as all other kagi-* skills.

1. Create an account at https://kagi.com/signup
2. Navigate to Settings → Advanced → API portal: https://kagi.com/settings/api
3. Generate an API Token
4. Add funds at: https://kagi.com/settings/billing_api
5. Add to your shell profile (`~/.profile` or `~/.zprofile`):
   ```bash
   export KAGI_API_KEY="your-api-key-here"
   ```
6. Install the binary — see [Installation](#installation) below

## Pricing

1.5¢ per query ($15 USD per 1000 queries). Cached responses are free.

## Usage

```bash
kagi fastgpt "query"                        # Ask a question (default)
kagi fastgpt "query" --format json           # JSON output
kagi fastgpt "query" --no-refs               # Answer only, no references
kagi fastgpt "query" --no-cache              # Bypass response cache
kagi fastgpt "query" --show-balance          # Show API balance for this call
kagi balance                                 # Show last cached API balance
kagi fastgpt "query" --timeout 60            # Custom timeout (default: 30s)
```

### Options

| Flag | Description |
|------|-------------|
| `--format <fmt>` | Output format: json (default), compact, pretty |
| `--no-refs` | Suppress references in text output |
| `--no-cache` | Bypass cached responses (use for time-sensitive queries) |
| `--show-balance` | Print API balance to stderr for this call |
| `--timeout <sec>` | HTTP timeout in seconds (default: 30) |

## Output

### Default (JSON)

Returns a JSON object with:

- `query` — the original query
- `output` — the synthesized answer
- `tokens` — tokens consumed
- `references[]` — array of `{ title, url, snippet }` objects
- `meta` — API metadata (`id`, `node`, `ms`)

### Pretty text (`--format pretty`)

Prints the synthesized answer, followed by a numbered reference list:

```
Python 3.11 was released on October 24, 2022 and introduced several improvements...

--- References ---
[1] What's New In Python 3.11 — Python 3.11.3 documentation
    https://docs.python.org/3/whatsnew/3.11.html
    The headline changes in Python 3.11 include significant performance improvements...
```

Token usage is printed to stderr. API balance is only shown when you pass `--show-balance`.

## When to Use

- **Use kagi fastgpt** when you need a direct answer synthesized from web sources (e.g. "What version of X was released last month?", "How do I configure Y?")
- **Use kagi search** when you need raw search results to scan, compare, or extract data from yourself
- **Use web-browser** when you need to interact with a page or the content is behind JavaScript

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

The binary has no external dependencies beyond the Go standard library and cobra.
