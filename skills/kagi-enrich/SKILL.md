---
name: kagi-enrich
description: Search Kagi's unique non-commercial web (Teclis) and non-mainstream news (TinyGem) indexes for independent, ad-free content you won't find in regular search results. Use when you want to discover small-web sites, independent blogs, niche discussions, or non-mainstream news on a topic.
---

# Kagi Enrichment

Search Kagi's proprietary enrichment indexes using the [Kagi Enrichment API](https://help.kagi.com/kagi/api/enrich.html). These are Kagi's "secret sauce" — curated indexes of non-commercial and independent content that complement mainstream search results.

Two indexes are available:

| Index | Backend | Best for |
|-------|---------|----------|
| `web` | **Teclis** | Independent websites, personal blogs, open-source projects, non-commercial content |
| `news` | **TinyGem** | Non-mainstream news sources, interesting discussions, off-the-beaten-path journalism |

This skill uses the unified `kagi` CLI binary for fast startup and zero runtime dependencies.

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

**$2 per 1,000 searches** ($0.002 per query). Billed only when non-zero results are returned.

## Usage

```bash
# Search the independent web (Teclis index)
kagi enrich web "rust async programming"

# Search non-mainstream news (TinyGem index)
kagi enrich news "open source AI"

# Limit number of results
kagi enrich web "sqlite internals" -n 5

# JSON output
kagi enrich web "zig programming language" --format json
kagi enrich news "climate change solutions" --format json

# Show balance only when needed
kagi enrich web "query" --show-balance
kagi balance

# Custom timeout
kagi enrich web "query" --timeout 30
```

### Options

| Flag | Description |
|------|-------------|
| `-n, --num <num>` | Max results to display (default: all returned) |
| `--format <fmt>` | Output format: json (default), compact, pretty |
| `--show-balance` | Print API balance to stderr for this call |
| `--timeout <sec>` | HTTP timeout in seconds (default: 15) |

## Output

### Default (JSON)

```json
{
  "query": "sqlite internals",
  "index": "web",
  "meta": { "id": "abc123", "node": "us-east4", "ms": 386, "api_balance": 9.998 },
  "results": [
    {
      "rank": 1,
      "title": "SQLite Internals: How The World's Most Used Database Works",
      "url": "https://www.compileralchemy.com/books/sqlite-internals/",
      "snippet": "A deep-dive into SQLite's B-tree...",
      "published": "2023-04-01T00:00:00Z"
    }
  ]
}
```

### Pretty text (`--format pretty`)

Prints readable text blocks with labeled fields.

## When to Use

- **Use `web`** when you want independent, non-commercial perspectives — personal blogs, indie projects, academic pages
- **Use `news`** when you want news from sources outside the mainstream media cycle
- **Combine with `kagi search`** for the most complete picture
- **Use `kagi fastgpt`** instead when you need a synthesized answer rather than a list of sources

### Note on result counts

The enrichment indexes are intentionally niche — they may return fewer results than general search. No results for a query means no relevant content was found in that index (and you won't be billed).

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
