---
name: kagi-quick
description: Get quick AI-generated answers using your Kagi subscriber session. Similar to FastGPT but uses subscriber credits instead of API credits.
---

# Kagi Quick Answer

Get AI-generated quick answers using your Kagi subscriber session.

Requires `KAGI_SESSION_TOKEN` environment variable.

## Setup

1. Log into your Kagi account
2. Get your session token from your browser cookies or token URL
3. Add to your shell profile:
   ```bash
   export KAGI_SESSION_TOKEN="your-session-token"
   ```

## Usage

```bash
kagi quick "What is the population of Tokyo?"
kagi quick "How to reverse a string in Go?" --format json
```

### Options

| Flag | Description |
|------|-------------|
| `--format <fmt>` | Output format: json (default), compact, pretty |
| `--timeout <sec>` | HTTP timeout in seconds (default: 30) |

## When to Use

- Quick factual questions using subscriber session (no API credits)
- Alternative to `kagi fastgpt` when you want to use subscriber access
