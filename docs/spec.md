# Gokesh — Product Specification (v1.0 target)

> This document is the authoritative source of truth for what gokesh does today, what is broken, and what must be built before shipping to customers. It supersedes `spec.md`.
>
> Confidence markers: **[HIGH]** = verified in code, **[MED]** = inferred from behaviour, **[LOW]** = assumption.

---

## 1. What gokesh is

Gokesh is a zero-dependency static site generator (SSG) written in Go. It takes a directory of Markdown files with YAML-like frontmatter, runs them through Go templates, and outputs a directory of static HTML files ready to deploy to any host.

**Design constraints that must not be broken:**
- Zero external Go module dependencies (stdlib only)
- Clean URL output (`/slug/` not `/slug.html`)
- Single binary distribution
- Works offline at runtime

**Target users:**
- Developer-bloggers who want full control without a framework
- Go developers who want to read and own their build pipeline
- NOT CMS users, NOT non-technical writers

---

## 2. Current state — observed behavior (code-backed)

### 2.1 Commands

| Command | What it actually does |
|---|---|
| `gokesh init` | Interactive wizard: writes `gokesh.toml`, copies embedded templates/styles/README, optionally injects Tailwind CDN into `header.tmpl`, writes `public/_headers` and `public/site.webmanifest` |
| `gokesh build` | Copies `./styles/` → `./public/`, builds all `.md` files in `./markdown/` recursively, generates `sitemap.xml` |
| `gokesh build page <name>` | Builds `./markdown/<name>.md` → `./public/<name>/index.html` only; does NOT copy styles or regenerate sitemap |
| `gokesh build dir <name>` | Builds all files in `./markdown/<name>/` (non-recursive); does NOT copy styles or regenerate sitemap |
| `gokesh watch` | Loads config once at startup, polls watched dirs every 1 s, triggers full rebuild on any mtime change; blocks until killed |
| `gokesh dev` | Serves `./public/` via `http.FileServer` on `:8000`; separate process from watch |

All paths (`./markdown/`, `./templates/`, `./public/`, `./styles/`, `./data/`) are hardcoded relative to the current working directory. You must run gokesh from the project root. **[HIGH]** — `main.go`, `build.go`, `watch.go`

### 2.2 Build pipeline (single page)

```
.md file
  → parser.ParseFrontmatter()      split ---delimiters--- from body
  → parser.ToHTML()                markdown body → HTML string
  → os.ReadFile("./data/"+file)    optional: load JSON data file
  → resolveTemplates()             glob templates/*.tmpl, entry last
  → template.ParseFiles() + Execute("Page")
  → WriteHTMLFile()                clean URL output
```

### 2.3 Data model

**`Config`** (loaded once from `gokesh.toml`):
```
Author, SiteTitle, BaseURL, Description
```

**`pageData`** (passed to every template render):
```
Body, SiteTitle, BaseURL, Description, Year, Author, Data, Pagematter.PageTitle
```

`Year` is always the *build year*, not the post date. **[HIGH]** — `build.go:BuildPageAt()`

### 2.4 Frontmatter fields

Flat key-value strings only. No arrays, no booleans, no dates.

| Field | Behaviour |
|---|---|
| `title` | Required by convention; parser returns empty string if missing; template renders empty `<title>` tag with no error |
| `template` | Entry template name (without `.tmpl`); default `page` |
| `data` | JSON filename read from `./data/<value>`; path is hardcoded |

---

## 3. Design flaws and bugs (must fix before ship)

### CRITICAL

**CF-1: CSP blocks the Tailwind CDN `init` just added**
`init` injects `<script src="https://unpkg.com/@tailwindcss/browser@4">` into `header.tmpl` AND writes a `_headers` file with `Content-Security-Policy: default-src 'self'`. The CSP blocks the external script. A user who answers "yes" to Tailwind during `init` ends up with a broken site. **[HIGH]** — `init.go:setupCSS()` and `init.go:writeHeaders()`

**CF-2: Tailwind browser CDN is not for production**
`@tailwindcss/browser` processes CSS at runtime in the browser. It is explicitly labelled "not for production" in Tailwind's documentation. Shipping this to users without a prominent warning is misleading.

