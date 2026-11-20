# Fix round r1 — slice `v2-06-animations`

> The slice contract still applies (resting frames identical to v2-05s; every
> intermediate frame exact size). Findings from
> `work/inbox/reviewer-v2-06-animations-r1.md`. **Fix every one.**

| id | severity | fix must achieve |
|---|---|---|
| F-r1-01 | major | `testutil.NewTestModel` / `NewTestModelWithRunner` build a Model with motion **off by themselves** (not via the caller's env). `internal/app` tests run with animations off and do not read the developer's real `~/.config/lazybrew/config.yml` (point them at a temp config / HOME). A test in `flows` and one covering the testutil helpers fail if a helper-built Model would animate. |
| F-r1-02 | minor | With animations **off**, `Update` does no motion work at all (no clone, no `motionTargets`). With them **on**, targets are recomputed only when their inputs change (panel, tab, size, modal, toast stack, op share) or computed without re-rendering the tab bar. Add a benchmark (kept in the repo, not run by the gate) for an unhandled message at 140×40 with animations off and on; record before/after numbers in the receipt; off must be within ~2× of 314dd3c's 2 allocs / 2.4 µs. Resting frames unchanged (re-run the harness, `diff -r` vs `v2-05s-services/after/` empty). |
| F-r1-03 | minor | `work/verification/v2-06-animations/result.md` records the deviation: at `7c79a51` `TestMovingValueTicksUntilItSettles` failed, fixed by `0ebc253`; why. |
| F-r1-04 | minor | The modal backdrop fade never brightens: the first fade frame is no brighter than the frame it replaces (start from the frame's own colours / subtle role, not the text role); the fade's settle threshold fits a 0..1 value (no snap from ≈0.8). Re-dump `motion/` modal frames; resting frame unchanged. Test: every fade step's foreground luminance ≤ the previous step's for a backdrop cell. |

## Rules
- One commit per finding, subject naming the id. Stage by name. Gate (CGO_ENABLED=1 for -race). **Every commit passes the tests.**
- Receipt `work/inbox/runner-v2-06-animations-fix-r1.md` (≤ 30 lines), committed exactly once. Worktree clean. No push. Never run brew writes.
