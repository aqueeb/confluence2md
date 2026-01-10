package converter

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"unicode/utf8"
)

// FuzzPreProcessHTML tests the preProcessHTML function with random inputs.
// This function has many regex patterns that could be vulnerable to:
// - Catastrophic backtracking (ReDoS)
// - Panics on malformed input
// - Unexpected behavior on edge cases
func FuzzPreProcessHTML(f *testing.F) {
	// Seed corpus with real-world examples
	seeds := []string{
		// Empty and minimal inputs
		"",
		" ",
		"\n",
		"\t",

		// Basic HTML
		"<html><body>Hello</body></html>",
		"<p>Simple paragraph</p>",
		"<div>Content</div>",

		// Confluence layout divs (should be removed)
		`<div class="contentLayout2"><p>Content</p></div>`,
		`<div class="columnLayout two-left-sidebar" data-layout="two-left-sidebar"><p>Content</p></div>`,
		`<div class="cell aside" data-type="aside"><div class="innerCell"><p>Content</p></div></div>`,
		`<div class="sectionColumnWrapper"><div class="sectionMacro"><p>Content</p></div></div>`,

		// Plugin elements
		`<div class="plugin_pagetree">Tree</div>`,
		`<fieldset class="hidden"><input type="hidden" name="test" value="123"></fieldset>`,
		`<ul class="plugin_pagetree_children"><li>Child</li></ul>`,

		// Empty paragraphs (should be removed)
		`<p></p>`,
		`<p><br></p>`,
		`<p>   </p>`,
		`<p><br/></p>`,

		// Style and data attributes (should be removed)
		`<p style="margin-left: 40.0px;">Indented</p>`,
		`<div data-layout="single" data-type="normal">Content</div>`,
		`<span tabindex="0" draggable="true">Span</span>`,

		// Images with various attributes
		`<img class="confluence-embedded-image" draggable="false" src="test.png" alt="Test">`,
		`<img src="image.jpg" data-image-src="/download/123" width="500">`,
		`<img class="emoticon" src="tick.png" alt="(tick)">`,

		// Tables (complex cleanup)
		`<table class="confluenceTable"><colgroup><col></colgroup><tr><td>Cell</td></tr></table>`,
		`<table><thead><tr><th class="confluenceTh" scope="col">Header</th></tr></thead></table>`,
		`<td>Line 1<br/>Line 2</td>`,
		`<td><p>Paragraph in cell</p></td>`,
		`<div class="table-wrap"><table><tr><td>Cell</td></tr></table></div>`,

		// Spans (various classes)
		`<span class="nolink">No link text</span>`,
		`<span class="status-macro aui-lozenge">STATUS</span>`,
		`<span class="icon aui-icon"></span>`,
		`<span class="aui-message">Message</span>`,

		// Double-encoded HTML
		`&lt;p&gt;This was double encoded&lt;/p&gt;`,
		`&lt;div class="test"&gt;Content&lt;/div&gt;`,

		// Content wrapper divs
		`<div class="content-wrapper"><p>Wrapped content</p></div>`,

		// Edge cases with nested structures
		`<div class="contentLayout2"><div class="columnLayout"><div class="cell"><div class="innerCell"><p>Deeply nested</p></div></div></div></div>`,

		// Malformed/edge case HTML
		`<div class="`,
		`<div class="test>`,
		`<img src="`,
		`<table><tr><td>Unclosed`,
		`>>>>>>>>>`,
		`<<<<<<<<<<`,
		`<div><div><div><div>`,
		`</div></div></div></div>`,

		// Unicode content
		`<p>日本語コンテンツ</p>`,
		`<p>Émoji: 🎉 ✅ ❌</p>`,
		`<div>Ñoño</div>`,

		// Large repetition (potential ReDoS)
		strings.Repeat("<div>", 100) + "Content" + strings.Repeat("</div>", 100),
		strings.Repeat(`<span class="test">`, 50) + "X" + strings.Repeat("</span>", 50),
	}

	for _, seed := range seeds {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, input string) {
		// The function should never panic
		result := preProcessHTML(input)

		// Invariant: if input is valid UTF-8, result should be too
		if utf8.ValidString(input) && !utf8.ValidString(result) {
			t.Errorf("Valid UTF-8 input produced invalid UTF-8 output")
		}

		// Invariant: should not increase size dramatically (no exponential blowup)
		if len(result) > len(input)*10+1000 {
			t.Errorf("Result size %d is much larger than input size %d", len(result), len(input))
		}
	})
}

