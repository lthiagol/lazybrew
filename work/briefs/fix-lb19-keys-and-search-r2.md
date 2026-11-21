# Fix brief — lb19-keys-and-search, round r2

> Slice contract: `work/briefs/lb19-keys-and-search.md`. Findings:
> `work/inbox/reviewer-lb19-keys-and-search-r1.md` (read in full). One commit
> per finding, staged by name; tests first (show the failing run, or for
> test-only findings show the test failing with the wiring reverted).

- **F-r1-01:** teatest flow cycling all 7 panels both ways through a focused
  Search, asserting the active panel in `View()` after each step; it must fail
  with the AC-01 fix reverted.
- **F-r1-02:** `TestSearchMarkClearsAfterUninstall` must drain the cmd batch
  returned by the `TaskCompletedMsg` update and check the mark clears; it must
  fail with `m.remarkSearch()` removed from that batch.
- **F-r1-03:** test where marks from an older result set arrive after a newer
  `SearchDoneMsg`; no row of the new set changes. It must fail with the name
  guard (`SearchMarksMsg`) removed.
- **F-r1-04:** when an installed list cannot be read, re-mark leaves that
  type's marks as they are (rows still installed stay installed); fix the doc
  comment; test.
- **F-r1-05:** ruling `Q-lb19-search-partial` (work/questions.md): keep the
  rows of the read that worked; error only if both fail. Test both cases.
- **F-r1-06:** extend `TestRetiredTabAndPageKeysDoNothing` to the open help
  screen (PgUp/PgDn leave `helpScroll` unchanged); it must fail with a
  `pgdown` case restored in `helpKey`.

Gate: `make test-fast PKG=...` per commit, `make ci` before the receipt. Never
launch the TUI. Receipt: append "Fix r2" (≤ 15 lines) to
`work/inbox/runner-lb19-keys-and-search.md`, commit
(`inbox: runner fix r2 lb19-keys-and-search`), reply with one done line.
