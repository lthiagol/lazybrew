# Fix round r1 — slice `v2-05-content-polish`

> The slice contract still applies. Findings from
> `work/inbox/reviewer-v2-05-content-polish-r1.md`. **Fix every one.**

| id | severity | fix must achieve |
|---|---|---|
| F-r1-01 | major | Upgrade-all progress counts only per-package `Upgrading` steps (the `==> Upgrading N outdated package(s):` header is not a package). The total comes from that header's N (or excludes pinned packages), so a finished run reads `total/total`. Test fixture starts with the real header; assert mid-run and final counts (incl. 1-package and pinned-skipped cases). |
| F-r1-02 | major | Retarget every expectation that guards an invariant through the dead dashboard text (timeout → ⚠, outdated unknown never `0`, doctor exit 1 = warnings, doctor failure ≠ "no issues", and the rest the reviewer lists) to `dMsg.Summary` and/or the rendered cards (`renderDashboard`). Then delete the dead text path (`DataLoadedMsg.Dashboard`, the `items` in `statusDashboard`, `FormatStatusDashboard`, `FormatDoctorStatus`, `OutdatedInfo`, `truncate`) and tests that only tested it. Presentation coverage floor kept. Name every moved expectation in the receipt (old test:line → new test). |
| F-r1-03 | minor | Test: help open + running `opState`: first Esc closes help and does not cancel; second Esc cancels. |
| F-r1-04 | minor | When the dashboard cards stack (80×24), Health comes right after Homebrew (before Packages/Taps/Services) so problems are visible without scrolling; when the doctor read failed, Health says `doctor unavailable` (with the reason), not `press d to run doctor`. Tests. Re-capture 01 at both sizes. |

## Rules
- One commit per finding, subject naming the id. Stage by name. Gate (CGO_ENABLED=1 for -race).
- Re-run the frame harness; update `after/` (+ `.ansi` 140x40) and add a `result.md` line per changed frame.
- Receipt `work/inbox/runner-v2-05-content-polish-fix-r1.md` (≤ 35 lines), committed once. Worktree clean. No push. Never run brew write commands.