// FuzzPostProcessMarkdown tests the postProcessMarkdown function.
// This function has many regex patterns for cleaning up Confluence-specific artifacts.
func FuzzPostProcessMarkdown(f *testing.F) {
	seeds := []string{
		// Empty and minimal
		"",
		" ",
		"\n",
		"# Heading",
		"Plain text",

		// Emoji images
		`<img src="test" alt="(tick)" />`,
		`<img src="test" alt="(error)" class="emoticon"/>`,
		`<img alt="(blue star)" src="test.png">`,
		`<img src="x" alt="(warning)">`,
		`<img class="expand-control-image" src="expand.png">`,

		// Text emojis
		":celebration:",
		":thumbsup:",
		":thumbsdown:",
		":check:",
		":cross:",
		":warning:",
		":info:",
		":star:",
		":fire:",
		":rocket:",
		":sparkles:",
		":question:",

		// Section1 div
		`<div class="Section1">Content</div>`,

		// TOC macro
		`<div class="toc-macro rbtoc1234">- [Link](#link)</div>`,

		// Confluence macros
		`<div class="confluence-information-macro confluence-information-macro-tip"><div class="confluence-information-macro-body">Tip text</div></div>`,
		`<div class="confluence-information-macro confluence-information-macro-note">Note</div>`,
		`<div class="confluence-information-macro confluence-information-macro-warning">Warning</div>`,
		`<div class="confluence-information-macro confluence-information-macro-information">Info</div>`,

		// AUI icons
		`<span class="aui-icon aui-icon-small"></span> Text`,

		// Panel divs
		`<div class="panel" style="border: 1px;"><div class="panelContent">Panel content</div></div>`,

		// Expander sections
		`<div id="expander-123"><div id="expander-control-123"><span class="expand-control-text">Title</span></div><div id="expander-content-123">Content</div></div>`,
		`<span class="expand-control-icon">+</span><span class="expand-control-text">Expand</span>`,

		// Code blocks
		"``` syntaxhighlighter-pre\ncode\n```",
		"```{.language}\ncode\n```",
		`<div class="code panel pdl"><div class="codeContent panelContent pdl">code</div></div>`,
		`<div class="codeHeader panelHeader">Header</div>`,

		// Links
		`<a href="https://example.com">Example</a>`,
		`<a href="https://example.com" class="external-link">Example</a>`,
		`<a href="url"><u>Underlined Link</u></a>`,

		// HTML entities
		"&lt;tag&gt;",
		"&amp;",
		"&quot;quoted&quot;",
		"&nbsp;",

		// Escaped HTML from pandoc
		`\<br\>`,
		`\<p\>paragraph\</p\>`,
		`\<div class="test"\>content\</div\>`,
		`\<img src="test.png" alt="Image"\>`,

		// Nested lists
		"- - Item",
		"  - - Nested",

		// Details tags
		"<details>\nContent\n</details>",
		"Content\n</details>",
		"<details>\n</details>\n</details>\n</details>",

		// Br tags
		"Line 1<br>Line 2",
		"Line 1<br/>Line 2",
		"Line 1<br />Line 2",

		// Multiple newlines
		"Line 1\n\n\n\n\nLine 2",

		// Trailing whitespace
		"Line with trailing spaces   ",
		"Line with trailing tabs\t\t",

		// Div cleanup
		"</div></div></div>",
		"<div>Open",
		"</div>Orphan",

		// Span tags
		`<span class="test">Content</span>`,
		`</span>orphan`,

		// Unicode
		"日本語",
		"Émoji 🎉",

		// Edge cases
		strings.Repeat("</details>", 100),
		strings.Repeat("<details>", 100),
		strings.Repeat("<br>", 100),
	}

	for _, seed := range seeds {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, input string) {
		// Should never panic
		result := postProcessMarkdown(input)

		// Invariant: if input is valid UTF-8, result should be too
		if utf8.ValidString(input) && !utf8.ValidString(result) {
			t.Errorf("Valid UTF-8 input produced invalid UTF-8 output")
		}

		// Invariant: result should end with newline (per implementation)
		if result != "" && !strings.HasSuffix(result, "\n") {
			t.Errorf("Result should end with newline, got: %q", result[max(0, len(result)-20):])
		}

		// Invariant: result size should not explode
		if len(result) > len(input)*10+1000 {
			t.Errorf("Result size %d is much larger than input size %d", len(result), len(input))
		}

		// Invariant: details tags should be balanced in output
		openCount := strings.Count(result, "<details>")
		closeCount := strings.Count(result, "</details>")
		if closeCount > openCount {
			t.Errorf("More closing </details> tags (%d) than opening (%d)", closeCount, openCount)
		}
	})
}

