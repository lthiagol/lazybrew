# Milestone 20 — Functional Completeness & UX Correctness

> **Status:** 🔜 Planned  
> **Size estimate:** L (5–7 days)  
> **Depends on:** M19 ✅ (TaskManager for 20.3+)  
> **Enables:** M17, daily-driver use  
> **Parallel track:** C (UX) — phase A can start during M19.8  
> **Gate criteria:** Tab content matches selection; Info tab ≠ sidebar; batch upgrade works

Execute **phases A → F in order**. Do not skip phases.

| Phase | Steps | Theme |
|---|---|---|
| **A — Data truth** | 20.1, 20.6, 20.9 | Cache keys, typed data, errors |
| **B — Display truth** | 20.2, 20.5, 20.10 | Info tab, empty states, scroll |
| **C — Action truth** | 20.3, 20.4 | Batch upgrade, pin |
| **D — Config truth** | 20.8 | Wire planned config fields |
| **E — Layout minimum** | 20.7 | Small terminal warning |
| **F — Verification** | 20.11 | Manual smoke |

---

## Goal

Fix user-visible incorrectness: stale tabs, wrong Info content, broken batch select, pin semantics, silent errors, and undocumented config behavior.

---

## Out of Scope

- Lazygit boxes / accordion (M17)
- Search info preview (M17.11)
- Controller package split (B-01)
- Lazy panel loading (B-02)

---

## Architecture Decisions (ADRs)

| ID | Decision |
|---|---|
| D20-1 | Tab cache key = `panel:tab:itemName` |
| D20-2 | Info tab uses list JSON data first; `Get()` only if field missing |
| D20-3 | Batch upgrade sequential through TaskManager (not parallel brew) |
| D20-4 | Viewport used for tab content > visible height |

---

## Step Index

| Step | Title | Size | Phase | Depends |
|---|---|---|---|---|
| 20.1 | Tab cache key + invalidation | M | A | — |
| 20.6 | Outdated typed data | S | A | — |
| 20.9 | Propagate fetch errors | S | A | — |
| 20.2 | Info tab formatters + render | M | B | 20.1 |
| 20.5 | Panel empty states | S | B | — |
| 20.10 | Wire viewport for tab content | M | B | 20.1 |
| 20.3 | Batch upgrade | M | C | M19, 20.6 |
| 20.4 | Pin toggle fix | S | C | M19 |
| 20.8 | Config field wiring | M | D | 18.9 |
| 20.7 | Small terminal warning | S | E | — |
| 20.11 | Manual smoke checklist | S | F | all |

---

## Phase A — Data Truth

### 20.1 — Tab Cache Key + Invalidation

**Size:** M · **Phase:** A

**Problem:** `tabKey(panel, tab)` ignores selected item → stale Deps/Used By/Files.

**Files:** `commands.go`, `gui.go`, `render.go`

**Implementation:**
1. Change signature:
   ```go
   func tabKey(panel PanelID, tab int, itemName string) string
   ```
2. Update all readers/writers of `tabContent`
3. Add `needsTabFetch(panel, tab) bool` — centralize logic from `loadTabContent`
4. On `j`/`k` in `Update`: if active tab needs fetch → `return m, m.loadTabContent()`
5. On `switchPanel` / `nextPanel` / `prevPanel`: reset `activeTab` or trigger fetch if tab > 0 needs data
6. On `RefreshMsg` and successful `DataLoadedMsg`: `clearTabContent()` or clear keys for affected panel
7. On selection change within same panel: invalidate keys for old item only (optional optimization)

**Acceptance criteria:**
- [ ] Select formula A → Deps tab → j/k to B → Deps shows B's deps (mock test)
- [ ] Refresh clears stale tab cache

**Tests:** `TestTabContentRefetchOnSelection`, `TestTabKeyIncludesItemName`

**Out of scope:** Prefetch on hover — not applicable

---

### 20.6 — Outdated Panel Typed Data

**Size:** S · **Phase:** A

**File:** `commands.go` — `fetchPanelData` Outdated case

**Implementation:**
1. Keep separate slices: `outdatedFormulae`, `outdatedCasks` on `panelData` OR reuse `formulae`/`casks` fields
2. Extend `DataLoadedMsg` if new fields needed
3. Populate typed data from `client.Formulae.Outdated` / `Casks.Outdated`

