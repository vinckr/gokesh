package build

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/template"
	"time"

	"github.com/vinckr/gokesh/internal/parser"
)

// Config holds site-wide configuration from gokesh.toml.
type Config struct {
	Author      string
	SiteTitle   string
	BaseURL     string
	Description string
	MarkdownDir string // default: "markdown"
	OutputDir   string // default: "public"
}

// LoadConfig reads and parses gokesh.toml. Returns an empty Config if the file does not exist.
func LoadConfig(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return Config{}, nil
		}
		return Config{}, fmt.Errorf("reading config %q: %w", path, err)
	}
	fields := parser.ParseConfig(data)
	cfg := Config{
		Author:      fields["author"],
		SiteTitle:   fields["site_title"],
		BaseURL:     fields["base_url"],
		Description: fields["description"],
		MarkdownDir: fields["markdown_dir"],
		OutputDir:   fields["output_dir"],
	}
	if cfg.MarkdownDir == "" {
		cfg.MarkdownDir = "markdown"
	}
	if cfg.OutputDir == "" {
		cfg.OutputDir = "public"
	}
	return cfg, nil
}

// PageSummary holds the metadata of one page, used to populate .Pages in templates.
type PageSummary struct {
	Title string
	URL   string
	Date  time.Time
	Tags  []string
}

// pageData holds the full context passed to page templates.
type pageData struct {
	Body        string
	SiteTitle   string
	BaseURL     string
	Description string
	Year        string
	Author      string
	Data        json.RawMessage
	Pages       []PageSummary
	Pagematter  struct {
		PageTitle   string
		Description string
		Date        time.Time
		Tags        []string
		Slug        string
		Draft       bool
	}
}

// GetFilesFromDirectory returns all directory entries at path.
func GetFilesFromDirectory(path string) ([]fs.DirEntry, error) {
	files, err := os.ReadDir(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("directory not found: %q", path)
		}
		return nil, fmt.Errorf("reading directory %q: %w", path, err)
	}
	return files, nil
}

// ReadMarkdownFileFromDirectory reads a markdown file from a directory.
func ReadMarkdownFileFromDirectory(path string, filename string) ([]byte, error) {
	md, err := os.ReadFile(path + filename)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("file not found: %q", path+filename)
		}
		return nil, fmt.Errorf("reading %q: %w", path+filename, err)
	}
	return md, nil
}

// SplitBodyAndFrontmatter extracts frontmatter and returns only the body.
func SplitBodyAndFrontmatter(md []byte) ([]byte, map[string]string) {
	matter, body := parser.ParseFrontmatter(md)
	return body, matter
}

// jsonify returns the raw JSON string for embedding in templates.
func jsonify(v json.RawMessage) string {
	return string(v)
}

// items parses a RawMessage JSON array into a slice for template ranging.
func items(v json.RawMessage) []map[string]any {
	var result []map[string]any
	json.Unmarshal(v, &result)
	return result
}

// dateFormat formats a time.Time using the given Go layout string.
// Returns "" for a zero time.
func dateFormat(t time.Time, layout string) string {
	if t.IsZero() {
		return ""
	}
	return t.Format(layout)
}

// sortBy returns a copy of pages sorted by "date" (newest-first) or "title" (A-Z).
func sortBy(pages []PageSummary, field string) []PageSummary {
	out := make([]PageSummary, len(pages))
	copy(out, pages)
	sort.Slice(out, func(i, j int) bool {
		switch field {
		case "title":
			return out[i].Title < out[j].Title
		default: // "date"
			if out[i].Date.IsZero() != out[j].Date.IsZero() {
				return !out[i].Date.IsZero()
			}
			return out[i].Date.After(out[j].Date)
		}
	})
	return out
}

// filterByTag returns pages that include the given tag.
func filterByTag(pages []PageSummary, tag string) []PageSummary {
	var out []PageSummary
	for _, p := range pages {
		for _, t := range p.Tags {
			if t == tag {
				out = append(out, p)
				break
			}
		}
	}
	return out
}

