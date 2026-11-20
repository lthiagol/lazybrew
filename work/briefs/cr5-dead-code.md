# Slice: cr5-dead-code — delete what nothing uses, tidy the tests

> **Contract.** This brief wins over any plan, note, or prior convention where
> they differ. `work/briefs/cr-common.md` applies. If the repo contradicts the
> brief, follow the brief and report the contradiction. Don't stop to ask.

**Depends on:** `cr4-gui-dedup` closed.

## Goal

After cr1–cr4 the code has settled. Delete the dead code from `code-review.md`
§3, move test support out of the production binary, and fix the test-hygiene
items from §5. The work is mechanical, but **re-check each symbol before
deleting it**: cr1–cr4 may have revived or already removed some. Record what
you found in the receipt as "already gone" or "now used, kept".

Run these first and paste their output into the receipt (as the "before"):

```bash
go run golang.org/x/tools/cmd/deadcode@latest ./cmd/...
go run golang.org/x/tools/cmd/deadcode@latest -test ./...
go run honnef.co/go/tools/cmd/staticcheck@latest ./...
```

## Acceptance criteria

To-do (one commit each, in this order):

- [ ] AC-01 — `MockRunner` moves to `internal/brew/brewtest` [D6]
- [ ] AC-02 — panel items become bare data [D3]
- [ ] AC-03 — unused brew symbols and interface methods [D1, D2, D4 brew part]
- [ ] AC-04 — unused gui symbols, fields and branches [D1, D4, D5 gui part]
- [ ] AC-05 — unused theme styles [D4 styles]
- [ ] AC-06 — version sentinel [Q7]
- [ ] AC-07 — test hygiene [§5]

### AC-01 — `MockRunner` moves to `internal/brew/brewtest` [D6]

1. Move `MockRunner`/`NewMockRunner` out of `internal/brew/runner.go` into a
   new package `internal/brew/brewtest`. Update every importer: the brew,
   gui, gui/flows, gui/testutil and app tests, and `internal/app` if it uses
   it outside tests. Check whether `internal/brew`'s own tests can import
   `brewtest` without an import cycle. If they can't, those tests use an
   external test package (`package brew_test`) or a local copy in a
   `_test.go` file; say which.
2. Also move the shared test error (`assertAnError`, defined twice) into
   `brewtest`, renamed `ErrTest` (staticcheck ST1012).
3. `go build ./cmd/...` no longer links `MockRunner`. Check with
   `go tool nm bin/lazybrew | grep -c MockRunner`, which should print `0`.

### AC-02 — panel items become bare data [D3]

1. Nothing parses `p.items` back into a name any more (cr1). Either remove
   `p.items` and use the typed slices directly, or keep `items` as bare names
   only, for the code that still needs a string per row (filtering, counts).
   Choose one and say which.
2. Delete `extractPackageName`, `selectedItemName` (if not already gone),
   and the `presentation.FormatFormula`, `FormatCask`, `FormatTap`,
   `FormatService`, `FormatOutdatedFormula`, `FormatOutdatedCask` functions,
   together with `padRight` if it is now unused. Delete their tests too and
   name them in the receipt. `FormatFormulaInfo`/`FormatCaskInfo` stay.
3. Frames are unchanged. The sidebar draws through `typedRow`.

### AC-03 — unused brew symbols and interface methods [D1, D2, D4 brew part]

1. Delete these unless they were revived:
   - `dependencyJSON`;
   - `ParseTrustStatus`, `ParseServiceStatus` and their tests;
   - `InvalidateAll` (test-only);
   - `sfCall.waiters`: make the singleflight test assert another way, e.g.
     by counting fetch calls;
   - `TypedCache`, if cr2 left it.
2. Interface methods with no production caller: `Taps.Get`,
   `Trust.GetTapTrustStatus`, `Search.SearchDesc`, `TapsWrite.TapWithURL`,
   `Diagnostics.Version`. Also `Formulae.Get`/`Casks.Get` **unless** cr1
   AC-03 revived them. Remove them from the interfaces and their
   implementations, and update the test fakes.
