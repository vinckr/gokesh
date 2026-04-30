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
	return Config{
		Author:      fields["author"],
		SiteTitle:   fields["site_title"],
		BaseURL:     fields["base_url"],
		Description: fields["description"],
	}, nil
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
	Pagematter  struct {
		PageTitle string
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

// BuildTemplate renders the named "Page" template with data and returns HTML.
func BuildTemplate(d pageData, templates ...string) (string, error) {
	funcMap := template.FuncMap{"jsonify": jsonify, "items": items}
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

func BuildPage(fileName string, dir string, outpath string, cfg Config, templates ...string) error {
	return BuildPageAt(fileName, dir, outpath, time.Now(), cfg, templates...)
}

func BuildPageAt(fileName string, dir string, outpath string, now time.Time, cfg Config, templates ...string) error {
	md, err := ReadMarkdownFileFromDirectory(dir, fileName)
	if err != nil {
		return err
	}
	body, matter := SplitBodyAndFrontmatter(md)

	var d pageData
	d.Body = string(parser.ToHTML(body))
	d.Author = cfg.Author
	d.SiteTitle = cfg.SiteTitle
	d.BaseURL = cfg.BaseURL
	d.Description = cfg.Description
	d.Year = now.Format("2006")
	d.Pagematter.PageTitle = matter["title"]

	if dataFile := matter["data"]; dataFile != "" {
		raw, err := os.ReadFile("./data/" + dataFile)
		if err != nil {
			return fmt.Errorf("reading data file %s: %w", dataFile, err)
		}
		if !json.Valid(raw) {
			return fmt.Errorf("invalid JSON in data file %s", dataFile)
		}
		d.Data = json.RawMessage(raw)
	}

	tmpl := templates
	if name := matter["template"]; name != "" {
		tmpl = make([]string, len(templates))
		copy(tmpl, templates)
		tmpl[0] = "./templates/" + strings.TrimSuffix(name, ".tmpl") + ".tmpl"
	}

	slog.Info("building page", "title", d.Pagematter.PageTitle, "template", tmpl[0])
	html, err := BuildTemplate(d, tmpl...)
	if err != nil {
		return err
	}
	return WriteHTMLFile(fileName, outpath, html)
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

func BuildPages(dir string, outpath string, cfg Config, templates ...string) error {
	files, err := GetFilesFromDirectory(dir)
	if err != nil {
		return err
	}
	for _, file := range files {
		if err := BuildPage(file.Name(), dir, outpath, cfg, templates...); err != nil {
			return err
		}
	}
	return nil
}

// BuildAll walks markdownDir recursively and builds every .md file it finds,
// preserving the directory structure in outpath.
func BuildAll(markdownDir string, outpath string, cfg Config, templates ...string) error {
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
		return BuildPage(filepath.Base(path), dir+string(filepath.Separator), fileOutpath, cfg, templates...)
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