// FuzzDecodeHTMLEntities tests HTML entity decoding.
func FuzzDecodeHTMLEntities(f *testing.F) {
	seeds := []string{
		// No entities (passthrough)
		"",
		"plain text",
		"no entities here",

		// Named entities
		"&lt;",
		"&gt;",
		"&amp;",
		"&quot;",
		"&apos;",
		"&nbsp;",
		"&#39;",
		"&#34;",

		// Numeric decimal entities
		"&#60;",
		"&#62;",
		"&#38;",
		"&#65;", // A
		"&#90;", // Z
		"&#97;", // a
		"&#122;", // z
		"&#126;", // ~
		"&#127;", // DEL (boundary)
		"&#128;", // above ASCII (not decoded)
		"&#200;", // È (not decoded)

		// Numeric hex entities
		"&#x3C;", // <
		"&#x3E;", // >
		"&#x26;", // &
		"&#x41;", // A
		"&#x7E;", // ~
		"&#x7F;", // DEL (boundary)
		"&#xC8;", // È (not decoded)
		"&#x3c;", // lowercase hex

		// Mixed content
		"&lt;p&gt;Hello &amp; world&lt;/p&gt;",
		"Price: &lt;$100 &amp; &gt;$50",
		"Quotes: &quot;hello&quot; and &#39;world&#39;",

		// Edge cases
		"&lt;&lt;&lt;",
		"&#60;&#60;&#60;",
		"&#x3C;&#x3C;&#x3C;",
		"&amp;lt;", // &lt; literally
		"&#38;lt;",
		"&unknown;",
		"&",
		"&#",
		"&#;",
		"&#x;",
		"&#xG;", // invalid hex
		"&#abc;", // invalid decimal
		"&#999999999999;", // overflow

		// Without trigger (passthrough - no &lt; or &#)
		"no entities at all",
		"& standalone ampersand",

		// Large numbers
		"&#0;",
		"&#1;",
		"&#65535;",
		"&#x0;",
		"&#xFFFF;",

		// Unicode
		"日本語 &lt;tag&gt;",
		"Émoji: 🎉 &amp;",
	}

	for _, seed := range seeds {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, input string) {
		// Should never panic
		result := decodeHTMLEntities(input)

		// Invariant: if input is valid UTF-8, result should be too
		if utf8.ValidString(input) && !utf8.ValidString(result) {
			t.Errorf("Valid UTF-8 input produced invalid UTF-8 output")
		}

		// Only check other invariants for valid UTF-8 input
		if utf8.ValidString(input) {
			// Invariant: if input has no &lt; or &#, output equals input
			if !strings.Contains(input, "&lt;") && !strings.Contains(input, "&#") {
				if result != input {
					t.Errorf("Expected passthrough for input without entities, got different result")
				}
			}
		}

		// Invariant: result length should not explode
		if len(result) > len(input)*2+100 {
			t.Errorf("Result size %d unexpectedly larger than input size %d", len(result), len(input))
		}
	})
}

// FuzzBalanceDetailsTags tests the details tag balancing function.
// NOTE: This fuzz test originally discovered a bug where removing a </details> tag
// could accidentally create a new one from surrounding characters. The bug was fixed
// by recounting tags after each removal instead of just decrementing a counter.
func FuzzBalanceDetailsTags(f *testing.F) {
	seeds := []string{
		// Balanced
		"",
		"no tags",
		"<details>content</details>",
		"<details><details>nested</details></details>",
		"<details>a</details><details>b</details>",

		// Orphan closing tags
		"</details>",
		"content</details>",
		"</details></details></details>",
		"<details>content</details></details>",
		"<details></details></details></details>",

		// Only opening
		"<details>",
		"<details><details>",
		"<details>content",

		// Mixed
		"text<details>content</details>more</details>end",
		"<details>a</details>b</details>c</details>",

		// Edge cases
		strings.Repeat("<details>", 50) + strings.Repeat("</details>", 100),
		strings.Repeat("</details>", 100),
		"<details>" + strings.Repeat("</details>", 100),

		// With other content
		"# Heading\n<details>\nContent\n</details>\n</details>",
		"<details>\n```code```\n</details>",
	}

	for _, seed := range seeds {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, input string) {
		// Should never panic
		result := balanceDetailsTags(input)

		// Invariant: if input is valid UTF-8, result should be too
		if utf8.ValidString(input) && !utf8.ValidString(result) {
			t.Errorf("Valid UTF-8 input produced invalid UTF-8 output")
		}

		// Invariant: result should not be longer than input
		// (we only remove tags, never add)
		if len(result) > len(input) {
			t.Errorf("Result length %d exceeds input length %d", len(result), len(input))
		}

		// Invariant: closing tags should not exceed opening tags in output
		// (this was previously disabled due to a bug, now fixed)
		openCount := strings.Count(result, "<details>")
		closeCount := strings.Count(result, "</details>")
		if closeCount > openCount {
			t.Errorf("Result has more </details> (%d) than <details> (%d)", closeCount, openCount)
		}
	})
}

