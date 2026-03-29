---
name: kagi-news
description: Browse Kagi's curated news feed with category filtering. Uses the public news.kagi.com API.
---

# Kagi News

Browse news from Kagi's curated news feed with category filtering.

Uses the public `news.kagi.com` API — no authentication required.

## Usage

```bash
kagi news                                # All news
kagi news --category technology          # Filter by category
kagi news --category science -n 5        # Limit results
kagi news --lang de                      # German news
kagi news --format json                  # JSON output
```

### Options

| Flag | Description |
|------|-------------|
| `--category <cat>` | Filter: world, business, technology, science, health, sports, entertainment |
| `-n, --num <num>` | Number of items (default: 20) |
| `--lang <code>` | Language code, e.g., en, de, fr (default: en) |
| `--format <fmt>` | Output format: json (default), compact, pretty |
| `--timeout <sec>` | HTTP timeout in seconds (default: 15) |

## When to Use

- Browsing curated news from Kagi
- Getting technology or science news
- Category-filtered news browsing from the terminal
