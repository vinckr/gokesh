package parser

import (
	"strings"
	"testing"
)

func TestToHTML(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		contains []string // substrings that must appear in output
		absent   []string // substrings that must NOT appear
	}{
		// Headings
		{
			name:     "h1",
			input:    "# Heading One",
			contains: []string{"<h1>Heading One</h1>"},
		},
		{
			name:     "h2",
			input:    "## Heading Two",
			contains: []string{"<h2>Heading Two</h2>"},
		},
		{
			name:     "h3",
			input:    "### Heading Three",
			contains: []string{"<h3>Heading Three</h3>"},
		},
		{
			name:     "h4",
			input:    "#### Heading Four",
			contains: []string{"<h4>Heading Four</h4>"},
		},
		{
			name:     "h5",
			input:    "##### Heading Five",
			contains: []string{"<h5>Heading Five</h5>"},
		},
		{
			name:     "h6",
			input:    "###### Heading Six",
			contains: []string{"<h6>Heading Six</h6>"},
		},
		// Inline formatting
		{
			name:     "bold",
			input:    "**bold text**",
			contains: []string{"<strong>bold text</strong>"},
		},
		{
			name:     "italic underscore",
			input:    "_italic text_",
			contains: []string{"<em>italic text</em>"},
		},
		{
			name:     "italic star",
			input:    "*italic text*",
			contains: []string{"<em>italic text</em>"},
		},
		{
			name:     "bold italic",
			input:    "**_bold italic_**",
			contains: []string{"<strong><em>bold italic</em></strong>"},
		},
		{
			name:     "strikethrough",
			input:    "~~struck~~",
			contains: []string{"<del>struck</del>"},
		},
		{
			name:     "inline code",
			input:    "`code here`",
			contains: []string{"<code>code here</code>"},
		},
		// Links and images
		{
			name:     "link",
			input:    "[example](https://example.com)",
			contains: []string{`<a href="https://example.com">example</a>`},
		},
		{
			name:     "image",
			input:    "![alt text](https://example.com/img.png)",
			contains: []string{`<img src="https://example.com/img.png" alt="alt text">`},
		},
		// Lists
		{
			name:  "unordered list",
			input: "- Item 1\n- Item 2\n- Item 3",
			contains: []string{
				"<ul>",
				"<li>Item 1</li>",
				"<li>Item 2</li>",
				"<li>Item 3</li>",
				"</ul>",
			},
		},
		{
			name:  "ordered list",
			input: "1. Item A\n2. Item B\n3. Item C",
			contains: []string{
				"<ol>",
				"<li>Item A</li>",
				"<li>Item B</li>",
				"<li>Item C</li>",
				"</ol>",
			},
		},
		// Blockquote
		{
			name:     "blockquote",
			input:    "> This is a quote",
			contains: []string{"<blockquote>", "<p>This is a quote</p>", "</blockquote>"},
		},
		// Fenced code block
		{
			name:     "fenced code block",
			input:    "```\nfunc main() {}\n```",
			contains: []string{"<pre><code>", "func main() {}", "</code></pre>"},
		},
		{
			name:   "fenced code block no inline processing",
			input:  "```\n**not bold**\n```",
			contains: []string{"**not bold**"},
			absent:   []string{"<strong>"},
		},
		// Table
		{
			name:  "table",
			input: "| Col A | Col B |\n| ----- | ----- |\n| R1C1  | R1C2  |",
			contains: []string{
				"<table>",
				"<thead>",
				"<th>Col A</th>",
				"<th>Col B</th>",
				"</thead>",
				"<tbody>",
				"<td>R1C1</td>",
				"<td>R1C2</td>",
				"</tbody>",
				"</table>",
			},
		},
		// Paragraph
		{
			name:     "paragraph",
			input:    "This is a paragraph.",
			contains: []string{"<p>This is a paragraph.</p>"},
		},
		// Horizontal rule
		{
			name:     "horizontal rule ---",
			input:    "---",
			contains: []string{"<hr>"},
			absent:   []string{"<p>"},
		},
		{
			name:     "horizontal rule ***",
			input:    "***",
			contains: []string{"<hr>"},
		},
		// Inline code protects content from other inline processing
		{
			name:     "inline code protects bold syntax",
			input:    "`**not bold**`",
			contains: []string{"<code>**not bold**</code>"},
			absent:   []string{"<strong>"},
		},
		{
			name:     "inline code protects strikethrough syntax",
			input:    "`~~not struck~~`",
			contains: []string{"<code>~~not struck~~</code>"},
			absent:   []string{"<del>"},
		},
		{
			name:     "inline code escapes HTML angle brackets",
			input:    "`<div class=\"foo\">`",
			contains: []string{`<code>&lt;div class="foo"&gt;</code>`},
			absent:   []string{"<div "},
		},
		{
			name:     "inline code escapes HTML comment",
			input:    "`<!-- comment -->`",
			contains: []string{"<code>&lt;!-- comment --&gt;</code>"},
		},
		// Paragraph wraps multiple consecutive lines
		{
			name:     "multi-line paragraph",
			input:    "Line one\nLine two",
			contains: []string{"<p>Line one Line two</p>"},
		},
		// Blank line separates blocks
		{
			name:  "two paragraphs",
			input: "First paragraph.\n\nSecond paragraph.",
			contains: []string{
				"<p>First paragraph.</p>",
				"<p>Second paragraph.</p>",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := string(ToHTML([]byte(tt.input)))
			for _, want := range tt.contains {
				if !strings.Contains(got, want) {
					t.Errorf("output missing %q\nfull output:\n%s", want, got)
				}
			}
			for _, bad := range tt.absent {
				if strings.Contains(got, bad) {
					t.Errorf("output should not contain %q\nfull output:\n%s", bad, got)
				}
			}
		})
	}
}
