# Fix brief — cr5-dead-code, round r1

Source: `work/inbox/reviewer-cr5-dead-code-r1.md`. Fix every finding, one
commit each, subject names the id. Briefs `work/briefs/cr5-dead-code.md` and
`work/briefs/cr-common.md` still apply. The seven AC commits are already on
`wip`; do not redo them. Comments, test names and the receipt only: no
production behaviour change, frames unchanged. The tools are installed (paths in
the cr5 brief); never `go run` or `go install`.

- **F-r1-01 (minor)** — `internal/gui/render.go:113` and `:77`. The doc of
  `contentWidthIn` says "mainContentWidth for the layout l", but
  `mainContentWidth` was deleted in c978a89 together with the only explanation
  of what the width means. `renderMainPanelIn` (:77) lost its "main column,
  exactly l.mainW × l.bodyH" text the same way. Fold the deleted doc text (read
  it at `c978a89^`) into these two comments so that no comment names a removed
  function. Then grep `internal/` for any other comment that names a symbol
  this slice deleted (`mainContentWidth`, `renderMainPanel`,
  `renderCachedTabContent`, `StatusCancelled`, `Status.String`,
  `ProgressLineMsg`, `padRight`, `extractPackageName`, `selectedItem`,
  `Format*` that were removed, `MockRunner` paths) and fix each.
- **F-r1-02 (minor)** — `internal/gui/task/manager_test.go:409` says the test
  "locks the StatusCancelled wiring … recorded as cancelled", but
  `StatusCancelled` is gone. `internal/gui/gui_test.go:1257`, `:1275`, `:1391`,
  `:1415` still talk about "the fast path" and keep the name
  `TestDepsReissuesFetchOnDataLoadedWhenFastPathDoesNotApply`; the fast path was
  `renderCachedTabContent`, deleted in c978a89. Reword the comments to say what
  the tests assert now (the completion carries `context.Canceled`; the Deps tab
  needs no fetch) and rename that test to match (rename only, same body; if its
  body is now identical to another test, say so in the receipt rather than
  deleting it).
- **F-r1-03 (minor)** — the receipt `work/inbox/runner-cr5-dead-code.md` does
  not give the full old → new rename mapping that AC-07.2 asks for. Put all 7 in
  the fix receipt, including `review_r2` → `batch_upgrade_test.go`,
  `runner_stream_r1` → `runner_stream_errors_test.go`, `runner_stream_r2` →
  `runner_stream_cancel_test.go`, and the four the first receipt names. Read the
  real mapping from `git show --stat -M d0a6a86`.

Gate: `make test-fast PKG=./internal/gui/...` per commit (plus
`./internal/gui/task/...` for F-r1-02), then `make ci` and
`GOOS=darwin go build ./...`. Keep disk use low: `/tmp` is a nearly full tmpfs,
so do not run `go clean -cache` and keep scratch files small.
Receipt `work/inbox/runner-cr5-dead-code-fix-r1.md` (at most 30 lines, committed
once), worktree clean, reply with exactly one line:
`done: work/inbox/runner-cr5-dead-code-fix-r1.md`.
