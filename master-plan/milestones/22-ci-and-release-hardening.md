# Milestone 22 — CI & Release Hardening

> **Status:** ⚠️ Partial (~60% done)  
> **Size estimate:** S remaining (22.1b, 22.3a/b, 22.4, 22.6)  
> **Depends on:** M18.1 (done), M19.5 (done), M21.3 (done)  
> **Enables:** Public v0.2.0 tag  
> **Parallel track:** E (Ops) — in progress

See [planning-challenge-2026-06-13.md](../archive/planning-challenge-2026-06-13.md) — minimal CI lands early.

---

## Goal

Automate verification on every push, provide optional brew integration runs, and make releases reproducible.

---

## Reality Check (2026-06-14)

Prerequisites are done:

- `LICENSE` exists (M18.1).
- `Makefile` has `test`, `test-integration`, `lint`, `cover-check` targets.
- `scripts/check-coverage.sh` exists.
- `.goreleaser.yml` exists and references `LICENSE`.
- `.github/workflows/ci.yml` exists (M22.1a done).
- `.github/dependabot.yml` exists (M22.5 done).
- Tests pass with `-race` locally.

Remaining:

- Verify CI green on push/PR (M22.1b).
- Validate goreleaser snapshot (M22.3a/b).
- Document branch protection recommendations (M22.6).
- Sign release checklist (M22.4).

This milestone is the **primary remaining engineering blocker** for v0.2.0.

---

## Out of Scope

- Homebrew formula tap for lazybrew (future)
- Signed releases / notarization (future macOS)
- Docker images

---

## Step Index

| Step | Title | Size | Status | Depends |
|---|---|---|---|---|
| 22.1a | Minimal CI workflow file (Ubuntu) | S | Done | M19 tests done |
| 22.1b | Verify CI green on push/PR | S | **Remaining** | CI file merged |
| 22.1c | Add CI badge to README | S | Done | Green CI run |
| 22.2 | Integration workflow | M | Done | M21.3 |
| 22.3a | Goreleaser snapshot local validation | S | **Remaining** | LICENSE |
| 22.3b | Fix goreleaser config if needed | S | **Remaining** | Snapshot result |
| 22.4 | Release checklist sign-off | S | **Remaining** | — |
| 22.5 | Dependabot | S | Done | — |
| 22.6 | Branch protection recommendations | S | **Remaining** | — |

---

## Steps

### 22.1a — Minimal CI Workflow File (Ubuntu)

**Size:** S · **Status:** Done

**File:** `.github/workflows/ci.yml`

```yaml
name: CI
on:
  push:
    branches: [main, first-version]
  pull_request:

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.24.x'
          cache: true
      - run: go mod verify
      - run: make lint
      - run: make test
      - run: go build -o /dev/null ./cmd/lazybrew
```

**Notes:**
- Ubuntu job won't run integration tests (no brew required).
- `make test` includes `-race` per existing Makefile.

**Acceptance criteria:**
- [x] File committed to `.github/workflows/ci.yml`
- [x] Workflow syntax valid (`gh workflow list` or GitHub UI shows it)

**Tests:** CI itself after merge

---

### 22.1b — Verify CI Green on Push/PR

**Size:** S · **Depends on:** 22.1a

**Implementation:**
1. Push the workflow file to `main` (or open a PR).
2. Confirm the `CI` workflow runs and passes.
3. If flaky, fix root cause; do not disable CI.

**Acceptance criteria:**
- [ ] Green run on push to current branch
- [ ] Matches local `make test` behavior

**Tests:** CI itself

---

### 22.1c — Add CI Badge to README

**Size:** S · **Status:** Done

**Implementation:**
1. Add Markdown badge near the top of `README.md`:
   ```markdown
   ![CI](https://github.com/lthiagol/lazybrew/actions/workflows/ci.yml/badge.svg)
   ```
2. Verify badge URL matches `.goreleaser.yml` owner/name.

**Acceptance criteria:**
- [x] Badge added to `README.md`
- [ ] Badge renders green on GitHub repo page (requires 22.1b green run)

