# Security Policy

## Supported Versions

| Version | Supported          |
| ------- | ------------------ |
| 0.3.x   | :white_check_mark: |
| 0.2.x   | :white_check_mark: |
| < 0.2   | :x:                |

## Reporting a Vulnerability

If you discover a security vulnerability in confluence2md, please report it responsibly.

### How to Report

1. **Do NOT open a public GitHub issue** for security vulnerabilities
2. Email the maintainer directly at: **[create a GitHub Security Advisory](https://github.com/aqueeb/confluence2md/security/advisories/new)**
3. Alternatively, use GitHub's private vulnerability reporting feature

### What to Include

Please include the following in your report:

- Description of the vulnerability
- Steps to reproduce
- Potential impact
- Suggested fix (if any)

### Response Timeline

- **Acknowledgment**: Within 48 hours
- **Initial Assessment**: Within 1 week
- **Resolution**: Depends on severity, typically within 30 days

### Security Considerations

#### File Handling

- confluence2md reads files from paths you specify
- Output files are written with standard permissions (0644)
- The tool does not follow symlinks specially; standard OS behavior applies

#### Pandoc Execution

- The embedded Pandoc binary is extracted to `~/.cache/confluence2md/`
- Pandoc is invoked with input via stdin (not command-line arguments)
- No user-controlled data is passed as command-line arguments to Pandoc

#### Input Validation

- Only processes files that match Confluence MIME export format
- HTML content is processed through Pandoc, not executed
- No JavaScript or active content execution

### Known Limitations

- The tool trusts that input files are what they claim to be (Confluence exports)
- Large files may consume significant memory during processing
- No sandboxing beyond standard OS process isolation

## Security Updates

Security updates will be released as patch versions (e.g., 0.3.1 -> 0.3.2) and announced in:

- GitHub Releases
- CHANGELOG.md