**Acceptance criteria:**
- [ ] `selectedFormula()` works on Outdated panel when applicable
- [ ] Versions tab (tab 1) can access `NewVersion`

**Tests:** `TestOutdatedPanelTypedData`

---

### 20.9 — Propagate fetchPanelData Errors

**Size:** S · **Phase:** A

**Files:** `commands.go` — Outdated, `fetchStatusData`

**Implementation:**
1. Replace `formulae, _ := client.Formulae.Outdated` with error handling
2. Status dashboard: surface partial failure (e.g. casks fail → show error badge on panel)
3. `renderList` already shows `p.err` — ensure it's set

**Acceptance criteria:**
- [ ] Mock runner error → panel shows error state, not empty list silently

**Tests:** `TestOutdatedFetchSurfacesError`, `TestStatusFetchPartialError`

---

## Phase B — Display Truth

### 20.2 — Info Tab Formatters + Render

**Size:** M · **Phase:** B · **Depends on:** 20.1

**Files:**
- `presentation/formatters.go` — add `FormatFormulaInfo`, `FormatCaskInfo`
- `render.go` — Formulae/Casks tab 0

**FormatFormulaInfo minimum fields:**
Name, Version, Tap, Status (installed/keg-only/pinned/outdated), License, Homepage, Description (truncated), Bottled

**Implementation:**
1. Add formatters with snapshot tests
2. `renderContent` case PanelFormulae tab 0:
   ```go
   f := panel.selectedFormula()
   if f == nil { return emptyState }
   return presentation.FormatFormulaInfo(*f, width)
   ```
3. Same for Casks tab 0
4. Outdated panel tab 0: info for selected outdated item (uses 20.6 typed data)

**Acceptance criteria:**
- [ ] Info tab text ≠ sidebar single-line entry
- [ ] Snapshot stable for fixture formula

**Tests:** `TestFormatFormulaInfoSnapshot`, `TestInfoTabNotDuplicateList`

---

### 20.5 — Panel Empty States

**Size:** S · **Phase:** B

**File:** `panel.go` — `renderList` or new `emptyMessage(panel PanelID) string`

| Panel | Message |
|---|---|
| Formulae | `No formulae installed` |
| Casks | `No casks installed` |
| Outdated | `Everything up to date!` |
| Taps | `No custom taps` (or `No taps` if list empty includes official) |
| Services | `No services configured` |
| Search | `No results` / `Press / to search` when never searched |

**Acceptance criteria:**
- [ ] Each panel shows distinct message (snapshot)

**Tests:** `TestEmptyStateMessages` table-driven

---

### 20.10 — Wire Viewport for Tab Content

**Size:** M · **Phase:** B · **Depends on:** 20.1

**Problem:** Root `viewport` unused; long Deps/Files clip without scroll.

**Implementation:**
1. **Option chosen (D20-4):** Use `m.viewport` for main tab content when content height > available
2. In `renderContent` for tabs 1,2,4 (Formulae), 1,2 (Casks), Status 1,2:
   - Set `viewport.SetContent(content)`
   - Return `m.viewport.View()`
3. Route scroll keys (`j`/`k` or `up`/`down`) to viewport when focused on scrollable tab — **or** use `ctrl+j/k` to avoid conflict with list navigation (document in help)
4. If scroll conflict too complex: viewport scroll only when content tab active AND modifier held — document decision in DESIGN

**Recommended:** When `activeTab != 0` on panels with list in sidebar, `j`/`k` scroll viewport; sidebar list uses when tab 0. Simpler: tab 0 = list in main was wrong; after 20.2 tab 0 is info text; **j/k always moves sidebar selection**; use `ctrl+d/u` or `pgup/pgdown` for viewport scroll.

**Acceptance criteria:**
- [ ] 100-line Files tab scrollable
- [ ] Documented scroll keys in help.go

**Tests:** `TestViewportRendersLongContent`

---

## Phase C — Action Truth

### 20.3 — Batch Upgrade

**Size:** M · **Phase:** C · **Depends on:** M19, 20.6

**Files:** `panel.go`, `render.go`, `commands.go`, `task.go`

