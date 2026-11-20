# Fix round r2 — slice `v2-04-overlays-styled`

> The slice contract, ruling Q-v2-04-wrap and the r1 fix brief still apply.
> Findings from `work/inbox/reviewer-v2-04-overlays-styled-r2.md`. **Fix every one.**

| id | severity | fix must achieve |
|---|---|---|
| F-r2-01 | minor | Every line of the input modal fits the inner width (fit the textinput, including prompt and cursor, to the inner width, or cut it with `fitLine`); chip, prompt and key line centred again. Test: every content row of the input modal ≤ inner width, chip centred (left/right gaps within 1). Also assert the same for confirm and menu so no dialog can regress this way. Re-capture frame 14 (both sizes, `.ansi` 140x40) and add a `result.md` line. |

## Rules
- One commit (plus the frames commit), subject naming the id. Stage by name. Gate (CGO_ENABLED=1 for -race).
- Receipt `work/inbox/runner-v2-04-overlays-styled-fix-r2.md` (≤ 20 lines), committed once. Worktree clean. No push. Never run brew write commands.
