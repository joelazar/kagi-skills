# Kagi CLI

This project exposes Kagi search, AI, translation, and feed workflows through a Go CLI.

## Product role

The CLI is the project's primary user surface for both human terminal use and shell-capable agents.

Humans can run direct subcommands or launch the interactive terminal mode, while agents typically invoke the same command surface through shell calls or skill wrappers. The internal structure lives in [[architecture#Architecture]], and agent-specific packaging lives in [[agents#Agent Skill Packaging]].

## Capability Families

The command set is grouped by the kind of backend contract each workflow uses.

### API-key commands

These commands use Kagi's paid API endpoints and authenticate with an API key.

`search`, `batch`, `fastgpt`, `summarize`, and `enrich` resolve credentials through [[internal/api/auth.go#ResolveAPIKey]] and usually refresh cached API-balance metadata after successful calls.

### Session-token commands

These commands depend on a logged-in subscriber session rather than the paid API key.

`quick`, `assistant`, `askpage`, and `translate` resolve a subscriber token through [[internal/api/auth.go#ResolveSessionToken]]. Assistant also exposes thread list, get, and delete workflows.

### Feed and utility commands

These commands round out the CLI with public feeds, local setup helpers, and build metadata.

`news` and `smallweb` surface Kagi-curated feeds, while `auth`, `config`, `completion`, `balance`, and `version` focus on setup, observability, and operator ergonomics.
