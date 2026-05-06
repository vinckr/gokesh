# Gokesh Release Roadmap: 0.1 → 1.0

> Updated 2026-05-06. All 21 implementation tasks from the previous plan are **done** — code is feature-complete per spec-final.md. What remains is hardening, documentation, and release mechanics.

---

## Current state summary

| Area | Status |
|---|---|
| All 21 BLOCKER features implemented | ✅ Done |
| Tests pass (`go test ./...`) | ✅ Pass |
| Golden tests pass | ✅ Pass |
| Uncommitted changes | ⚠️ 2 modified + 1 untracked |
| README documents new commands | ❌ Outdated |
| CLI smoke/integration tests | ❌ Missing |

Uncommitted: `internal/build/build.go` (LoadConfig defaults fix), `internal/build/build_test.go` (230 new tests), `cmd/gokesh/new_test.go` (untracked).

---

## v0.1 — Minimal releasable (ready now)

The only things blocking a v0.1.0 tag are housekeeping. Code is shippable.

### Checklist

- [ ] **Commit pending changes** — `build.go` fix + `build_test.go` + `new_test.go`
- [ ] **Update README** — document all new commands and frontmatter fields (see below)
- [ ] **Tag and release** — `make release VERSION=v0.1.0`

### README gaps (must fix before 0.1)

**Commands table is missing:**
- `gokesh serve` — watch + live reload in one process (the primary dev command)
- `gokesh new <name>` — create `markdown/<name>.md` with pre-filled frontmatter
- `gokesh clean` — delete the output directory
- `--version` / `-v` — print version and exit
- `--help` / `-h` — print usage and exit

**Frontmatter table is missing new fields:**
- `date: 2026-01-15` — post date, exposed as `.Pagematter.Date`
- `draft: true` — exclude page from all builds
- `description: "..."` — per-page meta description
- `tags: [go, web]` — taxonomy tags, exposed as `.Pagematter.Tags`
- `slug: "custom-url"` — override the output URL path

**Config table is missing:**
- `output_dir` — default `public`
- `markdown_dir` — default `markdown`

**Template variables table is missing:**
- `.Pages` — slice of all non-draft pages (`Title`, `URL`, `Date`, `Tags`)
- `.Pagematter.Date`, `.Pagematter.Tags`, `.Pagematter.Description`, `.Pagematter.Slug`, `.Pagematter.Draft`

**Template functions section needs adding:**
- `dateFormat .Pagematter.Date "Jan 2, 2006"` — format a date
- `sortBy .Pages "date"` — sort pages by date or title
- `filterByTag .Pages "go"` — filter pages by tag

**Remove the stale "1.0 Roadmap" section** — those features are already implemented.

---

## v0.x — Iterative hardening (before 1.0)

### v0.2 — Integration tests

The only item spec-final.md calls a "BLOCKER for v1.0" that is not yet done:

| Task | File(s) | Size |
|---|---|---|
| CLI smoke tests via `os/exec` — test `init`, `build`, `serve` in a temp dir | `cmd/gokesh/smoke_test.go` | M |

Test cases needed:
1. `gokesh --version` exits 0, output contains version string
2. `gokesh --help` exits 0, output contains "Commands:"
3. `gokesh init` in empty temp dir creates `gokesh.toml`, `templates/`, `styles/`
4. `gokesh build` after init produces `public/index.html`
5. `gokesh new my-post` creates `markdown/my-post.md` with frontmatter
6. `gokesh clean` deletes `public/`

### v0.3 — Bug fixes from real use (as discovered)

Placeholder for issues found after initial release. Expected areas:
- Live reload edge cases (large sites, binary files)
- Frontmatter parsing edge cases
- Windows path separator issues (if any)

---

## v1.0 — Feature-complete freeze

v1.0 is the last release before API stability. After v1.0, **no more breaking changes**.

### v1.0 success criteria (from spec-final.md §16)

- [ ] All CRITICAL and HIGH flaws fixed ← **done**
- [ ] All BLOCKER features implemented ← **done**
- [ ] `gokesh init` + `gokesh serve` produces working site in under 2 min in blank dir ← **needs verification**
- [ ] `gokesh build` on 50-page site completes in under 1 s ← **needs verification**
- [ ] CLI smoke tests pass in CI ← **v0.2 work**
- [ ] No page can silently render with empty `<title>` ← **done (slog warn implemented)**
- [ ] `gokesh --version` works ← **done**
- [ ] README is accurate ← **v0.1 work**

### v1.0 = v0.2 + verified perf + no regressions

Once v0.2 ships with smoke tests and there are no open critical bugs, tag v1.0.

---

## v1.x backlog (never for 1.0)

These were explicitly called "post-ship" in the original plan. They require design decisions and may involve breaking changes to templates:

| ID | Feature |
|---|---|
| V1X-1 | Section index auto-generation |
| V1X-2 | Pagination |
| V1X-3 | Nested lists in Markdown |
| V1X-4 | Asset fingerprinting / cache busting |
| V1X-5 | Server-side syntax highlighting |
| V1X-6 | Multiple config environments |
| V1X-7 | Structured CLI with `flag` package |

---

## Architecture decisions (locked at 1.0)

These cannot change after 1.0 without a major version bump:

- Zero external Go module dependencies
- Clean URL output: `foo.md` → `foo/index.html` → `/foo/`
- Config file: `gokesh.toml` (custom TOML parser, no third-party)
- Template data shape: `pageData` struct with `Pagematter`, `Pages`, `Body`, `SiteTitle`, etc.
- Template function names: `dateFormat`, `sortBy`, `filterByTag`, `jsonify`, `items`
- CLI command names: `init`, `build`, `new`, `clean`, `serve`, `watch`, `dev`
