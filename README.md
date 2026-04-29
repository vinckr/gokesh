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

- Go 1.22 or later

## Installation

### Option A: Go install (recommended if you have Go)

```bash
go install github.com/vinckr/gokesh/cmd/gokesh@latest
```

This installs the latest tagged release. Re-run the same command to update.

### Option B: Download a binary

Download a pre-built binary for your platform from the [Releases page](https://github.com/vinckr/gokesh/releases), extract it, and put it somewhere on your `$PATH`.

---

## Getting Started

This guide assumes you already have a folder of Markdown files and want to turn them into a blog.

### 1. Create a project directory

```bash
mkdir myblog && cd myblog
```

### 2. Configure your site

Edit the `.env` file at the project root:

```
AUTHOR=yourname
SITETITLE=myblog.com
```

### 3. Add frontmatter to your Markdown files

Every Markdown file needs a YAML frontmatter block at the top:

```markdown
---
pagetitle: "My First Post"
---

# My First Post

Content goes here.
```

### 4. Put your Markdown files in the right place

- **Index / standalone pages** → `markdown/<name>.md`
- **Blog posts** → `markdown/blog/<name>.md` (or any subdirectory name you choose)

```
markdown/
├── index.md          # your homepage
└── blog/
    ├── first-post.md
    └── second-post.md
```

### 5. Customize the templates

Edit the files in `templates/` to match your design. At minimum you may want to update `header.tmpl` to change the site name, stylesheet, or add navigation.

### 6. Build your site

Build the homepage:

```bash
gokesh build page index
```

Build all blog posts:

```bash
gokesh build dir blog
```

Generated HTML lands in `public/`.

### 7. Preview locally

```bash
gokesh dev
```

Open [http://localhost:8000](http://localhost:8000) in your browser.

### 8. Deploy

Copy the contents of `public/` to any static file host (GitHub Pages, Netlify, Cloudflare Pages, an S3 bucket, etc.). No server-side runtime required.

---

## Project Structure

```
gokesh/
├── cmd/
│   └── gokesh/main.go       # Single binary: build and dev commands
├── internal/
│   ├── build/               # Build logic and tests
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
gokesh build page <name>
```

This reads `./markdown/<name>.md` and writes `./public/<name>.html`.

```bash
# Example: build the index page
gokesh build page index
```

### Build all pages in a directory

```bash
gokesh build dir <directory>
```

This reads all `.md` files in `./markdown/<directory>/` and writes corresponding `.html` files to `./public/`.

```bash
# Example: build all blog posts
gokesh build dir blog
```

### Preview your site locally

```bash
gokesh dev
```

Starts a local server at [http://localhost:8000](http://localhost:8000) serving `public/`.

### Make commands

| Command                       | Description                                    |
|-------------------------------|------------------------------------------------|
| `make test`                   | Run tests                                      |
| `make vet`                    | Run go vet                                     |
| `make build`                  | Build binary to `bin/gokesh`                   |
| `make dev`                    | Build example pages and start preview server   |
| `make install`                | Install binary to `$GOPATH/bin`                |
| `make update-golden`          | Update golden test files                       |
| `make release VERSION=v0.1.0` | Tag, push, and trigger a release               |
| `make help`                   | Show all available Make commands               |

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

## Cutting a release

Make sure all changes are committed and tests pass:

```bash
go test ./...
git status
```

Tag the release and push:

```bash
make release VERSION=v0.1.0
```

This runs tests and vet, then tags and pushes. GitHub Actions will automatically build binaries for all platforms and publish them to the [Releases page](https://github.com/vinckr/gokesh/releases).

To update an existing installation after a release:

```bash
go install github.com/vinckr/gokesh/cmd/gokesh@latest
```

## License

Apache 2.0 - see [LICENSE](LICENSE) for details.
