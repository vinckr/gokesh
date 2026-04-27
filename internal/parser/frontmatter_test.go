package parser

import (
	"testing"
)

func TestParseFrontmatter(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantMatter  map[string]string
		wantBody    string
	}{
		{
			name:  "quoted value",
			input: "---\npagetitle: \"Hello World\"\n---\nbody content",
			wantMatter: map[string]string{"pagetitle": "Hello World"},
			wantBody:   "body content",
		},
		{
			name:  "unquoted value",
			input: "---\npagetitle: Hello World\n---\nbody content",
			wantMatter: map[string]string{"pagetitle": "Hello World"},
			wantBody:   "body content",
		},
		{
			name:  "single-quoted value",
			input: "---\npagetitle: 'Hello World'\n---\nbody content",
			wantMatter: map[string]string{"pagetitle": "Hello World"},
			wantBody:   "body content",
		},
		{
			name:       "no frontmatter",
			input:      "just body content",
			wantMatter: map[string]string{},
			wantBody:   "just body content",
		},
		{
			name:       "empty frontmatter block",
			input:      "---\n---\nbody",
			wantMatter: map[string]string{},
			wantBody:   "body",
		},
		{
			name:  "multiple keys",
			input: "---\npagetitle: \"My Page\"\nauthor: vinckr\n---\ncontent here",
			wantMatter: map[string]string{
				"pagetitle": "My Page",
				"author":    "vinckr",
			},
			wantBody: "content here",
		},
		{
			name:       "no closing delimiter",
			input:      "---\npagetitle: test\nbody without closing",
			wantMatter: map[string]string{},
			wantBody:   "---\npagetitle: test\nbody without closing",
		},
		{
			name:  "multiline body preserved",
			input: "---\npagetitle: \"Test\"\n---\nline one\nline two\nline three",
			wantMatter: map[string]string{"pagetitle": "Test"},
			wantBody:   "line one\nline two\nline three",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matter, body := ParseFrontmatter([]byte(tt.input))

			for k, want := range tt.wantMatter {
				got, ok := matter[k]
				if !ok {
					t.Errorf("key %q missing from matter", k)
					continue
				}
				if got != want {
					t.Errorf("matter[%q] = %q, want %q", k, got, want)
				}
			}
			for k := range matter {
				if _, ok := tt.wantMatter[k]; !ok {
					t.Errorf("unexpected key %q in matter", k)
				}
			}

			if string(body) != tt.wantBody {
				t.Errorf("body = %q, want %q", string(body), tt.wantBody)
			}
		})
	}
}