**CF-3: `site.webmanifest` references icons that are never created**
`init` writes `site.webmanifest` referencing `/img/icon-192.png` and `/img/icon-512.png`, but those files are never generated or copied. The manifest is broken on every new project. **[HIGH]** — `init.go:writeWebManifest()`

**CF-4: Config changes during `watch` do not take effect**
`cfg` is loaded once before `Watch()` is called. If `gokesh.toml` changes while watching, the watch loop detects the mtime change and rebuilds — but with the stale `cfg`. Config updates require restarting the process. **[HIGH]** — `watch.go:Watch()`

**CF-5: Fenced code blocks do not escape HTML**
Content inside `` ``` `` fences is passed to `escapeHTML()` (confirmed in code), so this is actually fine. *(Correction from the previous spec — the earlier spec was wrong about this being a vulnerability.) **[HIGH, VERIFIED]** — `markdown.go` line 85-86: `out.WriteString(escapeHTML(line))`*

**CF-6: Frontmatter values containing colons are silently truncated**
`parseYAMLLines` uses `strings.Cut(line, ":")` — splits on the first colon. `title: "Link: https://example.com"` parses as `title = '"Link'`. URLs and subtitles in frontmatter are broken. **[HIGH]** — `parser/frontmatter.go:parseYAMLLines()`

**CF-7: `BuildPages` builds non-markdown files**
`BuildPages` calls `GetFilesFromDirectory` (returns all DirEntry) and calls `BuildPage` on every entry including non-`.md` files and subdirectories. **[HIGH]** — `build.go:BuildPages()`

**CF-8: README.md is unconditionally overwritten by `init`**
`init` always writes the gokesh README to the project directory. Running `gokesh init` in an existing project silently destroys the project's own README. **[HIGH]** — `init.go:runInit()`

### HIGH SEVERITY

**HF-1: `build page` and `build dir` skip style copying and sitemap**
Partial builds leave `public/` in an inconsistent state (missing styles, stale sitemap). Users who use partial builds during development will be surprised. **[HIGH]** — `main.go`

**HF-2: `data/` directory is not watched**
`watch.go` watches `markdown/`, `templates/`, `styles/`, `gokesh.toml` but not `data/`. Changing a JSON data file does not trigger a rebuild. **[HIGH]** — `watch.go:Watch()`

**HF-3: `watch` and `dev` are separate processes with no live reload**
Typical workflow requires two terminals: one for `watch`, one for `dev`. No browser reload on rebuild. This is a significant DX friction point compared to every competing SSG.

**HF-4: No `--help` or `--version` flags**
`gokesh` with no args prints usage and exits 1 (error exit). `gokesh --help` prints "unknown command". There is no version info available from the binary. **[HIGH]** — `main.go`

**HF-5: No per-page `<meta name="description">`**
Only the global `description` from `gokesh.toml` exists. The default `header.tmpl` does not include any `<meta name="description">` tag. Every page has identical or no SEO description. Not shippable for a blog/docs site.

**HF-6: `itoa()` only works for single digits**
`markdown.go:itoa(n)` returns `string(rune('0' + n))`. This works for 0–9 but produces garbage for n≥10. Currently safe because heading levels are capped at 6, but the function is a time bomb if reused. **[HIGH]** — `markdown.go`

**HF-7: `gokesh build page` README documentation is wrong**
README says "`Build markdown/<name>.md → public/<name>.html`" but actual output is `public/<name>/index.html`. **[HIGH]** — `README.md:65`

**HF-8: README still references `.env`**
The Getting Started section includes `.env` in the example project structure. The `.env` approach was replaced by `gokesh.toml`. Confusing for new users. **[HIGH]** — `README.md:26`

### MEDIUM SEVERITY

**MF-1: No `title` validation**
Missing `title` in frontmatter silently produces `<title> · Site Name</title>`. Should warn or error.

**MF-2: Paragraph soft breaks are not preserved**
Multi-line paragraph text is joined with a space. Hard line breaks (`  \n` in CommonMark) are not supported. Not a bug per se, but undocumented and surprising.

**MF-3: Watch debounce missing**
Saving multiple files rapidly (e.g., auto-formatter on save) triggers one rebuild per changed file sequentially. Not harmful but wasteful.

**MF-4: `init` copies `templates-examples/` in addition to `templates/`**
The intent seems to be "example copy" vs "live copy", but the logic in `copyEmbedDir` is confusing and the `templates-examples/` directory adds noise to user projects. **[MED]** — `init.go:copyEmbedDir()`

---

## 4. Missing features required to ship to customers

Grouped by priority. Features marked **BLOCKER** must ship in v1.0. Others are v1.x.

### 4.1 DX / developer experience (BLOCKER)

| ID | Feature | Why |
|---|---|---|
| DX-1 | **Combined `serve` command** (watch + live reload in one process) | Every competing SSG has this. Running two terminals is unacceptable DX. |
| DX-2 | **`--version` flag** | Users must be able to report what version they have. |
| DX-3 | **`--help` flag** | Should not exit 1 on `--help`. |
| DX-4 | **`gokesh new <name>`** | Scaffold a new markdown file with frontmatter pre-filled (title, date, template). Without this, users copy-paste frontmatter manually. |
| DX-5 | **Configurable output directory** | `--out` flag or `output_dir` in `gokesh.toml`. Currently hardcoded to `./public/`. |
| DX-6 | **`gokesh clean`** | Delete the output directory. Needed before full rebuilds. |

### 4.2 Content model (BLOCKER)

| ID | Feature | Why |
|---|---|---|
| CM-1 | **`date` frontmatter field** | Every blog post needs a date. Currently impossible without hacks. |
| CM-2 | **`draft: true` frontmatter** | Drafts must be excluded from builds. Baseline feature of every SSG. |
| CM-3 | **Per-page `description` frontmatter** | SEO requirement. Template should expose `{{ .Pagematter.Description }}`. |
| CM-4 | **`slug` frontmatter field** | Allow overriding the output URL without renaming the file. |
| CM-5 | **Tag/taxonomy frontmatter** | `tags: [go, web]` as an array value. Requires frontmatter parser to support arrays. |

### 4.3 Build pipeline (BLOCKER)

| ID | Feature | Why |
|---|---|---|
| BP-1 | **Incremental builds** | Full rebuild on every watch tick is unusable for sites with 100+ pages. Only rebuild files newer than their output. |
| BP-2 | **`static/` directory copy** | CSS-only `styles/` is too narrow. Users need to copy images, fonts, favicons, JS. A generic `static/` → `public/` copy is standard. |
| BP-3 | **Watch `data/` directory** | Data file changes must trigger rebuild. |
| BP-4 | **Partial builds copy styles + regenerate sitemap** | `build page` and `build dir` should be usable standalone. |

### 4.4 Templates and output (BLOCKER)

| ID | Feature | Why |
|---|---|---|
| TP-1 | **`<meta name="description">` in default template** | Required for SEO. Use per-page description falling back to site description. |
| TP-2 | **Open Graph meta tags in default template** | Required for social sharing previews (`og:title`, `og:description`, `og:url`). |
| TP-3 | **`dateFormat` template function** | Format a date string for display. Without this `date` frontmatter is useless in templates. |
| TP-4 | **`pages` template function or global** | Access the list of all built pages in templates (title, URL, date). Needed for navigation and index pages. Without this you cannot build a blog post list. |
| TP-5 | **Code block language class** | Fenced code blocks should emit `<code class="language-go">` when a language specifier is given. Required for syntax highlighting libraries. |

### 4.5 RSS feed (BLOCKER for blog use case)

| ID | Feature | Why |
|---|---|---|
| RSS-1 | **Auto-generate `public/feed.xml`** | Baseline expectation for any blog. Requires `date` frontmatter (CM-1). |

### 4.6 Init command fixes (BLOCKER)

| ID | Feature | Why |
|---|---|---|
| IN-1 | **Fix CSP + Tailwind contradiction** | Either omit Tailwind from init or adjust CSP to allow the CDN. |
| IN-2 | **Do not overwrite existing README.md** | Check before writing. |
| IN-3 | **Remove or fix broken PWA icons** | Either bundle placeholder icons or omit `site.webmanifest` until icons exist. |
| IN-4 | **Document Tailwind CDN limitation** | Warn users at `init` time that the CDN build is for development only. |
| IN-5 | **Remove `templates-examples/` copy** | Unnecessary noise. Copy only to `templates/`. |

### 4.7 v1.x (post-ship)

| ID | Feature | Why |
|---|---|---|
| V1X-1 | Section index auto-generation | Auto-build `public/blog/index.html` listing all posts in `markdown/blog/` |
| V1X-2 | Pagination | Split long post lists across multiple pages |
| V1X-3 | Nested lists in markdown | `- item\n  - nested` not currently supported |
| V1X-4 | Asset fingerprinting | `style.abc123.css` for cache busting |
| V1X-5 | Syntax highlighting (server-side) | Emit highlighted HTML at build time, not in the browser |
| V1X-6 | Multiple config environments | `gokesh.dev.toml` vs `gokesh.prod.toml` |
| V1X-7 | Structured CLI with `flag` package | `--version`, `--help`, `--config`, `--out` as proper flags |

---

## 5. Frontmatter — target spec (v1.0)

The frontmatter parser must support:
- Flat string values (current)
- Array values: `tags: [go, web, ssr]` → `[]string`
- Boolean values: `draft: true` → `bool`
- Date values: `date: 2026-01-15` → `time.Time`
- Values containing colons: `title: "Go: The Missing Manual"` (fix CF-6 by splitting on `": "` or limiting to first `:` followed by space)

**Full frontmatter reference (v1.0 target):**

| Field | Type | Required | Description |
|---|---|---|---|
| `title` | string | Yes — warn if missing | Page `<title>` and `{{.Pagematter.PageTitle}}` |
| `description` | string | No | Per-page meta description; falls back to `Config.Description` |
| `date` | date | No | Post date (`YYYY-MM-DD`); exposed as `{{.Pagematter.Date}}` |
| `draft` | bool | No | If `true`, page is excluded from all builds |
| `slug` | string | No | Override output URL path |
| `template` | string | No | Entry template (default: `page`) |
| `data` | string | No | JSON file from `data/` |
| `tags` | []string | No | Taxonomy tags |

---

## 6. `pageData` — target spec (v1.0)

```go
type pageData struct {
    // from Config
    SiteTitle   string
    BaseURL     string
    Description string
    Author      string

    // build-time
    Year        string          // build year
    Body        string          // rendered HTML
    Data        json.RawMessage // optional JSON data file

    // per-page frontmatter
    Pagematter struct {
        PageTitle   string
        Description string    // NEW: per-page description
        Date        time.Time // NEW: parsed post date
        Tags        []string  // NEW: taxonomy tags
        Slug        string    // NEW: URL override
        Draft       bool      // NEW: exclude from build
    }

    // site-wide
    Pages []PageSummary // NEW: all non-draft pages (for nav/index templates)
}

