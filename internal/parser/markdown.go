package parser

import (
	"regexp"
	"strings"
)

// ToHTML converts a Markdown byte slice to HTML.
// Supports: headings (h1-h6), bold, italic, bold+italic, strikethrough,
// unordered lists, ordered lists, blockquotes, fenced code blocks,
// tables, links, images, inline code, and paragraphs.
func ToHTML(src []byte) []byte {
	lines := strings.Split(string(src), "\n")
	var out strings.Builder

	type blockKind int
	const (
		bNone blockKind = iota
		bParagraph
		bUL
		bOL
		bBlockquote
		bCode
		bTable
	)

	current := bNone
	var buf []string // accumulated lines for current block

	flush := func() {
		if current == bNone || len(buf) == 0 {
			current = bNone
			buf = buf[:0]
			return
		}
		switch current {
		case bParagraph:
			out.WriteString("<p>")
			out.WriteString(inlineHTML(strings.Join(buf, "\n")))
			out.WriteString("</p>\n")
		case bUL:
			out.WriteString("<ul>\n")
			for _, item := range buf {
				out.WriteString("<li>")
				out.WriteString(inlineHTML(item))
				out.WriteString("</li>\n")
			}
			out.WriteString("</ul>\n")
		case bOL:
			out.WriteString("<ol>\n")
			for _, item := range buf {
				out.WriteString("<li>")
				out.WriteString(inlineHTML(item))
				out.WriteString("</li>\n")
			}
			out.WriteString("</ol>\n")
		case bBlockquote:
			out.WriteString("<blockquote>\n<p>")
			out.WriteString(inlineHTML(strings.Join(buf, "\n")))
			out.WriteString("</p>\n</blockquote>\n")
		case bTable:
			renderTable(&out, buf)
		}
		current = bNone
		buf = buf[:0]
	}

	inFence := false

	for _, line := range lines {
		// Handle fenced code blocks
		if strings.HasPrefix(line, "```") {
			if !inFence {
				flush()
				inFence = true
				out.WriteString("<pre><code>")
				continue
			}
			// closing fence
			inFence = false
			out.WriteString("</code></pre>\n")
			continue
		}
		if inFence {
			out.WriteString(escapeHTML(line))
			out.WriteString("\n")
			continue
		}

		// Blank line: flush current block
		if strings.TrimSpace(line) == "" {
			flush()
			continue
		}

		// Headings
		if strings.HasPrefix(line, "#") {
			flush()
			level := 0
			for level < len(line) && line[level] == '#' {
				level++
			}
			if level > 6 {
				level = 6
			}
			text := strings.TrimSpace(line[level:])
			out.WriteString("<h" + itoa(level) + ">")
			out.WriteString(inlineHTML(text))
			out.WriteString("</h" + itoa(level) + ">\n")
			continue
		}

		// Unordered list item
		if strings.HasPrefix(line, "- ") || strings.HasPrefix(line, "* ") {
			if current != bUL {
				flush()
				current = bUL
			}
			buf = append(buf, strings.TrimSpace(line[2:]))
			continue
		}

		// Ordered list item (matches "1. ", "2. " etc.)
		if isOrderedItem(line) {
			if current != bOL {
				flush()
				current = bOL
			}
			buf = append(buf, orderedItemText(line))
			continue
		}

		// Blockquote
		if strings.HasPrefix(line, "> ") {
			if current != bBlockquote {
				flush()
				current = bBlockquote
			}
			buf = append(buf, strings.TrimSpace(line[2:]))
			continue
		}

		// Table row
		if strings.HasPrefix(line, "|") {
			// Skip separator rows (e.g. | --- | --- |)
			if isTableSeparator(line) {
				continue
			}
			if current != bTable {
				flush()
				current = bTable
			}
			buf = append(buf, line)
			continue
		}

		// Default: paragraph
		if current != bParagraph {
			flush()
			current = bParagraph
		}
		if len(buf) == 0 {
			buf = append(buf, line)
		} else {
			buf[0] = buf[0] + " " + line
		}
	}

	flush()
	return []byte(out.String())
}

