package build

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
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

func TestDraftPagesAreSkipped(t *testing.T) {
	t.Parallel()

	mdDir := t.TempDir()
	outDir := t.TempDir()
	tmplDir := t.TempDir()

	writeFile(t, filepath.Join(mdDir, "published.md"), "---\ntitle: \"Published\"\n---\nHello")
	writeFile(t, filepath.Join(mdDir, "draft.md"), "---\ntitle: \"Draft\"\ndraft: true\n---\nSecret")

	pageTmpl := `{{define "Page"}}{{.Body}}{{end}}`
	writeFile(t, filepath.Join(tmplDir, "page.tmpl"), pageTmpl)

	cfg := Config{SiteTitle: "Test"}
	if err := BuildAll(mdDir, outDir, cfg, tmplDir+string(filepath.Separator)); err != nil {
		t.Fatalf("BuildAll: %v", err)
	}

	if _, err := os.ReadFile(filepath.Join(outDir, "published", "index.html")); err != nil {
		t.Errorf("expected published/index.html: %v", err)
	}
	if _, err := os.ReadFile(filepath.Join(outDir, "draft", "index.html")); err == nil {
		t.Error("draft page should not produce output")
	}
}

func TestPagesGlobalContainsNonDrafts(t *testing.T) {
	t.Parallel()

	mdDir := t.TempDir()
	outDir := t.TempDir()
	tmplDir := t.TempDir()

	writeFile(t, filepath.Join(mdDir, "post.md"), "---\ntitle: \"Post\"\ndate: 2026-01-15\n---\nHello")
	writeFile(t, filepath.Join(mdDir, "draft.md"), "---\ntitle: \"Draft\"\ndraft: true\n---\nSecret")
	writeFile(t, filepath.Join(mdDir, "index.md"), "---\ntitle: \"Home\"\n---\n{{range .Pages}}{{.Title}}{{end}}")

	pageTmpl := `{{define "Page"}}{{range .Pages}}{{.Title}},{{end}}{{end}}`
	writeFile(t, filepath.Join(tmplDir, "page.tmpl"), pageTmpl)

	cfg := Config{SiteTitle: "Test"}
	if err := BuildAll(mdDir, outDir, cfg, tmplDir+string(filepath.Separator)); err != nil {
		t.Fatalf("BuildAll: %v", err)
	}

	// .Pages must not contain the draft
	content, err := os.ReadFile(filepath.Join(outDir, "index.html"))
	if err != nil {
		t.Fatalf("index.html missing: %v", err)
	}
	if strings.Contains(string(content), "Draft") {
		t.Errorf("draft page should not appear in .Pages, got: %s", string(content))
	}
	if !strings.Contains(string(content), "Post") {
		t.Errorf("post should appear in .Pages, got: %s", string(content))
	}
}

func TestCopyStatic(t *testing.T) {
	t.Parallel()

	staticDir := t.TempDir()
	outDir := t.TempDir()

	// nested file to verify subdirectory preservation
	if err := os.MkdirAll(filepath.Join(staticDir, "img"), 0755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(staticDir, "img", "logo.png"), "PNG")
	writeFile(t, filepath.Join(staticDir, "robots.txt"), "User-agent: *")

	if err := CopyStatic(staticDir, outDir); err != nil {
		t.Fatalf("CopyStatic: %v", err)
	}

	check := func(rel, want string) {
		t.Helper()
		data, err := os.ReadFile(filepath.Join(outDir, rel))
		if err != nil {
			t.Errorf("expected %s: %v", rel, err)
			return
		}
		if string(data) != want {
			t.Errorf("%s = %q, want %q", rel, string(data), want)
		}
	}
	check(filepath.Join("img", "logo.png"), "PNG")
	check("robots.txt", "User-agent: *")
}

