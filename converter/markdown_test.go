package converter

import (
	"strings"
	"testing"
)

func TestCheckPandoc(t *testing.T) {
	// This test assumes pandoc is installed (as required by the tool)
	err := CheckPandoc()
	if err != nil {
		t.Skipf("Pandoc not installed, skipping test: %v", err)
	}
}

func TestConvertHTMLToMarkdown(t *testing.T) {
	// Skip if pandoc is not available
	if err := CheckPandoc(); err != nil {
		t.Skipf("Pandoc not installed, skipping test: %v", err)
	}

	tests := []struct {
		name     string
		html     string
		contains []string
	}{
		{
			name:     "basic heading",
			html:     "<html><body><h1>Test Heading</h1></body></html>",
			contains: []string{"# Test Heading"},
		},
		{
			name:     "paragraph",
			html:     "<html><body><p>This is a paragraph.</p></body></html>",
			contains: []string{"This is a paragraph."},
		},
		{
			name:     "link",
			html:     `<html><body><a href="https://example.com">Example</a></body></html>`,
			contains: []string{"[Example](https://example.com)"},
		},
		{
			name:     "code block",
			html:     "<html><body><pre><code>func main() {}</code></pre></body></html>",
			contains: []string{"func main() {}"},
		},
		{
			name:     "unordered list",
			html:     "<html><body><ul><li>Item 1</li><li>Item 2</li></ul></body></html>",
			contains: []string{"- Item 1", "- Item 2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			md, err := ConvertHTMLToMarkdown(tt.html)
			if err != nil {
				t.Fatalf("ConvertHTMLToMarkdown failed: %v", err)
			}

			for _, want := range tt.contains {
				if !strings.Contains(md, want) {
					t.Errorf("Expected markdown to contain %q, got: %s", want, md)
				}
			}
		})
	}
}

func TestPostProcessMarkdown_Emojis(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		expect string
	}{
		{
			name:   "tick emoji",
			input:  `<img src="test" alt="(tick)" />`,
			expect: "✅",
		},
		{
			name:   "error emoji",
			input:  `<img src="test" alt="(error)" class="emoticon"/>`,
			expect: "❌",
		},
		{
			name:   "blue star emoji",
			input:  `<img alt="(blue star)" src="test.png">`,
			expect: "🚧",
		},
		{
			name:   "celebration text emoji",
			input:  "Great job! :celebration:",
			expect: "Great job! 🎉",
		},
		{
			name:   "thumbsup text emoji",
			input:  "Thanks :thumbsup:",
			expect: "Thanks 👍",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := postProcessMarkdown(tt.input)
			if !strings.Contains(result, tt.expect) {
				t.Errorf("Expected result to contain %q, got: %s", tt.expect, result)
			}
		})
	}
}

func TestPostProcessMarkdown_ConfluenceMacros(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		expect string
	}{
		{
			name:   "tip macro",
			input:  `<div class="confluence-information-macro confluence-information-macro-tip"><div class="confluence-information-macro-body">This is a tip</div></div>`,
			expect: "> **Tip:**",
		},
		{
			name:   "note macro",
			input:  `<div class="confluence-information-macro confluence-information-macro-note"><div class="confluence-information-macro-body">This is a note</div></div>`,
			expect: "> **Note:**",
		},
		{
			name:   "warning macro",
			input:  `<div class="confluence-information-macro confluence-information-macro-warning"><div class="confluence-information-macro-body">This is a warning</div></div>`,
			expect: "> **Warning:**",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := postProcessMarkdown(tt.input)
			if !strings.Contains(result, tt.expect) {
				t.Errorf("Expected result to contain %q, got: %s", tt.expect, result)
			}
		})
	}
}

func TestPostProcessMarkdown_Links(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		expect string
	}{
		{
			name:   "simple link",
			input:  `<a href="https://example.com">Example</a>`,
			expect: "[Example](https://example.com)",
		},
		{
			name:   "link with attributes",
			input:  `<a href="https://example.com" class="external-link" rel="nofollow">Example</a>`,
			expect: "[Example](https://example.com)",
		},
		{
			name:   "link with underline",
			input:  `<a href="https://example.com"><u>Example</u></a>`,
			expect: "[Example](https://example.com)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := postProcessMarkdown(tt.input)
			if !strings.Contains(result, tt.expect) {
				t.Errorf("Expected result to contain %q, got: %s", tt.expect, result)
			}
		})
	}
}

