# Project Name — Master Plan Status

> **Project:** _short name — one-line description_  
> **Stack:** _primary technologies_  
> **Platforms:** _target platforms_  
> **Created:** YYYY-MM-DD  
> **Last Updated:** YYYY-MM-DD  
> **Version target:** _e.g. v0.2.0, Homebrew 6.0.0+_  
> **Planning status:** _Planning | Ready for execution | In execution | Release candidate_

---

## Planning Documents

| Document | Purpose |
|---|---|
| [status.md](status.md) | This file — portfolio status (single source of truth) |
| [templates/](templates/) | Milestone and status templates + conventions |
| [review-template.md](review-template.md) | Project-agnostic audit checklist |
| [backlog.md](backlog.md) | Deferred scope (B-XX) |
| [architecture-review-YYYY-MM-DD.md](architecture-review-YYYY-MM-DD.md) | Findings report _(create on review)_ |
| [planning-challenge-YYYY-MM-DD.md](planning-challenge-YYYY-MM-DD.md) | Challenged decisions _(optional)_ |
| [coverage-audit.md](coverage-audit.md) | _Domain-specific audit — optional_ |
| [smoke-checklist.md](smoke-checklist.md) | Manual pre-release verification _(optional)_ |

---

## Overall Progress

<!-- Legend must appear directly under the block -->

```
[X]  Milestone 1   — Short name
[~]  Milestone 2   — Short name (gap → M__._)
[ ]  Milestone 3   — Short name
```

**Legend:** `[X]` complete · `[~]` partial · `[ ]` not started · 🚫 blocked

| Field | Value |
|---|---|
| **Current phase** | _e.g. M19 TaskManager — step 19.3_ |
| **Execution entry point** | _Next step(s) an agent should pick up_ |
| **Blockers** | _None | describe_ |
| **Target release** | _vX.Y.Z — date or “when M__ gates pass”_ |

---

## Parallel Execution Tracks

<!-- Delete section if single-threaded project -->

| Track | Milestones | Focus | Starts when | Blocks |
|---|---|---|---|---|
| **A —** | M__ | | | |
| **B —** | M__ | | | |

### Critical path

```
M__._ → M__._ → M__._ → release
```

Steps on the critical path should not slip without updating **Blockers** and dependent milestones.

---

## Milestone Index

| # | Milestone | Status | Steps | Size | Gate / Remaining |
|---|---|---|---|---|---|
| 1 | [Title](milestones/01-slug.md) | ✅ | 1.1–1.N | S/M/L | _gate or “—”_ |
| 2 | [Title](milestones/02-slug.md) | ⚠️ | 2.1–2.N | | _open gap summary_ |

**Status column values:** ✅ Complete · ⚠️ Partial · 🔜 Planned · 🚫 Blocked

**Rules:**

- One row per milestone; link to `milestones/NN-slug.md`
- **Remaining** column required for ⚠️ Partial
- **Gate** = what this milestone unlocks (short phrase)

---

## Active Milestone Detail

<!-- Optional: expand only milestones in flight to reduce noise -->

### M__ — Title (⚠️ In Progress)

| Step | Title | Status | Owner/notes |
|---|---|---|---|
| __.1 | | ☐ / ☑ | |
| __.2 | | ☐ | |

---

## Challenged Decisions

<!-- Summary only — full text in planning-challenge doc -->

| Decision | Outcome | Doc |
|---|---|---|
| | | [planning-challenge-…](planning-challenge-YYYY-MM-DD.md) |

---

## Metrics Baseline

<!-- Update on reviews; used to detect drift -->

| Metric | Value | Measured |
|---|---|---|
| Source lines | | YYYY-MM-DD |
| Test count | | |
| Coverage (_package_) | | |
| Integration tests | _N files_ | |
| E2E tests | _N_ | |
| CI workflows | _N_ | |

**Measure commands:** _(project-specific)_

```bash
# Example:
go test ./... -list . | grep -c '^Test'
go test ./... -cover
```

---

## Testing Strategy

| Tier | Framework / tag | Covers | When run |
|---|---|---|---|
| Unit | | | Every commit |
| Integration | `//go:build integration` | | `make test-integration` |
| E2E | teatest / playwright / etc. | | CI job name |
| Manual | smoke-checklist.md | | Pre-release |

---

## Architecture Reference

| Doc | Location |
|---|---|
| Design | `DESIGN.md` _(create if missing)_ |
| Agent/contributor guide | `AGENTS.md` |
| Domain audit | _link_ |

---

## Decision Log

Newest first.

| Date | Decision | Context |
|---|---|---|
| YYYY-MM-DD | | |

---

## Adoption / Execution Readiness

<!-- Check before telling contributors “ready to execute” -->

- [ ] Active milestones have Step Index + acceptance criteria + per-step tests
- [ ] No ⚠️/✅ mismatch between milestone headers and this file
- [ ] Critical path and entry point are current
- [ ] Backlog used for deferred scope (no hidden gaps)
- [ ] Review template available for next audit

**Ready for execution:** ☐ Yes ☐ No — _reason_

---

## Maintenance

**Update this file when:**

1. Starting or finishing a milestone step (at least **Current phase**)
2. Changing milestone Status emoji/header
3. Adding/removing milestones
4. Completing a review (metrics + decision log)
5. Changing critical path or blockers

**Do not update only the milestone file** without syncing this file.

---

## Version History

| Date | Change |
|---|---|
| YYYY-MM-DD | Created from [templates/status.md](templates/status.md) |
