# Fix round r1 — slice `v2-03-look-foundation`

> The slice contract `work/briefs/v2-03-look-foundation.md` still applies in
> full. Findings from `work/inbox/reviewer-v2-03-look-foundation-r1.md` plus
> three orchestrator rulings (`work/questions.md`, "Orchestrator rulings on
> Q-v2-03-a..e"). **Fix every one.**

| id | severity | fix must achieve |
|---|---|---|
| F-r1-01 | blocker | `TestCommandLogToggleKeepsCursorVisible` guards R3-4 again: a fixture where the **main** window is the binding clamp (sidebar visibleRows larger than the expanded contentH), asserting the offset only the main clamp gives. Prove it: mutate `normalizeToHeight(min(vc, mainH))` → `normalizeToHeight(vc)` and show the test fails (record the output in the receipt), then restore. |
| F-r1-02 | minor | Search textinput prompt / text / placeholder / cursor styled from theme roles (Accent, Text, Muted); readable on Mocha and Latte. Test that the styles come from the roles. |
| F-r1-03 | minor | Operation overlay title and Log status glyphs (`⟳ ✓ ✗`) come from the icon set (Refresh/OK/Error) and honour `show_icons: false`. |
| F-r1-04 | minor | Header drop loop reserves the gap between left and right (test at exact fit: never `…g1f335ae125 formulae`); the `headerBadges` comment matches the drop order. Land with F-orch-03. |
| F-r1-05 | minor | A loaded empty panel's label is never cut: show `0` (counts are always shown in full). Test at 80 columns. Land with F-orch-02. |
| F-orch-01 | ruling Q-v2-03-a | Status shows **no** count in its border label (loaded state shows nothing, or only the loading/error forms). Test. |
| F-orch-02 | ruling Q-v2-03-c | When the full label does not fit the border: `not loaded` → `–` (muted), loading → `…`, error → `!`; counts always in full. Test at 80 columns for each state. |
| F-orch-03 | ruling Q-v2-03-d | Header: a running operation outranks badges — badges drop first, then the operation is cut with `…`; the app name always stays. Test at 80 columns with an operation running. |

## Rules
- One commit per finding (F-r1-04 may share a commit with F-orch-03, F-r1-05 with F-orch-02 — say so), subject naming the id(s). Stage by name.
- Gate: `make build`, `make lint`, `CGO_ENABLED=1 go test ./... -race -count=1`, `make cover-check`.
- Re-run the frame harness; update `after/` (incl. `.ansi` 140x40) and add a line per changed frame to `result.md`. Frames exact size.
- Receipt `work/inbox/runner-v2-03-look-foundation-fix-r1.md` (≤ 30 lines), committed once. Worktree clean, `internal/gui/dump` absent. No push. Never run brew write commands.
