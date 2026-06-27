# Lazybrew — Release Checklist

> **When to use:** Before tagging a new version (v0.2.0+).

---

## Pre-release Verification

- [x] All M18–M22 DoD for target version
- [x] `go test -race ./...` passes
- [ ] `make cover-check` passes — ⚠️ 2 pre-existing floor failures (brew 61.7%<62%, presentation 90.6%<91%); needs decision before release
- [x] M24 code steps 24.1–24.12 complete and tested
- [ ] Manual smoke (`smoke-checklist.md`) signed off — M24.13, requires human
- [ ] Integration workflow green (manual dispatch) — requires macOS runner

---

## Release Actions

- [ ] Update README version/status — requires human
- [ ] Update CHANGELOG or release notes — requires human
- [ ] Tag `v0.2.0` (or next version) — requires human
- [x] `goreleaser release --snapshot --clean` succeeds
- [ ] `goreleaser release` (publish) — requires `GITHUB_TOKEN`
- [ ] Verify GitHub release artifacts

---

## Post-release

- [ ] Close milestones in `status.md`
- [ ] Archive planning review docs if superseded

---

## Branch Protection (GitHub Settings)

- [ ] Require CI pass before merge
- [ ] Require 1 review for human contributors
- [ ] Disallow force-push to main

## Agent Notes (2026-06-25)

**Automated verification complete:**
- `go test -race ./...` — all pass (including M24 tests: operation panel, batch upgrade, confirm modal, refresh loading, deps fetch)
- `go vet ./...` — clean
- `make lint` — clean
- `make cover-check` — ⚠️ 2 floors failing (pre-existing, not caused by M24): `internal/brew` 61.7% < 62%, `internal/gui/presentation` 90.6% < 91%. Both were already below floor at `fbd1543` (before M24 work); M24 actually improved brew (61.2%→61.7%). **Needs decision:** add coverage or recalibrate floors (decision log entry required per AGENTS.md).
- `goreleaser release --snapshot --clean` — succeeds, produces 4 tarballs (Linux/Darwin × amd64/arm64) + checksums
- M17.3 (update summary toast), M21.2a (install teatest), M21.2b (uninstall teatest) — all implemented and tested
- M24.1–M24.12 — all implemented and tested; all 14 Test Plan rows green
- M24.2 — confirmed wired (CommandLog callbacks in `app.go`); no code change needed
- CI is green on main (5/5 recent runs passing)
- `.goreleaser.yml` deprecated options fixed (`format` → `formats`, `snapshot.name_template` → `version_template`)

**Requires human:**
- Manual smoke M24.13 (`make run` + verify all `smoke-checklist.md` items)
- Integration workflow (needs macOS runner with Homebrew)
- Update README/CHANGELOG
- Tag `v0.2.0` and run `goreleaser release` with `GITHUB_TOKEN`