func TestPostProcessMarkdown_HTMLEntities(t *testing.T) {
	input := "Tom &amp; Jerry &lt;3 &gt; love &quot;cheese&quot;"
	result := postProcessMarkdown(input)

	expects := []string{"Tom & Jerry", "<3", ">", `"cheese"`}
	for _, expect := range expects {
		if !strings.Contains(result, expect) {
			t.Errorf("Expected result to contain %q, got: %s", expect, result)
		}
	}
}

func TestPostProcessMarkdown_Section1Cleanup(t *testing.T) {
	input := `<div class="Section1">
# Heading
Content here
</div>`
	result := postProcessMarkdown(input)

	if strings.Contains(result, "Section1") {
		t.Errorf("Expected Section1 div to be removed, got: %s", result)
	}
	if !strings.Contains(result, "# Heading") {
		t.Errorf("Expected content to be preserved, got: %s", result)
	}
}

func TestPostProcessMarkdown_TOCCleanup(t *testing.T) {
	input := `<div class="toc-macro rbtoc1234567">
- [Heading 1](#heading-1)
- [Heading 2](#heading-2)
</div>`
	result := postProcessMarkdown(input)

	if strings.Contains(result, "toc-macro") {
		t.Errorf("Expected toc-macro div to be removed, got: %s", result)
	}
	if !strings.Contains(result, "[Heading 1]") {
		t.Errorf("Expected TOC content to be preserved, got: %s", result)
	}
}

func TestBalanceDetailsTags(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		expect string
	}{
		{
			name:   "balanced tags",
			input:  "<details>\nContent\n</details>",
			expect: "<details>\nContent\n</details>",
		},
		{
			name:   "orphan closing tag",
			input:  "Content\n</details>",
			expect: "Content\n",
		},
		{
			name:   "multiple orphan closing tags",
			input:  "<details>\nContent\n</details>\n</details>\n</details>",
			expect: "<details>\nContent\n</details>\n\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := balanceDetailsTags(tt.input)
			if result != tt.expect {
				t.Errorf("Expected %q, got %q", tt.expect, result)
			}
		})
	}
}

func TestPostProcessMarkdown_WhitespaceNormalization(t *testing.T) {
	input := "Line 1\n\n\n\n\nLine 2"
	result := postProcessMarkdown(input)

	// Should normalize to max 2 newlines
	if strings.Contains(result, "\n\n\n") {
		t.Errorf("Expected max 2 consecutive newlines, got: %q", result)
	}
}

func TestPreProcessHTML_LayoutDivs(t *testing.T) {
	input := `<div class="contentLayout2">
<div class="columnLayout two-left-sidebar" data-layout="two-left-sidebar">
<div class="cell aside" data-type="aside">
<div class="innerCell">
<h2>Problem</h2>
<p>Some content here.</p>
</div>
</div>
</div>
</div>`

	result := preProcessHTML(input)

	// Should remove layout divs
	if strings.Contains(result, "contentLayout2") {
		t.Errorf("Expected contentLayout2 to be removed, got: %s", result)
	}
	if strings.Contains(result, "columnLayout") {
		t.Errorf("Expected columnLayout to be removed, got: %s", result)
	}
	if strings.Contains(result, "innerCell") {
		t.Errorf("Expected innerCell to be removed, got: %s", result)
	}

	// Should preserve actual content
	if !strings.Contains(result, "<h2>Problem</h2>") {
		t.Errorf("Expected heading to be preserved, got: %s", result)
	}
	if !strings.Contains(result, "Some content here") {
		t.Errorf("Expected paragraph content to be preserved, got: %s", result)
	}
}

func TestPreProcessHTML_EmptyParagraphs(t *testing.T) {
	input := `<p></p><p>Real content</p><p><br></p><p>   </p>`

	result := preProcessHTML(input)

	if strings.Contains(result, "<p></p>") {
		t.Errorf("Expected empty paragraphs to be removed, got: %s", result)
	}
	if !strings.Contains(result, "Real content") {
		t.Errorf("Expected real content to be preserved, got: %s", result)
	}
}

