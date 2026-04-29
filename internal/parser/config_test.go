package parser

import "testing"

func TestParseConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  map[string]string
	}{
		{
			name:  "double-quoted values",
			input: "author = \"vinckr\"\nsite_title = \"My Blog\"",
			want:  map[string]string{"author": "vinckr", "site_title": "My Blog"},
		},
		{
			name:  "single-quoted values",
			input: "author = 'vinckr'",
			want:  map[string]string{"author": "vinckr"},
		},
		{
			name:  "unquoted values",
			input: "author = vinckr",
			want:  map[string]string{"author": "vinckr"},
		},
		{
			name:  "comments and blank lines ignored",
			input: "# this is a comment\n\nauthor = \"vinckr\"",
			want:  map[string]string{"author": "vinckr"},
		},
		{
			name:  "all config fields",
			input: "author = \"vinckr\"\nsite_title = \"My Blog\"\nbase_url = \"https://example.com\"\ndescription = \"A blog\"",
			want: map[string]string{
				"author":      "vinckr",
				"site_title":  "My Blog",
				"base_url":    "https://example.com",
				"description": "A blog",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := ParseConfig([]byte(tt.input))
			for k, want := range tt.want {
				if got[k] != want {
					t.Errorf("key %q = %q, want %q", k, got[k], want)
				}
			}
		})
	}
}
