# Implementation Plan: Gokesh v1.0

> Derived from `spec-final.md`. Read-only planning pass — no code was written during this phase.
> All task IDs match `todo.md`.

---

## Architecture decisions

**AD-1: Frontmatter parser must be extended before anything else.**
Six features (date, draft, description, slug, tags, colon fix) all live in `parser/frontmatter.go`. Do this first so every downstream task builds on stable data.

**AD-2: Two-pass build for the `.Pages` global.**
Templates need access to _all_ pages (for nav, blog index, RSS). This requires: (1) walk all `.md` files and collect `PageSummary` metadata; (2) build each page passing the full list. This changes `BuildAll`'s signature and is the single biggest architectural change. Schedule it after the frontmatter expansion.

**AD-3: Draft filtering belongs in the build pipeline, not the parser.**
`draft: true` should be checked in `BuildAll`/`BuildPageAt` after frontmatter is parsed, not inside the parser. Keeps the parser a pure data extractor.

**AD-4: Live reload via SSE, not websocket.**
`net/http` supports SSE natively. Inject a small `<script>` into every served page that polls a `/~reload` endpoint. No new files, no external dependencies.

**AD-5: Config extended before CLI flags.**
`output_dir` and `markdown_dir` go into `gokesh.toml` first; CLI `--out` / `--src` flags come later (v1.x). Keeps CLI parsing simple during v1.0.

**AD-6: Zero new external dependencies, ever.**
Every solution must use Go stdlib only.

---

## Dependency graph

```
parser/frontmatter.go (colon fix, array/bool/date types)
    │
    ├── pageData struct expansion (description, date, tags, slug, draft)
    │       │
    │       ├── Draft filtering in BuildAll
    │       │
    │       ├── Two-pass build → .Pages global
    │       │       │
    │       │       ├── RSS feed generation
    │       │       └── sortBy / filterByTag template functions
    │       │
    │       └── dateFormat template function
    │               │
    │               └── default template update (OG, description, RSS link)
    │
    ├── (independent) Build pipeline: watch data/, static/, incremental, output dir
    ├── (independent) CLI: --version, --help, clean, new
    └── (independent) Combined serve command + live reload
```

---

## Phase 0: Critical bug fixes

All fixes are self-contained. No new features. All tests must pass after this phase.

---

### Task 0.1 — Fix frontmatter colon parsing

**Description:** `strings.Cut(line, ":")` splits on the first colon. `title: "Go: A Tour"` becomes `title = '"Go'`. Fix by splitting on `": "` (colon-space) and falling back to trimming only if the value is unquoted. Update tests.

**Acceptance criteria:**
- [ ] `title: "URL: https://example.com"` parses as `title = "URL: https://example.com"`
- [ ] `title: plain value` still parses correctly
- [ ] Existing frontmatter tests still pass

**Verification:** `go test ./internal/parser/...`

**Dependencies:** None

**Files:** `internal/parser/frontmatter.go`, `internal/parser/frontmatter_test.go`

**Size:** XS

---

### Task 0.2 — Filter `.md` files in `BuildPages`

**Description:** `BuildPages` currently calls `BuildPage` on every `DirEntry`, including non-`.md` files and subdirectories. Add a `strings.HasSuffix(file.Name(), ".md") && !file.IsDir()` guard.

**Acceptance criteria:**
- [ ] Non-`.md` files in a markdown directory are silently skipped
- [ ] Subdirectories are skipped
- [ ] All `.md` files still build correctly

**Verification:** `go test ./internal/build/...`

**Dependencies:** None

**Files:** `internal/build/build.go`

**Size:** XS

---

### Task 0.3 — Fix `init` README overwrite

**Description:** `init` unconditionally writes the embedded `README.md` over any existing file. Add a check: only write if `README.md` does not exist.

**Acceptance criteria:**
- [ ] Running `gokesh init` in a project with an existing `README.md` does not change it
- [ ] Running `gokesh init` in an empty directory still creates `README.md`

**Verification:** Manual test with and without existing README.md

**Dependencies:** None

**Files:** `cmd/gokesh/init.go`

**Size:** XS

---

