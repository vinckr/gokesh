package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunNewCreatesFile(t *testing.T) {
	t.Parallel()

	mdDir := t.TempDir()
	if err := runNew("my-post", mdDir+string(filepath.Separator)); err != nil {
		t.Fatalf("runNew: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(mdDir, "my-post.md"))
	if err != nil {
		t.Fatalf("file not created: %v", err)
	}
	content := string(data)

	if !strings.Contains(content, `title: "my-post"`) {
		t.Errorf("missing title in frontmatter:\n%s", content)
	}
	if !strings.Contains(content, "draft: true") {
		t.Errorf("missing draft: true in frontmatter:\n%s", content)
	}
	if !strings.Contains(content, "date: ") {
		t.Errorf("missing date in frontmatter:\n%s", content)
	}
}

func TestRunNewStripsExtension(t *testing.T) {
	t.Parallel()

	mdDir := t.TempDir()
	// passing "post.md" should not create "post.md.md"
	if err := runNew("post.md", mdDir+string(filepath.Separator)); err != nil {
		t.Fatalf("runNew: %v", err)
	}
	if _, err := os.ReadFile(filepath.Join(mdDir, "post.md")); err != nil {
		t.Fatalf("post.md not found: %v", err)
	}
	if _, err := os.ReadFile(filepath.Join(mdDir, "post.md.md")); err == nil {
		t.Error("double extension post.md.md should not exist")
	}
}

func TestRunNewFailsIfExists(t *testing.T) {
	t.Parallel()

	mdDir := t.TempDir()
	writeFileHelper(t, filepath.Join(mdDir, "existing.md"), "already here")

	err := runNew("existing", mdDir+string(filepath.Separator))
	if err == nil {
		t.Error("expected error when file already exists, got nil")
	}
}

func TestRunNewCreatesSubdirectory(t *testing.T) {
	t.Parallel()

	mdDir := t.TempDir()
	// name with path separator to ensure directory is created
	if err := runNew("blog/my-post", mdDir+string(filepath.Separator)); err != nil {
		t.Fatalf("runNew with subdir: %v", err)
	}
	if _, err := os.ReadFile(filepath.Join(mdDir, "blog", "my-post.md")); err != nil {
		t.Fatalf("blog/my-post.md not created: %v", err)
	}
}

func writeFileHelper(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("writeFileHelper %s: %v", path, err)
	}
}
