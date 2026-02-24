# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [v1.1.0] - 2026-02-24

### Added
- Dedicated `balance` command across all four kagi skills to check Kagi API credit balance
- `AGENTS.md` with multi-module monorepo layout and Makefile task reference

### Security
- Wrapper scripts now verify SHA-256 checksums after downloading binaries

## [v1.0.0] - 2025-02-21

### Added
- Four CLI tools for Kagi APIs: `kagi-search`, `kagi-fastgpt`, `kagi-summarizer`, and `kagi-enrich`
- Pre-built binaries for Linux (amd64/arm64), macOS (amd64/arm64), and Windows (amd64)
- Auto-download wrapper scripts in each `SKILL.md` for seamless agent integration
- `go install` support via canonical module paths
- `--version` / `-v` flag for all binaries
- CI workflow (lint, format, build) and automated release workflow
- `Makefile` with `build`, `lint`, `test`, `fmt`, and `clean` targets
- MIT license

[Unreleased]: https://github.com/joelazar/kagi-skills/compare/v1.1.0...HEAD
[v1.1.0]: https://github.com/joelazar/kagi-skills/compare/v1.0.0...v1.1.0
[v1.0.0]: https://github.com/joelazar/kagi-skills/releases/tag/v1.0.0