type PageSummary struct {
    Title string
    URL   string
    Date  time.Time
    Tags  []string
}
```

---

## 7. Template functions — target spec (v1.0)

| Function | Signature | Description |
|---|---|---|
| `jsonify` | `(json.RawMessage) string` | Embed raw JSON (existing) |
| `items` | `(json.RawMessage) []map[string]any` | Range over JSON array (existing) |
| `dateFormat` | `(time.Time, string) string` | Format a date: `{{ dateFormat .Pagematter.Date "Jan 2, 2006" }}` |
| `pages` | Available as `.Pages` on `pageData` | All non-draft pages, sorted newest-first |
| `sortBy` | `([]PageSummary, string) []PageSummary` | Sort pages by field name |
| `filterByTag` | `([]PageSummary, string) []PageSummary` | Filter pages by tag |

---

## 8. Default template — target spec (v1.0)

The embedded `header.tmpl` must include at minimum:

```html
<meta name="description" content="{{ if .Pagematter.Description }}{{ .Pagematter.Description }}{{ else }}{{ .Description }}{{ end }}">
<meta property="og:title" content="{{ .Pagematter.PageTitle }} · {{ .SiteTitle }}">
<meta property="og:description" content="{{ if .Pagematter.Description }}{{ .Pagematter.Description }}{{ else }}{{ .Description }}{{ end }}">
<meta property="og:url" content="{{ .BaseURL }}/{{ .Pagematter.Slug }}/">
<link rel="canonical" href="{{ .BaseURL }}/{{ .Pagematter.Slug }}/">
<link rel="alternate" type="application/rss+xml" title="{{ .SiteTitle }}" href="{{ .BaseURL }}/feed.xml">
```

---

## 9. Config — target spec (v1.0)

`gokesh.toml` gains two optional fields:

```toml
author       = "vinckr"
site_title   = "My Site"
base_url     = "https://example.com"
description  = "My personal site"
output_dir   = "public"      # NEW: default "public"
markdown_dir = "markdown"    # NEW: default "markdown"
```

The custom TOML parser in `parser/config.go` can handle these with no structural changes.

---

## 10. CLI — target spec (v1.0)

```
gokesh [--version] [--help]