func TestPreProcessHTML_StyleAttributes(t *testing.T) {
	input := `<p style="margin-left: 40.0px;">Indented text</p>`

	result := preProcessHTML(input)

	if strings.Contains(result, "style=") {
		t.Errorf("Expected style attribute to be removed, got: %s", result)
	}
	if !strings.Contains(result, "Indented text") {
		t.Errorf("Expected text content to be preserved, got: %s", result)
	}
}

func TestPreProcessHTML_DataAttributes(t *testing.T) {
	input := `<div data-layout="single" data-type="normal">Content</div>`

	result := preProcessHTML(input)

	if strings.Contains(result, "data-layout") {
		t.Errorf("Expected data-layout to be removed, got: %s", result)
	}
	if strings.Contains(result, "data-type") {
		t.Errorf("Expected data-type to be removed, got: %s", result)
	}
}

func TestPreProcessHTML_ImageSimplification(t *testing.T) {
	input := `<img class="confluence-embedded-image" draggable="false" width="468" src="abc123.png" data-image-src="/download/attachments/123/test.png" alt="Screenshot">`

	result := preProcessHTML(input)

	// Should simplify to basic img tag
	if strings.Contains(result, "confluence-embedded-image") {
		t.Errorf("Expected class to be removed, got: %s", result)
	}
	if strings.Contains(result, "draggable") {
		t.Errorf("Expected draggable to be removed, got: %s", result)
	}
	if !strings.Contains(result, `src="abc123.png"`) {
		t.Errorf("Expected src to be preserved, got: %s", result)
	}
	if !strings.Contains(result, `alt="Screenshot"`) {
		t.Errorf("Expected alt to be preserved, got: %s", result)
	}
}

func TestPostProcessMarkdown_EscapedHTML(t *testing.T) {
	tests := []struct {
		name             string
		input            string
		shouldNotContain []string
	}{
		{
			name:             "escaped br tags",
			input:            `Text\<br\>more text`,
			shouldNotContain: []string{`\<br\>`},
		},
		{
			name:             "escaped div tags",
			input:            `\<div class="test"\>content\</div\>`,
			shouldNotContain: []string{`\<div`, `\</div`},
		},
		{
			name:             "escaped p tags",
			input:            `\<p\>paragraph\</p\>`,
			shouldNotContain: []string{`\<p\>`, `\</p\>`},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := postProcessMarkdown(tt.input)
			for _, shouldNot := range tt.shouldNotContain {
				if strings.Contains(result, shouldNot) {
					t.Errorf("Expected %q to be removed, got: %s", shouldNot, result)
				}
			}
		})
	}
}

func TestDecodeHTMLEntities(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		expect string
	}{
		{
			name:   "no entities - passthrough",
			input:  "plain text without entities",
			expect: "plain text without entities",
		},
		{
			name:   "basic entities",
			input:  "&lt;div&gt;content&lt;/div&gt;",
			expect: "<div>content</div>",
		},
		{
			name:   "ampersand entity with trigger",
			input:  "&lt;Tom &amp; Jerry&gt;",
			expect: "<Tom & Jerry>",
		},
		{
			name:   "quote entities with trigger",
			input:  "&#60;&quot;quoted&quot;&#62;",
			expect: `<"quoted">`,
		},
		{
			name:   "apos entity with lt trigger",
			input:  "&lt;&apos;apostrophe&apos;&gt;",
			expect: "<'apostrophe'>",
		},
		{
			name:   "hex entity for less than",
			input:  "&#x3C;tag&#x3E;",
			expect: "<tag>",
		},
		{
			name:   "decimal entity for less than",
			input:  "&#60;tag&#62;",
			expect: "<tag>",
		},
		{
			name:   "hex entity uppercase",
			input:  "&#x3c;lower&#x3e;",
			expect: "<lower>",
		},
		{
			name:   "mixed entities",
			input:  "&lt;p&gt;Hello &amp; &#x27;world&#x27;&lt;/p&gt;",
			expect: "<p>Hello & 'world'</p>",
		},
		{
			name:   "nbsp entity with lt trigger",
			input:  "&lt;word&nbsp;word&gt;",
			expect: "<word word>",
		},
		{
			name:   "high codepoint unchanged",
			input:  "&#200;", // È - above ASCII range
			expect: "&#200;",
		},
		{
			name:   "hex high codepoint unchanged",
			input:  "&#xC8;", // È - above ASCII range
			expect: "&#xC8;",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := decodeHTMLEntities(tt.input)
			if result != tt.expect {
				t.Errorf("Expected %q, got %q", tt.expect, result)
			}
		})
	}
}

