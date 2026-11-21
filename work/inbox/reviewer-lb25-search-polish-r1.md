verdict: ok

# Review — lb25-search-polish, round r1

Range: `9687acb..b8bb20a` (AC-01 `b2885f6`, AC-02 `4c14fbf`, AC-03 `897f98f`, fix r1 `51bd21e`).

| AC | Verdict | Evidence |
|---|---|---|
| AC-01 ✔ installed mark | ok | `Installed: "✔"` in Nerd and Unicode sets (`style/icons.go`); icons-off frames draw no mark. Own frame diff base→HEAD: Formulae/Casks rows ✓/nerd glyph → ✔, alignment unchanged; outdated `↑`, trusted tap `✓`, service glyphs unchanged. |
| AC-02 Search rows | ok | Own `View()` harness at 80x24 and 140x40, dark and light, Unicode/Nerd/off: `│▸ ✔ alpha-cask … cask│`, `│    beta-formula … formula│`, long name `a-really-long-search-result-name-o… formula`, both-kinds `✔ both … formula` / `  both … cask`. Names start in one column, types end in one column, ✔ only on installed rows. Focused selected row carries the bg SGR. `SearchMarksMsg` with `Installed=false` clears the ✔. `i` opens `Install beta-formula?` (args path untouched: `selectedSearchRow`, `confirmMutation` not in the diff). |
| AC-03 drop `I` | ok | `case "I"` and `sInstallCap` gone; footer `i install  x uninstall` (no `I install`); help shows only `i Install selected`. Mutation (re-apply old keys.go/keybindings.go) fails 5 tests incl. `TestSearchDoesNotAnswerCapitalI` and `TestSearchInstallKeyIsLowercaseI`. |
| Fix r1 Completed row | ok | `operation.go:152` uses `m.icons.Installed`. Overlay frames, both themes, Unicode and Nerd: Completed row and every finished step draw U+2714 in the same SGR (dark `38;2;166;227;161`, light `38;2;64;160;43`); `InstalledBadge == SuccessText`. Reverting line 152 fails `TestCompletedRowMatchesTheStepTicks` at both sizes. |

## Gate (own runs)
- `GOFLAGS=-p=1 make ci`: exit 0, "All coverage floors met" (`-p=1`: /tmp tmpfs at 93%).
- `GOOS=darwin go build ./...`: ok.
- staticcheck `./...`: 5 findings at base, same 5 at HEAD (line shift only). deadcode `./cmd/...`: empty at both.
- No Makefile/scripts change; floors untouched. One commit per AC, in order.

## V5 frame diff (base vs HEAD, every panel, 2 sizes × 2 themes × 3 icon sets)
Only lines that differ: rows with the installed glyph (Formulae/Casks sidebar, the Info `Status: ✔ installed` word, dep/step ticks), Search rows, the Search footer, and the overlay Completed row. Nothing else.

## Findings
- F-r1-01 (info) — `work/inbox/runner-lb25-search-polish.md` — receipt is 68 lines; cap is 40 (+6 for Fix r1). — `wc -l` → 68. — No fix: receipts are not edited after commit; noted for the next brief.

## Notes for the orchestrator
- `Q-lb25-ok-uses` is open: dashboard "ready to brew" (`dashboard.go:111`, `:198`) and command-log success (`command_log.go:153`) still draw `✓` via `Icons.OK`. Within the fix brief; needs a ruling if "one check mark across the interface" should reach them.
- Real-terminal width of ✔ not checked (no TUI launch); `lipgloss.Width == 1` asserted in tests.
