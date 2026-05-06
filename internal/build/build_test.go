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

func TestCollectPagesSortedNewestFirst(t *testing.T) {
	t.Parallel()

	mdDir := t.TempDir()
	writeFile(t, filepath.Join(mdDir, "old.md"), "---\ntitle: \"Old\"\ndate: 2024-01-01\n---\n")
	writeFile(t, filepath.Join(mdDir, "new.md"), "---\ntitle: \"New\"\ndate: 2026-06-01\n---\n")
	writeFile(t, filepath.Join(mdDir, "undated.md"), "---\ntitle: \"Undated\"\n---\n")
	writeFile(t, filepath.Join(mdDir, "draft.md"), "---\ntitle: \"Draft\"\ndraft: true\ndate: 2026-01-01\n---\n")

	pages := CollectPages(mdDir, "https://example.com")

	// draft must be excluded
	for _, p := range pages {
		if p.Title == "Draft" {
			t.Error("draft page must not appear in CollectPages result")
		}
	}

	// must be 3 non-draft pages
	if len(pages) != 3 {
		t.Fatalf("got %d pages, want 3: %v", len(pages), pages)
	}

	// newest-first: New (2026) → Old (2024) → Undated (zero date last)
	if pages[0].Title != "New" {
		t.Errorf("pages[0] = %q, want \"New\"", pages[0].Title)
	}
	if pages[1].Title != "Old" {
		t.Errorf("pages[1] = %q, want \"Old\"", pages[1].Title)
	}
	if pages[2].Title != "Undated" {
		t.Errorf("pages[2] = %q, want \"Undated\"", pages[2].Title)
	}
}

func TestCollectPagesURL(t *testing.T) {
	t.Parallel()

	mdDir := t.TempDir()
	writeFile(t, filepath.Join(mdDir, "index.md"), "---\ntitle: \"Home\"\n---\n")
	writeFile(t, filepath.Join(mdDir, "about.md"), "---\ntitle: \"About\"\n---\n")

	pages := CollectPages(mdDir, "https://example.com")

	urls := map[string]string{}
	for _, p := range pages {
		urls[p.Title] = p.URL
	}

	if urls["Home"] != "https://example.com/" {
		t.Errorf("index URL = %q, want %q", urls["Home"], "https://example.com/")
	}
	if urls["About"] != "https://example.com/about/" {
		t.Errorf("about URL = %q, want %q", urls["About"], "https://example.com/about/")
	}
}

func TestBuildAllIncrementalSkipsUnchanged(t *testing.T) {
	t.Parallel()

	mdDir := t.TempDir()
	outDir := t.TempDir()
	tmplDir := t.TempDir()

	writeFile(t, filepath.Join(mdDir, "page.md"), "---\ntitle: \"Page\"\n---\nHello")
	writeFile(t, filepath.Join(tmplDir, "page.tmpl"), `{{define "Page"}}{{.Body}}{{end}}`)

	cfg := Config{SiteTitle: "Test"}

	// first build
	if err := BuildAllIncremental(mdDir, outDir, cfg, tmplDir+string(filepath.Separator)); err != nil {
		t.Fatalf("first BuildAllIncremental: %v", err)
	}
	outFile := filepath.Join(outDir, "page", "index.html")
	fi1, err := os.Stat(outFile)
	if err != nil {
		t.Fatalf("output not created: %v", err)
	}

	// second incremental build — output is newer than source, so mtime must be unchanged
	if err := BuildAllIncremental(mdDir, outDir, cfg, tmplDir+string(filepath.Separator)); err != nil {
		t.Fatalf("second BuildAllIncremental: %v", err)
	}
	fi2, err := os.Stat(outFile)
	if err != nil {
		t.Fatalf("output missing after second build: %v", err)
	}
	if !fi2.ModTime().Equal(fi1.ModTime()) {
		t.Error("incremental build rewrote unchanged output — should have skipped it")
	}
}

func TestLoadConfigDefaults(t *testing.T) {
	t.Parallel()

	// missing file → default dirs
	cfg, err := LoadConfig("/nonexistent-path/gokesh.toml")
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}
	if cfg.MarkdownDir != "markdown" {
		t.Errorf("MarkdownDir default = %q, want %q", cfg.MarkdownDir, "markdown")
	}
	if cfg.OutputDir != "public" {
		t.Errorf("OutputDir default = %q, want %q", cfg.OutputDir, "public")
	}
}

