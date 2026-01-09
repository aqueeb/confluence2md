# Contributing to confluence2md

Thanks for your interest in contributing! This document outlines how to get started.

## Getting Started

1. Fork the repository
2. Clone your fork: `git clone https://github.com/YOUR_USERNAME/confluence2md.git`
3. Create a branch: `git checkout -b my-feature`

## Development Setup

### Prerequisites

- Go 1.21 or later
- Pandoc (for running tests)

### Building

```bash
# Download pandoc binaries (required for embedded pandoc builds)
./scripts/download-pandoc.sh

# Build
go build -o confluence2md .

# Run tests
go test ./... -v
```

## Making Changes

1. Make your changes
2. Add tests if applicable
3. Run `go test ./...` to ensure tests pass
4. Run `go fmt ./...` to format code
5. Commit with a clear message

## Pull Requests

- Keep PRs focused on a single change
- Update documentation if needed
- Ensure tests pass
- Write a clear PR description

## Reporting Issues

When reporting bugs, please include:

- Your OS and Go version
- The command you ran
- Expected vs actual behavior
- Sample input file (if possible, sanitized of sensitive content)

## Code Style

- Follow standard Go conventions
- Run `go fmt` before committing
- Keep functions focused and small
- Add comments for non-obvious logic

## Questions?

Open an issue with your question — happy to help!
