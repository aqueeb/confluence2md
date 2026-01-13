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

## Commit Message Conventions

We follow [Conventional Commits](https://www.conventionalcommits.org/) for clear, consistent history.

### Format

```
<type>: <description>

[optional body]
```

### Types

| Type | Description |
|------|-------------|
| `feat` | New feature |
| `fix` | Bug fix |
| `docs` | Documentation only |
| `test` | Adding or updating tests |
| `refactor` | Code change that neither fixes a bug nor adds a feature |
| `chore` | Build process, CI, or auxiliary tool changes |
| `perf` | Performance improvement |

### Examples

```
feat: Add --output-dir flag for custom output location
fix: Handle empty MIME boundaries gracefully
docs: Add troubleshooting section to README
test: Add fuzz tests for MIME parser
refactor: Extract HTML cleaning into separate function
chore: Update golangci-lint to v1.55
```

### Guidelines

- Use lowercase for the description
- Don't end the description with a period
- Use imperative mood ("Add feature" not "Added feature")
- Keep the first line under 72 characters
- Reference issues when applicable: `fix: Handle edge case (#42)`

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

## Developer Certificate of Origin (DCO)

This project uses the [Developer Certificate of Origin](https://developercertificate.org/)
to ensure contributions are properly licensed. By contributing, you certify that you have
the right to submit your contribution under the Apache 2.0 license.

### Signing Your Commits

Add the `-s` flag when committing:

```bash
git commit -s -m "feat: Add new feature"
```

This adds a `Signed-off-by` line to your commit message:

```
feat: Add new feature

Signed-off-by: Your Name <your.email@example.com>
```

### Git Configuration

Ensure your Git config matches your GitHub account:

```bash
git config user.name "Your Name"
git config user.email "your.email@example.com"
```

### Fixing Unsigned Commits

If you forgot to sign a commit, amend it:

```bash
git commit --amend -s --no-edit
git push --force-with-lease
```

For multiple unsigned commits, use interactive rebase:

```bash
git rebase -i HEAD~N  # where N is the number of commits
# Mark commits as "edit", then for each:
git commit --amend -s --no-edit
git rebase --continue
```

## Questions?

Open an issue with your question — happy to help!