// BuildTemplate renders the named "Page" template with data and returns HTML.
func BuildTemplate(d pageData, templates ...string) (string, error) {
	funcMap := template.FuncMap{
		"jsonify":       jsonify,
		"items":         items,
		"dateFormat":    dateFormat,
		"sortBy":        sortBy,
		"filterByTag":   filterByTag,
	}
	t, err := template.New("").Funcs(funcMap).ParseFiles(templates...)
	if err != nil {
		return "", fmt.Errorf("parsing templates: %w", err)
	}
	build := new(strings.Builder)
	if err := t.ExecuteTemplate(build, "Page", d); err != nil {
		return "", fmt.Errorf("executing template: %w", err)
	}
	return build.String(), nil
}

// WriteHTMLFile writes the rendered page HTML to disk using clean URLs.
// index.md → outpath/index.html
// anything.md → outpath/anything/index.html (so /anything/ serves cleanly)
func WriteHTMLFile(fileName string, outpath string, page string) error {
	base := strings.TrimSuffix(fileName, ".md")
	var outPath string
	if base == "index" {
		if err := os.MkdirAll(outpath, 0755); err != nil {
			return fmt.Errorf("creating output directory %s: %w", outpath, err)
		}
		outPath = filepath.Join(outpath, "index.html")
	} else {
		dir := filepath.Join(outpath, base)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("creating output directory %s: %w", dir, err)
		}
		outPath = filepath.Join(dir, "index.html")
	}
	if err := os.WriteFile(outPath, []byte(page), 0644); err != nil {
		return fmt.Errorf("writing %s: %w", outPath, err)
	}
	slog.Info("file written", "file", outPath)
	return nil
}

// resolveTemplates loads all *.tmpl files from templatesDir and ensures the
// entry-point template (the one that defines "Page") is last so its definition
// wins when multiple files define the same template name.
func resolveTemplates(templatesDir, entryName string) ([]string, error) {
	all, err := filepath.Glob(filepath.Join(templatesDir, "*.tmpl"))
	if err != nil {
		return nil, fmt.Errorf("globbing templates in %q: %w", templatesDir, err)
	}
	if len(all) == 0 {
		return nil, fmt.Errorf("no .tmpl files found in %q", templatesDir)
	}

	if entryName == "" {
		entryName = "page"
	}
	entry := filepath.Join(templatesDir, strings.TrimSuffix(filepath.Base(entryName), ".tmpl")+".tmpl")

	rest := make([]string, 0, len(all))
	found := false
	for _, t := range all {
		if filepath.Clean(t) == filepath.Clean(entry) {
			found = true
		} else {
			rest = append(rest, t)
		}
	}
	if !found {
		return nil, fmt.Errorf("template %q not found in %q", entryName, templatesDir)
	}
	return append(rest, entry), nil
}

func BuildPage(fileName string, dir string, outpath string, cfg Config, templatesDir string) error {
	pages := CollectPages("./markdown/", cfg.BaseURL)
	return buildPageAt(fileName, dir, outpath, time.Now(), cfg, templatesDir, pages)
}

func BuildPageAt(fileName string, dir string, outpath string, now time.Time, cfg Config, templatesDir string) error {
	return buildPageAt(fileName, dir, outpath, now, cfg, templatesDir, nil)
}

func buildPageAt(fileName string, dir string, outpath string, now time.Time, cfg Config, templatesDir string, pages []PageSummary) error {
	md, err := ReadMarkdownFileFromDirectory(dir, fileName)
	if err != nil {
		return err
	}
	fm, body := parser.ParseTypedFrontmatter(md)

	if fm.Draft {
		slog.Info("skipping draft", "file", fileName)
		return nil
	}
	if fm.Title == "" {
		slog.Warn("page has no title", "file", fileName)
	}

	var d pageData
	d.Body = string(parser.ToHTML(body))
	d.Author = cfg.Author
	d.SiteTitle = cfg.SiteTitle
	d.BaseURL = cfg.BaseURL
	d.Description = cfg.Description
	d.Year = now.Format("2006")
	d.Pages = pages
	d.Pagematter.PageTitle = fm.Title
	d.Pagematter.Description = fm.Description
	d.Pagematter.Date = fm.Date
	d.Pagematter.Tags = fm.Tags
	d.Pagematter.Slug = fm.Slug
	d.Pagematter.Draft = fm.Draft

	if fm.Data != "" {
		raw, err := os.ReadFile("./data/" + fm.Data)
		if err != nil {
			return fmt.Errorf("reading data file %s: %w", fm.Data, err)
		}
		if !json.Valid(raw) {
			return fmt.Errorf("invalid JSON in data file %s", fm.Data)
		}
		d.Data = json.RawMessage(raw)
	}

	tmpl, err := resolveTemplates(templatesDir, fm.Template)
	if err != nil {
		return err
	}

	slog.Info("building page", "title", d.Pagematter.PageTitle, "template", tmpl[len(tmpl)-1])
	html, err := BuildTemplate(d, tmpl...)
	if err != nil {
		return err
	}
	return WriteHTMLFile(fileName, outpath, html)
}

