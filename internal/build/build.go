package build

import (
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"strings"
	"text/template"
	"time"

	"github.com/vinckr/gokesh/internal/parser"
)

// pageData holds the full context passed to page templates.
type pageData struct {
	Body       string
	SiteTitle  string
	Year       string
	Author     string
	Pagematter struct {
		PageTitle string
	}
}

// GetFilesFromDirectory returns all directory entries at path.
func GetFilesFromDirectory(path string) ([]fs.DirEntry, error) {
	files, err := os.ReadDir(path)
	if err != nil {
		return nil, fmt.Errorf("reading directory %s: %w", path, err)
	}
	return files, nil
}

// ReadMarkdownFileFromDirectory reads a markdown file from a directory.
func ReadMarkdownFileFromDirectory(path string, filename string) ([]byte, error) {
	md, err := os.ReadFile(path + filename)
	if err != nil {
		return nil, fmt.Errorf("reading %s%s: %w", path, filename, err)
	}
	return md, nil
}

// SplitBodyAndFrontmatter extracts frontmatter and returns only the body.
func SplitBodyAndFrontmatter(md []byte) ([]byte, map[string]string) {
	matter, body := parser.ParseFrontmatter(md)
	return body, matter
}

// BuildTemplate renders the named "Page" template with data and returns HTML.
func BuildTemplate(d pageData, templates ...string) (string, error) {
	t, err := template.ParseFiles(templates...)
	if err != nil {
		return "", fmt.Errorf("parsing templates: %w", err)
	}
	build := new(strings.Builder)
	if err := t.ExecuteTemplate(build, "Page", d); err != nil {
		return "", fmt.Errorf("executing template: %w", err)
	}
	return build.String(), nil
}

// WriteHTMLFile writes the rendered page HTML to disk.
func WriteHTMLFile(fileName string, outpath string, page string) error {
	outPath := outpath + strings.TrimSuffix(fileName, ".md") + ".html"
	if err := os.WriteFile(outPath, []byte(page), 0644); err != nil {
		return fmt.Errorf("writing %s: %w", outPath, err)
	}
	slog.Info("file written", "file", outPath)
	return nil
}

func BuildPage(fileName string, dir string, outpath string, templates ...string) error {
	return BuildPageAt(fileName, dir, outpath, time.Now(), templates...)
}

func BuildPageAt(fileName string, dir string, outpath string, now time.Time, templates ...string) error {
	author := os.Getenv("AUTHOR")
	sitetitle := os.Getenv("SITETITLE")
	currentYear := now.Format("2006")

	md, err := ReadMarkdownFileFromDirectory(dir, fileName)
	if err != nil {
		return err
	}
	body, matter := SplitBodyAndFrontmatter(md)

	var d pageData
	d.Body = string(parser.ToHTML(body))
	d.SiteTitle = sitetitle
	d.Year = currentYear
	d.Author = author
	d.Pagematter.PageTitle = matter["pagetitle"]

	slog.Info("building page", "title", d.Pagematter.PageTitle)
	html, err := BuildTemplate(d, templates...)
	if err != nil {
		return err
	}
	return WriteHTMLFile(fileName, outpath, html)
}

// CopyStyles copies all files from stylesDir into outpath.
func CopyStyles(stylesDir string, outpath string) error {
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

func BuildPages(dir string, outpath string, templates ...string) error {
	files, err := GetFilesFromDirectory(dir)
	if err != nil {
		return err
	}
	for _, file := range files {
		if err := BuildPage(file.Name(), dir, outpath, templates...); err != nil {
			return err
		}
	}
	return nil
}
