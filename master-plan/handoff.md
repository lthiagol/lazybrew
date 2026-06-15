# Lazybrew — Session Handoff

> **Date:** 2026-06-15  
> **Branch:** `wip` (tracking `origin/wip`)  
> **Status:** Testing the TUI on macOS — guided smoke test in progress.

---

## Where we are

Code is feature-complete for v0.2.0. History is clean (4 commits on top of `8015ad2`). Panels load, search works, sidebar scrolls, JSON parsing is compatible with Homebrew 6.0+.

We're currently doing manual smoke on a real macOS machine. The smoke checklist is at `master-plan/smoke-checklist.md`.

---

## Next steps

1. **Finish smoke test** — run through `smoke-checklist.md`, report any bugs
2. **M22.4 release sign-off** — once smoke passes:
   - `git tag -a v0.2.0 -m "v0.2.0" && git push origin v0.2.0`
   - `goreleaser release` (needs `GITHUB_TOKEN`)
3. **Post-release** — pick from backlog (`master-plan/backlog.md`):
   - **B-01** — Split `internal/gui/` per panel
   - **B-02** — Lazy panel loading
   - **B-07** — Runner SIGKILL on cancel
   - **B-09** — **Groom:** we already have `https://github.com/lthiagol/homebrew-tap` with a lazybrew formula. Installable via `brew install lthiagol/tap/lazybrew`.
   - **B-13** — Command log border

---

## Verification

```bash
go test -race ./...
go vet ./...
make lint
make cover-check
```

---

## Environment

- **SSH agent:** if pushes fail, find socket: `ls /tmp/ssh-*/agent.*`, then `export SSH_AUTH_SOCK=...`
- **Default branch:** `main`
- **Goreleaser:** binary at `/tmp/goreleaser` (v2.16.0)
