package converter

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/aqueeb/confluence2md/internal/pandoc"
)

const (
	// pandocTimeout is the maximum time allowed for pandoc conversion.
	pandocTimeout = 2 * time.Minute

	// maxASCIICodePoint is the maximum code point for ASCII characters.
	// Used when decoding numeric HTML entities to only decode printable ASCII.
	maxASCIICodePoint = 127
)

// htmlEntityMap maps HTML entities to their decoded characters.
// Used for decoding double-encoded HTML from Confluence exports.
var htmlEntityMap = map[string]string{
	"&lt;":   "<",
	"&gt;":   ">",
	"&amp;":  "&",
	"&quot;": "\"",
	"&#39;":  "'",
	"&apos;": "'",
	"&#x27;": "'",
	"&#34;":  "\"",
	"&#60;":  "<",
	"&#62;":  ">",
	"&#38;":  "&",
	"&nbsp;": " ",
}

// CheckPandoc verifies that pandoc is available (embedded or in PATH).
func CheckPandoc() error {
	// First try to use embedded pandoc
	if pandoc.IsEmbedded() {
		_, err := pandoc.EnsureExtracted()
		if err != nil {
			return fmt.Errorf("failed to extract embedded pandoc: %w", err)
		}
		return nil
	}

	// Fallback to system pandoc
	_, err := exec.LookPath("pandoc")
	if err != nil {
		return fmt.Errorf("pandoc not found in PATH. Please install pandoc: https://pandoc.org/installing.html")
	}
	return nil
}

// ConvertHTMLToMarkdown converts HTML content to Markdown using pandoc and applies post-processing.
func ConvertHTMLToMarkdown(html string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), pandocTimeout)
	defer cancel()

	// Pre-process HTML to remove Confluence layout markup
	html = preProcessHTML(html)

	// Try embedded pandoc first
	if pandoc.IsEmbedded() {
		mdBytes, err := pandoc.Convert(ctx, []byte(html), "html", "gfm", "--wrap=none")
		if err != nil {
			return "", fmt.Errorf("pandoc conversion failed: %w", err)
		}

		markdown := postProcessMarkdown(string(mdBytes))
		return markdown, nil
	}

	// Fallback to system pandoc using temp files
	tmpHTML, err := os.CreateTemp("", "confluence-*.html")
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmpHTML.Name())

	if _, err := tmpHTML.WriteString(html); err != nil {
		return "", fmt.Errorf("failed to write HTML to temp file: %w", err)
	}
	tmpHTML.Close()

	// Create temp file for Markdown output
	tmpMD, err := os.CreateTemp("", "confluence-*.md")
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmpMD.Name())
	tmpMD.Close()

	// Run system pandoc
	cmd := exec.Command("pandoc",
		"-f", "html",
		"-t", "gfm",
		"--wrap=none",
		tmpHTML.Name(),
		"-o", tmpMD.Name(),
	)

	if output, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("pandoc failed: %w\nOutput: %s", err, string(output))
	}

	// Read the converted markdown
	mdBytes, err := os.ReadFile(tmpMD.Name())
	if err != nil {
		return "", fmt.Errorf("failed to read converted markdown: %w", err)
	}

	markdown := postProcessMarkdown(string(mdBytes))
	return markdown, nil
}