Commands:
  init               Set up a new project
  build              Build all pages
  build page <name>  Build a single page (+ copy styles)
  build dir <name>   Build a directory (+ copy styles)
  new <name>         Create markdown/<name>.md with pre-filled frontmatter
  clean              Delete output directory
  serve              Watch + dev server + live reload in one command
  watch              Watch and rebuild (no server)
  dev                Serve output directory
```

`serve` is the new primary development command. `watch` and `dev` remain for scripting/advanced use.

---

## 11. Known correct behavior (do not regress)

- Clean URL output: `foo.md` → `foo/index.html` → served at `/foo/` **[HIGH]**
- `index.md` at any depth → `index.html` (not `index/index.html`) **[HIGH]**
- Recursive directory structure preserved in output **[HIGH]**
- All templates parsed together; entry template last so its `{{define "Page"}}` wins **[HIGH]**
- `resolveTemplates` returns error if entry template not found **[HIGH]**
- Fenced code blocks HTML-escape their content **[HIGH, correction from prior spec]**
- `gokesh.toml` missing = silently use empty config, no error **[HIGH]**
- Tests use fixed timestamp for reproducible golden output **[HIGH]**

---

## 12. Build / test / release commands

```bash
make build            # compile → bin/gokesh
make install          # go install → $GOPATH/bin
make test             # go test ./...
make update-golden    # regenerate golden files
make vet              # go vet ./...
make dev              # build examples + serve on :8000
make release VERSION=v0.1.0  # tag + push → GoReleaser
```

GoReleaser builds: Linux/macOS/Windows × amd64/arm64. `.tar.gz` on Unix, `.zip` on Windows.

---

## 13. Testing strategy

| Layer | Framework | Location | Coverage target |
|---|---|---|---|
| Parser unit tests | `testing` stdlib | `internal/parser/*_test.go` | All block/inline elements, edge cases |
| Build unit tests | `testing` stdlib | `internal/build/build_test.go` | `WriteHTMLFile`, `BuildTemplate`, `SplitBodyAndFrontmatter` |
| Golden (integration) | `testing` stdlib + `-update` flag | `internal/build/golden_test.go` | Full build output matches baseline HTML |
| CLI smoke tests | `testing` + `os/exec` | Not yet implemented | **BLOCKER for v1.0**: test `init`, `build`, `serve` commands end-to-end |

All tests must be `t.Parallel()`. New features require golden file updates before merging.

---

## 14. Deployment targets

| Host | Notes |
|---|---|
| Netlify | `_headers` file respected; `public/` as publish dir |
| Cloudflare Pages | `_headers` file respected |
| GitHub Pages | Push `public/` to `gh-pages` branch |
| AWS S3 + CloudFront | Sync `public/` |
| Any static host | `public/` contains everything needed |

---

## 15. Out of scope (never do)

- Database or server-side rendering
- User authentication
- CMS or admin UI
- External Go module dependencies
- Plugin/extension system
- Themes marketplace

---

## 16. Success criteria for v1.0 ship

A v1.0 release is ready to ship when all of the following are true:

- [ ] All CRITICAL and HIGH flaws (CF-*, HF-*) are fixed
- [ ] All BLOCKER features (DX-1–6, CM-1–5, BP-1–4, TP-1–5, RSS-1, IN-1–5) are implemented
- [ ] `gokesh init` + `gokesh serve` workflow produces a working site in under 2 minutes on a blank directory
- [ ] `gokesh build` on a 50-page site completes in under 1 s
- [ ] Default templates include description, OG tags, and RSS link
- [ ] No page can silently render with an empty `<title>` tag
- [ ] `gokesh --version` works
- [ ] CLI smoke tests pass in CI
- [ ] README is accurate (no references to `.env`, correct `build page` output path)