func TestLoadConfigCustomDirs(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "gokesh.toml")
	writeFile(t, path, "output_dir = \"dist\"\nmarkdown_dir = \"content\"\n")

	cfg, err := LoadConfig(path)
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}
	if cfg.OutputDir != "dist" {
		t.Errorf("OutputDir = %q, want %q", cfg.OutputDir, "dist")
	}
	if cfg.MarkdownDir != "content" {
		t.Errorf("MarkdownDir = %q, want %q", cfg.MarkdownDir, "content")
	}
}

func TestPageURLHelper(t *testing.T) {
	t.Parallel()

	cases := []struct {
		rel     string
		baseURL string
		want    string
	}{
		{"index.md", "https://example.com", "https://example.com/"},
		{"about.md", "https://example.com", "https://example.com/about/"},
		{"blog/post.md", "https://example.com", "https://example.com/blog/post/"},
		{"index.md", "", "/"},
		{"about.md", "", "/about/"},
	}

	for _, c := range cases {
		got := pageURL(c.rel, c.baseURL)
		if got != c.want {
			t.Errorf("pageURL(%q, %q) = %q, want %q", c.rel, c.baseURL, got, c.want)
		}
	}
}

func TestDateFormatZeroTime(t *testing.T) {
	t.Parallel()
	if got := dateFormat(time.Time{}, "2006-01-02"); got != "" {
		t.Errorf("dateFormat(zero) = %q, want empty string", got)
	}
}

func TestSortByTitle(t *testing.T) {
	t.Parallel()
	pages := []PageSummary{
		{Title: "Zebra"},
		{Title: "Apple"},
		{Title: "Mango"},
	}
	sorted := sortBy(pages, "title")
	if sorted[0].Title != "Apple" || sorted[1].Title != "Mango" || sorted[2].Title != "Zebra" {
		t.Errorf("sortBy title = %v", sorted)
	}
	// original slice must not be modified
	if pages[0].Title != "Zebra" {
		t.Error("sortBy must not modify the original slice")
	}
}

func TestFilterByTagNoMatch(t *testing.T) {
	t.Parallel()
	pages := []PageSummary{
		{Title: "A", Tags: []string{"go"}},
		{Title: "B", Tags: []string{"rust"}},
	}
	result := filterByTag(pages, "python")
	if len(result) != 0 {
		t.Errorf("filterByTag no-match = %v, want empty", result)
	}
}

func TestBuildPageAtNewFieldsWired(t *testing.T) {
	t.Parallel()

	mdDir := t.TempDir()
	outDir := t.TempDir()
	tmplDir := t.TempDir()

	writeFile(t, filepath.Join(mdDir, "post.md"),
		"---\ntitle: \"My Post\"\ndescription: \"Great post\"\ndate: 2026-03-15\ntags: [go, testing]\nslug: my-post\n---\nHello")

	pageTmpl := `{{define "Page"}}` +
		`title:{{.Pagematter.PageTitle}}|` +
		`desc:{{.Pagematter.Description}}|` +
		`date:{{dateFormat .Pagematter.Date "2006-01-02"}}|` +
		`tags:{{range .Pagematter.Tags}}{{.}},{{end}}|` +
		`slug:{{.Pagematter.Slug}}` +
		`{{end}}`
	writeFile(t, filepath.Join(tmplDir, "page.tmpl"), pageTmpl)

	cfg := Config{SiteTitle: "Test"}
	if err := BuildPageAt("post.md", mdDir+string(filepath.Separator), outDir+string(filepath.Separator), time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), cfg, tmplDir+string(filepath.Separator)); err != nil {
		t.Fatalf("BuildPageAt: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(outDir, "post", "index.html"))
	if err != nil {
		t.Fatalf("output missing: %v", err)
	}
	out := string(data)

	checks := map[string]string{
		"title":       "title:My Post",
		"description": "desc:Great post",
		"date":        "date:2026-03-15",
		"tags":        "tags:go,testing,",
		"slug":        "slug:my-post",
	}
	for field, want := range checks {
		if !strings.Contains(out, want) {
			t.Errorf("Pagematter.%s not wired: want %q in output\n%s", field, want, out)
		}
	}
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("writeFile %s: %v", path, err)
	}
}