// decodeHTMLEntities decodes HTML entities that represent actual HTML tags.
// Confluence exports sometimes double-encode HTML, resulting in &lt;p&gt; instead of <p>.
func decodeHTMLEntities(html string) string {
	// Check if content appears to be double-encoded (contains &lt; which represents <)
	if !strings.Contains(html, "&lt;") && !strings.Contains(html, "&#") {
		return html
	}

	// Decode common HTML entities using the shared map
	for entity, char := range htmlEntityMap {
		html = strings.ReplaceAll(html, entity, char)
	}

	// Handle numeric HTML entities for common characters.
	// These patterns decode character references like &#x3C; (hex) and &#60; (decimal).
	// Only ASCII characters (< 127) are decoded to avoid issues with extended characters.
	//
	// Hex format: &#xNN; where NN is a hexadecimal number
	// Pattern breakdown: &#x captures literal prefix, ([0-9a-fA-F]+) captures hex digits, ; captures literal suffix
	hexEntityPattern := regexp.MustCompile(`&#x([0-9a-fA-F]+);`)
	html = hexEntityPattern.ReplaceAllStringFunc(html, func(match string) string {
		submatches := hexEntityPattern.FindStringSubmatch(match)
		if len(submatches) > 1 {
			val, err := strconv.ParseInt(submatches[1], 16, 32)
			if err == nil && val > 0 && val < maxASCIICodePoint {
				return string(rune(val))
			}
		}
		return match
	})

	// Decimal format: &#NNN; where NNN is a decimal number
	// Pattern breakdown: &# captures literal prefix, (\d+) captures decimal digits, ; captures literal suffix
	decEntityPattern := regexp.MustCompile(`&#(\d+);`)
	html = decEntityPattern.ReplaceAllStringFunc(html, func(match string) string {
		submatches := decEntityPattern.FindStringSubmatch(match)
		if len(submatches) > 1 {
			val, err := strconv.ParseInt(submatches[1], 10, 32)
			if err == nil && val > 0 && val < maxASCIICodePoint {
				return string(rune(val))
			}
		}
		return match
	})

	return html
}