3. **`CasksWrite.Pin`/`Unpin` are removed regardless**, because `brew pin`
   works only on formulae. Check that no key binding offers pin on Casks. If
   one does, hide it there and name the golden that changes.
4. Remove the domain fields that are never filled in: `Cask.ConflictsWith`,
   `Tap.LastCommit`, `BrewConfig.Arch`, `Formula.Size`. First grep
   presentation and render for reads, because a read of an always-empty field
   is itself dead UI. Name any UI line removed with it.

### AC-04 — unused gui symbols, fields and branches [D1, D4, D5 gui part]

1. Delete these unless cr1–cr4 already removed or revived them:
   - `Model.newestToast`;
   - the `panelData` fields `icon`, `width` and `active` (switch the test
     that reads `active` to `m.activePanel`);
   - `ProgressLineMsg`, its `Update` case and its test;
   - `MutationResultMsg.Type` and `DepCheckMsg.Name`;
   - `thirdPartyTaps`;
   - `statusDashboard`'s `[]error` return. Update `fetch_timeouts_test.go`
     to assert on `summary.Problems`/`DoctorErr`.
2. [D5] Remove the unreachable branches:
   - the Taps arms of `executeMenuAction`. Give the Brewfile menu its own
     `pendingAction`, or call `executeBrewfileAction` directly;
   - the `searchInput.Focused()` checks inside the `esc`/`enter` handling;
   - `renderCachedTabContent` and the `formulaDepsTab` entry in
     `needsTabFetch`, together with the matching `fetchTabContentCmd` arm, if
     the Deps tab is drawn at render time. Prove it with a test: the Deps tab
     shows its content without a fetch.
3. `task.Status.String()` and `Task.Status`: if only tests read them, delete
   them and assert on `TaskCompletedMsg.Err` instead. Otherwise move the
   write in `RunNext` under the lock. Say which.

### AC-05 — unused theme styles [D4 styles]

1. Delete `ActiveBorder`, `InactiveBorder`, `HintKey`, `HintDesc` and
   `DocStyle` from `style/theme.go`, along with the stale `ActiveBorder`
   comment in `mouse.go`.
2. `PanelTitle`, `ActivePanelBg`, `LogBorder` and `CurrentTheme` are read
   only by tests. Delete them and update those tests. If a test uses one to
   check theming as a whole, switch it to a style that production uses.

### AC-06 — version sentinel [Q7]

1. `cmd/lazybrew/main.go`: `var version = ""`. When it is empty, fall back
   to `debug.ReadBuildInfo()`, then to `"dev"`. The repeated `"1.0.0-rc1"`
   literal is gone. Check that the Makefile's `-ldflags -X` still targets the
   variable. If the Makefile would need a change, make it in this AC's
   commit and say so in the receipt.
2. Test or check: `./bin/lazybrew --version` prints something sensible with
   and without ldflags.

### AC-07 — test hygiene [§5]

1. Fix every staticcheck finding in the tests. Each SA4006 (an assigned
   `m`/`updated` that is never used) is resolved by adding the assertion that
   was obviously meant. Only if no assertion makes sense, drop the
   assignment and explain why in the receipt. Fill or remove the empty branch
   in `style/theme_test.go`. Fix S1021 and S1039.
2. Rename the test files named after review rounds to describe their
   behaviour, using `git mv` only, with no content edits in that commit:
   `review_r1..r4_test.go`, `runner_stream_r1/r2_test.go` and
   `regression_test.go`. Name each file after what most of its tests cover;
   the receipt gives the old → new mapping. If a file mixes unrelated
   behaviour, keep a neutral name and note it under "seen, not done".
   Splitting `gui_test.go` is out of scope; note it as backlog.

## Done means

- `deadcode ./cmd/...` lists nothing, apart from symbols you kept on purpose,
  each one justified in the receipt.
- `staticcheck ./...` lists nothing.
- The receipt shows the output of both tools before and after, and
  `git diff --shortstat` for the whole slice.

## Invariants

The common invariants apply. In addition: frames are unchanged except where
AC-03.3 or AC-03.4 names a golden. Coverage floors are not lowered. Deleting
a test is allowed only for code that this slice deletes.