func TestCopyStaticMissingDirIsNoop(t *testing.T) {
	t.Parallel()
	if err := CopyStatic("/nonexistent-static-dir", t.TempDir()); err != nil {
		t.Errorf("missing static dir should not error: %v", err)
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

func TestTemplateFunctions(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	pageTmpl := `{{define "Page"}}` +
		`date:{{dateFormat .Pagematter.Date "2006-01-02"}}|` +
		`sorted:{{range sortBy .Pages "date"}}{{.Title}},{{end}}|` +
		`tagged:{{range filterByTag .Pages "go"}}{{.Title}},{{end}}` +
		`{{end}}`
	writeFile(t, filepath.Join(dir, "page.tmpl"), pageTmpl)

	t1 := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	t2 := time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC)

	var d pageData
	d.Pagematter.Date = t1
	d.Pages = []PageSummary{
		{Title: "Older", Date: t2, Tags: []string{"go"}},
		{Title: "Newer", Date: t1, Tags: []string{"web"}},
	}

	result, err := BuildTemplate(d, filepath.Join(dir, "page.tmpl"))
	if err != nil {
		t.Fatalf("BuildTemplate: %v", err)
	}

	if !strings.Contains(result, "date:2026-01-01") {
		t.Errorf("dateFormat not working, got: %s", result)
	}
	if !strings.Contains(result, "sorted:Newer,Older,") {
		t.Errorf("sortBy not working, got: %s", result)
	}
	if !strings.Contains(result, "tagged:Older,") {
		t.Errorf("filterByTag not working, got: %s", result)
	}
}

func TestGenerateRSS(t *testing.T) {
	t.Parallel()

	outDir := t.TempDir()
	pages := []PageSummary{
		{Title: "Post One", URL: "https://example.com/post-one/", Date: time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)},
		{Title: "Undated Post", URL: "https://example.com/undated/"},
	}

	if err := GenerateRSS(pages, outDir, "My Blog", "https://example.com"); err != nil {
		t.Fatalf("GenerateRSS: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(outDir, "feed.xml"))
	if err != nil {
		t.Fatalf("feed.xml missing: %v", err)
	}
	content := string(data)

	if !strings.Contains(content, "<rss") {
		t.Error("missing <rss> element")
	}
	if !strings.Contains(content, "Post One") {
		t.Error("missing Post One in feed")
	}
	if strings.Contains(content, "Undated Post") {
		t.Error("undated pages should not appear in RSS feed")
	}
}

func TestGenerateRSSSkipsWhenNoBaseURL(t *testing.T) {
	t.Parallel()
	outDir := t.TempDir()
	if err := GenerateRSS(nil, outDir, "Site", ""); err != nil {
		t.Errorf("should not error with empty base_url: %v", err)
	}
	if _, err := os.ReadFile(filepath.Join(outDir, "feed.xml")); err == nil {
		t.Error("feed.xml should not be written when base_url is empty")
	}
}

func TestShouldRebuild(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	src := filepath.Join(dir, "page.md")
	out := filepath.Join(dir, "page", "index.html")

	writeFile(t, src, "# hello")

	// output doesn't exist yet — should rebuild
	if !shouldRebuild(src, out) {
		t.Error("should rebuild when output missing")
	}

	// create output older than source
	if err := os.MkdirAll(filepath.Join(dir, "page"), 0755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, out, "<h1>hello</h1>")

	// backdate output so source is newer
	old := time.Now().Add(-time.Hour)
	if err := os.Chtimes(out, old, old); err != nil {
		t.Fatal(err)
	}
	if !shouldRebuild(src, out) {
		t.Error("should rebuild when source is newer than output")
	}

	// now make output newer than source
	future := time.Now().Add(time.Hour)
	if err := os.Chtimes(out, future, future); err != nil {
		t.Fatal(err)
	}
	if shouldRebuild(src, out) {
		t.Error("should not rebuild when output is newer than source")
	}
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("writeFile %s: %v", path, err)
	}
}
