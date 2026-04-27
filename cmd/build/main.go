package main

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"strings"
	"text/template"
	"time"

	"github.com/vinckr/gokesh/internal/parser"
)

// data holds the full context passed to page templates.
type data struct {
	Body       string
	SiteTitle  string
	Year       string
	Author     string
	Pagematter struct {
		PageTitle string
	}
}

// GetFilesFromDirectory returns all directory entries at path.
func GetFilesFromDirectory(path string) []fs.DirEntry {
	files, err := os.ReadDir(path)
	if err != nil {
		log.Fatalf("Error getting files: %s", err)
	}
	return files
}

// ReadMarkdownFileFromDirectory reads a markdown file from a directory.
func ReadMarkdownFileFromDirectory(path string, filename string) []byte {
	md, err := os.ReadFile(path + filename)
	if err != nil {
		log.Fatalf("Error reading markdown file: %s", err)
	}
	return md
}

// SplitBodyAndFrontmatter extracts frontmatter and returns only the body.
func SplitBodyAndFrontmatter(md []byte) ([]byte, map[string]string) {
	matter, body := parser.ParseFrontmatter(md)
	return body, matter
}

// BuildTemplate renders the named "Page" template with data and returns HTML.
func BuildTemplate(d data, templates ...string) string {
	t := template.Must(template.ParseFiles(templates...))
	build := new(strings.Builder)
	if err := t.ExecuteTemplate(build, "Page", d); err != nil {
		log.Fatalf("Error building the template: %s", err)
	}
	return build.String()
}

// WriteHTMLFile writes the rendered page HTML to disk.
func WriteHTMLFile(fileName string, outpath string, page string) {
	outPath := outpath + strings.TrimSuffix(fileName, ".md") + ".html"
	if err := os.WriteFile(outPath, []byte(page), 0644); err != nil {
		log.Fatalf("Error writing file: %s", err)
	}
	fmt.Printf("\n%s written to %s\n------------------------", fileName, outPath)
}

func BuildPage(fileName string, dir string, outpath string, templates ...string) {
	// Global config from environment
	author := os.Getenv("AUTHOR")
	sitetitle := os.Getenv("SITETITLE")
	currentYear := time.Now().Format("2006")

	// Read and parse markdown
	md := ReadMarkdownFileFromDirectory(dir, fileName)
	body, matter := SplitBodyAndFrontmatter(md)

	// Convert markdown body to HTML
	htmlBody := parser.ToHTML(body)

	// Assemble page data
	var d data
	d.Body = string(htmlBody)
	d.SiteTitle = sitetitle
	d.Year = currentYear
	d.Author = author
	d.Pagematter.PageTitle = matter["pagetitle"]

	fmt.Printf("\nBuilding page %s:", d.Pagematter.PageTitle)
	build := BuildTemplate(d, templates...)
	WriteHTMLFile(fileName, outpath, build)
}

func BuildPages(dir string, outpath string, templates ...string) {
	files := GetFilesFromDirectory(dir)
	for _, file := range files {
		BuildPage(file.Name(), dir, outpath, templates...)
	}
}

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: go run main.go <input_type> <input_path> [output_path]")
		fmt.Println("input_type: 'page' for single page or 'dir' for directory")
		return
	}
	inputType := os.Args[1]
	inputPath := os.Args[2]
	outputPath := "./public/"

	tmpl := []string{
		"./templates/page.tmpl",
		"./templates/header.tmpl",
		"./templates/footer.tmpl",
		"./templates/body.tmpl",
	}

	switch inputType {
	case "page":
		BuildPage(inputPath+".md", "./markdown/", outputPath, tmpl...)
	case "dir":
		BuildPages("markdown/"+inputPath+"/", outputPath, tmpl...)
	default:
		fmt.Println("Invalid input type. Use 'page' or 'dir'.")
	}
}
