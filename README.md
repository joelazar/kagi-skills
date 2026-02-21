# kagi-skills

[![CI](https://github.com/joelazar/kagi-skills/actions/workflows/ci.yml/badge.svg)](https://github.com/joelazar/kagi-skills/actions/workflows/ci.yml)
[![Latest Release](https://img.shields.io/github/v/release/joelazar/kagi-skills)](https://github.com/joelazar/kagi-skills/releases/latest)
[![Go Version](https://img.shields.io/badge/go-1.26+-blue)](https://go.dev/dl/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

CLI tools that give any AI agent access to [Kagi's](https://kagi.com) search and AI APIs: web search, AI-synthesized answers, page summarization, and independent-web enrichment.

Works with any agent that can call shell commands: [pi](https://github.com/mariozechner/pi), [Claude Code](https://github.com/anthropics/claude-code), [Gemini CLI](https://github.com/google-gemini/gemini-cli), [Codex](https://github.com/openai/codex), [Cursor](https://cursor.com), or your own custom agent.

## Why Kagi?

Kagi is a **user-funded search engine** with no ads, no tracking, and no data selling. Because its revenue comes directly from users rather than advertisers, results are optimized for relevance and quality instead of engagement and ad clicks. That alignment matters a lot when you're feeding search results into an AI agent.

Kagi's own research shows that **AI models perform up to 80% better when sourcing data through Kagi Search** compared to ad-supported engines ([source](https://blog.kagi.com/kagi-ai-search), [blog](https://blog.kagi.com/blog)). The results are cleaner, less polluted by SEO spam, and drawn from more authoritative sources. Kagi also runs its own non-commercial crawl indexes: Teclis surfaces high-quality content from smaller, independent sites that commercial indexes deprioritize, and TinyGem covers alternative and independent news sources. On top of that, Kagi offers a built-in summarizer API that works on any URL, PDF, or raw text, and uses simple pay-per-use pricing with no quota tiers or subscription traps.

## Tools

| Tool                | What it does                                               | Kagi API                                                     | Price                                        |
| ------------------- | ---------------------------------------------------------- | ------------------------------------------------------------ | -------------------------------------------- |
| **kagi-search**     | Web search + page content extraction                       | [Search](https://help.kagi.com/kagi/api/search.html)         | $0.025 / query                               |
| **kagi-fastgpt**    | AI answer synthesized from live web search                 | [FastGPT](https://help.kagi.com/kagi/api/fastgpt.html)       | $0.015 / query                               |
| **kagi-summarizer** | Summarize any URL, PDF, or text block                      | [Summarizer](https://help.kagi.com/kagi/api/summarizer.html) | $0.030 / 1k tokens ($0.025 on Ultimate plan) |
| **kagi-enrich**     | Search the independent web (Teclis) and alt-news (TinyGem) | [Enrichment](https://help.kagi.com/kagi/api/enrich.html)     | $0.002 / query                               |

## Get an API Key

1. Create a Kagi account at <https://kagi.com/signup>
2. Go to **Settings → API** → <https://kagi.com/settings?p=api>
3. Generate a token and add credit
4. Export the key in your shell profile:

```bash
# bash/zsh
export KAGI_API_KEY="your-api-key-here"

# fish
set -Ux KAGI_API_KEY "your-api-key-here"
```

One key works for all four tools.

## Quick Start

1. Set your API key (see [Get an API Key](#get-an-api-key) above)
2. Symlink the skill(s) into your agent's skills directory (see [Agent Integration](#agent-integration) below)
3. On first run, the wrapper auto-builds from source if Go is available. Otherwise it prompts you to confirm downloading a pre-built binary from GitHub releases

Each tool's full usage is documented in its `SKILL.md`.

## Agent Integration

Each skill folder contains a `SKILL.md` that agents read to understand how to invoke the tool. Symlink or copy the skill folder into your agent's skills directory, then install the binary inside it.

### [Pi](https://github.com/mariozechner/pi) / [Gemini CLI](https://github.com/google-gemini/gemini-cli) / [Codex](https://github.com/openai/codex) / other agents

Default skills directory: `~/.agents/skills/`

```bash
# bash/zsh
for skill in kagi-*; do ln -s "$(pwd)/$skill" ~/.agents/skills/"$skill"; done

# fish
for skill in kagi-*; ln -s (pwd)/$skill ~/.agents/skills/$skill; end
```

### [Claude Code](https://github.com/anthropics/claude-code)

Default skills directory: `~/.claude/skills/`

```bash
# bash/zsh
for skill in kagi-*; do ln -s "$(pwd)/$skill" ~/.claude/skills/"$skill"; done

# fish
for skill in kagi-*; ln -s (pwd)/$skill ~/.claude/skills/$skill; end
```

The binaries speak plain text and JSON (`--json` flag). No special integration is needed beyond dropping the skill folder in the right place.