// CopyStatic recursively copies all files from staticDir into outpath,
// preserving subdirectory structure. Silently succeeds if staticDir is missing.
func CopyStatic(staticDir string, outpath string) error {
	if _, err := os.Stat(staticDir); os.IsNotExist(err) {
		return nil
	}
	if err := os.MkdirAll(outpath, 0755); err != nil {
		return fmt.Errorf("creating output directory %s: %w", outpath, err)
	}
	return filepath.Walk(staticDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, relErr := filepath.Rel(staticDir, path)
		if relErr != nil {
			return relErr
		}
		dst := filepath.Join(outpath, rel)
		if info.IsDir() {
			return os.MkdirAll(dst, 0755)
		}
		if copyErr := copyFile(path, dst); copyErr != nil {
			return copyErr
		}
		slog.Info("static file copied", "src", path, "dst", dst)
		return nil
	})
}

// CopyStyles copies all files from stylesDir into outpath.
func CopyStyles(stylesDir string, outpath string) error {
	if err := os.MkdirAll(outpath, 0755); err != nil {
		return fmt.Errorf("creating output directory %s: %w", outpath, err)
	}
	entries, err := os.ReadDir(stylesDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("reading styles directory: %w", err)
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		src := stylesDir + entry.Name()
		dst := outpath + entry.Name()
		if err := copyFile(src, dst); err != nil {
			return err
		}
		slog.Info("style copied", "src", src, "dst", dst)
	}
	return nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("opening %s: %w", src, err)
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("creating %s: %w", dst, err)
	}
	defer out.Close()

	if _, err = io.Copy(out, in); err != nil {
		return fmt.Errorf("copying %s to %s: %w", src, dst, err)
	}
	return nil
}

func BuildPages(dir string, outpath string, cfg Config, templatesDir string) error {
	files, err := GetFilesFromDirectory(dir)
	if err != nil {
		return err
	}
	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".md") {
			continue
		}
		if err := BuildPage(file.Name(), dir, outpath, cfg, templatesDir); err != nil {
			return err
		}
	}
	return nil
}

// shouldRebuild returns true if srcPath is newer than outPath, or outPath doesn't exist.
func shouldRebuild(srcPath, outPath string) bool {
	outInfo, err := os.Stat(outPath)
	if err != nil {
		return true // output missing
	}
	srcInfo, err := os.Stat(srcPath)
	if err != nil {
		return true
	}
	return srcInfo.ModTime().After(outInfo.ModTime())
}

// CollectPages walks markdownDir and returns a PageSummary for every non-draft .md file.
// Pages with a Date are sorted newest-first; undated pages appear at the end.
func CollectPages(markdownDir string, baseURL string) []PageSummary {
	markdownDir = filepath.Clean(markdownDir)
	var pages []PageSummary
	_ = filepath.Walk(markdownDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(path, ".md") {
			return nil
		}
		raw, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		fm, _ := parser.ParseTypedFrontmatter(raw)
		if fm.Draft {
			return nil
		}
		rel, err := filepath.Rel(markdownDir, path)
		if err != nil {
			return nil
		}
		url := pageURL(rel, baseURL)
		pages = append(pages, PageSummary{
			Title: fm.Title,
			URL:   url,
			Date:  fm.Date,
			Tags:  fm.Tags,
		})
		return nil
	})
	// sort: dated pages newest-first, undated pages last
	sort.Slice(pages, func(i, j int) bool {
		if pages[i].Date.IsZero() != pages[j].Date.IsZero() {
			return !pages[i].Date.IsZero()
		}
		return pages[i].Date.After(pages[j].Date)
	})
	return pages
}

