# Lazybrew — Release Checklist

> **When to use:** Before tagging a new version (v0.2.0+).

---

## Pre-release Verification

- [ ] All M18–M22 DoD for target version
- [ ] `go test -race ./...` passes
- [ ] `make cover-check` exists and passes
- [ ] Manual smoke (`smoke-checklist.md`) signed off
- [ ] Integration workflow green (manual dispatch)

---

## Release Actions

- [ ] Update README version/status
- [ ] Update CHANGELOG or release notes
- [ ] Tag `v0.2.0` (or next version)
- [ ] `goreleaser release --snapshot --clean` succeeds
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