// FuzzExtractHTMLFromMIME tests the MIME parser with random inputs.
// This tests the robustness of the MIME parsing against malformed input.
func FuzzExtractHTMLFromMIME(f *testing.F) {
	// Valid MIME message seeds
	validMIME := `Date: Wed, 7 Jan 2026 01:29:00 +0000 (UTC)
Message-ID: <1234567890.123.1234567890123@test>
Subject: Exported From Confluence
MIME-Version: 1.0
Content-Type: multipart/related;
	boundary="----=_Part_123_456789.123456789"

------=_Part_123_456789.123456789
Content-Type: text/html; charset=UTF-8
Content-Transfer-Encoding: quoted-printable

<html><body><h1>Test</h1></body></html>
------=_Part_123_456789.123456789--
`

	seeds := []string{
		// Valid MIME
		validMIME,

		// Empty
		"",

		// Just headers, no body
		"Date: Wed, 7 Jan 2026 01:29:00 +0000\nMIME-Version: 1.0\n\n",

		// Plain text (not MIME)
		"Just plain text\nNo MIME here\n",

		// Partial headers
		"Date: Wed, 7 Jan 2026\n",
		"Content-Type: text/html\n\nContent",

		// Malformed boundary
		"Content-Type: multipart/related; boundary=\n\nBody",
		"Content-Type: multipart/related\n\nNo boundary param",

		// Non-multipart
		"Content-Type: text/plain\n\nPlain content",
		"Content-Type: application/json\n\n{}",

		// Truncated
		"Date: Wed, 7 Jan 2026\nContent-Type: multipart/related; boundary=\"test\"\n\n--test\nContent-Type: text/html\n\n<html>",

		// Binary-like content
		"\x00\x01\x02\x03",
		"Header: value\n\n\x00\x01\x02",

		// Very long lines
		"Header: " + strings.Repeat("x", 10000) + "\n\nBody",

		// Unicode in headers
		"Subject: 日本語\n\nBody",

		// Weird boundaries
		"Content-Type: multipart/related; boundary=\"\"\n\n",
		"Content-Type: multipart/related; boundary=\"---\"\n\n------\n",

		// Multiple parts with no HTML
		`Content-Type: multipart/related; boundary="b"

--b
Content-Type: text/plain

Plain
--b
Content-Type: image/png

PNG
--b--
`,
	}

	for _, seed := range seeds {
		f.Add([]byte(seed))
	}

	f.Fuzz(func(t *testing.T, data []byte) {
		// Create a temp file with the fuzzed content
		tmpDir := t.TempDir()
		tmpFile := filepath.Join(tmpDir, "fuzz.doc")

		if err := os.WriteFile(tmpFile, data, 0644); err != nil {
			t.Fatalf("Failed to write temp file: %v", err)
		}

		// The function should not panic, but errors are expected for invalid input
		_, _ = ExtractHTMLFromMIME(tmpFile)

		// We don't check the result - we just verify no panics occur
	})
}

// FuzzIsConfluenceMIME tests the MIME detection with random inputs.
func FuzzIsConfluenceMIME(f *testing.F) {
	seeds := []string{
		// Valid Confluence MIME headers
		"Date: Wed, 7 Jan 2026 01:29:00 +0000\nMIME-Version: 1.0\nSubject: Exported From Confluence\n\nBody",

		// Missing parts
		"Date: Wed, 7 Jan 2026\n\nBody",
		"MIME-Version: 1.0\n\nBody",
		"Subject: Exported From Confluence\n\nBody",
		"Date: x\nMIME-Version: 1.0\n\nBody",
		"Date: x\nSubject: Exported From Confluence\n\nBody",

		// Wrong subject
		"Date: x\nMIME-Version: 1.0\nSubject: Random Email\n\nBody",

		// Empty
		"",
		"\n",
		"\n\n\n",

		// Plain text
		"Just some text",

		// Binary
		"\x00\x01\x02",

		// Headers after limit (10 lines)
		strings.Repeat("X-Header: value\n", 15) + "Date: x\nMIME-Version: 1.0\nSubject: Exported From Confluence\n",

		// Very long lines
		"Date: " + strings.Repeat("x", 10000) + "\n\nBody",

		// Unicode
		"Subject: Exported From Confluence 日本語\nDate: x\nMIME-Version: 1.0\n",
	}

	for _, seed := range seeds {
		f.Add([]byte(seed))
	}

	f.Fuzz(func(t *testing.T, data []byte) {
		tmpDir := t.TempDir()
		tmpFile := filepath.Join(tmpDir, "fuzz.doc")

		if err := os.WriteFile(tmpFile, data, 0644); err != nil {
			t.Fatalf("Failed to write temp file: %v", err)
		}

		// Should not panic
		result, err := IsConfluenceMIME(tmpFile)

		// If no error, result should be boolean
		if err == nil {
			// Just verify it returns true or false without panic
			_ = result
		}
	})
}

// Helper function for max
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
