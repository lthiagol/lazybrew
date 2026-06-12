# Milestone 8 — P1 Features (Services, Pin, Cleanup, Doctor)

> **Status:** ✅ Done  
> **Depends on:** Milestone 6 (Package Mutations)  
> **Enables:** Milestone 9 (Polish & Release)

---

## Goal

Implement the "should-have" features that round out lazybrew beyond basic package management: service control, pin/unpin, cleanup, doctor diagnostics, leaves view, autoremove, and Brewfile support. After this milestone, lazybrew covers the full breadth of Homebrew operations.

---

## Steps

### 8.1 — Services Panel (Interactive)

**What:** Make the Services panel fully interactive — start, stop, restart services.

**Files:** Update `internal/gui/controllers/services_panel.go`

**Keybindings:**

| Key | Action |
|---|---|
| `s` | Start service |
| `S` | Stop service |
| `r` | Restart service |
| `f` | Run service (foreground, without auto-start on boot; 6.0.0 gap) |
| `c` | Cleanup stale service files (6.0.0 gap) |
| `Enter` | View service details in main panel |

**Start flow:**
```
Services panel → select "redis" (stopped) → press `s`
  → ConfirmModal: "Start redis?"
  → brew services start redis
  → Success: status changes to "started", indicator goes green
```

**Stop flow:**
```
Services panel → select "postgresql@16" (started) → press `S`
  → ConfirmModal: "Stop postgresql@16?"
  → brew services stop postgresql@16
  → Success: status changes to "stopped"
```

**Main panel content for Services:**
```
Service: postgresql@16
══════════════════════

Status:    ● started
User:      thiago
Plist:     /Users/thiago/Library/LaunchAgents/...
Exit code: 0

─────────────────────

Log file:  /opt/homebrew/var/log/postgresql@16.log

(Last 20 lines of log)
2026-06-11 12:00:01 UTC LOG: database system is ready...
2026-06-11 12:00:01 UTC LOG: listening on IPv4 address...
```

