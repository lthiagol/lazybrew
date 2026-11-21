# Receipt — runner lb30-taps-and-trust (LB-28)

## Commits (AC order, then one follow-up)

- AC-01 `7c878ea` — trust list reads the real shape; untrust writers
- AC-02 `d8655ab` — `Taps & Trust` pane: rows, Info tab, reloads, item actions
- AC-03 `d036d12` — taps refresh drops the trust list too; teatest flow
- follow-up `b750e10` — race fix in an AC-02 test helper (found by `make ci`)

Deviations, not amended (the brief forbids amend):
- The item actions (t/x/r/T, untrust confirm, footer hints) went in with AC-02, since
  they share the pane's code. AC-03 holds the cache fix and the flow.
- `b750e10` sits after AC-03 in history.

## AC-01 — brew layer

- Before: `undefined: w.UntrustFormula` (build failed). After: `go test ./internal/brew/` ok.
- Real shape, read-only `brew trust --json=v1`: `{"taps":[…],"formulae":[…],"casks":[…],"commands":[…]}`,
  string lists. Fixture `internal/brew/testdata/trust_v1.json` is anonymised.
- Tests: `TestTrustReaderParsesRealShape`, `TestTrustReaderTapDerivation` (3-part, 2-part, 1-part),
  `TestTrustReaderEmptyLists`, `TestTrustReaderMalformedJSONIsError`, `TestTrustWriterUntrustArgs`,
  `writes_invalidate_test` rows for untrust formula/cask/command.
- Choice: a tap entry's `Tap` is its own name; a non-tap 2-part name has `Tap` "".

## AC-02 — the pane

- Title `Taps & Trust` via `PanelID.Title()`; `String()` stays `Taps` (the key-dispatch test maps on it).
  Help page title too.
- `TestTapsPanelReadListsTrustedItemsBelowTheTaps` runs the real fetch path (140x40): taps, then
  formulae, casks, commands, alphabetical.
- `TestTapsReadFailureKeepsTheTapsAndToastsOnce`: taps usable, one toast, repeat failure silent.
- `TestTrustedItemCursorStaysOnItsNameAcrossReloads`, `…ClampsWhenItsRowIsGone`.
- `TestTrustedItemRowHasTheInfoTabAloneAndItsFacts` (140x40): Info only, h/l inert, facts.
- `TestTapsRowsFitAtEightyByTwentyFour`; theme frame `trusted item selected` runs in every theme at 80x24.
- Tests were written before the code; the first run failed to build (`undefined: trustedRows`).

## AC-03 — actions, footer, flow

- `TestTapsTrustedItemFlow` (teatest): `5`, `j`, `x` → toast `x untaps a tap; t untrusts octo/tools/zeta`,
  no untap run; `t` → confirm; `y` → `brew untrust --formula octo/tools/zeta` recorded;
  `R` → reload; final `View()`: row gone, cursor on `octo/tools`.
- Before the cache fix the flow failed: `taps_trust_test.go:121: the untrusted formula is still on the Taps list`.
  `R` kept the cached trust list. `TestTapsRefreshReadsTheTrustListAgain` failed with
  `trust list ran 1 times, want 2`. Fixed: `taps` cache group now includes `KeyTrustList`.
- Footer on an item: `t untrust` leads. Help `t` line: "Trust or untrust a tap or trusted item".
- README has no panel or Taps wording; left unchanged.

## Test expectations changed (named)

- `look_test.go` `TestSidebarLabelsAtEightyColumns`: `5 ⊕ Taps` → `Taps & Trust`; at this 24-cell
  width the icon gives way, the count still shows in full.
- `keymap_test.go` `helpOn` matches the help page on `Title()` (was `String()`).
- `cached_test.go`: trust fixture uses the real shape. `trust_test.go`: old list test replaced.
- `theme_frames_test.go`: one screen added. No golden frame files exist; no test deleted or skipped.

## Gate

- `GOFLAGS=-p=1 make ci` → exit 0: fmt-check, vet, `-race` all packages ok, `All coverage floors met`
  (gui 92.0 ≥ 83, brew 89.7 ≥ 78).
- `GOOS=darwin go build ./...` ok. `deadcode ./cmd/...`: no output.
- `staticcheck ./...`: only warnings that predate this slice (keymap_test `ParseDir` ×2, launch_check_test SA4006 ×3).
- Brew: only read-only `brew trust --json=v1`. No brew write command. TUI not launched.
- The first `make ci` failed on a race in my AC-02 test helper (concurrent appends to a call log); fixed in `b750e10`.

## Choices

- Item `t` does not depend on the tap-info read (`trustUnknown`).
- A trusted item whose tap is not on the list, or has no tap part: "Tap trusted: unknown".
- Cursor on a removed item stays at its index; normalize clamps it to the nearest row.
- `T` with no tap row: toast "T acts on a tap: select one first" (it used to open a menu with a blank name).

## Questions filed

- `Q-lb30-dashboard-taps` (work/questions.md): the Status dashboard card still says `Taps`.

## Seen, not done

- Real-terminal feel and a live `brew untrust`: human smoke, not run here.
- ⊕ icon at the default 28-cell sidebar: not checked in a real terminal.
- `wip` is not pushed by this lane.

## Fix r2

- F-r1-01 (`249eb6e`): `TestTapsFooterRowOnTrustedItemLeadsWithUntrust` reads the footer row (last row of `View`) and wants the panel group `t untrust · a add tap · y copy`. `TestTapsItemRowFooterIsExact` pins the item-row panel hints and checks each is answered by the Taps dispatch.
- Mutation (`selectedTrusted()` branch of `tapsHints` removed, then restored): both fail. Footer row `a add tap · y copy`, want `t untrust · a add tap · y copy`; item hints `[a y]`, want `[t a y]`. `TestPanelKeysAreOnFooterAndHelp` stays green, since tap rows still show `t trust`.
- Gate: `GOFLAGS=-p=1 make ci` exit 0, `All coverage floors met` (gui 92.0% ≥ 83%), `go vet` clean.
- F-r1-03: this section is 8 lines; the receipt above is not edited.
