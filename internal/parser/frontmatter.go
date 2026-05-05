package parser

import (
	"log/slog"
	"strings"
	"time"
)

// FrontmatterFields is the typed representation of a page's frontmatter.
// It is populated by ParseTypedFrontmatter.
type FrontmatterFields struct {
	Title       string
	Description string
	Slug        string
	Template    string
	Data        string
	Date        time.Time
	Tags        []string
	Draft       bool
}

// ParseTypedFrontmatter parses frontmatter into a typed struct plus the body.
// It builds on ParseFrontmatter for delimiter handling and adds type coercion.
func ParseTypedFrontmatter(content []byte) (FrontmatterFields, []byte) {
	matter, body := ParseFrontmatter(content)
	var f FrontmatterFields

	f.Title = matter["title"]
	f.Description = matter["description"]
	f.Slug = matter["slug"]
	f.Template = matter["template"]
	f.Data = matter["data"]

	if v := matter["draft"]; v == "true" {
		f.Draft = true
	}

	if v := matter["date"]; v != "" {
		t, err := time.Parse("2006-01-02", v)
		if err != nil {
			slog.Warn("invalid date in frontmatter", "value", v)
		} else {
			f.Date = t
		}
	}

	if v := matter["tags"]; v != "" {
		f.Tags = parseTags(v)
	}

	return f, body
}

// parseTags parses "[go, web, ssr]" into []string{"go", "web", "ssr"}.
func parseTags(s string) []string {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "[")
	s = strings.TrimSuffix(s, "]")
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	tags := make([]string, 0, len(parts))
	for _, p := range parts {
		if tag := strings.TrimSpace(p); tag != "" {
			tags = append(tags, tag)
		}
	}
	return tags
}

// Parse extracts YAML frontmatter from markdown content.
// Frontmatter is a block of key: value pairs between two --- delimiters at the
// start of the content. Only flat string values are supported.
// Returns a map of keys to values and the remaining body content.
// If no frontmatter is present, the map is empty and body is the full input.
func ParseFrontmatter(content []byte) (matter map[string]string, body []byte) {
	matter = make(map[string]string)
	s := string(content)

	// Must start with ---\n
	if !strings.HasPrefix(s, "---\n") {
		return matter, content
	}

	// Find the closing ---
	// After the opening ---\n, the closing may be at position 0 (empty block)
	// or preceded by a newline.
	rest := s[4:] // skip opening ---\n

	var yamlBlock string
	var afterClose string

	if strings.HasPrefix(rest, "---\n") {
		// Empty frontmatter: closing --- immediately follows
		yamlBlock = ""
		afterClose = rest[4:]
	} else if strings.HasPrefix(rest, "---") && len(rest) == 3 {
		// Closing --- at very end with no trailing newline
		yamlBlock = ""
		afterClose = ""
	} else {
		before, after, found := strings.Cut(rest, "\n---\n")
		if !found {
			// Try end of string: closing --- at very end (no trailing newline)
			if strings.HasSuffix(rest, "\n---") {
				yamlBlock = rest[:len(rest)-4]
				afterClose = ""
			} else {
				// No closing delimiter found — treat whole content as body
				return matter, content
			}
		} else {
			yamlBlock = before
			afterClose = after
		}
	}

	parseYAMLLines(yamlBlock, matter)
	body = []byte(afterClose)
	return matter, body
}

// parseYAMLLines parses simple flat key: value lines into the map.
// Supports quoted ("value") and unquoted values.
func parseYAMLLines(block string, matter map[string]string) {
	for _, line := range strings.Split(block, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		key, val, found := strings.Cut(line, ":")
		if !found {
			continue
		}
		key = strings.TrimSpace(key)
		val = strings.TrimSpace(val)
		// Strip surrounding quotes
		if len(val) >= 2 && val[0] == '"' && val[len(val)-1] == '"' {
			val = val[1 : len(val)-1]
		} else if len(val) >= 2 && val[0] == '\'' && val[len(val)-1] == '\'' {
			val = val[1 : len(val)-1]
		}
		matter[key] = val
	}
}
