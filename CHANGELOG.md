# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- `--version` / `-v` flag for all four binaries, injected at build time via `-ldflags`
- `Makefile` with `build`, `lint`, `test`, `fmt`, and `clean` targets
- `LICENSE` (MIT)
- `CHANGELOG.md` (this file)
- Version embedded in release binary filenames (`skill_vX.Y.Z_os_arch`)
- Checksum verification instructions in each `SKILL.md`
- Full `go install` support via canonical module paths (`github.com/joelazar/kagi-skills/...`)
- CI and release badges in `README.md`

[Unreleased]: https://github.com/joelazar/kagi-skills/compare/HEAD...HEAD
