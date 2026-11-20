# Fix brief — cr3-render-perf, round r1

Source: `work/inbox/reviewer-cr3-render-perf-r1.md`. Fix every finding, one
commit each, subject names the id. Briefs `work/briefs/cr3-render-perf.md` and
`work/briefs/cr-common.md` still apply (above all: **every frame byte-identical
to `dc4df13`** except C11's text). The seven AC commits are already on `wip`;
do not redo them.

- **F-r1-01 (blocker)** — `internal/gui/scroll.go:63` `mainRows`, also through
  `contentLast` (`scroll.go:110`). The Files/Used By wrap cache stores whatever
  `renderContent` returns whenever `cachedTab()` is true, but `renderContent`
  returns `renderOperation` when `m.opState != nil` and the spinner +
  "Loading..." when the panel is loading. The cache key ignores both:
  (a) an operation started over a cached tab shows the Files rows instead of
  the overlay for its whole run, including failed/cancelled;
  (b) an operation started while the tab is uncached has its first frame cached
  and frozen, and that frame replaces the Files tab after the operation closes;
  (c) loading over a cached tab shows the old rows, and a cached spinner never
  moves; `G`/End and `contentLast` measure the wrong rows during an operation.
  Fix: use the cache only when `renderContent` would really draw the Files /
  Used By tab (no `opState`, panel not loading), or key it on that state; in
  every other state fall through to the uncached path. Check every other
  early return in `renderContent` for the same hole and say what you found.
  Tests, written first, asserting on `View()` and failing on `c043ba2`'s
  `mainRows`: an operation over a **cached** Files tab (start, stream a line,
  fail, close) and over an **uncached** one; loading over a cached tab; after
  the operation closes the Files rows are back. Add these states to the frame
  comparison and run it at `dc4df13` and HEAD, 80x24 and 140x40, both themes:
  only the C11 frames may differ; paste the counts in the receipt.
- **F-r1-02 (minor)** — `internal/gui/motion.go:126` `toastName` still builds
  `"toast:" + FormatUint(id) + ":" + part` per call, several times per toast per
  frame. P12 asks for no string joins. Build each toast's motion keys once
  (keep the names on the toast, or memoise by id) so a frame allocates none for
  them; test with `testing.AllocsPerRun`. If `anim.Set`'s string keys make
  that unreasonable, append `Q-cr3-toastname` to `work/questions.md`, leave the
  code, and say so.

Gate: `make test-fast PKG=./internal/gui/...` per commit, then `make ci`.
Receipt `work/inbox/runner-cr3-render-perf-fix-r1.md` (at most 30 lines,
committed once), worktree clean, reply with exactly one line:
`done: work/inbox/runner-cr3-render-perf-fix-r1.md`.