func TestPreProcessHTML_Tables(t *testing.T) {
	input := `<table class="confluenceTable" data-layout="default">
<colgroup><col style="width: 50%"><col style="width: 50%"></colgroup>
<thead><tr><th class="confluenceTh" scope="col">Header 1</th><th class="confluenceTh">Header 2</th></tr></thead>
<tbody><tr><td class="confluenceTd">Cell 1</td><td class="confluenceTd">Cell 2</td></tr></tbody>
</table>`

	result := preProcessHTML(input)

	// Colgroup should be removed
	if strings.Contains(result, "colgroup") {
		t.Errorf("Expected colgroup to be removed, got: %s", result)
	}
	// Class attributes should be removed from table elements
	if strings.Contains(result, "confluenceTable") {
		t.Errorf("Expected confluenceTable class to be removed, got: %s", result)
	}
	if strings.Contains(result, "confluenceTh") {
		t.Errorf("Expected confluenceTh class to be removed, got: %s", result)
	}
	// Content should be preserved
	if !strings.Contains(result, "Header 1") {
		t.Errorf("Expected header content to be preserved, got: %s", result)
	}
	if !strings.Contains(result, "Cell 1") {
		t.Errorf("Expected cell content to be preserved, got: %s", result)
	}
}

func TestPreProcessHTML_TableCellBreaks(t *testing.T) {
	input := `<td>Line 1<br/>Line 2</td><th><br></th>`

	result := preProcessHTML(input)

	// br inside cells should be removed/converted to space
	if strings.Contains(result, "<br") {
		t.Errorf("Expected br tags in cells to be removed, got: %s", result)
	}
}

func TestPreProcessHTML_TableCellParagraphs(t *testing.T) {
	input := `<td><p>Paragraph content</p></td>`

	result := preProcessHTML(input)

	// p tags inside cells should be unwrapped
	if strings.Contains(result, "<p>") {
		t.Errorf("Expected p tags in cells to be unwrapped, got: %s", result)
	}
	if !strings.Contains(result, "Paragraph content") {
		t.Errorf("Expected content to be preserved, got: %s", result)
	}
}

func TestPreProcessHTML_SpanCleanup(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		expect string
	}{
		{
			name:   "nolink span",
			input:  `<span class="nolink">text</span>`,
			expect: "text",
		},
		{
			name:   "status macro span",
			input:  `<span class="status-macro aui-lozenge">STATUS</span>`,
			expect: "STATUS",
		},
		{
			name:   "empty icon span",
			input:  `<span class="icon aui-icon">  </span>`,
			expect: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := preProcessHTML(tt.input)
			result = strings.TrimSpace(result)
			if result != tt.expect {
				t.Errorf("Expected %q, got %q", tt.expect, result)
			}
		})
	}
}

func TestPreProcessHTML_PluginElements(t *testing.T) {
	input := `<div class="plugin_pagetree">Tree content</div>
<div class="plugin_pagetree_children">Child content</div>`

	result := preProcessHTML(input)

	// plugin_pagetree divs should be removed (opening tag only, content preserved)
	if strings.Contains(result, `class="plugin_pagetree"`) {
		t.Errorf("Expected plugin_pagetree class to be removed, got: %s", result)
	}
}

func TestPreProcessHTML_DoubleEncodedHTML(t *testing.T) {
	// Confluence sometimes double-encodes HTML
	input := `&lt;p&gt;This was double encoded&lt;/p&gt;`

	result := preProcessHTML(input)

	if !strings.Contains(result, "<p>") {
		t.Errorf("Expected double-encoded HTML to be decoded, got: %s", result)
	}
}

func TestPostProcessMarkdown_ExpanderSections(t *testing.T) {
	input := `<div id="expander-123"><div id="expander-control-123"><span class="expand-control-icon">+</span><span class="expand-control-text">Click to expand</span></div><div id="expander-content-123">Hidden content here</div></div>`

	result := postProcessMarkdown(input)

	if !strings.Contains(result, "<details>") {
		t.Errorf("Expected expander to be converted to details, got: %s", result)
	}
	if !strings.Contains(result, "<summary>") || !strings.Contains(result, "</summary>") {
		t.Errorf("Expected summary tags, got: %s", result)
	}
	if !strings.Contains(result, "Click to expand") {
		t.Errorf("Expected expand text to be preserved, got: %s", result)
	}
}

