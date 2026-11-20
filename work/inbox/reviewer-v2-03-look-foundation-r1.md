verdict: changes-required

# Reviewer r1 — v2-03-look-foundation

Reviewed `d3289ed..09426b9` (a6f22ab c569ee0 85f5e3f 0d4593e e56d97c d58fb96 4aaa3b0 09426b9).
Q-v2-03-a/c/d are F-orch-01..03 (not repeated here); b deferred; e accepted.

| Check | Result | Evidence |
|---|---|---|
| V1 gate | pass | build, lint (vet), `CGO_ENABLED=1 go test ./... -race -count=1` 9 ok, cover-check "All coverage floors met" (brew 72.9, pres 96.2, gui 75.1, modal 50.5) |
| per-commit | pass | each of the 6 code commits: `go build/vet/test ./...` exit 0, 9 ok |
| V2 frames | pass | harness re-run (real brew, read-only): 22 `.txt` + 11 `.ansi` byte-identical to `after/`; cell count: 22/22 frames exactly W×H; `internal/gui/dump` removed, tree clean |
| V3 AC-01 | pass | all 26×2 hex match official Mocha/Latte; raw-hex grep empty; BgColor hack gone; View sets bg/fg (gui.go:966) |
| V4 AC-02 | pass | default/env/off matrix tested (icons_test.go:61); PUA test real; `show_icons` off: `╭─ 1 Status ` no stray space; README + DESIGN document both keys |
| V5 AC-03 | pass | `panel()` (box.go:16) used by sidebar, main, Log; renderBox gone; labels visible at 80×24; mouse tests updated with header offset |
| V6 AC-04 | pass | counts omitted until loaded, outdated only >0; 80×24 header is 80 cells; op in middle (TestHeaderOperationInTheMiddle) |
| V7 AC-05 | pass | one row; `? more` iff cut (80/140 cut, 200 not); modal + help footers; `panelHints` still the panel source |
| V8 AC-06 | pass | before/ = v2-02 after/ (diff -r identical); every frame change matches result.md; no content restructure, modal drawn unchanged |
| V9 invariants | **fail** | tests 562 → 604 PASS, 0 skip; no brew/keybinding/go.mod change; but one test weakened (F-r1-01) |

## Findings

F-r1-01 (blocker) — internal/gui/review_r3_test.go:200-233 — TestCommandLogToggleKeepsCursorVisible
no longer guards R3-4. After the switchPanel fixture the active sidebar box (2 rows) binds the offset,
so the expected 28 is reached without the main-height clamp. — Evidence: in a scratch copy, mutating
gui.go:738 `normalizeToHeight(min(vc, mainH))` → `normalizeToHeight(vc)` leaves `go test ./internal/gui/...`
all ok at HEAD; the same mutation at d3289ed fails this test (`cursor 29 must be inside [offset 14, offset+13)`).
— Fix: keep a fixture where the main window is the binding one (e.g. sidebar visibleRows set larger than the
expanded contentH) and assert the offset that only the main clamp gives, so the mutation fails the suite.

F-r1-02 (minor) — internal/gui/gui.go:125 — 09-search: the search input keeps bubbles' default styles,
off the palette: prompt `\e[37m` (white), cursor `\e[7;37m`, placeholder `\e[38;5;240m`
(after/140x40/09-search.ansi row 4). Now that the app paints its canvas, on Latte (#eff1f5) the white
prompt is unreadable. — Fix: style prompt/text/placeholder/cursor from roles (Accent, Text, Muted).

F-r1-03 (minor) — internal/gui/render.go:510, internal/gui/command_log.go:113-123 — glyphs still literal:
the operation overlay title `⟳ ` and the Log's `⟳ ✓ ✗` ignore the icon set and `show_icons: false`, so with
the default Nerd set the header shows the Nerd refresh while the overlay shows Unicode `⟳`. — Fix: draw them
from `m.icons` (Refresh/OK/Error) via `withIcon`, or record the deferral (v2-04) in work/backlog.md.

F-r1-04 (minor) — internal/gui/header.go:31, :58-60 — the drop loop allows left+badges == width, leaving
`room` 0 and the badges glued to the version (`…g1f335ae125 formulae`); and the headerBadges comment says
"most important first, so the ones dropped … are the least useful" while the code drops badges[0].
— Fix: reserve `headerGap` between left and right in the drop condition (test at exact fit); make the
comment match the order. Can land together with F-orch-03.

F-r1-05 (minor) — internal/gui/render.go:50 — a loaded empty panel's label `(empty)` is cut to `(e…` at
80×24 (after/80x24/05-outdated.txt:11, 09-search, 11-deps-bash). The Q-v2-03-c ruling lists not loaded /
loading / error / counts but not empty. — Fix: give empty a form that is never cut (e.g. the count `0`,
shown in full) as part of F-orch-02, with a test at 80 columns.

## Notes (no action)

- AC-02 is two commits (c569ee0 + 85f5e3f follow-up); acknowledged in the receipt, not amendable.
- Sidebar rows still truncate with ASCII `...` (renderSidebarContent), the primitive with `…`; content, v2-03b.
- Toast on the Log's bottom border at 80×24 is pre-existing (v2-04).
