# runner lb29-footer-redesign — receipt

Commits, one per AC: `a918265` AC-01 · `0d87daa` AC-02 · `5379050` AC-03.
The 140-column sample is 167 cells and cannot fit; the Narrow-widths rules were followed (Q-lb29-sample-width).

## AC-01 renderer
- Tests first (`footer_test.go`). Red: `undefined: fitFooter, renderFooter, footerRow, footerDropOrder`.
- Green: `go test ./internal/gui/ -run TestFooter -v` → PASS: FitsAtEveryWidth (60–200), DropsInTheStatedOrder, SeparatorsAreMuted (dark + light), NeverClipsAPanel.
- Expectations changed: look_test (modal footer `·`; Status width 200 → 240), tap_actions_test (`·`).

## AC-02 key tables
- Tests first: TestFooterShowsTheArrowHintOnEveryPanel, TestHelpKeepsTheKeysTheFooterDropped. Red: `undefined: keyArrows`.
- Mutations, restored: no keyArrows → TestPanelKeysAreOnFooterAndHelp fails on up/down/left/right; no keyTabs help line → fails on h, l.
- footerExempt[Search] lost left/right; services_logs_test, tap_actions_test updated.

## AC-03 help, README, flows
- Help: `j/k up/down`, `h/l left/right` (ASCII; TestHelpSectionsLineUp counts bytes).
- flows/footer_test.go: TestFooterFlowStatusAt140, TestFooterFlowFormulaeAt140 — PASS on the final View(); TestFooterNeverClipsAPanel (7 panels × 80/100/140) — PASS.

## View() diff, base 8af0ef1 vs HEAD (plain text, 78 frames: 13 scenes × 3 sizes × 2 themes)
- Last row changed in 72; the 6 help frames keep `?/esc close`.
- Body rows also changed in 18: help (6), Tap Info and Services Info key lines (12). They call footerLine, so they take `·` (Q-lb29-info-keylines). The brief's "only the footer row differs" does not hold.

## 80-column rows as drawn (dark; 100x30 in the same dumps)
```
Status    R refresh · c cleanup · d doctor  │  ←↑↓→ navigate · / search · q quit · ? more
Formulae  i install · x remove · r reinstall  │  ←↑↓→ navigate · q quit · ? more
Casks     i install · x remove · X zap  │  ←↑↓→ navigate · / filter · q quit · ? more
Outdated  u upgrade · U upgrade all · Space select  │  ←↑↓→ navigate · q quit · ? more
Taps      t trust · x untap · r repair  │  ←↑↓→ navigate · / search · q quit · ? more
Services  S stop · r restart · c cleanup  │  ←↑↓→ navigate · / search · q quit · ? more
Logs tab  S stop · r restart · c cleanup  │  ←↑↓→ navigate · / search · q quit · ? more
Search    Enter search · i install · x uninstall  │  ←↑↓→ navigate · q quit · ? more
```
140 Formulae: `i install · x remove · r reinstall · u upgrade  │  ←↑↓→ navigate · Tab next · 1-7 jump · / filter · R refresh · C log · q quit · ? more`

## Gate (final tree)
- `GOFLAGS=-p=1 make ci`: exit 0 (fmt, vet, test -race, cover-check all floors met, build).
- `GOOS=darwin go build ./...` and `go vet ./...` OK.
- deadcode clean (it flagged `footerRow.all`; moved to the test file). staticcheck: 5 diagnostics, same as base. No brew commands; TUI not launched.

## Notes
- Modal dialog's own key line (`modal/frame.go`) untouched; sample's `x uninstall` kept as `x remove`. Seen, not done: Status shows `R refresh` twice.
- Don't set TMPDIR to a long path: TestLogsTabShowsLastLinesOfStdout and TestLogsEmptyStates depend on path length.
