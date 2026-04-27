# Gokesh

A simple, minimal static site builder written in Go. Gokesh converts Markdown files with YAML frontmatter into static HTML pages using Go templates.

## Features

- **Markdown to HTML** - Converts `.md` files to `.html` using an in-house parser with support for headings, bold, italic, strikethrough, lists, blockquotes, fenced code blocks, tables, links, and images.
- **YAML Frontmatter** - Extract per-page metadata (e.g. page title) from YAML frontmatter in your Markdown files.
- **Go Template System** - Composable templates using Go's `text/template` package. Templates are split into reusable components (header, body, footer) composed into a single page layout.
- **Single Page or Batch Builds** - Build a single page by name, or an entire directory of Markdown files at once.
- **Global Site Configuration** - Set site-wide variables (`AUTHOR`, `SITETITLE`) via a `.env` file.
- **Development Server** - Built-in local HTTP server on port 8000 for previewing your site.

## Requirements

- Go 1.19 or later

## Installation

```bash
git clone https://github.com/vinckr/gokesh.git
cd gokesh
```

## Project Structure

```
gokesh/
├── cmd/
│   ├── build/main.go        # Build command - converts markdown to HTML
│   └── dev/main.go          # Dev server - serves files on localhost:8000
├── internal/
│   └── parser/
│       ├── frontmatter.go   # In-house YAML frontmatter parser
│       └── markdown.go      # In-house Markdown-to-HTML converter
├── markdown/                # Source markdown files
│   ├── index.md
│   └── blog/
│       ├── lorem.md
│       ├── ipsum.md
│       └── markdown.md
├── templates/               # Go HTML templates
│   ├── page.tmpl            # Main layout (composes header + body + footer)
│   ├── header.tmpl          # HTML head and opening tags
│   ├── body.tmpl            # Article body wrapper
│   └── footer.tmpl          # Footer with copyright
├── public/                  # Generated HTML output
├── .env                     # Global site configuration
├── Makefile                 # Build shortcuts
└── go.mod
```

## Usage

### Build a single page

```bash
go run cmd/build/main.go page <name>
```

This reads `./markdown/<name>.md` and writes `./public/<name>.html`.

```bash
# Example: build the index page
go run cmd/build/main.go page index
```

### Build all pages in a directory

```bash
go run cmd/build/main.go dir <directory>
```

This reads all `.md` files in `./markdown/<directory>/` and writes corresponding `.html` files to `./public/`.

```bash
# Example: build all blog posts
go run cmd/build/main.go dir blog
```

### Preview your site locally

```bash
make dev
```

This builds all test pages and starts a local server at [http://localhost:8000](http://localhost:8000).

### Make commands

| Command     | Description                                         |
|-------------|-----------------------------------------------------|
| `make test` | Build test pages (index page + blog directory)      |
| `make dev`  | Build test pages and start preview server           |
| `make help` | Show all available Make commands                    |

## Writing Content

### Markdown files

Place your Markdown files in the `markdown/` directory. Each file must include YAML frontmatter with at least a `pagetitle` field:

```markdown
---
pagetitle: "My Page Title"
---

# Hello World

Your markdown content goes here.
```

The `pagetitle` is used in the HTML `<title>` tag and is accessible in templates as `{{ .Pagematter.PageTitle }}`.

### Supported Markdown syntax

- Headings (`#` through `######`)
- Paragraphs and blockquotes
- Ordered and unordered lists
- Fenced code blocks
- Tables
- Bold (`**text**`), italic (`_text_`), bold+italic (`**_text_**`), strikethrough (`~~text~~`)
- Inline code (`` `code` ``)
- Links (`[text](url)`) and images (`![alt](url)`)

## Configuration

### .env file

Set global site variables in the `.env` file at the project root:

```
AUTHOR=vinckr
SITETITLE=gokesh.com
```

These values are available in all templates:

| Variable   | Template access  | Description                          |
|------------|------------------|--------------------------------------|
| AUTHOR     | `{{ .Author }}`  | Author name, used in footer and title |
| SITETITLE  | `{{ .SiteTitle }}`| Site title                           |

The current year is automatically available as `{{ .Year }}`.

## Templates

Gokesh uses Go's `text/template` package with a composition pattern. The main layout (`page.tmpl`) composes three sub-templates:

```
page.tmpl
├── Header  - HTML doctype, head, meta tags, title, stylesheet link
├── Body    - Article wrapper around the converted markdown HTML
└── Footer  - Copyright notice and closing tags
```

### Template data

All templates receive a data object with the following fields:

| Field                    | Type   | Description                         |
|--------------------------|--------|-------------------------------------|
| `.Body`                  | string | HTML content converted from markdown |
| `.SiteTitle`             | string | From `SITETITLE` in `.env`          |
| `.Author`                | string | From `AUTHOR` in `.env`             |
| `.Year`                  | string | Current year (auto-generated)       |
| `.Pagematter.PageTitle`  | string | From `pagetitle` in frontmatter     |

### Customizing templates

Edit the files in `templates/` to change the HTML output. For example, to add a navigation bar, edit `header.tmpl` or create additional template blocks in `page.tmpl`.

## Dependencies

None. Gokesh uses only the Go standard library.

The frontmatter parser and Markdown-to-HTML converter are implemented in-house under `internal/parser/`.

## 1.0 Roadmap

Planned improvements for the 1.0 release:

1. **Watch mode** — `make watch` polls for file changes with `os.Stat` and auto-rebuilds (no external deps)
2. **Config file** — Replace `.env` with `gokesh.toml` or `gokesh.yaml` parsed in-house; add `baseurl`, `description` fields
3. **Rich frontmatter** — Support `date`, `description`, `tags`, `draft` fields; skip draft pages on build
4. **Incremental builds** — Only rebuild pages where source `.md` is newer than output `.html`
5. **Custom output path** — `--out <path>` CLI flag or per-directory config
6. **RSS feed** — Auto-generate `public/feed.xml` from blog directory (sorted by date)
7. **Sitemap** — Auto-generate `public/sitemap.xml` on build
8. **Static file copying** — Copy stylesheets and assets from `static/` into `public/`
9. **Better errors** — File and line number context in error messages instead of bare `log.Fatalf`
10. **Structured CLI** — Replace bare `os.Args` with `flag` stdlib; add `--version` and `--help`

## License

Apache 2.0 - see [LICENSE](LICENSE) for details.
