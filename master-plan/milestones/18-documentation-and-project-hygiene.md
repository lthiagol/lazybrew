# Milestone 18 — Documentation & Project Hygiene

> **Status:** 🔜 Planned  
> **Size estimate:** M (2–3 days total; steps are mostly S)  
> **Depends on:** Nothing  
> **Enables:** M19 (AGENTS concurrency rules), M22 (LICENSE, README)  
> **Parallel track:** A (Docs) — runs parallel to M19 after step 18.5  
> **Gate criteria:** DESIGN skeleton + config ADR exist; status.md is single source of truth

See [planning-challenge-2026-06-13.md](../planning-challenge-2026-06-13.md) — M19 may start after **18.5 only**, not all of M18.

---

## Goal

Make documentation, legal artifacts, and master-plan tracking **truthful and executable**. Agents starting M19+ must not guess conventions or re-audit completed milestones.

---

## Challenged Assumptions

| Assumption | Challenge | Decision |
|---|---|---|
| All docs before code | Delays TaskManager | Split: skeleton at 18.5, rest parallel |
| Remove unused config fields | Breaks user YAML | Mark planned/deferred; wire in M20.8 |
| Rewrite all M1–M17 files | Huge diff | Update headers + DoD checkboxes; link to audit |

---

## Out of Scope

- Implementing config features (M20.8)
- TaskManager (M19)
- CI workflows (M22.1 can start early but not part of M18)

---

## Architecture Decisions (ADRs)

| ID | Decision | Rationale |
|---|---|---|
| D18-1 | `DESIGN.md` ≤250 lines; details stay in milestones | Canonical, not duplicate |
| D18-2 | `AGENTS.md` is operational; `DESIGN.md` is architectural | Different audiences |
| D18-3 | Milestone header status must match `status.md` | Single source of truth |
| D18-4 | Config: wire in M20, document intent in M18 | No ambiguous "choose one" |

---

## Step Index

| Step | Title | Size | Depends | Deliverable |
|---|---|---|---|---|
| 18.1 | Add LICENSE | S | — | MIT file |
| 18.2 | Fix README | S | — | Accurate README |
| 18.3 | Milestone DoD audit worksheet | S | — | Audit table in status or review doc |
| 18.4 | Reconcile M1–M17 headers | M | 18.3 | Headers match reality |
| 18.5 | DESIGN.md skeleton + concurrency ADR | M | — | DESIGN.md v0.1 |
| 18.6 | DESIGN.md brew layer | S | 18.5 | Brew section complete |
| 18.7 | DESIGN.md GUI layer | S | 18.5 | GUI section complete |
| 18.8 | AGENTS.md | M | 18.5 | Contributor/agent guide |
| 18.9 | Config field ADR | S | 18.5 | Locked decisions in DESIGN |
| 18.10 | Cross-link and status finalization | S | 18.1–18.9 | All docs linked |

---

## Steps

### 18.1 — Add LICENSE

**Size:** S · **Depends on:** — · **Blocks:** M22 goreleaser

**Implementation:**
1. Add standard MIT `LICENSE` with copyright holder name/year
2. Verify `.goreleaser.yml` `files: LICENSE` resolves
3. Run `goreleaser release --snapshot --clean` (optional dry-run)

**Acceptance criteria:**
- [ ] `LICENSE` exists at repo root
- [ ] README License section matches

**Tests:** None (legal artifact)

---

### 18.2 — Fix README

**Size:** S · **Depends on:** —

**Implementation:**
1. **Status:** "Early development — TUI functional, not daily-driver reliable"
2. **Usage:** `lazybrew` launches TUI; document `--version`, `--config`, `--debug`
3. **Project structure:** Update tree (gui built, config exists)
4. **Keybindings:** Brief table or link to `?` help
5. **Config:** Path `~/.config/lazybrew/config.yml`
6. Remove "prints version and exits" unless `--version`

**Acceptance criteria:**
- [ ] New contributor can build, run, and find config from README alone
- [ ] No statement contradicting `cmd/lazybrew/main.go`

**Tests:** Manual — follow README from clean clone

---

### 18.3 — Milestone DoD Audit Worksheet

**Size:** S · **Depends on:** —

**Implementation:**
1. For each milestone M1–M17, walk Definition of Done checkboxes
2. Record in table: Plan claim | Code evidence | Test evidence | Verdict
3. Use [review-template.md](../review-template.md) Plan vs Code worksheet
4. Save as section in `architecture-review-2026-06-13.md` appendix or `status.md`

**Acceptance criteria:**
- [ ] Every ⚠️/✅ in status.md backed by worksheet row
- [ ] Discrepancies list matches [planning-challenge](../planning-challenge-2026-06-13.md)

**Tests:** N/A (audit artifact)

---

### 18.4 — Reconcile M1–M17 Milestone Headers

**Size:** M · **Depends on:** 18.3

