# Fix round r2 — slice `v2-03-look-foundation`

> The slice contract and the r1 fix brief still apply. Findings from
> `work/inbox/reviewer-v2-03-look-foundation-r2.md`. **Fix every one.**

| id | severity | fix must achieve |
|---|---|---|
| F-r2-01 | major | A count in a border label always appears **whole**. When it does not fit, the title yields: first drop the icon, then cut the name with `…` — never the count. Test at 80 columns with 4- and 5-digit Formulae counts (and Outdated/Services ≥ 1000): the full number is in the border and the row is exactly the box width. Reflow the comments the reviewer noted (`box.go:11`, `render.go:40`) while there. |

## Rules
- One commit, subject naming F-r2-01. Stage by name. Gate (CGO_ENABLED=1 for -race).
- Re-run the frame harness; if frames change update `after/` + a `result.md` line. Frames exact size.
- Receipt `work/inbox/runner-v2-03-look-foundation-fix-r2.md` (≤ 20 lines), committed once. Worktree clean. No push. Never run brew write commands.