// pageURL converts a relative markdown path to a clean URL string.
func pageURL(rel, baseURL string) string {
	base := strings.TrimSuffix(rel, ".md")
	base = filepath.ToSlash(base)
	if base == "index" {
		if baseURL != "" {
			return strings.TrimRight(baseURL, "/") + "/"
		}
		return "/"
	}
	if baseURL != "" {
		return strings.TrimRight(baseURL, "/") + "/" + base + "/"
	}
	return "/" + base + "/"
}

// BuildAll walks markdownDir recursively and builds every .md file it finds,
// preserving the directory structure in outpath.
// A first pass collects all page metadata so templates can access .Pages.
func BuildAll(markdownDir string, outpath string, cfg Config, templatesDir string) error {
	return buildAllWith(markdownDir, outpath, cfg, templatesDir, false)
}

// BuildAllIncremental is like BuildAll but skips pages whose output is newer than their source.
func BuildAllIncremental(markdownDir string, outpath string, cfg Config, templatesDir string) error {
	return buildAllWith(markdownDir, outpath, cfg, templatesDir, true)
}

func buildAllWith(markdownDir string, outpath string, cfg Config, templatesDir string, incremental bool) error {
	pages := CollectPages(markdownDir, cfg.BaseURL)
	markdownDir = filepath.Clean(markdownDir)
	return filepath.Walk(markdownDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || !strings.HasSuffix(path, ".md") {
			return nil
		}
		dir := filepath.Dir(path)
		relDir, err := filepath.Rel(markdownDir, dir)
		if err != nil {
			return err
		}
		fileOutpath := filepath.Join(outpath, relDir) + string(filepath.Separator)
		if incremental {
			base := strings.TrimSuffix(filepath.Base(path), ".md")
			var outFile string
			if base == "index" {
				outFile = filepath.Join(fileOutpath, "index.html")
			} else {
				outFile = filepath.Join(fileOutpath, base, "index.html")
			}
			if !shouldRebuild(path, outFile) {
				return nil
			}
		}
		return buildPageAt(filepath.Base(path), dir+string(filepath.Separator), fileOutpath, time.Now(), cfg, templatesDir, pages)
	})
}

type sitemapURL struct {
	Loc string `xml:"loc"`
}

type sitemapXML struct {
	XMLName xml.Name     `xml:"urlset"`
	XMLNS   string       `xml:"xmlns,attr"`
	URLs    []sitemapURL `xml:"url"`
}

// GenerateSitemap writes a sitemap.xml to outpath listing all .html files.
// Skips generation if baseURL is empty.
func GenerateSitemap(outpath string, baseURL string) error {
	if baseURL == "" {
		slog.Warn("skipping sitemap: base_url not set in gokesh.toml")
		return nil
	}
	baseURL = strings.TrimRight(baseURL, "/")

	var urls []sitemapURL
	err := filepath.Walk(outpath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || filepath.Base(path) != "index.html" {
			return err
		}
		relDir, err := filepath.Rel(outpath, filepath.Dir(path))
		if err != nil {
			return err
		}
		if relDir == "." {
			urls = append(urls, sitemapURL{Loc: baseURL + "/"})
		} else {
			urls = append(urls, sitemapURL{Loc: baseURL + "/" + filepath.ToSlash(relDir) + "/"})
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("building sitemap: %w", err)
	}

	sm := sitemapXML{
		XMLNS: "http://www.sitemaps.org/schemas/sitemap/0.9",
		URLs:  urls,
	}
	out, err := xml.MarshalIndent(sm, "", "  ")
	if err != nil {
		return fmt.Errorf("encoding sitemap: %w", err)
	}

	dest := filepath.Join(outpath, "sitemap.xml")
	content := append([]byte(xml.Header), out...)
	if err := os.WriteFile(dest, content, 0644); err != nil {
		return fmt.Errorf("writing sitemap: %w", err)
	}
	slog.Info("sitemap written", "file", dest)
	return nil
}
