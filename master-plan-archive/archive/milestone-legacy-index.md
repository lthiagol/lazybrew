# Legacy Milestone Index (M1–M17)

> **Purpose:** Map old-format milestones to current execution plan.  
> **Rule:** Do not mark legacy milestones ✅ unless DoD is verified **and** open items are routed below.  
> **New work:** Use [templates/milestone.md](templates/milestone.md) (M18+).

---

## Format status

| Range | Template | Action |
|---|---|---|
| M1–M17 | Legacy | Historical reference + **Active Work Routing** section |
| M18–M22 | Current | Execute from step index |

---

## Pending / partial — where to execute

| Milestone | Plan status | Open items | Execute in (do not use legacy steps) |
|---|---|---|---|
| **M17** | 🔜 Planned | Entire milestone | **M17** (refined file) after M19–M22 |

> **Note (2026-06-25):** All other legacy partials (M2, M4, M6, M7, M8, M11, M12, M14, M15, M16) have been closed — see "Complete — routed" below. Their routing targets (M18, M19, M20, M21) are all done. Final smoke verification for M7/M8/M11 items is consolidated under M24.13.

---

## Complete — routed

Previous partials whose remaining work was executed in M18–M21 (all `[X]` in `status.md`). Re-verify only during [review-template.md](../review-template.md) audits or the consolidated M24.13 smoke.

| Milestone | Was partial | Closed via | Notes |
|---|---|---|---|
| **M2** | Small terminal / accordion collapse | M20.7 (warning) + M17 | Final smoke in M24.13 |
| **M4** | Info tab, empty states | M20.2, M20.5 | — |
| **M6** | TaskManager, batch upgrade, reinstall confirm | M19, M20.3 | Batch upgrade re-verified in M24.9 |
| **M7** | Untap/trust/repair UX | M20.11 smoke | Smoke consolidated under M24.13 |
| **M8** | Cleanup/doctor/leaves keys | M20.11 smoke | Smoke consolidated under M24.13 |
| **M11** | Services/brewfile/vulns wiring | M20.11 smoke | Smoke consolidated under M24.13 |
| **M12** | E2E, integration, gui tests | M21 (all tiers) | 8 teatest flows + integration suite |
| **M14** | DoD verification | M18.4 audit | — |
| **M15** | TypedCache panic, Reader splits | M19.0, M18.4 | TypedCache moved to M19.0 |
| **M16** | Coverage targets, E2E | M21 | Coverage floors enforced via `make cover-check` |

---

## Complete — no routed work

## Complete — no routed work

| Milestone | Notes |
|---|---|
| M1 | Foundation — legacy format OK |
| M3 | Brew data layer |
| M5 | Modals & search |
| M9 | Polish & release (config/theme) |
| M10 | GUI architecture decomposition |
| M13 | Critical bug fixes |

Re-verify only during [review-template.md](review-template.md) audits.

---

## Refinement checklist (M18.4)

When touching a legacy milestone file:

- [ ] Header **Status** matches this index and [status.md](../status.md)
- [ ] **Active Work Routing** section present (if partial)
- [ ] **Remaining** line accurate or removed if fully routed
- [ ] Do not duplicate steps already in M18–M22

**M17 exception:** Full refinement to current template (pending execution milestone).

---

## Version History

| Date | Change |
|---|---|
| 2026-06-13 | Initial index; M17 refinement; routing blocks on M2,M4,M6,M7,M8,M11,M12,M14,M15,M16 |
| 2026-06-25 | Closed M2,M4,M6,M7,M8,M11,M12,M14,M15,M16 — routing targets M18–M21 all done; moved to "Complete — routed"; smoke items consolidated under M24.13 |