// preProcessHTML removes Confluence layout markup before Pandoc conversion.
// This ensures layout divs don't get escaped and pollute the output.
func preProcessHTML(html string) string {
	// First, decode HTML entities that represent actual HTML tags
	// Confluence sometimes double-encodes HTML, resulting in &lt;p&gt; instead of <p>
	html = decodeHTMLEntities(html)

	// Remove Confluence page layout containers (these wrap content in columns)
	layoutPatterns := []string{
		`<div class="contentLayout2"[^>]*>`,
		`<div class="columnLayout[^"]*"[^>]*>`,
		`<div class="cell[^"]*"[^>]*>`,
		`<div class="innerCell"[^>]*>`,
		`<div class="sectionColumnWrapper"[^>]*>`,
		`<div class="sectionMacro"[^>]*>`,
		`<div class="sectionMacroRow"[^>]*>`,
		`<div class="plugin_pagetree[^"]*"[^>]*>`,
		`<div class="plugin_pagetree_children[^"]*"[^>]*>`,
		`<div class="plugin-tabmeta-details"[^>]*>`,
	}
	for _, pattern := range layoutPatterns {
		html = regexp.MustCompile(pattern).ReplaceAllString(html, "")
	}

	// Remove Confluence plugin elements (page tree, hidden fieldsets, etc.)
	pluginPatterns := []string{
		`<fieldset class="hidden"[^>]*>[\s\S]*?</fieldset>`,
		`<input type="hidden"[^>]*>`,
		`<ul[^>]*class="[^"]*plugin_pagetree[^"]*"[^>]*>[\s\S]*?</ul>`,
	}
	for _, pattern := range pluginPatterns {
		html = regexp.MustCompile(pattern).ReplaceAllString(html, "")
	}

	// Remove empty paragraphs and excessive breaks
	html = regexp.MustCompile(`<p>\s*</p>`).ReplaceAllString(html, "")
	html = regexp.MustCompile(`<p>\s*<br\s*/?>\s*</p>`).ReplaceAllString(html, "")
	html = regexp.MustCompile(`<p[^>]*>\s*\\?<br\s*/?>\\?\s*</p>`).ReplaceAllString(html, "")

	// Remove style attributes that can cause issues
	html = regexp.MustCompile(`\s+style="[^"]*"`).ReplaceAllString(html, "")

	// Remove data-* attributes
	html = regexp.MustCompile(`\s+data-[a-z-]+="[^"]*"`).ReplaceAllString(html, "")

	// Remove tabindex attributes
	html = regexp.MustCompile(`\s+tabindex="[^"]*"`).ReplaceAllString(html, "")

	// Remove draggable attributes
	html = regexp.MustCompile(`\s+draggable="[^"]*"`).ReplaceAllString(html, "")

	// Convert Confluence image tags to simple img tags pandoc can handle better.
	// Extract src and alt attributes, discard all other attributes (data-*, class, etc.).
	//
	// Pattern breakdown:
	// <img[^>]*           - Match <img tag with any attributes before src
	// \ssrc="([^"]*)"     - Capture src attribute value (required)
	// [^>]*               - Match any attributes between src and alt
	// (?:\salt="([^"]*)"|) - Optionally capture alt attribute value (non-capturing group with alternation)
	// [^>]*>              - Match remaining attributes and closing >
	imgPattern := regexp.MustCompile(`<img[^>]*\ssrc="([^"]*)"[^>]*(?:\salt="([^"]*)"|)[^>]*>`)
	html = imgPattern.ReplaceAllStringFunc(html, func(match string) string {
		srcMatch := regexp.MustCompile(`src="([^"]*)"`).FindStringSubmatch(match)
		altMatch := regexp.MustCompile(`alt="([^"]*)"`).FindStringSubmatch(match)
		src := ""
		alt := ""
		if len(srcMatch) > 1 {
			src = srcMatch[1]
		}
		if len(altMatch) > 1 {
			alt = altMatch[1]
		}
		if src == "" {
			return ""
		}
		return fmt.Sprintf(`<img src="%s" alt="%s">`, src, alt)
	})

	// Clean up table markup so pandoc can convert to markdown tables
	// Remove colgroup/col elements (pandoc doesn't need them)
	html = regexp.MustCompile(`(?i)<colgroup[^>]*>[\s\S]*?</colgroup>`).ReplaceAllString(html, "")
	html = regexp.MustCompile(`(?i)<col[^>]*/?\s*>`).ReplaceAllString(html, "")

	// Remove class and scope attributes from table elements
	html = regexp.MustCompile(`(<(?:table|thead|tbody|tr|th|td)[^>]*)\s+class="[^"]*"`).ReplaceAllString(html, "$1")
	html = regexp.MustCompile(`(<(?:th|td)[^>]*)\s+scope="[^"]*"`).ReplaceAllString(html, "$1")

	// Remove table-wrap divs
	html = regexp.MustCompile(`<div class="table-wrap"[^>]*>`).ReplaceAllString(html, "")

	// Simplify any remaining attributes on table elements
	html = regexp.MustCompile(`<table[^>]*>`).ReplaceAllString(html, "<table>")
	html = regexp.MustCompile(`<thead[^>]*>`).ReplaceAllString(html, "<thead>")
	html = regexp.MustCompile(`<tbody[^>]*>`).ReplaceAllString(html, "<tbody>")
	html = regexp.MustCompile(`<tr[^>]*>`).ReplaceAllString(html, "<tr>")
	html = regexp.MustCompile(`<th[^>]*>`).ReplaceAllString(html, "<th>")
	html = regexp.MustCompile(`<td[^>]*>`).ReplaceAllString(html, "<td>")

	// Remove <br> tags inside table cells (pandoc can't handle them and falls back to HTML)
	// Match <td>...<br>...</td> and <th>...<br>...</th> and remove the br
	html = regexp.MustCompile(`(<t[dh]>)([^<]*)<br\s*/?>([^<]*)(</t[dh]>)`).ReplaceAllString(html, "$1$2 $3$4")
	// Handle cells that are just <br>
	html = regexp.MustCompile(`<td>\s*<br\s*/?>\s*</td>`).ReplaceAllString(html, "<td></td>")
	html = regexp.MustCompile(`<th>\s*<br\s*/?>\s*</th>`).ReplaceAllString(html, "<th></th>")

	// Remove <p> tags inside table cells (unwrap content)
	// First handle simple single-p cells
	html = regexp.MustCompile(`(<t[dh]>)\s*<p>([^<]*)</p>\s*(</t[dh]>)`).ReplaceAllString(html, "$1$2$3")
	// Handle multiple <p> tags in cells - convert to text with spaces
	html = regexp.MustCompile(`(<t[dh]>)([\s\S]*?)(</t[dh]>)`).ReplaceAllStringFunc(html, func(match string) string {
		// Remove <p> and </p> tags inside cells, replace with space
		inner := regexp.MustCompile(`<t[dh]>`).ReplaceAllString(match, "")
		inner = regexp.MustCompile(`</t[dh]>`).ReplaceAllString(inner, "")
		inner = regexp.MustCompile(`<p[^>]*>`).ReplaceAllString(inner, "")
		inner = regexp.MustCompile(`</p>`).ReplaceAllString(inner, " ")
		inner = strings.TrimSpace(inner)
		// Detect if it was th or td
		if strings.HasPrefix(match, "<th") {
			return "<th>" + inner + "</th>"
		}
		return "<td>" + inner + "</td>"
	})

	// Remove span tags inside table cells (especially nolink spans)
	html = regexp.MustCompile(`<span[^>]*class="[^"]*nolink[^"]*"[^>]*>([\s\S]*?)</span>`).ReplaceAllString(html, "$1")
	// Remove status-macro and aui-message spans (keep content)
	html = regexp.MustCompile(`<span[^>]*class="[^"]*(?:status-macro|aui-message|aui-lozenge)[^"]*"[^>]*>([\s\S]*?)</span>`).ReplaceAllString(html, "$1")
	// Remove empty icon spans
	html = regexp.MustCompile(`<span[^>]*class="[^"]*icon[^"]*"[^>]*>\s*</span>`).ReplaceAllString(html, "")
	// Remove remaining spans
	html = regexp.MustCompile(`<span[^>]*>([\s\S]*?)</span>`).ReplaceAllString(html, "$1")

	// Remove content-wrapper divs (keep content)
	html = regexp.MustCompile(`<div[^>]*class="[^"]*content-wrapper[^"]*"[^>]*>([\s\S]*?)</div>`).ReplaceAllString(html, "$1")

	// Remove closing divs that match the layout containers we removed
	// Count opens vs closes and balance
	openCount := strings.Count(html, "<div")
	closeCount := strings.Count(html, "</div>")
	for closeCount > openCount {
		html = strings.Replace(html, "</div>", "", 1)
		closeCount--
	}

	return html
}

