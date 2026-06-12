# Milestone 7 — Taps & Trust Management

> **Status:** 🔲 Not Started  
> **Depends on:** Milestone 5 (Modals), Milestone 6 (Package Mutations — for task manager)  
> **Enables:** Milestone 8 (P1 Features)

---

## Goal

Implement full tap lifecycle management (tap, untap, repair) and the trust system (`brew trust`, `brew untrust`) with a dedicated trust configuration UI. After this milestone, the Taps panel is fully interactive — users can add/remove taps and manage trust from the TUI.

---

## Steps

### 7.1 — Tap Action: Add a New Tap

**What:** Allow users to tap a new repository.

**Keybinding:** `a` in Taps panel

**Flow:**
```
Taps panel → press `a`
  → InputModal: "Tap repository (user/repo):"
    placeholder: "e.g., nicknisi/tap"
  → User enters "some-org/formulas", presses Enter
  → ConfirmModal: "Tap some-org/formulas?"
    "This will clone the repository and make its formulae available."
  → User confirms
  → TaskManager: brew tap some-org/formulas
  → ProgressModal: streaming clone output
  → On success: refresh taps list, toast "✓ Tapped some-org/formulas"
```

**Validation:**
- Input must match `user/repo` pattern
- Check if already tapped → toast warning
- Support custom URL variant: if input contains `://`, use `brew tap <name> <url>`

**Acceptance criteria:**
- [ ] Input modal accepts tap name
- [ ] Validation rejects invalid format
- [ ] Already-tapped check
- [ ] Custom URL support
- [ ] Progress shown during clone
- [ ] Taps list refreshes on success

---

### 7.2 — Untap Action

**What:** Remove a tapped repository.

**Keybinding:** `x` or `d` in Taps panel

**Flow:**
```
Taps panel → select "some-org/formulas" → press `x`
  → Check: any installed formulae from this tap?
    → If yes: ConfirmModal with warning
      "Untap some-org/formulas?
       ⚠ 3 installed formulae come from this tap:
         some-org/formulas/tool-a
         some-org/formulas/tool-b  
         some-org/formulas/tool-c
       These will become unavailable for updates."
    → If no: simple ConfirmModal
  → User confirms
  → TaskManager: brew untap some-org/formulas
  → On success: remove from list, toast
```

**Protection:**
- Cannot untap any `homebrew/*` tap (show error toast: "Cannot untap official taps")
- This includes `homebrew/core`, `homebrew/cask`, `homebrew/services`, `homebrew/bundle`, etc.
- Check is done by prefix match on tap name, not a hardcoded list

**Acceptance criteria:**
- [ ] Confirmation with installed-from-tap warning
- [ ] Official tap protection
- [ ] List updates on success
- [ ] Selection moves to next tap

---

### 7.3 — Trust Management UI

**What:** View and modify trust status for taps, formulae, and casks.

**Keybinding:** `t` in Taps panel (on a selected tap)

**Flow:**
```
Taps panel → select "nicknisi/tap" → press `t`
  → MenuModal: "Trust Configuration: nicknisi/tap"
    Current status: ⚠ untrusted (third-party)
    
    Options:
      [1] Trust entire tap
      [2] Trust specific formulae...
      [3] Trust specific casks...
      [4] Untrust (if currently trusted)
```

**Option 1 — Trust entire tap:**
```
  → ConfirmModal: "Trust all formulae and casks from nicknisi/tap?"
  → brew trust nicknisi/tap
  → Success: update trust indicator in list
```

**Note (Homebrew 6.0.0):** `brew trust --json=v1` provides machine-readable trust state. The TrustService (M3 §3.4) uses this to query current trust rather than parsing text. `brew tap-info --json` also returns a `trusted` boolean field, which is the preferred way to check trust for a specific tap.

**Option 2 — Trust specific formulae:**
```
  → Fetch list of formulae from tap (from tap-info)
  → Show loading spinner in menu while fetching
  → MenuModal: select formulae to trust (multi-select)
  → brew trust --formula nicknisi/tap/formula-name (for each)
  → Success: update trust info
```

> **Async note:** Fetching the list of formulae/casks from a tap can be slow for large taps (hundreds of items). The menu should show a loading state ("Loading formulae from nicknisi/tap...") while fetching, and only display the selection list after the data arrives. Use the cached `tap-info` data if available (M3 cache).