### Task 0.4 — Fix Tailwind CDN + CSP contradiction in `init`

**Description:** `init` injects `<script src="https://unpkg.com/@tailwindcss/browser@4">` into `header.tmpl` AND writes `_headers` with `Content-Security-Policy: default-src 'self'` which blocks external scripts. Fix: if user says yes to Tailwind, relax the CSP to allow `unpkg.com`. Also print a visible warning that the CDN build is development-only.

**Acceptance criteria:**
- [ ] When Tailwind is chosen, `_headers` CSP includes `script-src 'self' https://unpkg.com`
- [ ] When Tailwind is not chosen, CSP remains `default-src 'self'`
- [ ] A warning is printed: "Note: @tailwindcss/browser is for development only. Replace before deploying to production."

**Verification:** Manual: run `gokesh init`, choose yes to Tailwind, inspect `public/_headers`

**Dependencies:** None

**Files:** `cmd/gokesh/init.go`

**Size:** S

---

### Task 0.5 — Fix broken PWA webmanifest icons

**Description:** `site.webmanifest` references `/img/icon-192.png` and `/img/icon-512.png` that are never created. Fix: remove the `icons` array from the manifest, or bundle placeholder SVG icons in `embed.FS` and copy them during `init`.

**Acceptance criteria:**
- [ ] `site.webmanifest` references no files that don't exist after `init`
- [ ] If icons are bundled, they are copied to `public/img/` during `init`

**Verification:** Manual: run `gokesh init`, check that all files referenced in `site.webmanifest` exist

**Dependencies:** None

**Files:** `cmd/gokesh/init.go`

**Size:** S

---

### Task 0.6 — Fix watch config reload

**Description:** `Watch()` receives `cfg Config` by value, loaded once in `main.go` before the loop starts. Config changes while watching take no effect. Fix: reload config from disk inside `fullBuild()` on every rebuild cycle.

**Acceptance criteria:**
- [ ] Changing `gokesh.toml` while `gokesh watch` is running causes the next build to use the updated config
- [ ] Watch still runs if `gokesh.toml` is missing (empty config, no error)

**Verification:** Manual: start `gokesh watch`, change `site_title` in `gokesh.toml`, verify the next rebuild uses the new title

**Dependencies:** None

**Files:** `internal/build/watch.go`

**Size:** S

---

### Task 0.7 — Fix `itoa()`, partial builds, and README docs

**Description:** Three small independent fixes bundled to avoid trivial PRs:
1. `itoa(n)` in `markdown.go` only works for 0–9. Replace with `strconv.Itoa(n)` (stdlib, no new dep).
2. `gokesh build page` and `gokesh build dir` do not copy styles. Add `CopyStyles` call before partial builds in `main.go`.
3. README says `build page` outputs `public/<name>.html` — correct to `public/<name>/index.html`. Remove `.env` from the example project structure.

**Acceptance criteria:**
- [ ] `itoa` replaced with `strconv.Itoa` everywhere in markdown.go
- [ ] `gokesh build page foo` copies `styles/` to `public/` before building
- [ ] README example project structure has no `.env`
- [ ] README `build page` description says `public/<name>/index.html`

**Verification:** `go test ./...`; manual: run `gokesh build page`, confirm `public/style.css` exists

**Dependencies:** None

**Files:** `internal/parser/markdown.go`, `cmd/gokesh/main.go`, `README.md`

**Size:** S

---

### Task 0.8 — Remove `templates-examples/` from `init`

**Description:** `init` copies embedded templates to both `templates-examples/` and `templates/`. The `-examples` copy adds noise. Remove it; copy only to `templates/`.

**Acceptance criteria:**
- [ ] After `gokesh init`, only `templates/` exists (not `templates-examples/`)
- [ ] `styles-examples/` is also removed; only `styles/` is created

**Verification:** Manual: run `gokesh init` in a temp dir, confirm no `-examples` directories

**Dependencies:** None

**Files:** `cmd/gokesh/init.go`

**Size:** XS

---

### Checkpoint 0

- [ ] `go test ./...` passes with no failures
- [ ] `go vet ./...` clean
- [ ] `gokesh init` → `gokesh build` → `gokesh dev` works end-to-end in a fresh directory
- [ ] No files referenced in `site.webmanifest` are missing