// postProcessMarkdown cleans up Confluence-specific HTML artifacts from the converted Markdown.
func postProcessMarkdown(md string) string {
	// Replace emoji images with Unicode characters
	emojiReplacements := map[string]string{
		`(tick)`:        "✅ ",
		`(error)`:       "❌ ",
		`(blue star)`:   "🚧",
		`(warning)`:     "⚠️ ",
		`(info)`:        "ℹ️ ",
		`(question)`:    "❓ ",
		`(plus)`:        "➕ ",
		`(minus)`:       "➖ ",
		`(on)`:          "💡 ",
		`(off)`:         "⭕ ",
		`(star)`:        "⭐ ",
		`(thumbs up)`:   "👍 ",
		`(thumbs down)`: "👎 ",
	}

	// Match <img> tags with alt attributes containing emoticon names
	imgPattern := regexp.MustCompile(`<img[^>]*alt="([^"]*)"[^>]*/?>`)
	md = imgPattern.ReplaceAllStringFunc(md, func(match string) string {
		submatches := imgPattern.FindStringSubmatch(match)
		if len(submatches) > 1 {
			alt := submatches[1]
			if replacement, ok := emojiReplacements[alt]; ok {
				return replacement
			}
		}
		// Remove other img tags (like expand-control-image)
		if strings.Contains(match, "expand-control-image") {
			return ""
		}
		return match
	})

	// Clean up Section1 div wrapper
	md = regexp.MustCompile(`<div class="Section1">\s*`).ReplaceAllString(md, "")

	// Remove Confluence table of contents wrapper but keep the content
	md = regexp.MustCompile(`<div class="toc-macro[^"]*"[^>]*>\s*`).ReplaceAllString(md, "")

	// Convert Confluence info/tip/warning/note macros to blockquotes
	macroPatterns := []struct {
		pattern     *regexp.Regexp
		replacement string
	}{
		{
			regexp.MustCompile(`<div class="confluence-information-macro confluence-information-macro-tip"[^>]*>\s*`),
			"\n> **Tip:** ",
		},
		{
			regexp.MustCompile(`<div class="confluence-information-macro confluence-information-macro-note"[^>]*>\s*`),
			"\n> **Note:** ",
		},
		{
			regexp.MustCompile(`<div class="confluence-information-macro confluence-information-macro-warning"[^>]*>\s*`),
			"\n> **Warning:** ",
		},
		{
			regexp.MustCompile(`<div class="confluence-information-macro confluence-information-macro-information"[^>]*>\s*`),
			"\n> **Info:** ",
		},
	}

	for _, mp := range macroPatterns {
		md = mp.pattern.ReplaceAllString(md, mp.replacement)
	}

	// Remove aui-icon spans
	md = regexp.MustCompile(`<span class="aui-icon[^"]*"[^>]*></span>\s*`).ReplaceAllString(md, "")

	// Clean up confluence-information-macro-body divs
	md = regexp.MustCompile(`<div class="confluence-information-macro-body">\s*`).ReplaceAllString(md, "")

	// Convert panel divs to blockquotes
	md = regexp.MustCompile(`<div class="panel"[^>]*>\s*`).ReplaceAllString(md, "\n> ")
	md = regexp.MustCompile(`<div class="panelContent"[^>]*>\s*`).ReplaceAllString(md, "")

	// Handle expander/collapsible sections
	md = regexp.MustCompile(`<div id="expander-\d+"[^>]*>\s*`).ReplaceAllString(md, "\n<details>\n")
	md = regexp.MustCompile(`<div id="expander-control-\d+"[^>]*>\s*`).ReplaceAllString(md, "<summary>")
	md = regexp.MustCompile(`<span class="expand-control-icon">[^<]*</span><span class="expand-control-text">([^<]*)</span>\s*`).ReplaceAllString(md, "$1")
	md = regexp.MustCompile(`<span class="expand-control-text">([^<]*)</span>\s*`).ReplaceAllString(md, "$1")
	md = regexp.MustCompile(`<span class="expand-control-icon">[^<]*</span>\s*`).ReplaceAllString(md, "")
	md = regexp.MustCompile(`<div id="expander-content-\d+"[^>]*>\s*`).ReplaceAllString(md, "</summary>\n")

	// Fix nested details tags
	md = regexp.MustCompile(`</summary>\s*\n\s*<details>\s*\n`).ReplaceAllString(md, "</summary>\n\n")
	md = regexp.MustCompile(`<details>\s*\n\x60\x60\x60`).ReplaceAllString(md, "\n```")

	// Clean up code panel divs and code headers
	md = regexp.MustCompile(`<div class="code panel[^"]*"[^>]*>\s*`).ReplaceAllString(md, "")
	md = regexp.MustCompile(`<div class="codeContent[^"]*"[^>]*>\s*`).ReplaceAllString(md, "")
	md = regexp.MustCompile(`<div class="codeHeader[^"]*"[^>]*>\s*`).ReplaceAllString(md, "")

	// Fix code block language hints
	md = strings.ReplaceAll(md, "``` syntaxhighlighter-pre", "```")
	md = regexp.MustCompile("```\\s*\\{[^}]*\\}").ReplaceAllString(md, "```")

	// Convert remaining HTML links to Markdown
	linkPattern := regexp.MustCompile(`<a\s+href="([^"]*)"[^>]*>([^<]*)</a>`)
	md = linkPattern.ReplaceAllString(md, "[$2]($1)")

	// Handle links with underline tags
	linkUnderlinePattern := regexp.MustCompile(`<a\s+href="([^"]*)"[^>]*><u>([^<]*)</u></a>`)
	md = linkUnderlinePattern.ReplaceAllString(md, "[$2]($1)")

	// Remove underline tags
	md = regexp.MustCompile(`</?u>`).ReplaceAllString(md, "")

	// Clean up closing divs - try to match groups first
	md = regexp.MustCompile(`</div>\s*</div>\s*</div>\s*`).ReplaceAllString(md, "\n</details>\n\n")
	md = regexp.MustCompile(`</div>\s*</div>\s*`).ReplaceAllString(md, "\n\n")
	md = regexp.MustCompile(`</div>`).ReplaceAllString(md, "")

	// Remove any remaining span tags
	md = regexp.MustCompile(`</?span[^>]*>`).ReplaceAllString(md, "")

	// Clean up HTML entities using the shared map
	for entity, char := range htmlEntityMap {
		md = strings.ReplaceAll(md, entity, char)
	}

	// Remove escaped HTML that pandoc didn't convert
	// These appear as \<tag\> or \</tag\>
	md = regexp.MustCompile(`\\<br\\?/?>`).ReplaceAllString(md, "\n")
	md = regexp.MustCompile(`\\</?p\\?>`).ReplaceAllString(md, "\n")
	md = regexp.MustCompile(`\\</?div[^>]*\\?>`).ReplaceAllString(md, "")
	md = regexp.MustCompile(`\\</?span[^>]*\\?>`).ReplaceAllString(md, "")

	// Handle escaped img tags - convert to markdown images.
	// Pandoc sometimes escapes HTML tags as \<tag\>. This pattern matches escaped <img> tags
	// and converts them to proper Markdown image syntax: ![alt](src)
	//
	// Pattern breakdown:
	// \\<img           - Match escaped opening: \<img
	// [^>]*src="..."   - Capture src attribute
	// (?:alt="..."|)   - Optionally capture alt attribute
	// \\?>             - Match optional escaped closing: \> or just >
	escapedImgPattern := regexp.MustCompile(`\\<img[^>]*src="([^"]*)"[^>]*(?:alt="([^"]*)"|)[^>]*\\?>`)
	md = escapedImgPattern.ReplaceAllStringFunc(md, func(match string) string {
		srcMatch := regexp.MustCompile(`src="([^"]*)"`).FindStringSubmatch(match)
		altMatch := regexp.MustCompile(`alt="([^"]*)"`).FindStringSubmatch(match)
		src := ""
		alt := "image"
		if len(srcMatch) > 1 {
			src = srcMatch[1]
		}
		if len(altMatch) > 1 && altMatch[1] != "" {
			alt = altMatch[1]
		}
		if src == "" {
			return ""
		}
		return fmt.Sprintf("![%s](%s)", alt, src)
	})

	// Clean any remaining escaped tags
	md = regexp.MustCompile(`\\<[^>]*\\?>`).ReplaceAllString(md, "")

	// Fix double-dash in nested lists (pandoc sometimes produces "- - item")
	md = regexp.MustCompile(`^(\s*)- - `).ReplaceAllString(md, "$1  - ")
	md = regexp.MustCompile(`\n(\s*)- - `).ReplaceAllString(md, "\n$1  - ")

	// Clean up remaining HTML tags in output
	// Remove any stray <br> tags
	md = regexp.MustCompile(`<br\s*/?>`).ReplaceAllString(md, "\n")
	// Remove empty <div> tags
	md = regexp.MustCompile(`<div[^>]*>\s*</div>`).ReplaceAllString(md, "")
	// Remove standalone closing </div> tags
	md = regexp.MustCompile(`</div>`).ReplaceAllString(md, "")

	// Normalize multiple blank lines to max 2
	md = regexp.MustCompile(`\n{3,}`).ReplaceAllString(md, "\n\n")

	// Trim trailing whitespace from lines
	lines := strings.Split(md, "\n")
	for i, line := range lines {
		lines[i] = strings.TrimRight(line, " \t")
	}
	md = strings.Join(lines, "\n")

	// Trim leading/trailing whitespace from document
	md = strings.TrimSpace(md) + "\n"

	// Remove orphaned </details> tags (not matched with opening tags)
	md = balanceDetailsTags(md)

	// Convert text emoji shortcodes like :celebration:
	textEmojis := map[string]string{
		":celebration:": "🎉",
		":thumbsup:":    "👍",
		":thumbsdown:":  "👎",
		":check:":       "✅",
		":cross:":       "❌",
		":warning:":     "⚠️",
		":info:":        "ℹ️",
		":question:":    "❓",
		":star:":        "⭐",
		":fire:":        "🔥",
		":rocket:":      "🚀",
		":sparkles:":    "✨",
	}
	for code, emoji := range textEmojis {
		md = strings.ReplaceAll(md, code, emoji)
	}

	return md
}

// balanceDetailsTags removes orphaned </details> tags that don't have matching opening tags.
func balanceDetailsTags(md string) string {
	// Remove excess closing tags from the end, recounting after each removal
	// to handle edge cases where removal creates new tags from surrounding chars
	for {
		openCount := strings.Count(md, "<details>")
		closeCount := strings.Count(md, "</details>")

		if closeCount <= openCount {
			break
		}

		lastIdx := strings.LastIndex(md, "</details>")
		if lastIdx == -1 {
			break
		}
		md = md[:lastIdx] + md[lastIdx+len("</details>"):]
	}

	return md
}
