# confluence2md - Project Context

## Overview
CLI tool to convert Confluence MIME-encoded `.doc` exports to clean Markdown.

## Tech Stack
- **Language:** Go 1.21+
- **Dependencies:** Pandoc (embedded in release binaries, or system-installed for dev)
- **No Go dependencies** - uses only standard library

## Project Structure
```
confluence2md/
├── main.go                      # CLI entry point, flag parsing
├── converter/
│   ├── mime.go                  # MIME parsing, HTML extraction
│   ├── mime_test.go             # MIME parser tests
│   ├── markdown.go              # Pandoc invocation, post-processing
│   └── markdown_test.go         # Markdown converter tests
├── internal/
│   └── pandoc/
│       ├── pandoc.go            # Embedded pandoc extraction logic
│       ├── pandoc_test.go       # Pandoc extraction tests
│       ├── embed_linux_amd64.go # Linux amd64 embed
│       ├── embed_darwin_amd64.go # macOS Intel embed
│       ├── embed_darwin_arm64.go # macOS Apple Silicon embed
│       ├── embed_windows_amd64.go # Windows amd64 embed
│       └── bin/                 # Downloaded pandoc binaries (gitignored)
├── scripts/
│   └── download-pandoc.sh       # Download pandoc binaries for embedding
├── go.mod
├── .gitignore
├── README.md
└── .goreleaser.yaml             # Release configuration
```

## Key Functions

### main.go
- `main()` - CLI entry, flag parsing
- `convertFile()` - Single file conversion
- `convertDirectory()` - Batch conversion
- `generateOutputPath()` - Output naming (`.doc` → `.md`, `+` → `-`)
- `printStarPrompt()` - Post-conversion GitHub star prompt

### converter/mime.go
- `ExtractHTMLFromMIME()` - Parse MIME, decode quoted-printable, return HTML
- `IsConfluenceMIME()` - Validate file is Confluence export

### converter/markdown.go
- `CheckPandoc()` - Verify pandoc is installed
- `ConvertHTMLToMarkdown()` - Pandoc conversion + post-processing
- `postProcessMarkdown()` - Confluence-specific cleanup
- `balanceDetailsTags()` - Remove orphaned `</details>` tags

## Commands
```bash
# Download pandoc binaries (required before building with embedded pandoc)
./scripts/download-pandoc.sh

# Build (with embedded pandoc)
go build -o confluence2md .

# Build for specific platform
GOOS=darwin GOARCH=arm64 go build -o confluence2md-darwin-arm64 .

# Test (requires pandoc binaries downloaded)
go test ./... -v

# Run
./confluence2md input.doc
./confluence2md --dir /path/to/docs
```

## Embedded Pandoc
Release binaries include an embedded Pandoc binary (~85-100 MB per platform).
The binary is extracted to `~/.cache/confluence2md/pandoc-{version}/` on first run.

To update the embedded Pandoc version:
1. Edit `internal/pandoc/pandoc.go` - update `Version` constant
2. Run `./scripts/download-pandoc.sh {new-version}`
3. Test and release

## Current State
- ✅ Core functionality complete
- ✅ Unit tests passing
- ✅ README written
- ✅ GitHub repo created (confluence2md_private)
- ✅ GoReleaser configured for automated builds
- ✅ Embedded Pandoc support (zero external dependencies)

## Releasing
To create a release with multi-platform binaries:
```bash
git tag v0.1.0
git push origin v0.1.0
```

This triggers GitHub Actions to build binaries for:
- macOS (arm64, amd64)
- Linux (amd64)
- Windows (amd64)

## Pending Features (discussed but not implemented)
1. Image extraction (`--extract-images` flag)
2. Custom output directory (`--output-dir`)
3. Recursive directory scanning

## GitHub Info
- **Repo URL:** https://github.com/aqueeb/confluence2md
- **Owner:** aqueeb
