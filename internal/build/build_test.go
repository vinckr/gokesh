package build

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSplitBodyAndFrontmatter(t *testing.T) {
	t.Parallel()

	md := []byte("---\ntitle: \"Test Page\"\n---\n# Hello\n\nWorld")
	body, matter := SplitBodyAndFrontmatter(md)

	if matter["title"] != "Test Page" {
		t.Errorf("title = %q, want %q", matter["title"], "Test Page")
	}
	if !strings.Contains(string(body), "# Hello") {
		t.Errorf("body missing heading: %s", string(body))
	}
}

func TestWriteHTMLFile(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	if err := WriteHTMLFile("test.md", dir+string(filepath.Separator), "<html>hello</html>"); err != nil {
		t.Fatalf("WriteHTMLFile: %v", err)
	}

	// non-index pages are written as <name>/index.html for clean URLs
	outPath := filepath.Join(dir, "test", "index.html")
	content, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("output file not written: %v", err)
	}
	if string(content) != "<html>hello</html>" {
		t.Errorf("file content = %q, want %q", string(content), "<html>hello</html>")
	}
}

func TestBuildTemplate(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	pageTmpl := `{{define "Page"}}{{template "Body" .}}{{end}}`
	bodyTmpl := `{{define "Body"}}<article>{{.Body}}</article>{{end}}`
	writeFile(t, filepath.Join(dir, "page.tmpl"), pageTmpl)
	writeFile(t, filepath.Join(dir, "body.tmpl"), bodyTmpl)

	var d pageData
	d.Body = "<p>hello</p>"
	result, err := BuildTemplate(d, filepath.Join(dir, "page.tmpl"), filepath.Join(dir, "body.tmpl"))
	if err != nil {
		t.Fatalf("BuildTemplate: %v", err)
	}

	if !strings.Contains(result, "<article><p>hello</p></article>") {
		t.Errorf("unexpected template output: %s", result)
	}
}

func TestBuildPagesSkipsNonMarkdown(t *testing.T) {
	t.Parallel()

	mdDir := t.TempDir()
	outDir := t.TempDir()
	tmplDir := t.TempDir()

	// valid markdown file
	writeFile(t, filepath.Join(mdDir, "post.md"), "---\ntitle: \"Post\"\n---\nHello")
	// non-markdown files that must be silently skipped
	writeFile(t, filepath.Join(mdDir, "style.css"), "body{}")
	writeFile(t, filepath.Join(mdDir, "data.json"), `{"k":"v"}`)

	pageTmpl := `{{define "Page"}}{{.Body}}{{end}}`
	writeFile(t, filepath.Join(tmplDir, "page.tmpl"), pageTmpl)

	cfg := Config{SiteTitle: "Test"}
	if err := BuildPages(mdDir+string(filepath.Separator), outDir+string(filepath.Separator), cfg, tmplDir+string(filepath.Separator)); err != nil {
		t.Fatalf("BuildPages: %v", err)
	}

	// only post/index.html should exist; css and json must not produce output
	if _, err := os.ReadFile(filepath.Join(outDir, "post", "index.html")); err != nil {
		t.Errorf("expected post/index.html to exist: %v", err)
	}
	// non-.md files must be skipped — no output directory should be created for them
	if _, err := os.ReadFile(filepath.Join(outDir, "style.css", "index.html")); err == nil {
		t.Error("style.css should not produce output")
	}
	if _, err := os.ReadFile(filepath.Join(outDir, "data.json", "index.html")); err == nil {
		t.Error("data.json should not produce output")
	}
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("writeFile %s: %v", path, err)
	}
}
