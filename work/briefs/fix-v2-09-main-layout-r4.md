# Fix brief — v2-09-main-layout, round r4 (last allowed round: r5 reviews it)

Source: `work/inbox/reviewer-v2-09-main-layout-r4.md`. One commit for the fix
(subject names the id) and include the regenerated frames and result.md in it;
receipt `work/inbox/runner-v2-09-main-layout-fix-r4.md` (≤ 30 lines, committed
**once** — count the lines before committing). Gate per ruling
`Q-v2-builder-speed` (per commit: touched packages, no `-race`; full gate
with `-race` once before the receipt).

- **F-r4-01 (major)** — at 140x40 the split breaks paths and URLs mid-token
  in Tap Info (06) and Services Info (19). Use this rule (the reviewer's first
  option — keep it simple): **a view splits only when every non-prose value
  (paths, URLs, commands — anything without spaces) placed in the right column
  fits that column whole; otherwise that view keeps today's single column.**
  Prose still wraps. Formulae and Cask Info must keep splitting at 140x40.
  Tests at 87 cells on the Services Info and Tap Info fixtures: no path/URL
  split across rows (single column kept); a Formulae Info fixture still
  splits. Regenerate frames and name the changed screens in `result.md`.

Do not touch anything else.

Reply with exactly one line: `done: work/inbox/runner-v2-09-main-layout-fix-r4.md`.
