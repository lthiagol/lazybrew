# Lazybrew — Deferred Backlog

> Items identified during reviews but **explicitly out of scope** for M18–M22.  
> Prevents scope creep during execution.
>
> **Last cleaned:** 2026-06-13 — resolved B-03, B-04, B-05, B-06, B-08, B-10 into milestones or closed.

---

## Active Backlog (post-M22)

| ID | Item | Source | Suggested when | Priority |
|----|------|--------|---------------|----------|
| B-01 | Split `internal/gui/controllers/` per panel | Architecture review | Sprint after v0.2.0 tag | Medium |
| B-02 | Lazy panel loading (fetch active panel first) | Performance review | Performance pass after M22 | Low |
| B-07 | Runner SIGKILL after 5s on cancel | M6 plan | After M19 TaskManager stable + release | Medium |
| B-09 | Homebrew formula for lazybrew | User install | Post v1.0 | Low |
| B-11 | TypedCache serialization for config hot-reload | Planning | If config reload is added | Low |
| B-12 | Config migration path for renamed fields | Planning | If `ShowIcons` semantics change | Low |

---

## Resolved / Closed

| ID | Item | Resolution |
|----|------|-----------|
| B-03 | `tabContent` LRU / max entries | Conditional — no memory issues observed. Close. |
| B-04 | Search info preview in main panel | Covered by **M17.11** — tracked in milestone. |
| B-05 | Adopt `testify` assertions | Decided: stdlib sufficient. Won't fix. |
| B-06 | M7 remainder verification (untap UX, trust granularity) | Routed to **M20.11** smoke checklist. |
| B-08 | CHANGELOG.md automation | Covered by **M22.4** release checklist. |
| B-10 | `ShowIcons` in sidebar | Covered by **M17** — tracked in M18.9 ADR as deferred. |

---

**Rule:** New findings during M18–M22 execution → add row at bottom of Active Backlog unless severity is Critical (then insert into current milestone).
