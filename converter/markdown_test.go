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
