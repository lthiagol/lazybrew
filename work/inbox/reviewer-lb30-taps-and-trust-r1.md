verdict: changes-required

# Review — lb30-taps-and-trust r1 (7c878ea..9914491, base a0844a8)

| AC | Verdict | Evidence |
|---|---|---|
| AC-01 | ok | Real `brew trust --json=v1` (2 taps, 3 formulae, empty casks/commands) parses: taps → `tap` with Tap=self, `lthiagol/tap/frond` → Tap `lthiagol/tap`; empty lists → `[]`; a missing list → its kind absent, no error; `{"taps":[` and an object list → error. Parser reverted to the old shape: `TestTrustReaderParsesRealShape`, `…TapDerivation` fail. Writers `untrust --formula/--cask/--command x` + `untrust` invalidation in `writes_invalidate_test`. |
| AC-02 | ok (code) | Own frame diff base→HEAD, 9 themes × 80x24/140x40 × 7 panels × tabs/help (1680 frames): only the `Taps & Trust` title (key `5`, icon and count fit at 80x24), the item row `✔ octo/tools/zeta … formula`, and the help `t` line differ. Tap rows, tabs and footers unchanged. Item row: `Taps & Trust › Info`, single `Info` tab. |
| AC-03 | changes-required | Own teatest (mock runner, no `R`): formula/cask/command each → confirm "Homebrew will not load <name>…" → exactly `untrust --formula|--cask|--command <name>`, trust list re-read on its own, row gone, cursor on the same index / nearest. Failing untrust: row kept, toast `Untrust: octo/tools/zeta: boom untrust`. `x` toast, no untap. Footer test is vacuous: F-r1-01. |

## Gate

- `GOFLAGS=-p=1 make ci` → exit 0, `All coverage floors met` (gui race tests ok). A first run with `GOTMPDIR` set
  failed `TestLogsTabShowsLastLinesOfStdout`/`TestLogsEmptyStates` (long temp path is
  truncated in the header); same failure at base a0844a8 with `GOTMPDIR`, passes without:
  environment, not this slice.
- `GOOS=darwin go build ./...` ok. `deadcode ./cmd/...` empty. `staticcheck ./...` identical
  to base (keymap_test SA1019 ×2, launch_check_test SA4006 ×3).
- Floors unchanged; no `program.Send`; the only new goroutines are in `runConcurrently`
  inside the read `tea.Cmd` (existing helper). Writes go through `m.enqueue(errTask…)`.
- README has no panel/Taps wording (checked); unchanged is right.

## Findings

F-r1-01 (medium) — internal/gui/trusted_test.go:353 `TestTapsFooterOnTrustedItemLeadsWithUntrust`,
internal/gui/keymap_test.go:322 — the item-row footer is not pinned by any test — removing the
`selectedTrusted()` branch of `tapsHints` (keybindings.go:168-170; footer becomes `a add tap · y copy`)
leaves `go test ./internal/gui/` green: the footer test greps the whole `View()` for "untrust",
which the Info body's own `t untrust` line supplies; the keymap test only checks footer hints have
help entries — fix must achieve: a test that asserts on the footer row itself (`t untrust` leads,
no `x`/`r`/`T`) and fails with the branch removed; the brief's V6 "keymap test fails if the
item-row hints are dropped" holds.

F-r1-02 (low, process; no history rewrite) — commits d8655ab/d036d12/b750e10 — AC-02's commit
carries AC-03's actions (t/x/r/T, untrust confirm, footer); AC-03's commit holds a cache-group
fix to AC-02's reload plus the flow; b750e10 is an AC-02 fix after AC-03. At d8655ab, `R` did not
re-read the trust list (receipt's own failing run) — fix must achieve: nothing to rewrite;
fix-round commits carry the finding id in the subject.

F-r1-03 (low) — work/inbox/runner-lb30-taps-and-trust.md — 80 lines, brief caps it at 45 —
fix must achieve: the fix-round receipt within its 30-line cap.

## Notes (no action)

- `TestSidebarLabelsAtEightyColumns` now matches `Taps & Trust` without `5 ⊕`; in real 80x24
  frames the title keeps `5 ⊕` and the count (`╭─ 5 ⊕ Taps & Trust ─── 4 ─╮`).
- `Q-lb30-dashboard-taps` is open for the orchestrator.
