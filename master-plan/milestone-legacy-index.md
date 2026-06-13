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
| **M2** | ⚠️ Partial | Small terminal (full lazygit collapse) | **M20.7** (warning); accordion collapse → **M17** |
| **M4** | ⚠️ Partial | Info tab, empty states | **M20.2**, **M20.5** |
| **M6** | ⚠️ Partial | TaskManager, batch upgrade, reinstall confirm gaps | **M19**, **M20.3** |
| **M7** | ⚠️ Partial | Verify untap/trust/repair UX | **M20.11** smoke + backlog **B-06** |
| **M8** | ⚠️ Partial | Verify cleanup/doctor/leaves keys | **M20.11** smoke |
| **M11** | ⚠️ Partial | Verify services/brewfile/vulns wiring | **M20.11** smoke |
| **M12** | ⚠️ Partial | E2E, integration, gui tests | **M21** (all tiers) |
| **M14** | ⚠️ Mostly done | Verify all DoD | **M18.4** audit |
| **M15** | ⚠️ Mostly done | TypedCache panic (→ M19.0), Reader splits verify | **M19.0**, **M18.4** |
| **M16** | ⚠️ Partial | Coverage targets, E2E | **M21** |
| **M17** | 🔜 Planned | Entire milestone | **M17** (refined file) after M19–M22 |

---

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

- [ ] Header **Status** matches this index and [status.md](status.md)
- [ ] **Active Work Routing** section present (if partial)
- [ ] **Remaining** line accurate or removed if fully routed
- [ ] Do not duplicate steps already in M18–M22

**M17 exception:** Full refinement to current template (pending execution milestone).

---

## Version History

| Date | Change |
|---|---|
| 2026-06-13 | Initial index; M17 refinement; routing blocks on M2,M4,M6,M7,M8,M11,M12,M14,M15,M16 |
