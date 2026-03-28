---
name: kagi-smallweb
description: Browse Kagi's Small Web feed — recent content from the non-commercial web, crafted by individuals. Free API, no authentication required.
---

# Kagi Small Web

Browse recent content from the "small web" — non-commercial content crafted by individuals to express themselves or share knowledge without seeking financial gain.

**This is a free API endpoint — no API key or subscription required.**

## Usage

```bash
kagi smallweb                           # Latest 20 entries
kagi smallweb -n 5                      # Limit to 5 entries
kagi smallweb --format json             # JSON output
```

### Options

| Flag | Description |
|------|-------------|
| `-n, --num <num>` | Number of entries (default: 20) |
| `--format <fmt>` | Output format: json (default), compact, pretty |
| `--timeout <sec>` | HTTP timeout in seconds (default: 15) |

## Output

### JSON

```json
{
  "items": [
    {
      "title": "Article Title",
      "url": "https://example.com/article",
      "author": "Author Name",
      "published": "2024-01-01T00:00:00+00:00",
      "summary": "Brief description..."
    }
  ],
  "count": 20
}
```

## When to Use

- Discovering independent blogs and personal websites
- Finding non-commercial perspectives on topics
- Browsing curated content from the small web
