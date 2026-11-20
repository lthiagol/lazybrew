# Fix round r1 — slice `v2-03b-sidebar-lists`

> The slice contract still applies in full. Findings from
> `work/inbox/reviewer-v2-03b-sidebar-lists-r1.md` plus three orchestrator
> rulings (`work/questions.md`, "Orchestrator rulings on Q-v2-03b-a..d").
> **Fix every one.**

| id | severity | fix must achieve |
|---|---|---|
| F-orch-01 | ruling Q-v2-03b-a | Batch markers are drawn only on the Outdated list. Test: a batch on Outdated leaves Formulae/Casks/Taps/Services rows at the same indices unmarked. |
| F-orch-02 | ruling Q-v2-03b-c | Taps `Trust` tab shows only the trust story for the selected tap: trusted / untrusted / official, what that means for its formulae and casks, and the `t` key hint. `Tap Info` keeps the details. Test. |
| F-orch-03 | ruling Q-v2-03b-d | `PgUp`/`PgDn` scroll the main content by a page; `Home`/`End` to top/bottom (add `g`/`G` only if unbound — check `gui.go`). Help and the relevant hints document them. Tests for each key, clamped at both ends. |
| F-r1-01 | minor | No code change: state in the receipt that `0fa39ff` is AC-01. |
| F-r1-02 | minor | Status sidebar first row is label `Homebrew` + version in the same `%-10s` column as the others; update `TestStatusRowsAreCompactSummary`. |
| F-r1-03 | minor | Stale comments: `gui_test.go:1336` names Info, not List; the comment at `panel.go:74` describes the field it sits on. |

## Rules
- One commit per finding (F-r1-01 none), subject naming the id. Stage by name.
- Gate: `make build`, `make lint`, `CGO_ENABLED=1 go test ./... -race -count=1`, `make cover-check`.
- Re-run the frame harness; update `after/` (+ `.ansi` 140x40) and add a `result.md` line per changed frame. Frames exact size.
- Receipt `work/inbox/runner-v2-03b-sidebar-lists-fix-r1.md` (≤ 30 lines), committed once. Worktree clean. No push. Never run brew write commands.
