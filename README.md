# Gokesh

Minimal static site builder in Go. Converts Markdown + YAML frontmatter to HTML using Go templates. Zero dependencies.

## Installation

```bash
go install github.com/vinckr/gokesh/cmd/gokesh@latest
```

Or download a pre-built binary from the [Releases page](https://github.com/vinckr/gokesh/releases).

## Getting Started

**1. Initialize a new project:**

```bash
mkdir myblog && cd myblog
gokesh init
```

This creates `gokesh.toml`, `templates/`, `styles/`, and example Markdown files.

**2. Start the development server:**

```bash
gokesh serve
```

Opens a file watcher + local server with live reload at `http://localhost:8000`. Edit any file and the browser refreshes automatically.

**3. Deploy** by copying `public/` to any static host (GitHub Pages, Netlify, Cloudflare Pages, S3, etc.).

---

## Commands

| Command | Description |
| --- | --- |
| `gokesh init` | Set up a new project (templates, styles, config) |
| `gokesh serve` | Watch for changes + serve with live reload **(recommended for development)** |
| `gokesh build` | Build all `.md` files in `markdown/` recursively |
| `gokesh build page <name>` | Build `markdown/<name>.md` → `public/<name>/index.html` |
| `gokesh build dir <name>` | Build all `.md` files in `markdown/<name>/` |
| `gokesh new <name>` | Create `markdown/<name>.md` with pre-filled frontmatter |
| `gokesh clean` | Delete the output directory |
| `gokesh watch` | Watch for changes and rebuild (no server) |
| `gokesh dev` | Serve `public/` at http://localhost:8000 (no watch) |
| `--version`, `-v` | Print version and exit |
| `--help`, `-h` | Print usage and exit |

---

## Frontmatter

Each Markdown file starts with a `---`-delimited frontmatter block:

```markdown
---
title: "My Post"
date: 2026-01-15
description: "A short summary shown in search results and social previews."
tags: [go, web, ssr]
draft: false
slug: my-post
template: post
data: posts.json
---

# My Post

Content goes here.
```

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `title` | string | yes | Page title — `{{ .Pagematter.PageTitle }}` |
| `date` | date (`YYYY-MM-DD`) | no | Post date — `{{ .Pagematter.Date }}` |
| `description` | string | no | Per-page meta description — `{{ .Pagematter.Description }}` |
| `tags` | `[tag1, tag2]` | no | Taxonomy tags — `{{ .Pagematter.Tags }}` |
| `draft` | bool | no | If `true`, page is excluded from all builds |
| `slug` | string | no | Override the output URL path |
| `template` | string | no | Layout from `templates/`. Default: `page` |
| `data` | string | no | JSON file from `data/` passed as `.Data` to the template |

---

## Configuration

`gokesh.toml` at the project root. All fields are optional.

```toml
author       = "yourname"
site_title   = "myblog.com"
base_url     = "https://myblog.com"
description  = "My personal blog"
output_dir   = "public"    # default: public
markdown_dir = "markdown"  # default: markdown
```

| Key | Description |
| --- | --- |
| `author` | Author name — `{{ .Author }}` |
| `site_title` | Site title — `{{ .SiteTitle }}` |
| `base_url` | Base URL — `{{ .BaseURL }}` |
| `description` | Site description — `{{ .Description }}` |
| `output_dir` | Output directory (default: `public`) |
| `markdown_dir` | Source Markdown directory (default: `markdown`) |

---

## Templates

Edit files in `templates/`. The default layout (`page.tmpl`) composes `header.tmpl`, `body.tmpl`, and `footer.tmpl`. Specify a different layout per page with `template:` in frontmatter.

### Template variables

| Variable | Type | Description |
| --- | --- | --- |
| `{{ .Body }}` | string | Rendered HTML from Markdown |
| `{{ .SiteTitle }}` | string | From `gokesh.toml` |
| `{{ .BaseURL }}` | string | From `gokesh.toml` |
| `{{ .Description }}` | string | Site description from `gokesh.toml` |
| `{{ .Author }}` | string | From `gokesh.toml` |
| `{{ .Year }}` | string | Current build year |
| `{{ .Data }}` | JSON | Data file contents (if `data:` set in frontmatter) |
| `{{ .Pages }}` | `[]PageSummary` | All non-draft pages, sorted newest-first |
| `{{ .Pagematter.PageTitle }}` | string | `title` frontmatter field |
| `{{ .Pagematter.Description }}` | string | `description` frontmatter field |
| `{{ .Pagematter.Date }}` | `time.Time` | `date` frontmatter field |
| `{{ .Pagematter.Tags }}` | `[]string` | `tags` frontmatter field |
| `{{ .Pagematter.Slug }}` | string | `slug` frontmatter field |
| `{{ .Pagematter.Draft }}` | bool | `draft` frontmatter field |

### Template functions

| Function | Example | Description |
| --- | --- | --- |
| `dateFormat` | `{{ dateFormat .Pagematter.Date "Jan 2, 2006" }}` | Format a `time.Time` value |
| `sortBy` | `{{ sortBy .Pages "date" }}` | Sort pages by `"date"` (newest-first) or `"title"` (A–Z) |
| `filterByTag` | `{{ filterByTag .Pages "go" }}` | Filter pages to those with a given tag |
| `jsonify` | `{{ jsonify .Data }}` | Embed raw JSON |
| `items` | `{{ range items .Data }}` | Range over a JSON array |

### Building a post list

```html
{{ range sortBy .Pages "date" }}
  <article>
    <a href="{{ .URL }}">{{ .Title }}</a>
    <time>{{ dateFormat .Date "Jan 2, 2006" }}</time>
  </article>
{{ end }}
```

---

## Markdown

Headings (`#`–`######`), bold (`**`), italic (`_`), bold+italic (`**_`), strikethrough (`~~`), inline code (`` ` ``), fenced code blocks (with optional language class), blockquotes, ordered/unordered lists, tables, links, images, horizontal rules.

Fenced code blocks with a language specifier emit `<code class="language-go">`, compatible with syntax highlighting libraries like Prism or highlight.js.

---

## Make Commands

| Command | Description |
| --- | --- |
| `make test` | Run tests |
| `make vet` | Run go vet |
| `make build` | Build binary to `bin/gokesh` |
| `make install` | Install to `$GOPATH/bin` |
| `make dev` | Build example pages and start preview server |
| `make update-golden` | Update golden test files |
| `make release VERSION=v0.1.0` | Tag, push, and trigger a GitHub release |

## Releasing

```bash
make release VERSION=v0.1.0
```

Runs tests and vet, then tags and pushes. GitHub Actions builds binaries for all platforms and publishes to the [Releases page](https://github.com/vinckr/gokesh/releases).

## License

Apache 2.0 - see [LICENSE](LICENSE) for details.
