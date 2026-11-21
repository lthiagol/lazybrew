verdict: changes-required

# reviewer lb26-list-filter r1 — `db1513d..6502005`

| AC | Verdict | Evidence |
|---|---|---|
| AC-01 filter state on panelData | ok | Probe (mock data): reload with query keeps cursor on `fzf`; removal clamps to `fx`; empty result no panic, `sel=""`; leaves+query in both orders, clear restores full list, `unfiltered*` nil. |
| AC-02 the `/` input | changes-required | F-r1-01. Rest ok: typing, up/down, Enter, Esc, no-match, `m/total`, click (filter row selects nothing) and wheel map to filtered rows. |
| AC-03 keymap, help, footer, README, flows | ok, README hit by F-r1-01 | Keymap test passes; mutation (drop `keyFilter` from Formulae hints) → `keymap_test.go:358: Formulae: the dispatch answers "/", which its footer and help do not both show`. Flow asserts `Uninstall fzf?` + `uninstall fzf`. |

## Verify
- V1: `GOFLAGS=-p=1 make ci` → exit 0, all floors met (gui 91.2% ≥ 83). `GOOS=darwin go build ./...` ok.
  (A run with `TMPDIR`/`GOTMPDIR` moved off /tmp fails 3 `services_logs_test` cases on path truncation: environment, not this slice.)
- V3: with query `f` + Enter + `j j` → `fx`: `r` "Reinstall fx?", `u` "Upgrade fx?", `i` "Install fx?", content key `1:0:fx` for Info. Actions read `selectedFormula/Cask/nameAt` on the narrowed lists; no raw index into the full list remains (grep).
- V5: frames base `db1513d` vs HEAD, panels 1–7 and `?`, 80x24 + 140x40, dark + light, no filter: only Formulae/Casks footers and the help pages differ.
- V6: no `go func`, no `program.Send`, no go.mod/Makefile/scripts change; staticcheck = only the 5 pre-existing hits; deadcode empty.

## Findings
- **F-r1-01 (medium)** — `internal/gui/panels.go` `switchPanel` (focuses the empty Search box) + `keys.go` `handleKey` search-input branch; README.md:80; test `filter_ui_test.go` `TestSlashStillSearchesOnOtherPanels`.
  What: the documented path "`7`, then `/`" from Formulae types `/` into the Search box. Evidence: probe `2`, `7`, `/` → `searchInput.Value()="/"`, focused. The test checks focus only, not the value.
  Fix must achieve: from Formulae/Casks, `7` then `/` lands on Search with the box focused and **no `/` typed** (e.g. `/` on a focused, empty Search box only focuses it); test asserts `searchInput.Value()==""`. README stays true.
- **F-r1-02 (low)** — `internal/gui/header.go` `headerCount`, the Formulae/Casks cases.
  What: the header badge counts the narrowed list. Evidence: 5 formulae, query `rip` → header `2 formulae`; 60 formulae, query `pkg1` at 80x24 → `10 formulae`.
  Fix must achieve: the badge shows the installed total while a query (and leaves) narrows the list (`fullFormulae()`/`fullCasks()`), with a test.
- **F-r1-03 (info, no fix)** — AC-02 has two commits, and `381ac43` lands after AC-03 (`c298ca6`): not one commit per AC in AC order. History is not rewritten; record only.

## Seen, not judged
- 80x24: the active Formulae box has 2 rows, so with the filter row it lists 1 row; an inactive box with a query shows only `/ q`. Layout, not a defect of this slice.