**Option 3 — Trust specific casks:**
```
  → Same as option 2 but with --cask flag
```

**Option 4 — Untrust:**
```
  → ConfirmModal: "Remove trust from nicknisi/tap?"
  → brew untrust nicknisi/tap
  → Success: update trust indicator
```

**Trust Tab in Main Panel (when Taps panel is active):**
```
Trust Status: nicknisi/tap
═══════════════════════════

Overall: ⚠ Untrusted (third-party)

Trusted items:
  📦 nicknisi/tap/some-formula     ✓ trusted
  📦 nicknisi/tap/another-formula  ✓ trusted

Untrusted items:
  📦 nicknisi/tap/risky-tool       ⚠ untrusted
  🖥  nicknisi/tap/some-cask       ⚠ untrusted

Press 't' to manage trust configuration
```

**Acceptance criteria:**
- [ ] Trust menu shows current status
- [ ] Trust entire tap works
- [ ] Trust specific formula works
- [ ] Trust specific cask works
- [ ] Untrust works
- [ ] Trust indicators update in taps list after changes
- [ ] Trust tab in main panel shows detailed trust state
- [ ] Official taps show "always trusted" (no trust actions available)

---

### 7.4 — Tap Repair

**What:** Run `brew tap --repair` for a tap.

**Keybinding:** `r` in Taps panel

**Flow:**
```
Taps panel → select tap → press `r`
  → ConfirmModal: "Repair nicknisi/tap?"
  → TaskManager: brew tap --repair nicknisi/tap
  → ProgressModal: output
  → Success: toast
```

**Acceptance criteria:**
- [ ] Repair runs with progress
- [ ] Success/failure reported

---

### 7.5 — Taps Panel Keybinding Summary

Final keybindings for the Taps panel:

| Key | Action |
|---|---|
| `Enter` | Show tap info in main panel |
| `a` | Add (tap) a new repository |
| `x` / `d` | Untap (with confirmation) |
| `t` | Open trust configuration menu |
| `r` | Repair tap |
| `o` | Open tap's GitHub page in browser |
| `y` | Copy tap name to clipboard |

**Bottom bar:** `a: add  x: remove  t: trust  r: repair  ?: help`

---

## Tests for This Milestone

| Test | Type | File | What It Validates |
|---|---|---|---|
| `TestTapAdd` | E2E (teatest) | `internal/gui/flows/tap_test.go` | Full tap flow: input → confirm → progress → list update |
| `TestTapAdd_Invalid` | E2E (teatest) | `internal/gui/flows/tap_test.go` | Invalid format rejected |
| `TestTapAdd_AlreadyTapped` | E2E (teatest) | `internal/gui/flows/tap_test.go` | Duplicate detected |
| `TestUntap` | E2E (teatest) | `internal/gui/flows/tap_test.go` | Untap with confirmation |
| `TestUntap_WithInstalled` | E2E (teatest) | `internal/gui/flows/tap_test.go` | Warning shows installed packages |
| `TestUntap_OfficialBlocked` | E2E (teatest) | `internal/gui/flows/tap_test.go` | Cannot untap homebrew/core |
| `TestTrustMenu` | E2E (teatest) | `internal/gui/flows/trust_test.go` | Menu opens with correct options |
| `TestTrustEntireTap` | E2E (teatest) | `internal/gui/flows/trust_test.go` | Trust tap command executed |
| `TestTrustFormula` | E2E (teatest) | `internal/gui/flows/trust_test.go` | Trust specific formula |
| `TestUntrust` | E2E (teatest) | `internal/gui/flows/trust_test.go` | Untrust command executed |
| `TestTrustIndicators` | Snapshot | `internal/gui/presentation/taps_test.go` | All trust states render correctly |
| `TestTrustTab` | E2E (teatest) | `internal/gui/flows/trust_test.go` | Trust tab shows detailed state |
| `TestTapRepair` | E2E (teatest) | `internal/gui/flows/tap_test.go` | Repair flow |

---

## Definition of Done

- [ ] Tap add works (with input validation)
- [ ] Untap works (with dependency warning)
- [ ] Trust menu allows trust/untrust at tap and formula/cask granularity
- [ ] Trust indicators accurate in the taps list
- [ ] Trust tab in main panel shows detailed trust state
- [ ] Tap repair works
- [ ] All keybindings active and shown in bottom bar
- [ ] All tests pass
- [ ] Official taps protected from untap/trust-change
