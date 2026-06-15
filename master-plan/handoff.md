# Lazybrew — Session Handoff

> **Date:** 2026-06-14  
> **Branch:** `main` (origin/main is in sync)  
> **Goal:** Complete release readiness and tag `v0.2.0`.

`first-version` has been retired. All planning is up to date in `master-plan/status.md` and the milestone files. This handoff lists the executable tasks for the next session.

---

## Code tasks (implement first)

### 1. M17.3 — Update summary toast

- **Files:** `internal/gui/commands.go`, `internal/gui/gui.go`, `internal/gui/commands_test.go`
- **What:** Add `parseUpdateSummary(lines []string) string`, call it in the `UpdateCompleteMsg` handler, and show a success toast.
- **Details:** See `master-plan/milestones/17-lazygit-tui-and-auto-update.md` §17.3.
- **Acceptance:**
  - [ ] Parser unit tests pass for all listed cases.
  - [ ] Successful `brew update` shows a toast like "Updated 3 formulae" or "Already up to date".
  - [ ] Failed update keeps the existing error toast.
  - [ ] No new `program.Send` introduced.

### 2. M21.2a — Install flow teatest

- **File:** `internal/gui/flows/install_test.go`
- **What:** Use `teatest` to search for a package, press `i`, and assert the mock runner recorded an `install` command.
- **Details:** See `master-plan/milestones/21-test-strategy-v2.md` §21.2a.
- **Acceptance:**
  - [ ] Test compiles and passes with `go test -race ./internal/gui/flows/...`.
  - [ ] Test asserts the install command was recorded by the mock runner.

### 3. M21.2b — Uninstall flow teatest

- **File:** `internal/gui/flows/uninstall_test.go`
- **What:** Seed the Formulae panel with one item, press `x`, confirm the modal, and assert the mock runner recorded an `uninstall` command.
- **Details:** See `master-plan/milestones/21-test-strategy-v2.md` §21.2b.
- **Acceptance:**
  - [ ] Test compiles and passes with `go test -race ./internal/gui/flows/...`.
  - [ ] Test asserts the uninstall command was recorded by the mock runner.

---

## Verification tasks

### 4. M22.1b — Confirm CI green

- **Action:** Check the GitHub Actions tab for the latest `CI` run.
- **If green:** mark done in `master-plan/status.md` and `master-plan/milestones/22-ci-and-release-hardening.md`.
- **If red:** fix root cause (likely timeout or environment issue). Do not disable CI.

### 5. M22.3a/b — Goreleaser snapshot validation

- **Prerequisite:** Install `goreleaser` CLI (e.g. `brew install goreleaser/tap/goreleaser`).
- **Action:** Run `goreleaser release --snapshot --clean`.
- **Details:** See `master-plan/milestones/22-ci-and-release-hardening.md` §22.3a/b.
- **Acceptance:**
  - [ ] Build succeeds.
  - [ ] `dist/` contains binaries, tarballs, LICENSE, and checksums.
  - [ ] `./dist/.../lazybrew --version` works and matches the expected tag pattern.

### 6. M22.4 — Release checklist sign-off

- **Action:** Run through `master-plan/release-checklist.md` and check every box.
- **Acceptance:**
  - [ ] All pre-release verification items completed.
  - [ ] Tag `v0.2.0` pushed and goreleaser publish succeeded (requires `GITHUB_TOKEN`).

---

## Recommended execution order

```
Phase 1 — Code (parallelizable)
├── M17.3  parseUpdateSummary + toast
├── M21.2a install teatest
└── M21.2b uninstall teatest
        │
        ▼
Phase 2 — Commit/push and verify
├── Commit/push all Phase 1 work
└── M22.1b wait for GitHub Actions CI to go green
        │
        ▼
Phase 3 — Release artifact validation
├── Install goreleaser CLI
├── M22.3a run goreleaser snapshot
└── M22.3b fix config if needed
        │
        ▼
Phase 4 — Release
└── M22.4 sign release checklist and tag v0.2.0
```

---

## Environment notes for next session

- **SSH agent:** The current zellij/SSH session may have a stale `SSH_AUTH_SOCK`. If pushes fail with "Permission denied (publickey)", find the active socket with `ls /tmp/ssh-*/agent.*`, then:
  ```bash
  export SSH_AUTH_SOCK=/tmp/ssh-<id>/agent.<pid>
  ```
- **Default branch:** `main` is now the default branch on GitHub; `first-version` is deleted.
- **Verification commands:** Before any commit, run:
  ```bash
  go test -race ./...
  go vet ./...
  make lint
  make cover-check
  ```