func TestPostProcessMarkdown_PanelDivs(t *testing.T) {
	input := `<div class="panel" style="border-width: 1px;"><div class="panelContent">Panel content here</div></div>`

	result := postProcessMarkdown(input)

	if !strings.Contains(result, ">") {
		t.Errorf("Expected panel to be converted to blockquote, got: %s", result)
	}
	if !strings.Contains(result, "Panel content here") {
		t.Errorf("Expected panel content to be preserved, got: %s", result)
	}
}

func TestPostProcessMarkdown_CodeBlocks(t *testing.T) {
	tests := []struct {
		name             string
		input            string
		shouldNotContain string
	}{
		{
			name:             "syntaxhighlighter class",
			input:            "``` syntaxhighlighter-pre\ncode here\n```",
			shouldNotContain: "syntaxhighlighter-pre",
		},
		{
			name:             "code panel divs",
			input:            `<div class="code panel pdl"><div class="codeContent panelContent pdl">code</div></div>`,
			shouldNotContain: "code panel",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := postProcessMarkdown(tt.input)
			if strings.Contains(result, tt.shouldNotContain) {
				t.Errorf("Expected %q to be removed, got: %s", tt.shouldNotContain, result)
			}
		})
	}
}

func TestPostProcessMarkdown_AuiIcons(t *testing.T) {
	input := `<span class="aui-icon aui-icon-small aui-iconfont-approve"></span> Approved`

	result := postProcessMarkdown(input)

	if strings.Contains(result, "aui-icon") {
		t.Errorf("Expected aui-icon span to be removed, got: %s", result)
	}
	if !strings.Contains(result, "Approved") {
		t.Errorf("Expected text to be preserved, got: %s", result)
	}
}

func TestPostProcessMarkdown_EscapedImages(t *testing.T) {
	input := `\<img src="test.png" alt="Test Image"\>`

	result := postProcessMarkdown(input)

	if strings.Contains(result, `\<img`) {
		t.Errorf("Expected escaped img to be converted, got: %s", result)
	}
	if !strings.Contains(result, "![") {
		t.Errorf("Expected markdown image syntax, got: %s", result)
	}
}

func TestPostProcessMarkdown_NestedListFix(t *testing.T) {
	input := "- - Item 1\n- - Item 2"

	result := postProcessMarkdown(input)

	if strings.Contains(result, "- - ") {
		t.Errorf("Expected double dash to be fixed, got: %s", result)
	}
}

func TestPostProcessMarkdown_BrTagCleanup(t *testing.T) {
	input := "Line 1<br>Line 2<br/>Line 3<br />Line 4"

	result := postProcessMarkdown(input)

	if strings.Contains(result, "<br") {
		t.Errorf("Expected br tags to be converted to newlines, got: %s", result)
	}
}

func TestPostProcessMarkdown_InfoMacro(t *testing.T) {
	input := `<div class="confluence-information-macro confluence-information-macro-information"><div class="confluence-information-macro-body">Info content</div></div>`

	result := postProcessMarkdown(input)

	if !strings.Contains(result, "> **Info:**") {
		t.Errorf("Expected info macro to be converted, got: %s", result)
	}
}

func TestConvertHTMLToMarkdown_ComplexDocument(t *testing.T) {
	// Skip if pandoc is not available
	if err := CheckPandoc(); err != nil {
		t.Skipf("Pandoc not installed, skipping test: %v", err)
	}

	html := `<html>
<body>
<h1>Document Title</h1>
<p>Introduction paragraph.</p>
<h2>Section 1</h2>
<ul>
<li>Item 1</li>
<li>Item 2</li>
<li>Item 3</li>
</ul>
<h2>Section 2</h2>
<table>
<tr><th>Header A</th><th>Header B</th></tr>
<tr><td>Cell 1</td><td>Cell 2</td></tr>
</table>
<p>Final paragraph with <strong>bold</strong> and <em>italic</em> text.</p>
</body>
</html>`

	md, err := ConvertHTMLToMarkdown(html)
	if err != nil {
		t.Fatalf("ConvertHTMLToMarkdown failed: %v", err)
	}

	// Verify key elements are present
	expects := []string{"# Document Title", "## Section 1", "- Item 1", "## Section 2", "**bold**", "*italic*"}
	for _, want := range expects {
		if !strings.Contains(md, want) {
			t.Errorf("Expected markdown to contain %q, got: %s", want, md)
		}
	}
}

