# Agent Skill Packaging

This repo publishes per-capability skill folders that let external agents call the shared `kagi` binary.

## Shared binary model

Each skill delegates to the same CLI so behavior stays consistent across human shell use and agent automation.

The skill markdown under `skills/kagi-*` documents setup and usage, but execution still funnels through the shared executable and command tree described in [[architecture#Executable and packaging model]]. This keeps auth, output formats, and feature flags centralized in one binary.

## Installation expectations

Users install the binary once, then symlink or copy the desired skill folders into their agent's skills directory.

The README documents the default targets for `~/.agents/skills/` and `~/.claude/skills/`. The skill docs also describe source builds and local binary builds so agents can run without a separate runtime beyond Go or a downloaded release.
