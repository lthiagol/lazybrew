verdict: changes-required

# Reviewer r2 — v2-03-look-foundation

Reviewed `91863ae..7452e79` (fix commits f6c950a 02fcbae 18796ad b79b0e3 96a5508 6736abd 56e78c0 f4f4abf,
frames a835b6a, result.md 7f3e978, receipt 7452e79) on top of the r1 scope.

| Check | Result | Evidence |
|---|---|---|
| V1 gate | pass | build, lint, `CGO_ENABLED=1 go test ./... -race -count=1` 9 ok, cover-check "All coverage floors met" (gui 75.3) |
| per-commit | pass | each of the 8 fix commits: `go build/vet/test ./...` exit 0, 9 ok |
| V2 frames | pass | harness re-run (real brew, read-only): 22 `.txt` + 11 `.ansi` byte-identical to `after/`; 22/22 exactly W×H; dump dir removed, tree clean |
| V8 result.md | pass | "Fix round r1" section explains every changed frame (Status count gone, `–`, `0`, prompt colour) |
| V9 invariants | pass | 613 PASS (was 604), 0 skip; replaced assertions equal strength; no brew/keybinding/go.mod change; one commit per finding |
| ANSI 09-search | pass | off-palette `37`, `7;37`, `38;5;240` gone; prompt mauve, placeholder overlay0, cursor reverse mauve |

## Earlier findings

| id | status | evidence |
|---|---|---|
| F-r1-01 | fixed | my r1 mutation (`normalizeToHeight(vc); _ = mainH`, gui.go:741) on a HEAD copy: `review_r3_test.go:232: cursor 29 must be inside [offset 14, offset+14)` — FAIL; unmutated passes |
| F-r1-02 | fixed | `searchStyles()` gui.go:1218 from roles; TestSearchInputDrawsFromRoles; frame colours above |
| F-r1-03 | fixed | render.go:518 and command_log.go:127-141 use `m.icons`/`cl.icons` via `withIcon`; TestOperationAndLogGlyphsFollowIcons (nerd/unicode/off) |
| F-r1-04 | fixed | header.go:33 drop condition counts the gap (`badgesSpan`); comment matches drop order; TestHeaderKeepsAGapBeforeBadges |
| F-r1-05 | fixed | empty loaded → `0` (after/80x24/05-outdated: `4 ↑ Outdated ─── 0 ─╮`) |
| F-orch-01 | fixed | Status label empty when loaded (all frames `╭─ 1 ◉ Status ───…╮`); TestStatusBoxLabelShowsNoCount |
| F-orch-02 | **partly** | `–`/`…`/`!` short forms work (after/80x24 Casks/Services/Search `–`); counts are not always shown — F-r2-01 |
| F-orch-03 | fixed | probe at 80 cols with all badges + long op: `◆ lazybrew  Homebrew 4.5.2-123-gabcdef  ⟳ Upgrade: a-very-long-formula-name-for…` — badges dropped, op cut, name kept; TestHeaderOperationOutranksBadges |

## Findings

F-r2-01 (major) — internal/gui/box.go:30-39, internal/gui/render.go:57-60 — F-orch-02 requires counts
"always shown in full", but a count wider than the room left after the title has no short form and is
**left out**. At 80 columns the box is 24 cells and `2 ∘ Formulae` leaves 3 cells, so any install with
≥ 1000 formulae shows no count at all. — Evidence: probe test in a scratch copy, 80×24, Formulae loaded
with 1234 items: border renders `"╭─ 2 ∘ Formulae ───────╮"` (no count; Q-v2-small state lost). Same for
`6 ◎ Services` / `4 ↑ Outdated` at ≥ 1000. TestSidebarLabelsAtEightyColumns only uses 1-digit counts.
— Fix: a count always appears whole; when it does not fit, the title yields (e.g. drop the icon, then cut
the name with `…`) rather than the count. Test at 80 columns with a 4- and 5-digit Formulae count,
asserting the full number in the border and the row exactly 24 cells.

## Notes (no action)

- box.go:11 and render.go:40 comments were edited without reflowing (box.go:11 is a 110-column line); tidy when
  touching F-r2-01.
- The Log's task label is left out, not cut, when it does not fit (runner's stated decision); consistent
  with the border rule, acceptable.
