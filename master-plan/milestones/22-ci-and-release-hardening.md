# Milestone 22 — CI & Release Hardening

> **Status:** 🔜 Planned  
> **Size estimate:** M (2–3 days, split across execution)  
> **Depends on:** M18.1 (LICENSE), M19.5 (minimal CI), M21.3 (integration workflow)  
> **Enables:** Public v0.2.0 tag  
> **Parallel track:** E (Ops) — 22.1 early; rest after M21

See [planning-challenge-2026-06-13.md](../planning-challenge-2026-06-13.md) — minimal CI lands early.

---

## Goal

Automate verification on every push, provide optional brew integration runs, and make releases reproducible.

---

## Out of Scope

- Homebrew formula tap for lazybrew (future)
- Signed releases / notarization (future macOS)
- Docker images

---

## Step Index

| Step | Title | Size | When | Depends |
|---|---|---|---|---|
| 22.1 | Minimal CI (Ubuntu) | S | After M19.5 | M19 tests |
| 22.2 | Integration workflow | M | After M21.3 | 21.3 |
| 22.3 | Goreleaser validation | S | After M18.1 | LICENSE |
| 22.4 | Release checklist | S | After 22.3 | — |
| 22.5 | Dependabot (optional) | S | Anytime | — |
| 22.6 | Branch protection recommendations | S | After 22.1 | — |

---

## Steps

### 22.1 — Minimal CI (Ubuntu)

**Size:** S · **Start after:** M19.5

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
- Ubuntu job won't run integration tests (no brew required)
- `make test` includes `-race` per existing Makefile

**Acceptance criteria:**
- [ ] Green run on push to current branch
- [ ] Matches local `make test` behavior

**Tests:** CI itself

---

### 22.2 — Integration Workflow

**Size:** M · **Depends on:** M21.3

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
        with: { go-version: '1.24.x' }
      - run: brew --version
      - run: make test-integration
```

**Optional:** `ubuntu-latest` + Linuxbrew install step if Linux support critical.

**Acceptance criteria:**
- [ ] Manual dispatch runs integration suite
- [ ] Failures reported with brew version in log

---

### 22.3 — Goreleaser Validation

**Size:** S · **Depends on:** M18.1

**Steps:**
1. `goreleaser release --snapshot --clean`
2. Verify tarball contains: binary, LICENSE, checksums
3. Fix `.goreleaser.yml` if repo owner/name wrong for actual GitHub remote

**Acceptance criteria:**
- [ ] Snapshot build succeeds locally
- [ ] `main.version` ldflag matches tag pattern

**Optional CI job:** on tag push `v*` — run goreleaser (needs `GITHUB_TOKEN` secret).

---

### 22.4 — Release Checklist

**Size:** S

**File:** `master-plan/release-checklist.md`

**Sections:**
1. **Pre-release verification**
   - [ ] All M18–M21 DoD for target version
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

---

### 22.5 — Dependabot (Optional)

**Size:** S

**File:** `.github/dependabot.yml`

```yaml
version: 2
updates:
  - package-ecosystem: gomod
    directory: /
    schedule: { interval: weekly }
```

**Acceptance criteria:**
- [ ] PRs created for charmbracelet updates (or document why skipped)

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

- [ ] 22.1–22.4 complete (22.5 optional)
- [ ] CI green on default branch
- [ ] Integration workflow runnable
- [ ] Goreleaser snapshot succeeds
- [ ] release-checklist.md exists
- [ ] status.md updated

---

## Post-Milestone Gate

- [ ] Tag v0.2.0 permitted
- [ ] M17 may begin on release branch or main

---

## Rollback

If CI flaky:
1. Remove `-race` from CI only temporarily (document debt)
2. Keep race in local `make test`
3. Fix root cause before release tag

Do not disable CI entirely.