**Implementation:**
1. Visual: prefix `● ` for indices in `batch.selected` when rendering Outdated list
2. Key `a` on Outdated: select all (M6 spec) — verify if missing, add
3. Key `u` on Outdated:
   - If any batch selected → enqueue Task per selected name (sequential via manager queue)
   - Else → single upgrade current item (existing)
4. Detect formula vs cask from typed outdated data
5. On all TaskCompletedMsg success → `batch.selected = clear`, RefreshMsg

**Acceptance criteria:**
- [ ] Select 3 outdated → u → 3 upgrade commands in order (mock records args)
- [ ] Selection cleared after success

**Tests:** `TestBatchUpgradeSelected`, `TestBatchUpgradeSequential`

---

### 20.4 — Pin Toggle Fix

**Size:** S · **Phase:** C · **Depends on:** M19

**File:** `commands.go` — `togglePin`

**Implementation:**
```go
if m.activePanel == PanelCasks {
    c := panel.selectedCask()
    if c == nil { return m, nil }
    if c.Pinned { err = Unpin } else { err = Pin }
} else {
    f := panel.selectedFormula()
    // same with f.Pinned
}
```

Submit via TaskManager (short task).

**Acceptance criteria:**
- [ ] Unpinned package → Pin called once, not Unpin-then-Pin

**Tests:** `TestPinRespectsPinnedFlag`

---

## Phase D — Config Truth

### 20.8 — Wire Config Fields (per M18.9 ADR)

**Size:** M · **Phase:** D

| Field | Implementation |
|---|---|
| `AutoRefreshSeconds` | If > 0: `tea.Tick` in Init/refresh re-arm → `RefreshMsg`. If 0: disabled. |
| `Brew.Path` | `app.New`: if set, pass to `brew.NewRunnerWithPath(path)` (new constructor) |
| `ShowIcons` | No-op; comment `// deferred M17`; ignore in render |

**Files:** `config.go`, `app.go`, `runner.go`, `gui.go`

**Acceptance criteria:**
- [ ] Config example in README works for custom brew path (mock test)
- [ ] Auto-refresh tick fires (test with fake clock or injectable tick — or test handler only)

**Tests:** `TestAutoRefreshTick`, `TestCustomBrewPath`, `TestConfigShowIconsIgnored`

---

## Phase E — Layout Minimum

### 20.7 — Small Terminal Warning

**Size:** S · **Phase:** E

**Implementation:**
1. Constants: `minWidth=80`, `minHeight=24`
2. On `WindowSizeMsg`: if below min → set `m.terminalTooSmall=true`
3. `View()`: render warning banner above body or replace body with centered warning
4. Do **not** implement accordion (M17)

**Acceptance criteria:**
- [ ] 79×24 shows warning; 80×24 does not

**Tests:** `TestSmallTerminalWarning`

---

## Phase F — Verification

### 20.11 — Manual Smoke Checklist

**Size:** S · **Phase:** F

Create `master-plan/smoke-checklist.md`:

- [ ] Launch TUI; all panels load
- [ ] Formulae Info tab shows details
- [ ] j/k on Formulae → Deps tab updates
- [ ] Outdated batch select + upgrade
- [ ] Pin/unpin correct
- [ ] Doctor shows output in modal
- [ ] Config auto-refresh (set to 30s, observe refresh)
- [ ] 79-col terminal shows warning

**Acceptance criteria:**
- [ ] All items checked on real Homebrew machine

---

## Test Plan (milestone-level)

| Test | Phase | Step |
|---|---|---|
| Tab refetch | A | 20.1 |
| Outdated typed + errors | A | 20.6, 20.9 |
| Info snapshot | B | 20.2 |
| Empty states | B | 20.5 |
| Viewport scroll | B | 20.10 |
| Batch upgrade | C | 20.3 |
| Pin semantics | C | 20.4 |
| Config wiring | D | 20.8 |
| Small terminal | E | 20.7 |

---

## Definition of Done

- [ ] Phases A–F complete
- [ ] All step tests exist and pass
- [ ] smoke-checklist.md executed
- [ ] DESIGN.md updated (viewport keys, tab cache)
- [ ] status.md updated

---

## Post-Milestone Gate

- [ ] M21 T2 E2E flows may assert Info tab + tabs
- [ ] M17 visual work won't invalidate tab logic
