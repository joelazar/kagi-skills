---
name: kagi-summarizer
description: Summarize any URL or text using Kagi's Universal Summarizer API. Supports multiple engines (including the enterprise-grade Muriel model), bullet-point takeaways, and output translation to 28 languages. Use when you need a high-quality summary of an article, paper, video transcript, or any document.
---

# Kagi Universal Summarizer

Summarize any URL or block of text using [Kagi's Universal Summarizer API](https://help.kagi.com/kagi/api/summarizer.html). Handles articles, papers, PDFs, video transcripts, forum threads, and more. Supports multiple summarization engines and can translate output to 28 languages.

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

Token-based, billed per 1,000 tokens processed. Cached requests are free.

| Plan | Price per 1k tokens |
|------|---------------------|
| Standard (Cecil / Agnes) | **$0.030** |
| Kagi Ultimate subscribers | **$0.025** (automatically applied) |
| Muriel (enterprise-grade) | higher — check [API pricing page](https://kagi.com/settings?p=api) |

## Usage

```bash
# Summarize a URL
kagi summarize https://example.com/article

# Summarize raw text
kagi summarize --text "Paste your article text here..."

# Pipe text from stdin
cat paper.txt | kagi summarize
echo "Long text..." | kagi summarize --type takeaway

# Choose engine
kagi summarize https://arxiv.org/abs/1706.03762 --engine muriel

# Get bullet-point takeaways instead of prose
kagi summarize https://example.com/article --type takeaway

# Translate summary to another language
kagi summarize https://example.com/article --lang DE

# JSON output
kagi summarize https://example.com/article --format json

# Show balance only when needed
kagi summarize https://example.com/article --show-balance
kagi balance

# Combined options
kagi summarize https://arxiv.org/abs/1706.03762 --engine muriel --type takeaway --lang EN --format json
```

### Options

| Flag | Description |
|------|-------------|
| `--text <text>` | Summarize raw text instead of a URL |
| `--engine <name>` | Summarization engine (see below, default: `cecil`) |
| `--type <type>` | Output type: `summary` (prose) or `takeaway` (bullets) |
| `--lang <code>` | Translate output to a language code (e.g. `EN`, `DE`, `FR`, `JA`) |
| `--format <fmt>` | Output format: json (default), compact, pretty |
| `--no-cache` | Bypass cached responses |
| `--show-balance` | Print API balance to stderr for this call |
| `--timeout <sec>` | HTTP timeout in seconds (default: 120) |

### Engines

| Engine | Description |
|--------|-------------|
| `cecil` | Friendly, descriptive, fast summary **(default)** |
| `agnes` | Formal, technical, analytical summary |
| `muriel` | Best-in-class, enterprise-grade model — highest quality, slower |

### Language Codes

Common codes: `EN` English · `DE` German · `FR` French · `ES` Spanish · `IT` Italian · `PT` Portuguese · `JA` Japanese · `KO` Korean · `ZH` Chinese (simplified) · `ZH-HANT` Chinese (traditional) · `RU` Russian · `AR` Arabic

Full list: BG CS DA DE EL EN ES ET FI FR HU ID IT JA KO LT LV NB NL PL PT RO RU SK SL SV TR UK ZH ZH-HANT

## Output

### Default (JSON)

Returns a JSON object with `input`, `output`, `tokens`, `engine`, `type`, `meta`.

### Pretty text (`--format pretty`)

Prints the summary to stdout. Token usage printed to stderr.

## When to Use

- **Use kagi summarize** when you have a URL or document and need a concise summary
- **Use `--type takeaway`** for structured bullet points — ideal for research papers, long articles
- **Use `--engine muriel`** when quality matters most (longer documents, academic papers)
- **Use `--lang`** when you need the summary in a different language
- **Use kagi fastgpt** instead when you have a question requiring synthesis from multiple sources
- **Use kagi search** instead when you need raw search results to scan or compare

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