**Acceptance criteria:**
- [ ] Start/stop/restart work
- [ ] Status indicators update after action
- [ ] Service details shown in main panel
- [ ] Log tail shown (if log file exists)
- [ ] Context-appropriate actions (can't stop a stopped service)
- [ ] Works on macOS (launchctl) and Linux (systemd)

---

### 8.2 — Pin / Unpin

**What:** Pin formulae and casks to prevent them from being upgraded. **Homebrew 6.0.0 added cask pinning support.**

**Keybinding:** `p` in Formulae **and Casks** panels (toggles pin/unpin)

**Flow:**
```
Formulae panel → select "python@3.12" → press `p`
  → If not pinned:
    → brew pin python@3.12
    → Badge changes to "⊘ pinned"
    → Toast: "✓ python@3.12 pinned"
  → If already pinned:
    → brew unpin python@3.12
    → Badge removed
    → Toast: "✓ python@3.12 unpinned"

Casks panel → select "google-chrome" → press `p`
  → Same flow, uses brew pin/unpin --cask flag
```
**No confirmation needed** — pin/unpin is non-destructive and easily reversible.

**Acceptance criteria:**
- [ ] Pin toggles on/off for formulae
- [ ] Pin toggles on/off for casks (6.0.0)
- [ ] Pinned badge appears/disappears in list for both panels
- [ ] Pinned packages excluded from "Upgrade All" in Outdated panel
- [ ] Toast notification

**FormulaeService/CasksService update (6.0.0):**
- `Pin()` / `Unpin()` methods pass `--formula` / `--cask` flags as appropriate
- Cask type needs `Pinned` field in `types.go` (6.0.0 addition)

---

### 8.3 — Cleanup

**What:** Run `brew cleanup` with optional dry-run preview.

**Keybinding:** `c` in Status panel

**Flow:**
```
Status panel → press `c`
  → First: run `brew cleanup -n` (dry-run)
  → Show results in a menu/info modal:
    "Cleanup Preview:
     Would remove: 15 old versions
     Would free: 2.3 GB
     
     [1] Run cleanup
     [2] Cancel"
  → If user selects "Run cleanup":
    → ProgressModal: brew cleanup
    → Success: toast with space freed
```

**Acceptance criteria:**
- [ ] Dry-run preview shown first
- [ ] Space savings displayed
- [ ] Actual cleanup runs only after explicit approval
- [ ] Progress shown during cleanup
- [ ] Success toast with results

---

### 8.4 — Doctor Diagnostics

**What:** Run `brew doctor` and display results.

**Keybinding:** `d` in Status panel

**Flow:**
```
Status panel → press `d`
  → ProgressModal: "Running brew doctor..."
  → brew doctor
  → Results displayed in main panel (Doctor tab):
    
    ✓ Your system is ready to brew.
    
    ── or ──
    
    ⚠ 2 warnings found:
    
    Warning: Your Homebrew is outdated.
    You haven't updated in the last 30 days.
    Run `brew update` to get the latest.
    
    Warning: Broken symlinks:
    /opt/homebrew/bin/old-tool → (missing)
```

**Acceptance criteria:**
- [ ] Doctor runs with progress indicator
- [ ] Clean system shows green checkmark
- [ ] Warnings shown with structure (title + details)
- [ ] Results persist in the Doctor tab until next run
- [ ] Status panel summary updates ("Doctor: ✓ No issues" or "⚠ 2 warnings")

---

### 8.5 — Leaves View

**What:** Show top-level packages (not depended on by others) in the Formulae panel.

**Keybinding:** `L` in Formulae panel (toggle filter)

**Behavior:**
- `L` toggles between "All formulae" and "Leaves only"
- Panel title changes: "📦 Formulae (124)" → "🌿 Leaves (87)"
- Useful for seeing which packages you explicitly installed vs transitive deps

**Implementation:**
- Calls `brew leaves` to get the list
- Filters the existing formulae list to show only leaves
- Toggle state persists during the session

**Acceptance criteria:**
- [ ] `L` toggles leaves filter
- [ ] Panel title and icon change
- [ ] Filter is accurate (matches `brew leaves` output)
- [ ] Toggle back shows full list

---

### 8.6 — Autoremove

**What:** Remove orphaned dependencies that are no longer needed.

**Keybinding:** `A` in Status panel

**Flow:**
```
Status panel → press `A`
  → First: brew autoremove -n (dry-run)
  → Show preview:
    "Autoremove Preview:
     Would remove 5 orphaned dependencies:
       libfoo 1.2.3
       libbar 4.5.6
       ...
     
     [1] Remove orphans
     [2] Cancel"
  → If confirmed: brew autoremove
  → Success: refresh formulae list, toast
```

**Acceptance criteria:**
- [ ] Dry-run preview first
- [ ] List of orphans shown
- [ ] Actual removal after confirmation
- [ ] Formulae list refreshes
- [ ] Empty case: "No orphaned dependencies found"

---

### 8.7 — Brewfile Support

**What:** Export and import packages via Brewfile.

**Keybinding:** `B` in Status panel opens Brewfile menu

**Menu options:**
```
Brewfile Actions:
  [1] Export to Brewfile (brew bundle dump)
  [2] Install from Brewfile (brew bundle install)
  [3] Cleanup (remove packages not in Brewfile)
  [4] Check (verify Brewfile is satisfied)
  [5] List (list Brewfile entries)
```

**Note (Homebrew 6.0.0):**
- `brew bundle dump` records `trusted:` entries for taps that have been trusted
- `brew bundle install` now runs formula installations in parallel by default
- Brewfile supports npm, krew, winget extensions (handled by brew, no TUI changes needed)
- `brew bundle check` and `brew bundle list` added to close coverage audit gaps

**Export flow:**
```
→ InputModal: "Save Brewfile to:" (default: ~/Brewfile)
→ brew bundle dump --file=<path>
→ Success: toast "✓ Brewfile saved to ~/Brewfile"
```

**Install flow:**
```
→ InputModal: "Brewfile path:" (default: ~/Brewfile)
→ ConfirmModal: "Install packages from ~/Brewfile?"
→ ProgressModal: brew bundle install --file=<path>
→ Success: refresh all panels
```

**Cleanup flow:**
```
→ InputModal: "Brewfile path:" (default: ~/Brewfile)
→ brew bundle cleanup --file=<path> (dry-run first)
→ Show packages that would be removed
→ Confirm → brew bundle cleanup --force --file=<path>
```

**Check flow:**
```
→ InputModal: "Brewfile path:" (default: ~/Brewfile)
→ brew bundle check --file=<path>
→ Result: "Brewfile is satisfied" or list of unsatisfied dependencies
```

**List flow:**
```
→ brew bundle list --file=<path>
→ Display entries (formulae, casks, taps, trusted entries) in main panel
→ 6.0.0: show `trusted:` annotations for taps
```

**Acceptance criteria:**
- [ ] Export creates a valid Brewfile (with `trusted:` entries on 6.0.0+)
- [ ] Install from Brewfile works
- [ ] Cleanup with dry-run preview
- [ ] Check verifies Brewfile satisfaction
- [ ] List displays Brewfile contents
- [ ] Custom file path support
- [ ] Default path (~/Brewfile)

---

### 8.8 — Vulnerability Check (`brew vulns`)

**What:** Check installed packages for known CVEs using `brew vulns` (from the `homebrew-brew-vulns` tap).

**Keybinding:** `v` in Status panel

**Prerequisite:** The `homebrew-brew-vulns` tap must be installed. If not, show a toast: "Install homebrew/homebrew-brew-vulns to enable vulnerability checks."

**Flow:**
```
Status panel → press `v`
  → ProgressModal: "Checking for known vulnerabilities..."
  → brew vulns
  → Results displayed in main panel (Vulns tab):
    
    ✓ No known vulnerabilities found.
    
    ── or ──
    
    ⚠ 3 packages with known vulnerabilities:
    
    openssl@3  3.2.1
      CVE-2024-1234: Buffer overflow in X509 verification
      https://nvd.nist.gov/vuln/detail/CVE-2024-1234
    
    curl  8.5.0
      CVE-2024-5678: HTTP/2 rapid reset
      https://nvd.nist.gov/vuln/detail/CVE-2024-5678
```

**Acceptance criteria:**
- [ ] `brew vulns` runs with progress indicator
- [ ] Clean result shows green checkmark
- [ ] Vulnerabilities shown with CVE ID, description, and link
- [ ] Results persist in the Vulns tab until next run
- [ ] Missing tap shows helpful install message
- [ ] Status panel summary updates ("Vulns: ✓ Clean" or "⚠ 3 vulnerabilities")

---

### 8.9 — Missing Dependencies (`brew missing`)

**What:** Check for missing dependencies of installed formulae and casks.

**Keybinding:** `m` in Status panel

**Flow:**
```
Status panel → press `m`
  → ProgressModal: "Checking for missing dependencies..."
  → brew missing
  → Results displayed in main panel (Missing tab):
    
    ✓ All dependencies are satisfied.
    
    ── or ──
    
    ⚠ 2 missing dependencies:
    
    python@3.12: missing libsqlite
    node: missing icu4c
```

**Acceptance criteria:**
- [ ] `brew missing` runs with progress indicator
- [ ] Clean result shows green checkmark
- [ ] Missing deps shown with parent formula and missing dep name
- [ ] Results persist in the Missing tab until next run
- [ ] Status panel summary updates ("Missing: ✓ All satisfied" or "⚠ 2 missing")

---

### 8.10 — Status Panel Keybinding Summary

Updated keybindings for Status panel:

| Key | Action |
|---|---|
| `u` | Run `brew update` |
| `U` | Update + upgrade all |
| `d` | Run `brew doctor` |
| `c` | Cleanup (dry-run preview first) |
| `A` | Autoremove orphans (dry-run first) |
| `B` | Brewfile menu |
| `v` | Run `brew vulns` (vulnerability check) |
| `m` | Run `brew missing` (missing deps check) |

---

### 8.11 — Services Keybinding Conflict Resolution

> **Issue:** In M8 §8.1, `R` was assigned to "Run service" and `r` to "Restart service". These are too similar and easy to confuse.

**Resolution:** Change "Run service" from `R` to `f` (foreground). This is more mnemonic and avoids the case-sensitivity confusion.

**Updated Services keybindings:**

| Key | Action |
|---|---|
| `s` | Start service |
| `S` | Stop service |
| `r` | Restart service |
| `f` | Run service (foreground, without auto-start on boot) |
| `c` | Cleanup stale service files |
| `Enter` | View service details in main panel |

---

## Tests for This Milestone

| Test | Type | File | What It Validates |
|---|---|---|---|
| `TestServiceStart` | E2E | `internal/gui/flows/services_test.go` | Start flow + indicator update |
| `TestServiceStop` | E2E | `internal/gui/flows/services_test.go` | Stop flow + indicator update |
| `TestServiceRestart` | E2E | `internal/gui/flows/services_test.go` | Restart flow |
| `TestServiceContextActions` | Unit | `internal/gui/controllers/services_panel_test.go` | Can't stop a stopped service |
| `TestPinToggle` | E2E | `internal/gui/flows/pin_test.go` | Pin/unpin toggles badge (formula + cask) |
| `TestPinCask` | E2E | `internal/gui/flows/pin_test.go` | Pin cask (6.0.0 feature) |
| `TestPinnedExcludedFromUpgrade` | E2E | `internal/gui/flows/pin_test.go` | Pinned skipped in upgrade-all |
| `TestServiceRun` | E2E | `internal/gui/flows/services_test.go` | Run service (foreground, no auto-start) |
| `TestServiceCleanup` | E2E | `internal/gui/flows/services_test.go` | Cleanup stale service files |
| `TestVulnsClean` | E2E | `internal/gui/flows/vulns_test.go` | No vulnerabilities shows green check |
| `TestVulnsFound` | E2E | `internal/gui/flows/vulns_test.go` | Vulnerabilities displayed with CVE info |
| `TestVulnsMissingTap` | E2E | `internal/gui/flows/vulns_test.go` | Missing tap shows install message |
| `TestMissingClean` | E2E | `internal/gui/flows/missing_test.go` | No missing deps shows green check |
| `TestMissingFound` | E2E | `internal/gui/flows/missing_test.go` | Missing deps displayed |
| `TestCleanupDryRun` | E2E | `internal/gui/flows/cleanup_test.go` | Preview shown first |
| `TestCleanupExecute` | E2E | `internal/gui/flows/cleanup_test.go` | Actual cleanup after confirm |
| `TestDoctorClean` | E2E | `internal/gui/flows/doctor_test.go` | Green checkmark on clean system |
| `TestDoctorWarnings` | E2E | `internal/gui/flows/doctor_test.go` | Warnings displayed structured |
| `TestLeavesToggle` | E2E | `internal/gui/flows/leaves_test.go` | Filter toggles on/off |
| `TestAutoremoveDryRun` | E2E | `internal/gui/flows/autoremove_test.go` | Preview shown |
| `TestBrewfileExport` | E2E | `internal/gui/flows/bundle_test.go` | Brewfile created |
| `TestBrewfileInstall` | E2E | `internal/gui/flows/bundle_test.go` | Install from Brewfile |
| `TestBrewfileCleanup` | E2E | `internal/gui/flows/bundle_test.go` | Cleanup with preview |
| `TestBrewfileCheck` | E2E | `internal/gui/flows/bundle_test.go` | Check verifies Brewfile |
| `TestBrewfileList` | E2E | `internal/gui/flows/bundle_test.go` | List displays entries |
| `TestServiceLogTail` | Unit | `internal/gui/presentation/services_test.go` | Log rendering |
| `TestDoctorWarningParsing` | Unit | `internal/brew/doctor_test.go` | Warning blocks parsed correctly |

---

## Definition of Done

- [ ] Services panel fully interactive (start/stop/restart/run/cleanup)
- [ ] Pin/unpin with visual feedback for formulae and casks (6.0.0)
- [ ] Cleanup with dry-run preview
- [ ] Doctor diagnostics displayed
- [ ] Vulnerability check (`brew vulns`) integrated
- [ ] Missing dependencies (`brew missing`) integrated
- [ ] Leaves filter in Formulae panel
- [ ] Autoremove with dry-run preview
- [ ] Brewfile export/install/cleanup/check/list
- [ ] All tests pass
- [ ] All P1 features from the design doc covered
