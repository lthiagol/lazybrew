# Fix round r1 — slice `v2-04-overlays-styled`

> The slice contract still applies. Findings from
> `work/inbox/reviewer-v2-04-overlays-styled-r1.md`. **Fix every one.**

| id | severity | fix must achieve |
|---|---|---|
| F-r1-01 | major | The toast stack layer covers only the toast boxes: no padding cells. Every cell outside the union of the boxes equals the base frame. Test with a stack of mixed widths at 80×24 and 140×40. `toastAt` (click hit-test) stays consistent with the new layout. Re-capture frames. |
| F-r1-02 | major | Every confirm call site gives a consequence line (reuse existing texts where they exist). Extend `TestConfirmKindsByCallSite` (or equal) so each call site's dialog is asserted to carry a non-empty detail line. |
| F-r1-03 | minor | The chip names the **action** (`UNINSTALL`, `ZAP`, `UPGRADE`, `UPGRADE ALL`, `REINSTALL`, `INSTALL`, `UNTAP`, `REPAIR`, `CLEANUP`, `AUTOREMOVE`, …); the package/tap name stays in the question. Test per call site. |
| F-r1-04 | minor | Resolved by the orchestrator ruling `Q-v2-04-wrap` in `work/questions.md`: prose wraps, single-line parts (chip, key line, menu items) are cut with `…`; the box fits 80×24 including height. Make the code match that reading and test it (a long question at 80×24 wraps and the box fits; a long menu item is cut). |
| F-r1-05 | minor | The backdrop test also asserts that cells inside the modal box carry the modal's own styling (e.g. the accent border) and not the muted role. |

## Rules
- One commit per finding, subject naming the id. Stage by name. Gate (CGO_ENABLED=1 for -race).
- Re-run the frame harness; update `after/` (+ `.ansi` 140x40) and add a `result.md` line per changed frame.
- Receipt `work/inbox/runner-v2-04-overlays-styled-fix-r1.md` (≤ 30 lines), committed once. Worktree clean. No push. Never run brew write commands.
