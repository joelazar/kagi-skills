---
name: kagi-news
description: Browse Kagi's curated news feed with category filtering. Requires Kagi subscriber session token.
---

# Kagi News

Browse news from Kagi's curated news feed with category filtering.

Requires `KAGI_SESSION_TOKEN` environment variable.

## Setup

1. Log into your Kagi account
2. Get your session token
3. Add to your shell profile:
   ```bash
   export KAGI_SESSION_TOKEN="your-session-token"
   ```

## Usage

```bash
kagi news                                # All news
kagi news --category technology          # Filter by category
kagi news --category science -n 5        # Limit results
kagi news --format json                  # JSON output
```

### Options

| Flag | Description |
|------|-------------|
| `--category <cat>` | Filter: world, business, technology, science, health, sports, entertainment |
| `-n, --num <num>` | Number of items (default: 20) |
| `--format <fmt>` | Output format: json (default), compact, pretty |
| `--timeout <sec>` | HTTP timeout in seconds (default: 15) |

## When to Use

- Browsing curated news from Kagi
- Getting technology or science news
- Category-filtered news browsing from the terminal
