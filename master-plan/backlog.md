# Lazybrew — Deferred Backlog

> Items identified during reviews but **explicitly out of scope** for M18–M22.  
> Prevents scope creep during execution.

| ID | Item | Source | Suggested when |
|---|---|---|---|
| B-01 | Split `internal/gui/controllers/` per panel | Architecture review | Post v0.2.0 refactor sprint |
| B-02 | Lazy panel loading (fetch active panel first) | Performance review | After M22 |
| B-03 | `tabContent` LRU / max entries | Architecture review | If memory issues observed |
| B-04 | Search info preview in main panel | M17.11 | M17 |
| B-05 | Adopt `testify` assertions | M1 plan | Optional; stdlib sufficient |
| B-06 | M7 remainder verification (untap UX, trust granularity) | status.md | During M20 smoke or dedicated pass |
| B-07 | Runner SIGKILL after 5s on cancel | M6 plan | Hardening after TaskManager stable |
| B-08 | CHANGELOG.md automation | M22 | First public release |
| B-09 | Homebrew formula for lazybrew | User install | Post v1.0 |
| B-10 | `ShowIcons` in sidebar | M18.9 deferred | M17 or later |

**Rule:** New findings during M18–M22 execution → add row here unless severity is Critical (then insert into current milestone).