---

### 22.2 — Integration Workflow

**Size:** M · **Status:** Done

**File:** `.github/workflows/integration.yml`

```yaml
name: Integration
on:
  workflow_dispatch:
  schedule:
    - cron: '0 6 * * 1'  # weekly Monday 06:00 UTC

jobs:
  brew:
    runs-on: macos-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.24.x'
          cache: true
      - run: brew --version
      - run: make test-integration
```

**Acceptance criteria:**
- [x] File committed to `.github/workflows/integration.yml`
- [ ] Manual dispatch runs integration suite green
- [ ] Failures reported with brew version in log

---

### 22.3a — Goreleaser Snapshot Local Validation

**Size:** S · **Depends on:** 22.1b, M18.1

**Steps:**
1. Install/run `goreleaser release --snapshot --clean`
2. Verify tarball contains: binary, LICENSE, checksums
3. Record any errors

**Acceptance criteria:**
- [ ] Snapshot build succeeds locally
- [ ] `main.version` ldflag matches tag pattern

---

### 22.3b — Fix Goreleaser Config if Needed

**Size:** S · **Depends on:** 22.3a

**Steps:**
1. If 22.3a failed, fix `.goreleaser.yml` (owner/name, paths, ldflags).
2. Re-run snapshot until clean.
3. Add optional tag-push CI job for releases (needs `GITHUB_TOKEN` secret).

**Acceptance criteria:**
- [ ] Snapshot build succeeds after any fixes
- [ ] Config matches actual GitHub remote

---

### 22.4 — Release Checklist Sign-off

**Size:** S · **Depends on:** 22.3b

**File:** `master-plan/release-checklist.md` (exists; verify completeness)

**Sections:**
1. **Pre-release verification**
   - [ ] M18.8 AGENTS.md done
   - [ ] M17.3 update summary toast done
   - [ ] M21.2 ≥8 teatest flows done
   - [ ] `go test -race ./...`
   - [ ] `make cover-check`
   - [ ] Manual smoke (`smoke-checklist.md`)
   - [ ] Integration workflow green (manual dispatch)

2. **Release actions**
   - [ ] Update README version/status
   - [ ] Update CHANGELOG or release notes
   - [ ] Tag `v0.2.0`
   - [ ] goreleaser publish
   - [ ] Verify GitHub release artifacts

3. **Post-release**
   - [ ] Close milestones in status.md
   - [ ] Archive planning review docs if superseded

**Acceptance criteria:**
- [ ] Checklist usable without reading other docs
- [ ] All checkboxes checked before tag push

---

### 22.5 — Dependabot

**Size:** S · **Status:** Done

**File:** `.github/dependabot.yml`

```yaml
version: 2
updates:
  - package-ecosystem: gomod
    directory: /
    schedule: { interval: weekly }
```

**Acceptance criteria:**
- [x] File committed to `.github/dependabot.yml`
- [x] Dependabot enabled for gomod (PRs will be created by GitHub)

---

### 22.6 — Branch Protection Recommendations

**Size:** S · **Documentation only**

Add to `AGENTS.md` or `release-checklist.md`:

- Require CI pass before merge
- Require 1 review for human contributors
- Disallow force-push to main

Not enforced by agent — human GitHub settings.

---

## Definition of Done

- [x] 22.1a CI workflow file committed
- [ ] 22.1b CI green on push/PR
- [x] 22.1c CI badge added to README
- [x] 22.2 integration workflow file committed
- [ ] 22.3a goreleaser snapshot succeeds locally
- [ ] 22.3b goreleaser config fixed if needed
- [ ] 22.4 release checklist signed off
- [x] 22.5 dependabot configured
- [ ] 22.6 branch protection recommendations documented
- [ ] CI green on default branch
- [ ] Integration workflow runnable
- [ ] Goreleaser snapshot succeeds
- [ ] release-checklist.md exists and verified
- [x] status.md updated

---

## Post-Milestone Gate

- [ ] Tag v0.2.0 permitted
- [ ] All release readiness items complete
