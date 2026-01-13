# Third-Party Software Notices

This software includes or depends on the following third-party components:

## Pandoc

**Version:** 3.6.4
**License:** GNU General Public License v2.0 or later (GPL-2.0-or-later)
**Copyright:** 2006-2025 John MacFarlane
**Source Code:** https://github.com/jgm/pandoc
**License Text:** See [THIRD_PARTY_LICENSES/GPL-2.0.txt](THIRD_PARTY_LICENSES/GPL-2.0.txt)

### Description

Pandoc is a universal document converter. This software bundles Pandoc
executables for Windows, macOS, and Linux. These executables are extracted
to `~/.cache/confluence2md/pandoc-3.6.4/` at runtime and invoked as
separate processes via `exec.CommandContext()`.

The Pandoc executable is distributed as a separate program and is NOT linked
with confluence2md. This constitutes "mere aggregation" under GPL section 2.

### Source Code Availability

Complete source code for the bundled Pandoc version is available at:
https://github.com/jgm/pandoc/releases/tag/3.6.4

For questions about Pandoc licensing or source code, see the
[Pandoc GitHub repository](https://github.com/jgm/pandoc) or the
[pandoc-discuss mailing list](https://groups.google.com/g/pandoc-discuss).

We will provide a copy of the Pandoc source code upon request for three years
from the date of distribution, in accordance with GPL Section 3(b).

---

## Go Module Dependencies

This project uses Go modules. Run `go mod graph` to see all dependencies.
Each dependency's license is available in its respective repository.

---

*This notice file is provided for informational purposes regarding third-party
components. It does not constitute legal advice.*
