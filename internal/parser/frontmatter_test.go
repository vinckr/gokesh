package parser

import (
	"testing"
	"time"
)

func TestParseFrontmatter(t *testing.T) {
	t.Parallel()

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
		{
			name:  "value containing colon",
			input: "---\ntitle: \"Go: A Tour\"\n---\nbody",
			wantMatter: map[string]string{"title": "Go: A Tour"},
			wantBody:   "body",
		},
		{
			name:  "unquoted value containing colon",
			input: "---\nbase_url: https://example.com\n---\nbody",
			wantMatter: map[string]string{"base_url": "https://example.com"},
			wantBody:   "body",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
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

func TestParseTypedFrontmatter(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		check func(t *testing.T, f FrontmatterFields)
	}{
		{
			name:  "bool draft true",
			input: "---\ndraft: true\n---\n",
			check: func(t *testing.T, f FrontmatterFields) {
				if !f.Draft {
					t.Error("Draft should be true")
				}
			},
		},
		{
			name:  "bool draft false",
			input: "---\ndraft: false\n---\n",
			check: func(t *testing.T, f FrontmatterFields) {
				if f.Draft {
					t.Error("Draft should be false")
				}
			},
		},
		{
			name:  "date valid",
			input: "---\ndate: 2026-01-15\n---\n",
			check: func(t *testing.T, f FrontmatterFields) {
				want := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)
				if !f.Date.Equal(want) {
					t.Errorf("Date = %v, want %v", f.Date, want)
				}
			},
		},
		{
			name:  "date invalid is zero",
			input: "---\ndate: not-a-date\n---\n",
			check: func(t *testing.T, f FrontmatterFields) {
				if !f.Date.IsZero() {
					t.Errorf("Date should be zero for invalid input, got %v", f.Date)
				}
			},
		},
		{
			name:  "tags array",
			input: "---\ntags: [go, web, ssr]\n---\n",
			check: func(t *testing.T, f FrontmatterFields) {
				want := []string{"go", "web", "ssr"}
				if len(f.Tags) != len(want) {
					t.Fatalf("Tags len = %d, want %d: %v", len(f.Tags), len(want), f.Tags)
				}
				for i, tag := range want {
					if f.Tags[i] != tag {
						t.Errorf("Tags[%d] = %q, want %q", i, f.Tags[i], tag)
					}
				}
			},
		},
		{
			name:  "tags empty array",
			input: "---\ntags: []\n---\n",
			check: func(t *testing.T, f FrontmatterFields) {
				if len(f.Tags) != 0 {
					t.Errorf("Tags should be empty, got %v", f.Tags)
				}
			},
		},
		{
			name:  "all fields together",
			input: "---\ntitle: \"My Post\"\ndate: 2026-03-01\ndraft: true\ntags: [go, testing]\ndescription: \"A great post\"\nslug: my-post\n---\nbody",
			check: func(t *testing.T, f FrontmatterFields) {
				if f.Title != "My Post" {
					t.Errorf("Title = %q, want %q", f.Title, "My Post")
				}
				if f.Draft != true {
					t.Error("Draft should be true")
				}
				if f.Date.IsZero() {
					t.Error("Date should not be zero")
				}
				if len(f.Tags) != 2 {
					t.Errorf("Tags len = %d, want 2", len(f.Tags))
				}
				if f.Description != "A great post" {
					t.Errorf("Description = %q", f.Description)
				}
				if f.Slug != "my-post" {
					t.Errorf("Slug = %q", f.Slug)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			f, _ := ParseTypedFrontmatter([]byte(tt.input))
			tt.check(t, f)
		})
	}
}
