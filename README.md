# kagi-skills

[![CI](https://github.com/joelazar/kagi-skills/actions/workflows/ci.yml/badge.svg)](https://github.com/joelazar/kagi-skills/actions/workflows/ci.yml)
[![Latest Release](https://img.shields.io/github/v/release/joelazar/kagi-skills)](https://github.com/joelazar/kagi-skills/releases/latest)
[![Go Version](https://img.shields.io/badge/go-1.26+-blue)](https://go.dev/dl/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

CLI tools that give any AI agent access to [Kagi's](https://kagi.com) search and AI APIs — web search, AI-synthesized answers, page summarization, and independent-web enrichment.

Works with any agent that can call shell commands: [pi](https://github.com/mariozechner/pi), [Claude Code](https://github.com/anthropics/claude-code), [Gemini CLI](https://github.com/google-gemini/gemini-cli), [Codex](https://github.com/openai/codex), [Cursor](https://cursor.com), or your own custom agent.

## Why Kagi?

| | Kagi | Google / Bing / Brave / Tavily |
|---|---|---|
| **Privacy** | No ads, no tracking, no data selling | Ad-driven, user profiled |
| **Result quality** | Curated, SEO-spam filtered | Algorithm subject to SEO gaming |
| **Unique indexes** | Teclis (indie web) + TinyGem (alt news) | Not available |
| **Summarizer API** | Built-in, URL + text | Not available |
| **Pricing** | Pay-per-use, no surprise rate limits | Quota tiers, subscription traps |

## Tools

| Tool | What it does | Price |
|---|---|---|
| **kagi-search** | Web search + page content extraction | $0.025 / query |
| **kagi-fastgpt** | AI answer synthesized from live web search | $0.015 / query |
| **kagi-summarizer** | Summarize any URL, PDF, or text block | $1 / 1k tokens |
| **kagi-enrich** | Search the independent web (Teclis) and alt-news (TinyGem) | $0.002 / query |

## Get an API Key

1. Create a Kagi account at <https://kagi.com/signup>
2. Go to **Settings → API** → <https://kagi.com/settings?p=api>
3. Generate a token and add credit
4. Export the key in your shell profile:

```bash
export KAGI_API_KEY="your-api-key-here"
```

One key works for all four tools.

## Quick Start

1. Set your API key (see [Get an API Key](#get-an-api-key) above)
2. Symlink the skill(s) into your agent's skills directory (see [Agent Integration](#agent-integration) below)
3. On first run, the wrapper auto-builds from source if Go is available — otherwise it prompts you to confirm downloading a pre-built binary from GitHub releases

Each tool's full usage is documented in its `SKILL.md`.

## Agent Integration

Each skill folder contains a `SKILL.md` that agents read to understand how to invoke the tool. Symlink or copy the skill folder into your agent's skills directory, then install the binary inside it.

### [Pi](https://github.com/mariozechner/pi) / [Gemini CLI](https://github.com/google-gemini/gemini-cli) / [Codex](https://github.com/openai/codex) / other agents

Default skills directory: `~/.agents/skills/`

```bash
ln -s $(pwd)/kagi-search ~/.agents/skills/kagi-search
```

### [Claude Code](https://github.com/anthropics/claude-code)

Default skills directory: `~/.claude/skills/`

```bash
ln -s $(pwd)/kagi-search ~/.claude/skills/kagi-search
```

The binaries speak plain text and JSON (`--json` flag) — no special integration beyond dropping the skill folder in the right place.

