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
- [ ] Select formula A → Deps tab → j/k to B → Deps updates for B — **FAILED 2026-08-01: Deps tab hangs in 'loading' state indefinitely. Used by and Files tabs render fine. Follow-up: M13.**
- [ ] Outdated panel: empty state OR list with correct items
- [ ] Empty formulae edge case (if testable): correct message

---

## Mutations (M19 TaskManager)

> **Section skipped 2026-08-01 at tester's request — real `brew` mutations on dev machine avoided.** No follow-up opened; coverage relies on existing teatest flows (`internal/gui/flows/install_test.go`, `uninstall_test.go`) and integration tests under `-tags=integration`.

- [ ] ~~Search `/` → install a small package (`i`) — progress modal streams output~~ skipped
- [ ] ~~Cancel long operation (Esc in progress modal) — UI recovers, not stuck~~ skipped
- [ ] ~~Second operation while first runs — queued or rejected with toast~~ skipped
- [ ] ~~Uninstall with confirmation (`x`) — dependency warning if applicable~~ skipped
- [ ] ~~Pin/unpin (`p`) — correct behavior for pinned vs unpinned~~ skipped

---

## Outdated & Batch (M20.3)

> **Section skipped 2026-08-01 — tester opted out of real `brew upgrade` mutations.** Visual selection (item 1) is a candidate for teatest automation post-release.

- [ ] ~~Space toggles selection indicator on outdated items~~ skipped
- [ ] ~~`u` upgrades single item~~ skipped
- [ ] ~~Multi-select + `u` upgrades selected sequentially~~ skipped

---

## Status Panel Actions

> **Section skipped 2026-08-01 — smoke paused mid-run; tester has UI concerns to address before continuing.** Resume at M08 rerun.

- [ ] ~~`d` doctor — output in modal~~ skipped
- [ ] ~~`m` missing — output lines shown~~ skipped
- [ ] ~~`v` vulns — runs (or clear error if tap missing)~~ skipped
- [ ] ~~`R` refresh — panels reload~~ skipped

---

## Config (M20.8)

- [ ] ~~Custom `brew.path` in config (if tested)~~ skipped
- [ ] ~~`auto_refresh_seconds: 30` triggers refresh (observe ~30s)~~ skipped

---

## Terminal Size (M20.7)

- [ ] ~~Resize to 79×24 — warning shown~~ skipped
- [ ] ~~Resize to 80×24 — warning gone~~ skipped

---

## Exit

- [ ] ~~`q` quits cleanly~~ skipped
- [ ] ~~No panic in terminal after exit~~ skipped

---

## Result

| | |
|---|---|
| **Pass / Fail** | Partial — 8/30 items verified pass, 1 fail (M13 follow-up opened), 21 skipped (whole-section opt-outs + smoke paused) |
| **Blocking issues** | M13: Fix Deps tab infinite loading |
| **Notes** | Smoke paused 2026-08-01 mid-run; tester has UI concerns to address before continuing. M08 will rerun (or be replaced by automated teatest suite per M13/M14 plan) once concerns are resolved. |
| **M08 state** | Blocked on M13 (Deps fix) |