---

## Phase 1: Frontmatter + content model

The foundational expansion. Everything in Phase 4 (templates, RSS) depends on these types being available.

---

### Task 1.1 — Extend frontmatter parser: arrays, booleans, dates

**Description:** The current parser only handles flat string values. Add support for:
- **Arrays:** `tags: [go, web, ssr]` → `[]string`
- **Booleans:** `draft: true` / `draft: false` → `bool`
- **Dates:** `date: 2026-01-15` → `time.Time` (parsed as RFC3339 date, `2006-01-02` layout)

The parser's public API should remain `ParseFrontmatter(content []byte) (matter map[string]string, body []byte)` for backward compat. Add a second function `ParseFrontmatterTyped(content []byte) (FrontmatterFields, []byte)` that returns a typed struct.

**Acceptance criteria:**
- [ ] `tags: [go, web]` → `FrontmatterFields.Tags = []string{"go", "web"}`
- [ ] `draft: true` → `FrontmatterFields.Draft = true`
- [ ] `date: 2026-01-15` → `FrontmatterFields.Date = time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)`
- [ ] Invalid date value → warning logged, zero `time.Time` (no fatal error)
- [ ] Existing `ParseFrontmatter` API unchanged; all existing tests pass
- [ ] New tests cover all new value types

**Verification:** `go test ./internal/parser/...`

**Dependencies:** Task 0.1 (colon fix must be in before extending)

**Files:** `internal/parser/frontmatter.go`, `internal/parser/frontmatter_test.go`

**Size:** M

---

### Task 1.2 — Expand `pageData` and `pageData.Pagematter`

**Description:** Add new fields to `pageData` and `Pagematter` as specified in `spec-final.md` §6. Wire them up from the typed frontmatter parsed in 1.1. `Year` remains the build year (not post year — this is intentional for footer copyrights).

**New fields:**
```go
Pagematter struct {
    PageTitle   string
    Description string   // frontmatter "description"
    Date        time.Time
    Tags        []string
    Slug        string
    Draft       bool
}
Pages []PageSummary  // populated in Phase 1.4
```

