# Runner fix receipt — v2-03-look-foundation, round r1

One commit per finding (subject names the id); no amend, no push.
- F-r1-01 f6c950a · F-r1-02 02fcbae · F-r1-03 18796ad · F-r1-04 b79b0e3
- F-orch-03 96a5508 (operation outranks badges) · F-orch-01 6736abd · F-orch-02 56e78c0
- F-r1-05 f4f4abf · frames a835b6a · result.md fix 7f3e978 · this receipt

**F-r1-01 mutation (blocker).** The literal `normalizeToHeight(min(vc, mainH))` → `normalizeToHeight(vc)` does not compile (`mainH` becomes unused), so I used the equivalent `normalizeToHeight(vc + 0*mainH)`. Output on the mutated tree:
`review_r3_test.go:232: cursor 29 must be inside [offset 14, offset+14) after the C toggle` — FAIL. Restored: gui.go identical to HEAD, test passes. Fixture: sidebar window 20 rows > expanded main 14; expected offset 16 (was 28).

**Per finding.**
- F-r1-02: search prompt/cursor = accent, text = text role, placeholder/suggestion = muted, blurred text = subtle. `TestSearchInputDrawsFromRoles` (both flavours). Frame: prompt `ESC[37m` → `ESC[38;2;203;166;247m`.
- F-r1-03: operation title and Log marks use Refresh/OK/Error via `withIcon`; the Log takes the set via `SetIcons`. `TestOperationAndLogGlyphsFollowIcons` (nerd, unicode, off).
- F-r1-04: drop loop reserves the gap; `headerBadges` comment states drop order. `TestHeaderKeepsAGapBeforeBadges` sweeps 40–140 columns; reverting the gap fails it at width 51.
- F-orch-03: operation span counted in the drop condition; badges drop first, operation cut only after the last. `TestHeaderOperationOutranksBadges` (80 columns).
- F-orch-01: loaded Status label is empty; loading/error/not-loaded still show. `TestStatusBoxLabelShowsNoCount`.
- F-orch-02: `borderLabel{full, short}`: full if it fits, else short, else left out, never cut. Not loaded → `–`; loading → `…` as the short form (the spinner is one cell and always fits, so it is only a fallback); error `!`; counts have no short form. `TestSidebarLabelsAtEightyColumns`.
- F-r1-05: empty loaded panel shows `0`. Covered by the same 80-column test (Services, ` 0 ─╮`).
- F-r1-05 and F-orch-02 are separate commits (not shared).

**Gate** (code at f4f4abf, unchanged by the later docs commits): `make build` 0 · `make lint` 0 · `CGO_ENABLED=1 go test ./... -race -count=1 -v` 0, PASS 613 / FAIL 0 / SKIP 0 (was 604; +9 new, none deleted) · `make cover-check`: All coverage floors met.

**Frames.** Re-captured with the same harness (real Homebrew, read-only, Unicode set). after/ has 22 .txt frames and 11 .ansi (140x40); each frame is exactly its W×H. Changed rows per frame are listed in result.md, "Fix round r1".

**Tests changed this round (old → new):** `review_r3` `TestCommandLogToggleKeepsCursorVisible` (fixture + offset 28 → 16); `panel_loaded` `TestNeverLoadedPanelSaysNotLoaded` and `look_test` `TestSidebarLabelStates` (read `.full`; `(empty)` → `0`); `look_test` `panel(...)` calls take `borderLabel{}`.

**Decisions for the reviewer.** The Log's task title is left out, not cut, when it does not fit (same rule as the border). `not loaded` short form is `–` in MutedColor. Worktree clean; `internal/gui/dump` and `internal/zzharness` absent; no brew write ran.
