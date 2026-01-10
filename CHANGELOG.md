# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.4.0] - 2026-01-10

### Added
- Comprehensive unit tests for CLI functions (`main_test.go` with 15 tests)
- CHANGELOG.md for version history tracking
- SECURITY.md with vulnerability reporting guidelines
- CODE_OF_CONDUCT.md (Contributor Covenant)
- GitHub issue templates (bug report, feature request)
- Pull request template
- CodeQL security scanning workflow
- Test coverage reporting with Codecov
- golangci-lint integration in CI

### Changed
- `IsConfluenceMIME()` now returns `(bool, error)` for proper error handling
- Extracted HTML entity replacements to shared `htmlEntityMap`
- Added named constants for magic numbers (`pandocTimeout`, `maxASCIICodePoint`, `mimeHeaderScanLimit`)
- Improved documentation with regex pattern comments
- Replaced `fmt.Sscanf` with `strconv.ParseInt` for proper error handling

### Fixed
- Replaced custom `contains()` helper with `strings.Contains()` in tests

## [0.3.1] - 2026-01-09

### Changed
- Updated logo with resized shiny version

### Added
- Added logo to README

## [0.3.0] - 2026-01-08

### Fixed
- Fixed module path for Go proxy compatibility

## [0.2.2] - 2026-01-08

### Fixed
- Fixed CI lint job to download pandoc binaries before vet
- Improved macOS Gatekeeper workaround instructions
- Used GitHub alert block for macOS warning in README

## [0.2.1] - 2026-01-07

### Added
- Added macOS Gatekeeper workaround instructions to README

### Changed
- Updated README with accurate installation information

## [0.2.0] - 2026-01-07

### Added
- Dual license (free for personal/open-source, commercial license available)
- Buy Me a Coffee badge and support link

### Changed
- Improved README formatting

## [0.1.0] - 2026-01-07

### Added
- Initial release
- Convert Confluence MIME-encoded `.doc` exports to clean Markdown
- Embedded Pandoc binary (zero external dependencies)
- Single file conversion mode
- Directory batch conversion mode (`--dir` flag)
- Dry-run mode (`--dry-run` flag)
- Verbose output mode (`-v` / `--verbose` flag)
- Custom output path (`-o` / `--output` flag)
- Automatic filename cleanup (replaces `+` with `-`)
- Cross-platform support (macOS, Linux, Windows)
- Confluence-specific post-processing:
  - Emoji conversion (tick, error, warning icons)
  - Macro handling (tip, note, warning, info blocks)
  - Layout div removal
  - TOC cleanup
  - HTML entity decoding

[Unreleased]: https://github.com/aqueeb/confluence2md/compare/v0.4.0...HEAD
[0.4.0]: https://github.com/aqueeb/confluence2md/compare/v0.3.1...v0.4.0
[0.3.1]: https://github.com/aqueeb/confluence2md/compare/v0.3.0...v0.3.1
[0.3.0]: https://github.com/aqueeb/confluence2md/compare/v0.2.2...v0.3.0
[0.2.2]: https://github.com/aqueeb/confluence2md/compare/v0.2.1...v0.2.2
[0.2.1]: https://github.com/aqueeb/confluence2md/compare/v0.2.0...v0.2.1
[0.2.0]: https://github.com/aqueeb/confluence2md/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/aqueeb/confluence2md/releases/tag/v0.1.0
