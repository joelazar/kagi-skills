---
name: kagi-assistant
description: Chat with Kagi Assistant for AI-powered answers. Supports conversation threads. Requires Kagi subscriber session token.
---

# Kagi Assistant

Chat with Kagi's AI Assistant with support for conversation threads.

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
# Ask a question
kagi assistant "Explain quantum computing"

# Continue a conversation
kagi assistant "Tell me more" --thread <thread-id>

# JSON output
kagi assistant "Compare Go and Rust" --format json
```

### Thread Management

```bash
kagi assistant thread list                # List all threads
kagi assistant thread get <id>            # View a thread
kagi assistant thread delete <id>         # Delete a thread
```

### Options

| Flag | Description |
|------|-------------|
| `--thread <id>` | Continue an existing conversation thread |
| `--format <fmt>` | Output format: json (default), compact, pretty |
| `--timeout <sec>` | HTTP timeout in seconds (default: 60) |

## When to Use

- Extended conversations with AI that maintain context
- Complex questions requiring back-and-forth dialogue
- When you want to reference previous conversation history