**Acceptance criteria:**
- [ ] `{{.Pagematter.Description}}` available in templates
- [ ] `{{.Pagematter.Date}}` available as `time.Time`
- [ ] `{{.Pagematter.Tags}}` available as `[]string`
- [ ] Existing golden tests still pass (backward compat — old templates don't use new fields)

**Verification:** `go test ./internal/build/...`

**Dependencies:** Task 1.1

**Files:** `internal/build/build.go`

**Size:** S

---

### Task 1.3 — Draft filtering in build pipeline

**Description:** Pages with `draft: true` in frontmatter must be excluded from all builds. Add a check in `BuildPageAt` that returns `nil` (skip, no error) if `matter.Draft == true`. Also skip in the `BuildAll` walk. Add a `slog.Info("skipping draft", "file", fileName)` log line.

**Acceptance criteria:**
- [ ] A `.md` file with `draft: true` produces no output in `public/`
- [ ] The draft file does not appear in `sitemap.xml`
- [ ] A build with only drafts succeeds without error (zero pages built)
- [ ] Pages without `draft` field or with `draft: false` build normally

**Verification:** Add a draft test fixture; `go test ./internal/build/...`

**Dependencies:** Task 1.2

**Files:** `internal/build/build.go`, `internal/build/build_test.go`

**Size:** S

---

### Task 1.4 — Two-pass build: `.Pages` global

**Description:** This is the biggest architectural change in Phase 1. Templates need access to all pages for navigation menus and blog post indexes.

**Implementation:**
1. Add `CollectPages(markdownDir string) ([]PageSummary, error)` that walks the markdown directory and parses only frontmatter (no HTML rendering, fast). Returns non-draft pages sorted newest-first by `Date`.
2. Thread `[]PageSummary` through `BuildAll` → `BuildPageAt` → `pageData.Pages`.
3. `BuildPage` (single page build) calls `CollectPages` first, passes result to `BuildPageAt`.

```go
type PageSummary struct {
    Title string
    URL   string
    Date  time.Time
    Tags  []string
}
```

**Acceptance criteria:**
- [ ] `{{range .Pages}}{{.Title}} — {{.URL}}{{end}}` renders all non-draft page titles in templates
- [ ] Pages are sorted newest-first (zero `Date` pages appear last)
- [ ] Draft pages are excluded from `.Pages`
- [ ] Single-page build (`build page`) also populates `.Pages`
- [ ] All existing golden tests still pass

**Verification:** Add a golden test that uses `.Pages`; `go test ./internal/build/...`

**Dependencies:** Task 1.3

**Files:** `internal/build/build.go`, `internal/build/build_test.go`, `internal/build/golden_test.go`

**Size:** M

---

### Checkpoint 1

- [ ] `go test ./...` clean
- [ ] Template `{{.Pagematter.Date}}`, `{{.Pagematter.Tags}}`, `{{.Pages}}` all resolve without error
- [ ] Draft pages excluded from output and sitemap
- [ ] `.Pages` global populated on every build

---

## Phase 2: Build pipeline improvements

Independent of Phase 1's content model changes (can be developed in parallel by a second agent, but must merge before Phase 4).

---

### Task 2.1 — Watch `data/` directory

**Description:** `watch.go` does not watch `./data/`. Add it to the `watched` slice. This is a one-line change.

**Acceptance criteria:**
- [ ] Changing a file in `data/` while `gokesh watch` is running triggers a rebuild

**Verification:** Manual: start watch, modify `data/foo.json`, verify rebuild fires

**Dependencies:** None

**Files:** `internal/build/watch.go`

**Size:** XS

---

### Task 2.2 — `static/` directory copy

**Description:** `styles/` is too narrow — users also need to deploy images, fonts, and JS. Add a `CopyStatic(staticDir, outpath string) error` function that copies all files from `static/` → `public/` recursively (preserving subdirectory structure). Call it from `fullBuild()` in `watch.go` and from the `build` command in `main.go`. Silently skip if `static/` doesn't exist (same behaviour as `CopyStyles`).

**Acceptance criteria:**
- [ ] `static/img/logo.png` is copied to `public/img/logo.png`
- [ ] Subdirectories inside `static/` are created in `public/`
- [ ] Missing `static/` dir does not error
- [ ] `gokesh watch` triggers re-copy when static files change (add `./static/` to watched paths)

**Verification:** `go test ./internal/build/...`; manual test with image in `static/`

**Dependencies:** None

**Files:** `internal/build/build.go`, `internal/build/watch.go`, `cmd/gokesh/main.go`

**Size:** S

---

### Task 2.3 — Incremental builds

**Description:** Every `watch` tick triggers a full rebuild. Compare each source file's `ModTime` against its corresponding output file. Only call `BuildPageAt` if the source is newer than the output or the output doesn't exist.

Add `shouldRebuild(srcPath, outPath string) bool` helper. Also skip style/static copy if nothing in `styles/` or `static/` is newer than `public/style.css`.

**Acceptance criteria:**
- [ ] A site with 50 unchanged pages triggers 0 `BuildPageAt` calls on a second watch tick
- [ ] Modifying one `.md` file rebuilds only that page
- [ ] Deleting an output file causes it to be rebuilt on the next tick
- [ ] `gokesh build` (full build) always rebuilds all (ignore mtimes)
- [ ] `gokesh watch` uses incremental builds

**Verification:** Add unit test for `shouldRebuild`; manual: watch a multi-page site, edit one file, observe log shows only one page rebuilt

**Dependencies:** None (can be added without Phase 1)

**Files:** `internal/build/build.go`, `internal/build/watch.go`

**Size:** M

---

### Task 2.4 — Configurable output and markdown directories

**Description:** Add `output_dir` and `markdown_dir` config keys to `gokesh.toml`. Default to `"public"` and `"markdown"` respectively to preserve backward compat. Thread through `Config` struct and replace all hardcoded `"./public/"` and `"./markdown/"` path constructions in `main.go`.

**Acceptance criteria:**
- [ ] Setting `output_dir = "dist"` in `gokesh.toml` causes output to go to `dist/`
- [ ] Default behaviour (no config key) still uses `public/`
- [ ] `gokesh watch` uses the configured directories

**Verification:** `go test ./...`; manual test with non-default `output_dir`

**Dependencies:** None

**Files:** `internal/build/build.go`, `internal/parser/config.go`, `cmd/gokesh/main.go`

**Size:** S

---

### Checkpoint 2

- [ ] `go test ./...` clean
- [ ] `gokesh watch` only rebuilds changed files
- [ ] `static/` images appear in `public/`
- [ ] `data/` changes trigger rebuild during watch

---

## Phase 3: CLI improvements

All tasks in this phase are independent of each other and of Phases 1–2.

---

### Task 3.1 — `--version` and `--help` flags

**Description:** Handle `--version` / `-v` and `--help` / `-h` before the command switch in `main.go`. Version string is injected at build time via `ldflags` (`-X main.version=v0.1.0`). Default to `dev` if not set. `--help` prints usage and exits 0 (not 1).

**Acceptance criteria:**
- [ ] `gokesh --version` prints `gokesh v0.1.0` and exits 0
- [ ] `gokesh --help` prints usage and exits 0
- [ ] `gokesh` with no args still prints usage and exits 1 (unchanged)
- [ ] GoReleaser `.goreleaser.yml` passes version ldflags

**Verification:** `./bin/gokesh --version`; `./bin/gokesh --help`; `echo $?`

**Dependencies:** None

**Files:** `cmd/gokesh/main.go`, `.goreleaser.yml`

**Size:** S

---

### Task 3.2 — `gokesh clean`

**Description:** Add a `clean` command that deletes the output directory (`./public/` by default, or the configured `output_dir`). Prompt for confirmation if the directory exists and is non-empty: `"Delete public/ (N files)? [y/N]"`. Exit 0 if directory doesn't exist.

**Acceptance criteria:**
- [ ] `gokesh clean` deletes `public/`
- [ ] Prompts for confirmation before deleting
- [ ] Exits 0 cleanly if `public/` doesn't exist
- [ ] After Phase 2.4: uses configured output dir

**Verification:** Manual: create `public/`, run `gokesh clean`, confirm deletion

**Dependencies:** Task 2.4 (for configured output dir), but can be implemented with hardcoded path first

**Files:** `cmd/gokesh/main.go`

**Size:** S

---

### Task 3.3 — `gokesh new <name>`

**Description:** Scaffold a new markdown file at `markdown/<name>.md` with pre-filled frontmatter. Prompt for title (required) and optionally date (defaults to today). Fail if file already exists.

```markdown
---
title: "My Post"
date: 2026-05-05
draft: true
---
```

**Acceptance criteria:**
- [ ] `gokesh new my-post` creates `markdown/my-post.md`
- [ ] File includes `title`, `date` (today's date), `draft: true` frontmatter
- [ ] Fails with clear error if `markdown/my-post.md` already exists
- [ ] Prints the path of the created file

**Verification:** Manual test in a project directory

**Dependencies:** None

**Files:** `cmd/gokesh/main.go` or new `cmd/gokesh/new.go`

**Size:** S

---

### Checkpoint 3

- [ ] `go test ./...` clean
- [ ] `gokesh --version` works
- [ ] `gokesh new test-post` creates a valid markdown file
- [ ] `gokesh clean` deletes output without leaving files behind

---

## Phase 4: Templates, SEO, and RSS

Depends on Phase 1 (for `pageData.Pagematter.Description`, `.Date`, `.Tags`, `.Pages`).

---

### Task 4.1 — Code block language class

**Description:** Fenced code blocks should emit `<code class="language-go">` when a language specifier follows the opening fence (` ```go `). This allows syntax highlighting libraries (Prism, highlight.js) to work without changes to gokesh itself.

Change the fenced code block handler in `markdown.go` to capture the language hint and emit it as a class.

**Acceptance criteria:**
- [ ] ` ```go ` emits `<pre><code class="language-go">`
- [ ] ` ``` ` (no language) emits `<pre><code>` (unchanged)
- [ ] Content inside the fence is still HTML-escaped

**Verification:** Add test cases to `markdown_test.go`; `go test ./internal/parser/...`

**Dependencies:** None

**Files:** `internal/parser/markdown.go`, `internal/parser/markdown_test.go`

**Size:** XS

---

### Task 4.2 — Template functions: `dateFormat`, `sortBy`, `filterByTag`

**Description:** Add three new functions to the `funcMap` in `BuildTemplate`:

- `dateFormat(t time.Time, layout string) string` — e.g. `{{ dateFormat .Pagematter.Date "Jan 2, 2006" }}`
- `sortBy(pages []PageSummary, field string) []PageSummary` — sort by `"date"` or `"title"`
- `filterByTag(pages []PageSummary, tag string) []PageSummary` — filter to pages that include the tag

**Acceptance criteria:**
- [ ] `{{ dateFormat .Pagematter.Date "2006-01-02" }}` renders the date correctly in templates
- [ ] `{{ range sortBy .Pages "date" }}` returns pages newest-first
- [ ] `{{ range filterByTag .Pages "go" }}` returns only pages tagged `go`
- [ ] Zero-value `time.Time` with `dateFormat` returns `""` (empty string, no error)

**Verification:** Add test cases to `build_test.go`; `go test ./internal/build/...`

**Dependencies:** Task 1.4 (`.Pages` available), Task 1.2 (`.Date` on Pagematter)

**Files:** `internal/build/build.go`, `internal/build/build_test.go`

**Size:** S

---

### Task 4.3 — Update default templates (SEO meta tags)

**Description:** Update the embedded `templates/header.tmpl` to include:
- `<meta name="description">` (per-page description falling back to site description)
- `og:title`, `og:description`, `og:url` Open Graph tags
- `<link rel="canonical">`
- `<link rel="alternate" type="application/rss+xml">` (pointing to `/feed.xml`)

Also fix the `<title>` tag: if `Pagematter.PageTitle` is empty, use `SiteTitle` alone (not `· Site Name`).

**Acceptance criteria:**
- [ ] Built pages include `<meta name="description" content="...">` with non-empty content
- [ ] Built pages include `og:title` and `og:url`
- [ ] `<title>` on a page without frontmatter `title` renders `SiteTitle` (not ` · SiteTitle`)
- [ ] Golden tests updated to reflect new template output

**Verification:** `go test ./internal/build/...` (update goldens with `make update-golden`)

**Dependencies:** Task 1.2 (Pagematter.Description), Task 4.2 (for RSS link) — but can be done first with static `/feed.xml` href

**Files:** `templates/header.tmpl`, `internal/build/testdata/golden/*.html`

**Size:** S

---

### Task 4.4 — RSS feed generation

**Description:** Add `GenerateRSS(pages []PageSummary, outpath, siteTitle, baseURL string) error` in a new `internal/build/rss.go`. Generate a valid RSS 2.0 `feed.xml` at `outpath/feed.xml`. Call it at the end of `fullBuild()` and `gokesh build`. Skip (with a warning) if `baseURL` is empty.

RSS item fields: `<title>`, `<link>`, `<pubDate>` (from `Date`), `<description>` (page description).

**Acceptance criteria:**
- [ ] `public/feed.xml` is a valid RSS 2.0 document
- [ ] Only non-draft pages with a `date` appear in the feed
- [ ] Items are sorted newest-first
- [ ] Skipped (warning logged) if `base_url` is not set in config

**Verification:** `go test ./internal/build/...`; validate `feed.xml` with an RSS validator

**Dependencies:** Task 1.4 (`.Pages`), Task 2.4 (output dir)

**Files:** `internal/build/rss.go`, `internal/build/build.go`, `internal/build/watch.go`

**Size:** M

---

### Task 4.5 — Warn on missing `title` frontmatter

**Description:** In `BuildPageAt`, after parsing frontmatter, log a warning if `PageTitle` is empty: `slog.Warn("page has no title", "file", fileName)`. Do not error — allow the build to continue. Template authors can choose to handle empty titles.

**Acceptance criteria:**
- [ ] Building a page without `title:` logs a warning
- [ ] Build does not fail

**Verification:** `go test ./internal/build/...`; check log output in a test

**Dependencies:** Task 1.2

**Files:** `internal/build/build.go`

**Size:** XS

---

### Checkpoint 4

- [ ] `go test ./...` clean; goldens updated
- [ ] Built pages have valid `<meta name="description">` and OG tags
- [ ] `public/feed.xml` validates as RSS 2.0
- [ ] Syntax-highlighted pages work (emit language class)
- [ ] A template using `{{range sortBy .Pages "date"}}` renders a post list

---

## Phase 5: Combined `serve` command with live reload

This is the highest-complexity task. Schedule last to avoid blocking other phases.

---

### Task 5.1 — `gokesh serve`: watch + dev server in one process

**Description:** Add a `serve` command that combines `watch` and `dev` into a single blocking process. Run the file watcher in a goroutine; run the HTTP server in the main goroutine (or vice versa).

**Acceptance criteria:**
- [ ] `gokesh serve` starts both the watcher and the HTTP server
- [ ] File changes trigger a rebuild; server continues serving without restart
- [ ] `gokesh watch` and `gokesh dev` continue to work independently

**Verification:** Manual: start `gokesh serve`, edit a markdown file, verify rebuild and continued serving

**Dependencies:** None (can be done without live reload)

**Files:** `cmd/gokesh/main.go`, `internal/build/watch.go`

**Size:** M

---

### Task 5.2 — Live reload via SSE

**Description:** Inject a small `<script>` into every page served by the dev server (not in built output) that subscribes to an SSE endpoint (`/~reload`). When a rebuild completes, broadcast a reload event. Browser reloads the current page automatically.

Implementation:
- Add `/~reload` SSE handler to the dev server
- After every successful `fullBuild()`, send an event on the SSE channel
- The dev server wraps `http.FileServer` and injects a `<script>` tag into HTML responses

**Acceptance criteria:**
- [ ] Editing a markdown file causes the browser to reload automatically (no manual F5)
- [ ] The injected script is NOT present in the `public/` output files (only injected at serve time)
- [ ] SSE script does not block page load (added at end of `<body>` or as non-blocking script)

**Verification:** Manual: open browser, edit a `.md` file, observe browser auto-reload

**Dependencies:** Task 5.1

**Files:** `cmd/gokesh/main.go` or new `cmd/gokesh/serve.go`

**Size:** M

---

### Final Checkpoint

- [ ] `go test ./...` clean
- [ ] `gokesh init` → `gokesh serve` workflow works in a blank directory in under 2 minutes
- [ ] `gokesh build` on a 10-page site completes in under 500 ms
- [ ] `gokesh --version` returns correct version
- [ ] No page renders with empty `<title>` without a warning
- [ ] `public/feed.xml` exists after `gokesh build` (when `base_url` is set)
- [ ] Browser auto-reloads on file change during `gokesh serve`
- [ ] README is accurate; no references to `.env`

---

## Risks and mitigations

| Risk | Impact | Mitigation |
|---|---|---|
| Two-pass build (Task 1.4) breaks golden tests | High | Update golden tests immediately; keep `BuildAll` signature backward-compatible |
| Live reload SSE injection corrupts non-HTML responses | Medium | Only inject into responses with `Content-Type: text/html` |
| Incremental build (Task 2.3) misses template or style changes | High | If any template/style file changes, invalidate all pages (not just changed .md) |
| Frontmatter colon fix (Task 0.1) changes existing parse behaviour | Medium | Add regression test for current `key: value` before changing; run full golden suite |
| `pageData` struct change breaks user-authored templates | Low | Only add fields; never remove or rename existing ones |

---

## Parallelization opportunities

Tasks 0.1–0.8 are all independent — a single session can knock them out sequentially in one pass.

After Checkpoint 0, these tracks are independent and can run in parallel:
- **Track A:** Phase 1 (frontmatter + content model) → Phase 4 (templates + RSS)
- **Track B:** Phase 2 (build pipeline) + Phase 3 (CLI)
- **Track C:** Phase 5 (serve + live reload) — can start after Checkpoint 0

Track A must complete before Phase 4. Track B and Track C are independent of Track A.
