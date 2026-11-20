# Fix round r1 — slice `v2-02-layout-bugs`

> The slice contract `work/briefs/v2-02-layout-bugs.md` still applies in full.
> Findings from `work/inbox/reviewer-v2-02-layout-bugs-r1.md`. **Fix every one.**

| id | severity | fix must achieve |
|---|---|---|
| F-r1-01 | major | The main-area list (`renderList` panels: Status, Services, Outdated, Search, …) scrolls within **its own** height: a render-time window derived from `selected` and the main content height (as the Formulae table already does), independent of the sidebar's `p.offset`/`visibleCount`. Rendering must not write model state. At 80×24 with 30 rows, after 6× `j` the main list shows rows 00..N with the cursor on row-06 (rows above the cursor visible while they fit). Same at 140×40 past the sidebar window. The sidebar keeps its own scrolling as today. Test at 80×24 and 140×40: move the cursor down, assert earlier rows **and** the cursor row are drawn; move past the main height, assert the cursor row is still drawn. |
| F-r1-02 | minor | `TestModalsShowContentOnFirstFrame` asserts, per case, a string only the modal draws: the confirm's question (`This cannot be undone.` or the real question text), the trust menu's own text/item (not the tab name `Trust`), the tap-add input's prompt, the Brewfile menu's item. |

## Rules
- One commit per finding, in order, subject naming the id. Stage by name.
- Gate: `make build`, `make lint`, `CGO_ENABLED=1 go test ./... -race -count=1`, `make cover-check`.
- Re-run the frame harness. If any `after/` frame changes because of F-r1-01, update `after/` **and** add one line per changed frame to `result.md` saying why. Frames must stay exact size.
- Receipt `work/inbox/runner-v2-02-layout-bugs-fix-r1.md` (≤ 30 lines), committed once. Worktree clean, `internal/gui/dump` absent. Do not push.
- Never run brew write commands (`install`, `uninstall`, `services start|stop|...`).
