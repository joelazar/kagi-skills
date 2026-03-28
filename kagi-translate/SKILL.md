---
name: kagi-translate
description: Translate text using Kagi's translation service with language detection, formality control, and alternatives. Requires Kagi subscriber session token.
---

# Kagi Translate

Translate text between languages using Kagi's translation service.

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
# Basic translation
kagi translate --to DE "Hello, world!"

# Specify source language
kagi translate --from EN --to JA "Good morning"

# Pipe from stdin
echo "Bonjour le monde" | kagi translate --to EN

# With formality
kagi translate --to ES --formality formal "How are you?"

# JSON output
kagi translate --to DE "Hello" --format json
```

### Options

| Flag | Description |
|------|-------------|
| `--to <code>` | Target language code (required) |
| `--from <code>` | Source language code (auto-detect if omitted) |
| `--formality <level>` | Formality: formal, informal |
| `--format <fmt>` | Output format: json (default), compact, pretty |
| `--timeout <sec>` | HTTP timeout in seconds (default: 30) |

### Language Codes

Common: EN, DE, FR, ES, IT, PT, JA, KO, ZH, RU, AR

## When to Use

- Translating text between languages
- When formality level matters (formal letters, casual chat)
- Quick translations from the command line
