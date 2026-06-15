# Lazybrew — Release Checklist

> **When to use:** Before tagging a new version (v0.2.0+).

---

## Pre-release Verification

- [x] All M18–M22 DoD for target version
- [x] `go test -race ./...` passes
- [x] `make cover-check` exists and passes
- [ ] Manual smoke (`smoke-checklist.md`) signed off — requires human
- [ ] Integration workflow green (manual dispatch) — requires macOS runner

---

## Release Actions

- [ ] Update README version/status
- [ ] Update CHANGELOG or release notes
- [ ] Tag `v0.2.0` (or next version)
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

## Agent Notes (2026-06-14)

**Automated verification complete:**
- `go test -race ./...` — all pass
- `go vet ./...` — clean
- `make lint` — clean
- `make cover-check` — all floors met
- `goreleaser release --snapshot --clean` — succeeds, produces 4 tarballs (Linux/Darwin × amd64/arm64) + checksums
- M17.3 (update summary toast), M21.2a (install teatest), M21.2b (uninstall teatest) — all implemented and tested
- CI is green on main (5/5 recent runs passing)
- `.goreleaser.yml` deprecated options fixed (`format` → `formats`, `snapshot.name_template` → `version_template`)

**Requires human:**
- Manual smoke (`make run` + verify panels)
- Integration workflow (needs macOS runner with Homebrew)
- Update README/CHANGELOG
- Tag `v0.2.0` and run `goreleaser release` with `GITHUB_TOKEN`