**Implementation:**
1. Use [milestone-legacy-index.md](../milestone-legacy-index.md) as checklist
2. Update each `milestones/0N-*.md` header status emoji
3. Ensure **Active Work Routing** section exists on all partial milestones (done 2026-06-13 for M2,M4,M6–M8,M11,M12,M14–M16; M17 refined)
4. Unchecked DoD items that aren't done
5. Do **not** rewrite legacy step bodies — only status + routing + Remaining

**Files:** `master-plan/milestones/01-*.md` through `17-*.md`

**Acceptance criteria:**
- [ ] No milestone header says ✅ Complete with open DoD items
- [ ] M6 header reflects TaskManager missing
- [ ] M16 header reflects actual test count (~162) and coverage

**Tests:** N/A

---

### 18.5 — DESIGN.md Skeleton + Concurrency ADR

**Size:** M · **Depends on:** — · **Blocks:** M19.1

**Implementation:**
1. Create `DESIGN.md` with sections (empty stubs OK except marked):
   - Overview, Goals, Non-goals
   - System diagram (cmd → app → gui/brew)
   - Layer rules (dependency direction)
   - **Concurrency ADR (D19-1–D19-5)** — copy from M19
   - Config schema (pointer to M18.9)
   - Testing tiers (unit / integration / teatest)
   - Decision log table
2. Link from README and `status.md`

**Acceptance criteria:**
- [ ] M19 executor can implement TaskManager without reading 17 milestone files
- [ ] Concurrency ADR explicitly forbids `program.Send` from handlers

**Tests:** N/A

---

### 18.6 — DESIGN.md Brew Layer

**Size:** S · **Depends on:** 18.5

**Content to document:**
- Runner interface + env vars (`HOMEBREW_NO_ASK`, `HOMEBREW_NO_AUTO_UPDATE`)
- Client composition (Reader/Writer fields)
- Cache keys + `InvalidateGroups`
- Typed errors
- Test fixture policy (synthetic JSON)

**Acceptance criteria:**
- [ ] New brew command checklist: service method, cache invalidation, unit test, coverage audit row

---

### 18.7 — DESIGN.md GUI Layer

**Size:** S · **Depends on:** 18.5

**Content to document:**
- Model ownership (`gui.Model` fields summary)
- Message types index (link to `messages.go`)
- Modal flow (activeModal, handleModalResult)
- Panel/tab/content model
- TaskManager role (forward reference M19)
- Rendering pipeline: `View()` → renderSidebar / renderMainPanel / renderBottomBar

**Acceptance criteria:**
- [ ] Explains where to add a new keybinding (keybindings.go + gui.go Update + help.go)

---

### 18.8 — AGENTS.md

**Size:** M · **Depends on:** 18.5 · **Blocks:** M19.6 merge

**Sections:**
1. **Quick start** — clone, `make build`, `make test`, `make run`
2. **Repository map** — table of packages
3. **Change guidelines** — minimize scope, match conventions, no drive-by refactors
4. **Bubble Tea rules** — all state via Update; TaskManager for writes; no raw goroutines in commands
5. **Testing rules** — test in same step; tiers; `make test-integration` tag
6. **Planning rules** — update status.md; use [templates/](../templates/); review-template for audits
7. **Git rules** — no secrets; commit only when asked
8. **Verification commands** — `go test -race ./...`, `go vet ./...`

**Acceptance criteria:**
- [ ] Agent can execute M19.6 knowing concurrency rules
- [ ] Links to DESIGN.md, review-template, status.md

---

### 18.9 — Config Field ADR

**Size:** S · **Depends on:** 18.5

**Locked decisions** (from planning-challenge):

| Field | Status | M20 action |
|---|---|---|
| `AutoRefreshSeconds` | Planned | Wire tick |
| `Brew.Path` | Planned | Wire runner |
| `ShowIcons` | Deferred P2 | No-op + comment |

Add YAML comments in `config.go` struct tags or file header comment block.

**Acceptance criteria:**
- [ ] No "choose one" ambiguity in M20.8
- [ ] DESIGN decision log entries D18-4

---

### 18.10 — Cross-Link and Status Finalization

**Size:** S · **Depends on:** 18.1–18.9

**Implementation:**
1. README links: DESIGN, AGENTS, master-plan/status
2. status.md links: review-template, planning-challenge, architecture-review
3. Verify no broken "see design doc" references

**Acceptance criteria:**
- [ ] Every doc referenced from status.md exists

---

## Test Plan

| Step | Verification |
|---|---|
| 18.2 | Manual README walkthrough |
| 18.1 | goreleaser snapshot (optional) |
| All | Grep for "TUI is not yet built" → zero hits |

---

## Definition of Done

- [ ] Steps 18.1–18.10 complete
- [ ] DESIGN.md + AGENTS.md exist and linked
- [ ] LICENSE exists
- [ ] M1–M17 headers match status.md
- [ ] Config ADR locked (D18-4)
- [ ] status.md updated

---

## Post-Milestone Gate

- [ ] M19.1 may start (DESIGN concurrency ADR present)
- [ ] M22.3 goreleaser unblocked (LICENSE present)
