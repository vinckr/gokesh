# Gokesh

Minimal static site builder in Go. Converts Markdown + YAML frontmatter to HTML using Go templates. Zero dependencies.

## Installation

```bash
go install github.com/vinckr/gokesh/cmd/gokesh@latest
```

Or download a pre-built binary from the [Releases page](https://github.com/vinckr/gokesh/releases).

## Getting Started

**1. Create a project directory:**

```
myblog/
├── markdown/        # source .md files
│   ├── index.md
│   └── blog/
│       └── post.md
├── templates/       # copy from gokesh repo
├── styles/          # optional CSS
├── data/            # optional JSON data files
└── .env
```

**2. Configure `gokesh.toml`:**

```toml
author      = "yourname"
site_title  = "myblog.com"
base_url    = "https://myblog.com"
description = "My personal blog"
```

**3. Add frontmatter to each Markdown file:**

```markdown
---
title: "My Post"
---

# My Post

Content goes here.
```

**4. Build and preview:**

```bash
gokesh build   # builds all pages into public/
gokesh dev     # serves public/ at http://localhost:8000
```

**5. Deploy** by copying `public/` to any static host (GitHub Pages, Netlify, Cloudflare Pages, S3, etc.).

## Commands

| Command                    | Description                                       |
| -------------------------- | ------------------------------------------------- |
| `gokesh build`             | Build all `.md` files in `markdown/` recursively  |
| `gokesh build page <name>` | Build `markdown/<name>.md` → `public/<name>.html` |
| `gokesh build dir <name>`  | Build all `.md` files in `markdown/<name>/`       |
| `gokesh watch`             | Watch for changes and rebuild automatically       |
| `gokesh dev`               | Serve `public/` at http://localhost:8000          |

## Frontmatter

| Field      | Required | Description                                                                                 |
| ---------- | -------- | ------------------------------------------------------------------------------------------- |
| `title`    | yes      | Page title — `{{ .Pagematter.PageTitle }}` in templates                                     |
| `template` | no       | Layout from `templates/`. Default: `page`. E.g. `template: post` uses `templates/post.tmpl` |
| `data`     | no       | JSON file from `data/` passed as `.Data` to the template                                    |

## Markdown

Headings (`#`–`######`), bold (`**`), italic (`_`), bold+italic (`**_`), strikethrough (`~~`), inline code (`` ` ``), fenced code blocks, blockquotes, ordered/unordered lists, tables, links, images.

## Configuration

`gokesh.toml` at the project root. All fields are optional.

| Key           | Template             | Description             |
| ------------- | -------------------- | ----------------------- |
| `author`      | `{{ .Author }}`      | Author name             |
| `site_title`  | `{{ .SiteTitle }}`   | Site title              |
| `base_url`    | `{{ .BaseURL }}`     | Base URL of the site    |
| `description` | `{{ .Description }}` | Site description        |
| —             | `{{ .Year }}`        | Current year (auto)     |
| —             | `{{ .Body }}`        | Page HTML from markdown |

## Templates

Edit files in `templates/`. The default layout (`page.tmpl`) composes `header.tmpl`, `body.tmpl`, and `footer.tmpl`. Specify a different layout per page with `template:` in frontmatter.

## Make Commands

| Command                       | Description                                  |
| ----------------------------- | -------------------------------------------- |
| `make test`                   | Run tests                                    |
| `make vet`                    | Run go vet                                   |
| `make build`                  | Build binary to `bin/gokesh`                 |
| `make install`                | Install to `$GOPATH/bin`                     |
| `make dev`                    | Build example pages and start preview server |
| `make update-golden`          | Update golden test files                     |
| `make release VERSION=v0.1.0` | Tag, push, and trigger a GitHub release      |

## Releasing

```bash
make release VERSION=v0.1.0
```

Runs tests and vet, then tags and pushes. GitHub Actions builds binaries for all platforms and publishes to the [Releases page](https://github.com/vinckr/gokesh/releases).

## 1.0 Roadmap

1. **Incremental builds** — only rebuild pages newer than their output
1. **RSS feed** — auto-generate `public/feed.xml` from blog directory
1. **Structured CLI** — `flag` stdlib with `--version` and `--help`

## License

Apache 2.0 - see [LICENSE](LICENSE) for details.
