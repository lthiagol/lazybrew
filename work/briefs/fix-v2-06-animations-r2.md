# Fix round r2 — slice `v2-06-animations`

> The slice contract and the r1 fix brief still apply. Findings from
> `work/inbox/reviewer-v2-06-animations-r2.md`. **Fix every one.**

| id | severity | fix must achieve |
|---|---|---|
| F-r2-01 | minor | **Orchestrator choice:** blend **each backdrop cell from its own foreground colour** to the muted role, so no cell is ever brighter than in the frame it replaces — in **both** themes (Mocha and Latte, where the luminance order differs). If a per-cell blend is not feasible in the compositor, drop the fade and dim instantly (keep the modal slide) — state which in the receipt. Tests, run under **both** themes: for a backdrop cell that was a border and one that was text, fade step 1 is no brighter than the replaced frame, and steps never brighten. Fix the comment that claims "both themes" so it is true. Re-dump the `motion/` modal frames; resting frames unchanged (`diff -r` vs `v2-05s-services/after/` empty). |

## Rules
- One commit (plus frames), subject naming the id. Stage by name. Gate (CGO_ENABLED=1 for -race); every commit passes the tests.
- Receipt `work/inbox/runner-v2-06-animations-fix-r2.md` (≤ 20 lines), committed exactly once. Worktree clean. No push. Never run brew writes.
