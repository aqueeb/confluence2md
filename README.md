<p align="center">
  <img src="logo.png" alt="confluence2md logo" width="256">
</p>

<h1 align="center">confluence2md</h1>

<p align="center">
  <a href="https://goreportcard.com/report/github.com/aqueeb/confluence2md"><img src="https://goreportcard.com/badge/github.com/aqueeb/confluence2md?v=2" alt="Go Report Card"></a>
  <a href="https://codecov.io/gh/aqueeb/confluence2md"><img src="https://codecov.io/gh/aqueeb/confluence2md/branch/main/graph/badge.svg?token=unused" alt="Coverage"></a>
  <a href="LICENSE"><img src="https://img.shields.io/badge/License-Dual%20(Non--Commercial%20Free)-blue.svg" alt="License: Dual"></a>
  <a href="https://github.com/aqueeb/confluence2md/releases"><img src="https://img.shields.io/github/v/release/aqueeb/confluence2md" alt="Release"></a>
  <a href="https://buymeacoffee.com/aqueeb"><img src="https://img.shields.io/badge/Buy%20Me%20A%20Coffee-support-yellow?logo=buymeacoffee" alt="Buy Me A Coffee"></a>
</p>

A CLI tool to convert Confluence MIME-encoded `.doc` exports to clean Markdown.

## The Problem

Confluence's "Export to Word" feature doesn't create real Word documents—it creates **MIME-encoded HTML files** with a `.doc` extension. Only Microsoft Word can open them. This has been [a known issue for over 10 years](https://community.atlassian.com/forums/Confluence-questions/Why-is-confluence-cloud-s-export-to-Word-feature-creating-an/qaq-p/2325894).

**What doesn't work:**
- LibreOffice, Google Docs, and other word processors
- Programmatic document parsers (python-docx, mammoth, etc.)
- Any tool expecting a real `.doc` or `.docx` file

**Why this matters:**
You can't convert Confluence exports to Markdown for version control, static site generators, or LLM/RAG pipelines—until now.

## Features

- **Zero dependencies** - release binaries include embedded pandoc
- **LLM/RAG-ready output** - clean Markdown optimized for chunking and embedding
- Parses MIME-encoded Confluence exports (not binary `.doc` files)
- Uses pandoc for high-quality HTML-to-Markdown conversion
- Cleans up Confluence-specific HTML artifacts
- Converts emoji images to Unicode (✅ ❌ 🚧 ⚠️)
- Converts info/tip/warning boxes to blockquotes
- Handles collapsible sections, code blocks, and tables
- Batch convert entire directories

## Use Cases

- **Migrate to Git-based docs** — Move Confluence content to GitBook, Docusaurus, MkDocs, or any static site generator
- **Build RAG/LLM knowledge bases** — Feed your Confluence docs to LangChain, LlamaIndex, or custom embedding pipelines
- **Create portable backups** — Store documentation in a format that doesn't require Confluence or MS Word to read
- **Power AI coding assistants** — Add your team's documentation context to Copilot, Cursor, or Claude

## Installation

### From releases (recommended)

Download the binary for your platform from [Releases](https://github.com/aqueeb/confluence2md/releases). Release binaries include an embedded pandoc, so there are **no external dependencies**.

> [!IMPORTANT]
> **macOS users:** If you see "Apple could not verify" warning, either:
> - Run `xattr -d com.apple.quarantine /path/to/confluence2md` in Terminal, or
> - Go to **System Settings → Privacy & Security** and click "Open Anyway"

### From source

```bash
go install github.com/aqueeb/confluence2md@latest
```

> **Note:** Building from source requires [pandoc](https://pandoc.org/installing.html) to be installed on your system.

## Usage

```bash
# Convert a single file
confluence2md document.doc

# Convert with custom output path
confluence2md -o output.md document.doc

# Convert all .doc files in a directory
confluence2md --dir /path/to/docs

# Preview what would be converted (dry run)
confluence2md --dir /path/to/docs --dry-run

# Verbose output
confluence2md -v document.doc
```

## Flags

| Flag | Description |
|------|-------------|
| `-o, --output` | Output file path (default: input with `.md` extension) |
| `--dir` | Convert all `.doc` files in directory |
| `-v, --verbose` | Show detailed processing info |
| `--dry-run` | Show what would be converted without writing |
| `--version` | Show version |

## What it converts

This tool specifically handles **Confluence MIME exports** - files that look like `.doc` but are actually MIME-encoded HTML. These are created when exporting pages from Confluence to Word format.

It does **not** handle:
- Binary Microsoft Word `.doc` files
- `.docx` files (use pandoc directly for these)

## How it works

1. **MIME parsing**: Extracts HTML content from the multipart MIME message
2. **Pandoc conversion**: Converts HTML to GitHub-flavored Markdown
3. **Post-processing**: Cleans up Confluence-specific artifacts:
   - Removes wrapper divs (`Section1`, `toc-macro`)
   - Converts info boxes to blockquotes (`> **Tip:**`, `> **Note:**`)
   - Replaces emoji images with Unicode characters
   - Fixes code block language hints
   - Balances orphaned HTML tags

## Support

If this tool saved you time, consider buying me a coffee:

[!["Buy Me A Coffee"](https://www.buymeacoffee.com/assets/img/custom_images/orange_img.png)](https://www.buymeacoffee.com/aqueeb)

Or just star the repo — it helps others discover this tool!

## Contributing

Contributions are welcome! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## License

**Dual License** — Free for non-commercial use. Commercial use requires a paid license. See [LICENSE](LICENSE) for details.