func TestConvertHTMLToMarkdown_WithExpanders(t *testing.T) {
	if err := CheckPandoc(); err != nil {
		t.Skipf("Pandoc not installed, skipping test: %v", err)
	}

	html := `<html><body>
<div id="expander-1">
<div id="expander-control-1">
<span class="expand-control-icon">+</span>
<span class="expand-control-text">Show More</span>
</div>
<div id="expander-content-1">
<p>Hidden content that can be expanded.</p>
</div>
</div>
</body></html>`

	md, err := ConvertHTMLToMarkdown(html)
	if err != nil {
		t.Fatalf("ConvertHTMLToMarkdown failed: %v", err)
	}

	if !strings.Contains(md, "<details>") || !strings.Contains(md, "<summary>") {
		t.Errorf("Expected expander to be converted to details/summary, got: %s", md)
	}
}

func TestConvertHTMLToMarkdown_WithInfoMacros(t *testing.T) {
	if err := CheckPandoc(); err != nil {
		t.Skipf("Pandoc not installed, skipping test: %v", err)
	}

	html := `<html><body>
<div class="confluence-information-macro confluence-information-macro-tip">
<div class="confluence-information-macro-body">
<p>This is a tip for users.</p>
</div>
</div>
</body></html>`

	md, err := ConvertHTMLToMarkdown(html)
	if err != nil {
		t.Fatalf("ConvertHTMLToMarkdown failed: %v", err)
	}

	if !strings.Contains(md, "> **Tip:**") {
		t.Errorf("Expected tip macro to be converted, got: %s", md)
	}
}

func TestConvertHTMLToMarkdown_WithCodeBlock(t *testing.T) {
	if err := CheckPandoc(); err != nil {
		t.Skipf("Pandoc not installed, skipping test: %v", err)
	}

	html := `<html><body>
<pre><code class="language-go">package main

func main() {
    fmt.Println("Hello, World!")
}
</code></pre>
</body></html>`

	md, err := ConvertHTMLToMarkdown(html)
	if err != nil {
		t.Fatalf("ConvertHTMLToMarkdown failed: %v", err)
	}

	if !strings.Contains(md, "func main()") {
		t.Errorf("Expected code to be preserved, got: %s", md)
	}
}

func TestBalanceDetailsTags_EdgeCases(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		expect string
	}{
		{
			name:   "no details tags",
			input:  "Just plain text without any details tags",
			expect: "Just plain text without any details tags",
		},
		{
			name:   "only opening tag",
			input:  "<details>Content without closing",
			expect: "<details>Content without closing",
		},
		{
			name:   "multiple balanced pairs",
			input:  "<details>First</details><details>Second</details>",
			expect: "<details>First</details><details>Second</details>",
		},
		{
			name:   "nested details",
			input:  "<details><details>Nested</details></details>",
			expect: "<details><details>Nested</details></details>",
		},
		{
			name:   "extra orphan at middle and end",
			input:  "<details>Content</details></details>More text</details>",
			expect: "<details>Content</details>More text",
		},
		{
			name:   "removal creates new tag from surrounding chars",
			input:  "<</details>/details>",
			expect: "", // First removal creates "</details>", second removal clears it
		},
		{
			name:   "multiple removals create new tags",
			input:  "<</details>/details></details>",
			expect: "", // All orphaned closing tags are removed
		},
		{
			name:   "preserves content around removed tags",
			input:  "Hello</details>World",
			expect: "HelloWorld",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := balanceDetailsTags(tt.input)
			if result != tt.expect {
				t.Errorf("Expected %q, got %q", tt.expect, result)
			}
		})
	}
}

