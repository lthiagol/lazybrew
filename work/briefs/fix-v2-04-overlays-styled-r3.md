# Fix round r3 — slice `v2-04-overlays-styled`

> The slice contract, ruling Q-v2-04-wrap and earlier fix briefs still apply.
> Findings from `work/inbox/reviewer-v2-04-overlays-styled-r3.md`. **Fix every one.**

| id | severity | fix must achieve |
|---|---|---|
| F-r3-01 | minor | The confirm body after the question (dependents heading + list, timeout note) is laid out as **one left-aligned block**, the block centred in the box — the way the menu pads its items. Chip, question, detail line and key line stay centred per row. Every row still fits the inner width (keep the r2 test passing). Test: an uninstall confirm with dependents of different lengths draws every dependent name at the same start column. Correct the `result.md` line for 07 (and the r2 line that claimed content was unchanged), re-capture 07 (both sizes, `.ansi` 140x40). |

## Rules
- One fix commit (plus the frames commit), subject naming the id. Stage by name. Gate (CGO_ENABLED=1 for -race).
- Before reporting, read **every** dialog frame (07, 12, 14, and a menu) with your own eyes and say in the receipt that each reads aligned.
- Receipt `work/inbox/runner-v2-04-overlays-styled-fix-r3.md` (≤ 20 lines), committed once. Worktree clean. No push. Never run brew write commands.
