# Slice: cr3-render-perf — take repeated work out of `View()` and the output stream

> **Contract.** This brief wins over any plan, note, or prior convention where
> they differ. `work/briefs/cr-common.md` applies. If the repo contradicts the
> brief, follow the brief and report the contradiction. Don't stop to ask.

**Depends on:** `cr2-brew-cache` closed.

## Goal

`View()` runs after every message. It must not parse JSON, re-parse the whole
operation log, re-render thousands of rows, or copy the command log on every
frame. Streaming output must not cost one full Update+View per line. **Every
frame must stay byte-identical** to the current output, except for C11 (error
text) and the cases an AC names. Golden and teatest expectations are
unchanged.

## Acceptance criteria

To-do (one commit each, in this order):

- [ ] AC-01 — search info parsed once; real error shown [P4, C11]
- [ ] AC-02 — operation steps kept incrementally; output lines bounded [P5]
- [ ] AC-03 — batched task output; no lines lost at completion [P5, C10]
- [ ] AC-04 — cached main-view content [P6, P7]
- [ ] AC-05 — layout and command log [P8]
- [ ] AC-06 — shared text styles and cheaper sidebar rows [P9, P10]
- [ ] AC-07 — small hot-path costs [P11, P12, P13]

### AC-01 — search info parsed once; real error shown [P4, C11]

1. When the search info result arrives (`SearchDoneMsg` / the info message
   handled near `gui.go` ~:366), parse it once, using the parser from cr1
   AC-03. Store `*pkgInfo` and a separate error on the model.
   `renderSearchInfo` draws from those and never calls `json.Unmarshal`.
2. [C11] An error from brew is shown as that error (e.g.
   `Error: No available formula …`), not as `Parse error: invalid character
   'E'`. Test written first.
3. Benchmark: `View()` of the search panel with info loaded, before and
   after.

### AC-02 — operation steps kept incrementally; output lines bounded [P5]

1. Keep the parsed step state (steps, total, begun, whatever `parseSteps`
   returns) on `Operation` and update it in `AppendLine`. `renderOperation`,
   `operationBar`, `operationBody` and `motionProgress` read the cached state.
   None of them calls `parseSteps` over all lines on a render.
2. Cap `Operation.Lines` at a named constant (e.g. 5000). Older lines are
   dropped, and the cached step state stays correct (it was computed when the
   lines arrived). The body view shows the tail as it does now.
3. Tests: the step state after N appended lines equals `parseSteps(all
   lines)` on the same input, as a table over the existing step fixtures.
   Lines beyond the cap are dropped and the progress is unchanged. Benchmark:
   append 10 000 lines and render after each one, before and after.

### AC-03 — batched task output; no lines lost at completion [P5, C10]

1. `TaskOutputMsg` carries `Lines []string`. `Manager.readChunk`
   (`internal/gui/task/manager.go`) sends the first line, plus whatever else
   is already buffered without blocking, up to a named constant (e.g. 64), in
   one message. Update every consumer of `TaskOutputMsg.Line`.
2. [C10] When the error channel is ready, `readChunk` drains `outCh` until it
   closes (or until ctx is cancelled) and delivers those lines before
   `TaskCompletedMsg`. No buffered line is ever lost. Test written first: a
   fake task sends K lines and then the error at once; over 500 iterations
   every line arrives, under `-race`.
3. The existing task-manager and operation tests stay green.

### AC-04 — cached main-view content [P6, P7]

1. [P6] The rendered content of the Files and Used By tabs (`renderFiles`,
   `depviews.go`), and the wrapped content used by `scrolledContent` /
   `contentLast` (`scroll.go`), is cached on the model, keyed by
   (`contentKey`, width, theme). The cache is cleared when `tabContent` for
   that key changes, the selection changes, or the window resizes.
   Scrolling (PgDn, wheel, `G`) does not re-render or re-wrap.
2. [P7] The set built by `installedList` is computed once when panel data
   arrives (`DataLoadedMsg`) and stored on the panel, not on every render.
3. Tests: frames are identical with and without the cache (render twice and
   compare with the current output). Scrolling 50 times calls the renderer
   once; count it through a test hook. Benchmark: a Files tab with 5000 files,
   scrolling with PgDn, before and after.

### AC-05 — layout and command log [P8]

1. `CommandLog` gets `Len()` and `Last(n)`. `logHeight` uses `Len()`, and
   `CommandLog.View` uses `Last(h)`. Neither copies the whole ring any more.
2. `screenLayout()` is computed once per `View()` and passed to the helpers
   that need it. Cut the extra calls from `motionSidebar` in `Update` where
   you can, and say what is left.
3. Tests for `Len`/`Last` (empty ring, partly full, wrapped). Frames
   unchanged.

### AC-06 — shared text styles and cheaper sidebar rows [P9, P10]

1. [P9] `style.ApplyTheme` builds the shared foreground styles once:
   `MutedText`, `SuccessText`, `WarningText`, `ErrorText` and `TextStyle`
   (names are up to you). Replace the per-render
   `lipgloss.NewStyle().Foreground(style.X)` calls in `internal/gui` and its
   subpackages with them. At `c1aa732` grep finds about 28 sites; the receipt
   lists the count before and after. `OutdatedBadge`/`InstalledBadge` reuse
   the new styles where they are the same.
2. [P10] `layoutSidebarRow` (`sidebar_rows.go`) and `panel` (`box.go`) render
   the edge glyphs, separator and padding cells once per call, not once per
   row, and build the highlight styles once.
3. Frames are byte-identical across all themes (an existing or new test
   renders each theme). Benchmark: `View()` of a 300-row formulae panel,
   before and after.

### AC-07 — small hot-path costs [P11, P12, P13]

1. [P11] `needsTabFetch` (`gui.go`) uses a `switch` or a package-level table
   and allocates nothing per call.
2. [P12] `helpRows` (`help.go`) is cached per (width, theme, panel).
   `presentation.Info` (`detail.go`) no longer renders `Detail(left)` twice
   just to measure it. `toastName` (`motion.go`) no longer uses
   `fmt.Sprintf("%p")` and string joins: key the motion state on the `*Toast`
   or on a stable toast id.
3. [P13] `sortFormulae` (`gui/formulae.go`) lowercases each name once (a
   decorate-sort or precomputed keys) and uses `slices.SortStableFunc`. The
   order is unchanged, and a test locks the ordering, including names that
   differ only in case.

## Invariants

The common invariants apply. In addition: no frame changes except C11's error
text, and every frame stays exactly the terminal size at 80x24 and 140x40.
Animations and motion behaviour are unchanged.
