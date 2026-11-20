# Fix round r2 — slice `v2-03b-sidebar-lists`

> The slice contract and the r1 fix brief still apply. Findings from
> `work/inbox/reviewer-v2-03b-sidebar-lists-r2.md`. **Fix every one.**

| id | severity | fix must achieve |
|---|---|---|
| F-r2-01 | major | `g`/`G` must not steal type-to-search: on the Search panel with the input unfocused, `g`/`G` start a search with that letter, exactly as before this slice. Orchestrator choice: keep `g`/`G` as top/bottom **everywhere except** the Search panel (Home/End work everywhere). Tests: `g` and `G` on unfocused Search → input focused, value `g` / `G`; `g`/`G` on another panel still scroll. Help row says so (e.g. `Home/End (g/G outside Search)`). |
| F-r2-02 | minor | Trust story pluralises: `1 formula`, `1 cask`, `2 formulae`, `0 casks`. Update the test with singular and plural cases. |

## Rules
- One commit per finding, subject naming the id. Stage by name. Gate (CGO_ENABLED=1 for -race).
- Re-run the frame harness; if frames change update `after/` + a `result.md` line.
- Receipt `work/inbox/runner-v2-03b-sidebar-lists-fix-r2.md` (≤ 20 lines), committed once. Worktree clean. No push. Never run brew write commands.
