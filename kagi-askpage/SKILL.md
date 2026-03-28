---
name: kagi-askpage
description: Ask questions about a specific URL using Kagi's AI. Requires Kagi subscriber session token.
---

# Kagi Ask Page

Ask questions about the content of a specific URL using Kagi's AI.

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
kagi askpage https://golang.org/doc/go1.22 "What are the new features?"
kagi askpage https://arxiv.org/abs/1706.03762 "Summarize the main contribution" --format json
```

### Options

| Flag | Description |
|------|-------------|
| `--format <fmt>` | Output format: json (default), compact, pretty |
| `--timeout <sec>` | HTTP timeout in seconds (default: 60) |

## When to Use

- Asking specific questions about a web page's content
- Getting AI analysis of an article, paper, or documentation page
- When you need targeted answers rather than a general summary