func TestPreProcessHTML_ComplexTable(t *testing.T) {
	input := `<table class="confluenceTable wrapped" data-table-width="100%">
<colgroup>
<col style="width: 33.333%;">
<col style="width: 33.333%;">
<col style="width: 33.334%;">
</colgroup>
<thead>
<tr role="row">
<th class="confluenceTh" scope="col" data-highlight-colour="#F0F0F0">
<p>Column A</p>
</th>
<th class="confluenceTh" scope="col">
<p>Column B</p>
</th>
<th class="confluenceTh" scope="col">
<p>Column C</p>
</th>
</tr>
</thead>
<tbody>
<tr role="row">
<td class="confluenceTd">
<p>Data 1<br/>Line 2</p>
</td>
<td class="confluenceTd">
<p>Data 2</p>
</td>
<td class="confluenceTd">
<p>Data 3</p>
</td>
</tr>
</tbody>
</table>`

	result := preProcessHTML(input)

	// Verify cleanup happened
	if strings.Contains(result, "colgroup") {
		t.Error("Expected colgroup to be removed")
	}
	if strings.Contains(result, "confluenceTable") {
		t.Error("Expected confluenceTable class to be removed")
	}
	if strings.Contains(result, "data-table-width") {
		t.Error("Expected data-table-width to be removed")
	}

	// Verify content preserved
	if !strings.Contains(result, "Column A") || !strings.Contains(result, "Data 1") {
		t.Error("Expected table content to be preserved")
	}
}

func TestPostProcessMarkdown_AllEmojis(t *testing.T) {
	// Test all text emoji shortcodes defined in the textEmojis map
	tests := []struct {
		input  string
		expect string
	}{
		{`:thumbsup:`, "👍"},
		{`:thumbsdown:`, "👎"},
		{`:star:`, "⭐"},
		{`:fire:`, "🔥"},
		{`:rocket:`, "🚀"},
		{`:warning:`, "⚠️"},
		{`:check:`, "✅"},
		{`:cross:`, "❌"},
		{`:celebration:`, "🎉"},
		{`:sparkles:`, "✨"},
		{`:info:`, "ℹ️"},
		{`:question:`, "❓"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := postProcessMarkdown(tt.input)
			if !strings.Contains(result, tt.expect) {
				t.Errorf("Expected %q to contain %q, got: %s", tt.input, tt.expect, result)
			}
		})
	}
}

func TestPreProcessHTML_UserIcons(t *testing.T) {
	input := `<span class="confluence-userlink" data-username="john.doe">
<span class="user-icon">
<span class="aui-avatar aui-avatar-small"><span class="aui-avatar-inner"><img src="avatar.png" alt=""></span></span>
</span>
<span class="user-name">John Doe</span>
</span>`

	result := preProcessHTML(input)

	// User name should be preserved
	if !strings.Contains(result, "John Doe") {
		t.Errorf("Expected user name to be preserved, got: %s", result)
	}
}

func TestPreProcessHTML_Emoticons(t *testing.T) {
	input := `<img class="emoticon emoticon-tick" src="tick.png" alt="(tick)" data-emoticon-name="tick">
<img class="emoticon emoticon-cross" src="cross.png" alt="(error)" data-emoticon-name="cross">`

	result := preProcessHTML(input)

	// Emoticons with alt text should be preserved (for later emoji conversion)
	if !strings.Contains(result, "(tick)") && !strings.Contains(result, "✅") {
		t.Errorf("Expected tick emoticon reference to be preserved, got: %s", result)
	}
}

func TestDecodeHTMLEntities_EdgeCases(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		expect string
	}{
		{
			name:   "entity at end of string",
			input:  "&lt;test&gt;",
			expect: "<test>",
		},
		{
			name:   "multiple consecutive entities",
			input:  "&lt;&gt;&amp;",
			expect: "<>&",
		},
		{
			name:   "entity with no match",
			input:  "&unknown; &lt;test&gt;",
			expect: "&unknown; <test>",
		},
		{
			name:   "numeric entity at boundary",
			input:  "&#126; &#127; &#128;", // 126 is ~ (decoded), 127+ not decoded (val < 127 check)
			expect: "~ &#127; &#128;",
		},
		{
			name:   "low ascii numeric entities",
			input:  "&#65; &#66; &#67;", // A, B, C
			expect: "A B C",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := decodeHTMLEntities(tt.input)
			if result != tt.expect {
				t.Errorf("Expected %q, got %q", tt.expect, result)
			}
		})
	}
}
