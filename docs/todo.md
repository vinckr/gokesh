# Gokesh — Release Task List

> See `docs/plan.md` for the full roadmap and rationale.
> All 21 original implementation tasks are complete. These are the remaining release tasks.

---

## v0.1 — Ship it

- [ ] **Commit pending changes**
  - `internal/build/build.go` — LoadConfig defaults fix
  - `internal/build/build_test.go` — 230 new tests
  - `cmd/gokesh/new_test.go` — new command tests (untracked)
  - Verify: `go test ./...` passes

- [x] **Update README**
  - Add `serve`, `new`, `clean`, `--version`, `--help` to commands table
  - Add `date`, `draft`, `description`, `tags`, `slug` to frontmatter table
  - Add `output_dir`, `markdown_dir` to config table
  - Add `.Pages` and new `.Pagematter.*` fields to template variables
  - Add `dateFormat`, `sortBy`, `filterByTag` template functions section
  - Remove stale "1.0 Roadmap" section (those features are done)

- [ ] **Tag and release**
  - `make release VERSION=v0.1.0`

---

## v0.2 — Integration tests

- [x] **CLI smoke tests** (`cmd/gokesh/smoke_test.go`)
  - `--version` exits 0, output contains version string
  - `--help` exits 0, output contains "Commands:"
  - `init` in empty temp dir creates `gokesh.toml`, `templates/`, `styles/`
  - `build` after init produces `public/index.html`
  - `new my-post` creates `markdown/my-post.md` with frontmatter
  - `clean` deletes `public/`

---

## v1.0 — Feature-complete freeze

- [ ] Smoke tests pass in CI (v0.2 prerequisite)
- [ ] Manual verification: `init` + `serve` produces working site in under 2 min on blank dir
- [ ] Manual verification: `build` on 50-page site completes in under 1 s
- [ ] No open critical or high severity bugs
- [ ] Tag v1.0.0

---

## v1.x backlog (never for 1.0, decide later)

- [ ] V1X-1: Section index auto-generation
- [ ] V1X-2: Pagination
- [ ] V1X-3: Nested lists in Markdown
- [ ] V1X-4: Asset fingerprinting / cache busting
- [ ] V1X-5: Server-side syntax highlighting
- [ ] V1X-6: Multiple config environments
- [ ] V1X-7: Structured CLI with `flag` package