// renderTable writes a <table> element from accumulated row lines.
// The first row becomes <thead>, remaining rows become <tbody>.
func renderTable(out *strings.Builder, rows []string) {
	if len(rows) == 0 {
		return
	}
	out.WriteString("<table>\n<thead>\n<tr>")
	for _, cell := range splitTableRow(rows[0]) {
		out.WriteString("<th>")
		out.WriteString(inlineHTML(cell))
		out.WriteString("</th>")
	}
	out.WriteString("</tr>\n</thead>\n")
	if len(rows) > 1 {
		out.WriteString("<tbody>\n")
		for _, row := range rows[1:] {
			out.WriteString("<tr>")
			for _, cell := range splitTableRow(row) {
				out.WriteString("<td>")
				out.WriteString(inlineHTML(cell))
				out.WriteString("</td>")
			}
			out.WriteString("</tr>\n")
		}
		out.WriteString("</tbody>\n")
	}
	out.WriteString("</table>\n")
}

// splitTableRow splits a table row string into cell strings.
func splitTableRow(line string) []string {
	line = strings.Trim(line, "|")
	parts := strings.Split(line, "|")
	cells := make([]string, len(parts))
	for i, p := range parts {
		cells[i] = strings.TrimSpace(p)
	}
	return cells
}

// isTableSeparator returns true for divider rows like | --- | --- |
func isTableSeparator(line string) bool {
	line = strings.Trim(line, "| \t")
	for _, ch := range line {
		if ch != '-' && ch != ':' && ch != '|' && ch != ' ' {
			return false
		}
	}
	return true
}

// isOrderedItem returns true if line starts with a number followed by ". "
func isOrderedItem(line string) bool {
	i := 0
	for i < len(line) && line[i] >= '0' && line[i] <= '9' {
		i++
	}
	return i > 0 && i < len(line)-1 && line[i] == '.' && line[i+1] == ' '
}

// orderedItemText strips the leading "N. " prefix from an ordered list item.
func orderedItemText(line string) string {
	i := 0
	for i < len(line) && line[i] >= '0' && line[i] <= '9' {
		i++
	}
	return strings.TrimSpace(line[i+2:])
}

// Inline patterns — order matters
var (
	reImage        = regexp.MustCompile(`!\[([^\]]*)\]\(([^)]+)\)`)
	reLink         = regexp.MustCompile(`\[([^\]]+)\]\(([^)]+)\)`)
	reInlineCode   = regexp.MustCompile("`([^`]+)`")
	reBoldItalic1  = regexp.MustCompile(`\*\*_(.+?)_\*\*`)
	reBoldItalic2  = regexp.MustCompile(`_\*\*(.+?)\*\*_`)
	reBold         = regexp.MustCompile(`\*\*(.+?)\*\*`)
	reItalicUScore = regexp.MustCompile(`_(.+?)_`)
	reItalicStar   = regexp.MustCompile(`\*(.+?)\*`)
	reStrike       = regexp.MustCompile(`~~(.+?)~~`)
)

// inlineHTML applies inline Markdown transformations to a string.
func inlineHTML(s string) string {
	s = reImage.ReplaceAllString(s, `<img src="$2" alt="$1">`)
	s = reLink.ReplaceAllString(s, `<a href="$2">$1</a>`)
	s = reInlineCode.ReplaceAllString(s, `<code>$1</code>`)
	s = reBoldItalic1.ReplaceAllString(s, `<strong><em>$1</em></strong>`)
	s = reBoldItalic2.ReplaceAllString(s, `<strong><em>$1</em></strong>`)
	s = reBold.ReplaceAllString(s, `<strong>$1</strong>`)
	s = reItalicUScore.ReplaceAllString(s, `<em>$1</em>`)
	s = reItalicStar.ReplaceAllString(s, `<em>$1</em>`)
	s = reStrike.ReplaceAllString(s, `<del>$1</del>`)
	return s
}

// escapeHTML escapes characters that are significant in HTML.
func escapeHTML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	return s
}

// itoa converts a small integer to its string representation without fmt.
func itoa(n int) string {
	return string(rune('0' + n))
}
