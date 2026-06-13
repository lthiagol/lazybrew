# Lazybrew — Manual Smoke Checklist

> **Status:** Ready for M20 verification. Run before M22 release tag.  
> **Requires:** Real Homebrew installation (macOS or Linux).  
> **Duration:** ~15 minutes

---

## Environment

| Field | Value |
|---|---|
| Date | |
| Tester | |
| lazybrew version | `lazybrew --version` |
| brew version | `brew --version` |
| Terminal size | cols × rows |
| OS | |

---

## Launch & Navigation

- [ ] `lazybrew` starts without error
- [ ] All 7 sidebar panels populate within 10s
- [ ] Tab / Shift+Tab cycles panels
- [ ] Keys 1–7 jump to correct panel
- [ ] `[` / `]` switch tabs within panel
- [ ] `?` help opens; Esc closes

---

## Data Display (M20)

- [ ] Formulae **Info** tab shows package details (not duplicate sidebar list)
- [ ] Select formula A → Deps tab → j/k to B → Deps updates for B
- [ ] Outdated panel: empty state OR list with correct items
- [ ] Empty formulae edge case (if testable): correct message

---

## Mutations (M19 TaskManager)

- [ ] Search `/` → install a small package (`i`) — progress modal streams output
- [ ] Cancel long operation (Esc in progress modal) — UI recovers, not stuck
- [ ] Second operation while first runs — queued or rejected with toast
- [ ] Uninstall with confirmation (`x`) — dependency warning if applicable
- [ ] Pin/unpin (`p`) — correct behavior for pinned vs unpinned

---

## Outdated & Batch (M20.3)

- [ ] Space toggles selection indicator on outdated items
- [ ] `u` upgrades single item
- [ ] Multi-select + `u` upgrades selected sequentially

---

## Status Panel Actions

- [ ] `d` doctor — output in modal
- [ ] `m` missing — output lines shown
- [ ] `v` vulns — runs (or clear error if tap missing)
- [ ] `R` refresh — panels reload

---

## Config (M20.8)

- [ ] Custom `brew.path` in config (if tested)
- [ ] `auto_refresh_seconds: 30` triggers refresh (observe ~30s)

---

## Terminal Size (M20.7)

- [ ] Resize to 79×24 — warning shown
- [ ] Resize to 80×24 — warning gone

---

## Exit

- [ ] `q` quits cleanly
- [ ] No panic in terminal after exit

---

## Result

| | |
|---|---|
| **Pass / Fail** | |
| **Blocking issues** | |
| **Notes** | |
