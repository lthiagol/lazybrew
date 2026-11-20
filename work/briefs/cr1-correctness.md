# Slice: cr1-correctness — user-visible bugs from the code review

> **Contract.** This brief wins over any plan, note, or prior convention where
> they differ. `work/briefs/cr-common.md` applies (source, test rules, safety,
> invariants, gate, report). If the repo contradicts the brief, follow the
> brief and report the contradiction. Don't stop to ask.

## Goal

Fix the bugs where lazybrew does the wrong thing for the user: actions sent to
the wrong (truncated) package, failed writes reported as successes, a search
info pane that shows empty data, a modal nil panic, timer chains that
multiply, and a few small edge cases. Each fix gets a regression test that
fails before it. Refactoring beyond what a fix needs is **out of scope**:
`cr4` does the dedup and `cr5` deletes dead code.

## Acceptance criteria

To-do (one commit each, in this order):

- [ ] AC-01 — actions use the full package name [C1]
- [ ] AC-02 — `LoggingRunner` stream error has a single reader [C2]
- [ ] AC-03 — search info is decoded by the real parsers [C6]
- [ ] AC-04 — no nil dereference after a modal update [C5]
- [ ] AC-05 — at most one auto-refresh chain and one update ticker [C3, C4]
- [ ] AC-06 — small edge cases [C12, C13]

### AC-01 — actions use the full package name [C1]

1. Add one method on `panelData` (`internal/gui/panel.go`), say
   `selectedName() string`. It returns the selected row's real name from the
   typed slice for the panel (`formulae`, `casks`, outdated rows, `taps`,
   `services`). It returns `""` when nothing is selected or the index is out
   of range.
2. Replace **every** production call of `extractPackageName(...)` and
   `selectedItemName(...)` with it. That includes the leaves filter in the
   `Update` switch (it matches item names) and the tab-fetch key in
   `render.go`. At `c1aa732`, grep finds about 17 call sites in `commands.go`,
   `gui.go` and `render.go`; list them in the receipt. After this AC,
   `extractPackageName` has no production caller. Keep the function and
   `p.items` as they are: `cr5` deletes them (D3).
3. Tests (written first, failing before the fix): for each of Formulae, Casks,
   Outdated, Taps and Services, a row whose name is ≥ 30 runes. Examples:
   `font-jetbrains-mono-nerd-font`, `someuser/homebrew-a-very-long-tap`. Each
   test drives an action through the model (uninstall/upgrade/pin for
   packages, untap or trust for taps, start/stop for services, and a tab
   fetch) and asserts on the `MockRunner`'s recorded args that the **full**
   name was used.

### AC-02 — `LoggingRunner` stream error has a single reader [C2]

1. `LoggingRunner.ExecuteStream` (`internal/brew/runner.go`) reads the inner
   error channel **once**, calls `OnExec(args, err)`, then forwards `err` on a
   new buffered channel that it returns, closing that channel when done. Use
   the pattern in `sudoRunner.ExecuteStream` (`sudo.go`). When `OnExec` is
   nil, the behaviour is unchanged.
2. Tests (written first): with `OnExec` set and an inner stream that fails,
   (a) the caller receives the non-nil error, (b) `OnExec` receives the same
   error, and (c) this holds over 200 iterations under `-race`. The success
   path is covered too.

### AC-03 — search info is decoded by the real parsers [C6]

1. `parsePackageInfo` (`internal/gui/commands.go`) stops decoding into
   `brew.Formula`/`brew.Cask` through json tags. Decode with the brew
   package's parsers instead: export one helper, e.g.
   `brew.ParseInfo([]byte) ([]Formula, []Cask, error)`, built on
   `formulaeJSON`/`casksJSON` + `parseFormula`/`parseCask`. If it's simpler,
   revive `Formulae.Get`/`Casks.Get` on top of the same helper. Say which you
   chose.
2. Once nothing decodes JSON into them, delete the json tags from the domain
   types in `internal/brew/types.go` and the `Formula.Installed` field, which
   only served this path. Grep first: if another production path still relies
   on a tag, keep that tag and say so.
3. Tests (written first): a realistic `brew info --json=v2` fixture for one
   formula. Copy the shape from `brew info --json=v2 jq` (read-only) and keep
   only the fields you need, including `versions.stable`, `bottle.stable` and
   `installed[].version`. Also one cask fixture. Assert that `pkgInfo`
   `Version`, `Bottled`, `Installed` and `InstallPath` (if brew gives it) are
   right. Assert also on the rendered search info `View()`.

### AC-04 — no nil dereference after a modal update [C5]

1. At every place in `internal/gui/gui.go` where `m.activeModal.Update(...)`'s
   result goes through a `modal.Modal` type assertion (the paste path and the
   key path), the not-a-modal branch must not fall through to
   `m.activeModal.Done()`. Return early, or guard the call.
2. Test (written first): a fake modal whose `Update` returns a
   non-`modal.Modal` `tea.Model`. Paste and press a key. No panic, and the
   modal is cleared.

### AC-05 — at most one auto-refresh chain and one update ticker [C3, C4]

1. Add a model flag, e.g. `autoRefreshTicking`. `autoRefreshCmd()` is armed
   only from `Init` and from the `autoRefreshTickMsg`/`autoRefreshPausedMsg`
   handlers, never from the `RefreshMsg` handler. The pause-on-keypress
   behaviour (M12 AC-03) stays as it is.
2. The same goes for `UpdateTickMsg`: one ticker, guarded by a flag. The
   extra `updateTickerCmd()` armed by the `R` key handler is removed.
3. Tests (written first): with `AutoRefreshSeconds > 0`, send 5 `RefreshMsg`s
   (task completions, `R`) and count the tick commands in flight. Do this by
   running the returned cmds through a helper that collects messages, or by
   counting tick arms with a test hook. Exactly one tick chain may exist. Do
   the same for the update ticker after 3 `R` presses with
   `update_on_start`. The existing auto-refresh and pause tests stay green.

### AC-06 — small edge cases [C12, C13]

1. [C12] The error text in `panel.go` (~:302, `msg[:width-5]`) is cut with
   the existing `clipLine` (or `ansi.Truncate`), never by bytes. Test: a
   multibyte error message at a width that would split a rune; the output is
   valid UTF-8.
2. [C13] `internal/app/app.go` never sends debug logging to **stderr** while
   the TUI runs. If the debug log file can't be opened, debug logging is
   turned off (or goes to `io.Discard`) and a single line is printed to
   stderr **before** the TUI starts. Test: `UserHomeDir` (or the log path)
   fails, and nothing is written to the stderr writer after setup. Inject the
   writer if needed.

C11 (search error shown as a JSON parse error) is fixed in `cr3`, together
with parsing the search result once.

## Invariants

The common invariants apply. In addition: `p.items`, the `presentation.Format*`
functions and `extractPackageName` stay in the code (cr5 removes them).
Footer, help and goldens are unchanged.
