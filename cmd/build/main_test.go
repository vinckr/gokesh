package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSplitBodyAndFrontmatter(t *testing.T) {
	md := []byte("---\npagetitle: \"Test Page\"\n---\n# Hello\n\nWorld")
	body, matter := SplitBodyAndFrontmatter(md)

	if matter["pagetitle"] != "Test Page" {
		t.Errorf("pagetitle = %q, want %q", matter["pagetitle"], "Test Page")
	}
	if !strings.Contains(string(body), "# Hello") {
		t.Errorf("body missing heading: %s", string(body))
	}
}

func TestWriteHTMLFile(t *testing.T) {
	dir := t.TempDir()
	WriteHTMLFile("test.md", dir+string(filepath.Separator), "<html>hello</html>")

	outPath := filepath.Join(dir, "test.html")
	content, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("output file not written: %v", err)
	}
	if string(content) != "<html>hello</html>" {
		t.Errorf("file content = %q, want %q", string(content), "<html>hello</html>")
	}
}

func TestBuildTemplate(t *testing.T) {
	// Write minimal templates to a temp dir
	dir := t.TempDir()

	pageTmpl := `{{define "Page"}}{{template "Body" .}}{{end}}`
	bodyTmpl := `{{define "Body"}}<article>{{.Body}}</article>{{end}}`
	writeFile(t, filepath.Join(dir, "page.tmpl"), pageTmpl)
	writeFile(t, filepath.Join(dir, "body.tmpl"), bodyTmpl)

	var d data
	d.Body = "<p>hello</p>"
	result := BuildTemplate(d, filepath.Join(dir, "page.tmpl"), filepath.Join(dir, "body.tmpl"))

	if !strings.Contains(result, "<article><p>hello</p></article>") {
		t.Errorf("unexpected template output: %s", result)
	}
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("writeFile %s: %v", path, err)
	}
}
