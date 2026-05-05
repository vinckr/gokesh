# Gokesh v1.0 — Task List

> Derived from `tasks/plan.md`. Check off tasks as they are completed.
> All tasks require `go test ./...` to pass before marking done.

---

## Phase 0: Critical bug fixes

- [x] **0.1** Fix frontmatter colon parsing — `parser/frontmatter.go` _(XS)_
- [x] **0.2** Filter `.md` files in `BuildPages` — `build/build.go` _(XS)_
- [x] **0.3** Fix `init` README overwrite — `cmd/gokesh/init.go` _(XS)_
- [x] **0.4** Fix Tailwind CDN + CSP contradiction — `cmd/gokesh/init.go` _(S)_
- [x] **0.5** Fix broken PWA webmanifest icons — `cmd/gokesh/init.go` _(S)_
- [x] **0.6** Fix watch config reload — `build/watch.go` _(S)_
- [x] **0.7** Fix `itoa()`, partial build styles, README docs — `markdown.go`, `main.go`, `README.md` _(S)_
- [x] **0.8** Remove `templates-examples/` from init — `cmd/gokesh/init.go` _(XS)_

**Checkpoint 0:** `go test ./...` clean · `init` → `build` → `dev` works end-to-end

---

## Phase 1: Frontmatter + content model

- [ ] **1.1** Extend frontmatter parser: arrays, booleans, dates — `parser/frontmatter.go` _(M)_
- [ ] **1.2** Expand `pageData` struct — `build/build.go` _(S)_
- [ ] **1.3** Draft filtering in build pipeline — `build/build.go` _(S)_
- [ ] **1.4** Two-pass build: `.Pages` global — `build/build.go`, `build_test.go` _(M)_

**Checkpoint 1:** `.Pagematter.Date`, `.Pagematter.Tags`, `.Pages` all available in templates · drafts excluded

---

## Phase 2: Build pipeline improvements

- [ ] **2.1** Watch `data/` directory — `build/watch.go` _(XS)_
- [ ] **2.2** `static/` directory copy — `build/build.go`, `main.go` _(S)_
- [ ] **2.3** Incremental builds — `build/build.go`, `build/watch.go` _(M)_
- [ ] **2.4** Configurable output + markdown dirs — `build/build.go`, `parser/config.go`, `main.go` _(S)_

**Checkpoint 2:** Watch rebuilds only changed files · static files copied · `data/` changes trigger rebuild

---

## Phase 3: CLI improvements

- [ ] **3.1** `--version` and `--help` flags — `cmd/gokesh/main.go`, `.goreleaser.yml` _(S)_
- [ ] **3.2** `gokesh clean` command — `cmd/gokesh/main.go` _(S)_
- [ ] **3.3** `gokesh new <name>` command — `cmd/gokesh/main.go` or `new.go` _(S)_

**Checkpoint 3:** `--version` works · `clean` deletes output · `new` scaffolds a markdown file

---

## Phase 4: Templates, SEO, and RSS

- [ ] **4.1** Code block language class — `parser/markdown.go` _(XS)_
- [ ] **4.2** Template functions: `dateFormat`, `sortBy`, `filterByTag` — `build/build.go` _(S)_
- [ ] **4.3** Update default templates with SEO meta tags — `templates/header.tmpl` _(S)_
- [ ] **4.4** RSS feed generation — `build/rss.go`, `build/build.go` _(M)_
- [ ] **4.5** Warn on missing `title` frontmatter — `build/build.go` _(XS)_

**Checkpoint 4:** Built pages have OG tags + description · `feed.xml` valid RSS 2.0 · code blocks emit language class

---

## Phase 5: Combined serve + live reload

- [ ] **5.1** `gokesh serve`: watch + dev server in one process — `cmd/gokesh/main.go` _(M)_
- [ ] **5.2** Live reload via SSE injection — `cmd/gokesh/main.go` or `serve.go` _(M)_

**Final checkpoint:** `init` → `serve` works in blank dir in <2 min · browser auto-reloads on file change

---

## Post-ship (v1.x backlog)

- [ ] Section index auto-generation (V1X-1)
- [ ] Pagination (V1X-2)
- [ ] Nested lists in markdown (V1X-3)
- [ ] Asset fingerprinting / cache busting (V1X-4)
- [ ] Server-side syntax highlighting (V1X-5)
- [ ] Multiple config environments (V1X-6)
- [ ] Structured CLI with `flag` package (V1X-7)
