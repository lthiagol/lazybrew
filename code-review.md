# Lazybrew code review

**Date:** 2026-11-20 · **Branch:** `wip` @ `c1aa732`
**Focus:** code quality, performance, code reuse, dead code
**Scope:** all non-test Go code (~13.4k lines) and a light pass over the tests (~20.9k lines)

**Method:** `staticcheck -checks all` and `deadcode` (from `golang.org/x/tools`), then a manual read of every package. All findings were checked with grep. Findings marked ✅ were also re-read and confirmed by hand during consolidation.

Severity: **H** = fix soon (wrong behaviour or a large cost) · **M** = worth doing · **L** = cleanup.

---

## 0. Correctness bugs found along the way

These are outside the requested focus. They are listed first because each one makes lazybrew do the wrong thing for the user.

| # | Sev | Where | Finding | Fix |
|---|---|---|---|---|
| C1 ✅ | H | `presentation/formatters.go:181` (`padRight`), `gui.go:1422` (`extractPackageName`), ~15 callers in `commands.go` | Panel `items` are display strings built with `padRight(name, 22)`, which **cuts** names longer than 22 runes. Every action then gets the name back from that string with `extractPackageName`: uninstall, upgrade, pin, untap, trust, tab fetches and so on. For a cask such as `font-jetbrains-mono-nerd-font`, the command runs against `font-jetbrains-mono-ne`. | Take the name from the typed rows already on the panel (`p.formulae`, `p.casks`, `p.taps`, `p.services`) through one `p.selectedName()`. Then stop storing formatted strings in `items` (see D3). |
| C2 ✅ | H | `brew/runner.go:325-337` (`LoggingRunner.ExecuteStream`) | When `OnExec` is set (always, from `app.go:75`), a goroutine reads `<-errChan` and the **same** channel is also returned to the caller. That gives one buffered error two receivers. The loser reads the closed channel and gets `nil`, so either the task overlay or the command log shows a failed install or upgrade as a success. No test covers `LoggingRunner`. | Read the inner error once, call `OnExec`, then forward it on a new channel. `sudoRunner.ExecuteStream` (`sudo.go:33`) already works this way. Add a test. |
| C3 ✅ | M | `gui.go:728`, `gui.go:333-348` | Every `RefreshMsg` arms a new `autoRefreshCmd()` tick, and nothing stops a second tick chain from starting. Each task completion, `R` press or mutation sends a `RefreshMsg`, so each one adds a **permanent** extra chain. After N tasks, N+1 full refreshes (including `brew doctor` and `brew config`) run every interval. This only happens when `auto_refresh_seconds > 0`; the default is 0. | Add an `autoRefreshTicking` flag (the same pattern as `toastTicking`/`motionTicking`). Arm the tick only in `Init` and in the tick/paused handlers. |
| C4 | M | `gui.go:418`, `gui.go:866` | Same chain problem for `UpdateTickMsg`: each `R` press with `update_on_start` adds another 10 s ticker. The cost is only a redraw. | Use one ticker behind a flag. Drop the extra `updateTickerCmd()` in the `R` handler. |
| C5 ✅ | M | `gui.go:763-773` (and the twin path in the key handler) | If `activeModal.Update` returns something that is not a `modal.Modal`, the code sets `m.activeModal = nil` and then immediately calls `m.activeModal.Done()`, which is a nil dereference. | Return right after setting it to nil, or guard the `Done()` call. |
| C6 ✅ | M | `commands.go:718-736` (`parsePackageInfo`) and the json tags in `brew/types.go` | Search info decodes `brew info --json=v2` straight into the domain `brew.Formula` through ad-hoc json tags. Brew's JSON has no top-level `version`, `bottled` or `install_path`, so on the search info view **Version and InstallPath are always empty and Bottled is always false**. `formulaJSON` + `parseFormula` already exist for exactly this decoding. | Decode with the existing parsers. Revive `Formulae.Get`/`Casks.Get` (currently dead, see D1) or add a shared `parseInfo([]byte)`. Then remove the json tags from the domain types and `Formula.Installed`. |
| C7 | M | `formulae.go:265-300`, `casks.go:243-288`, `taps.go:222-246`, `trust.go:92-114`, `services.go:116-150` | Writers invalidate the cache **before** the command runs. During a minutes-long streaming install, any read (preload, Status, auto-refresh) puts pre-install data back in the cache for its full TTL: 5 min for lists, 30 min for outdated. | Invalidate after completion, or both before and after. For streams, wrap `errChan` as in C2. |
| C8 | M | `doctor.go:225-239` | `Cleanup`/`Autoremove` never invalidate the cache, although `InvalidateGroups["autoremove"]` exists. Autoremove removes formulae, so `KeyFormulaeList` stays stale. | Use the invalidation group (see R7). |
| C9 | M | `brew/cache.go:63-82` | `Get` finds an expired entry under the read lock, drops it, takes the write lock and deletes the key **without re-checking**. A `Set` that lands in that gap is lost. | Re-check the timestamp under the write lock, or don't delete at all (a miss is enough; `Set` overwrites). |
| C10 | M | `gui/task/manager.go:177-197` | Plausible: the blocking `select` races `outCh` against `errCh`. If a line and the error become ready between the non-blocking drain and the blocking select, Go can pick the error and drop buffered lines, which the comment above says must not happen. | After receiving from `errCh`, keep draining `outCh` until it closes, then complete (also see P5). |
| C11 | L | `render.go:572`, `gui.go:366` | Search errors are stored as `"Error: …"` in the field that holds raw JSON, so the info pane shows `Parse error: invalid character 'E'…` instead of the real error. | Keep the error in a separate field (see P4). |
| C12 | L | `panel.go:302` | Error text is cut by bytes (`msg[:width-5]`), which can split a UTF-8 rune. | Use `clipLine`. |
| C13 | L | `app/app.go:30-38` | `SetDebug(true)` logs to **stderr** before the switch to the log file. If `UserHomeDir` fails, debug output goes to stderr and corrupts the TUI. | Never send debug output to stderr while the TUI is running. |

---

## 1. Performance

### Refresh and brew-call cost

| # | Sev | Where | Finding | Fix |
|---|---|---|---|---|
| P1 | M | `gui.go:702-724`, `commands.go:1909-1919` | One refresh runs each `brew` list command **twice**. `fetchStatusData` reads `Formulae.List`, `Casks.List`, `Taps.List` and `Services.List`, while `fetchPanelData` reads the same lists for each loaded panel. Both start right after the cache is cleared, and only `Outdated` goes through singleflight. | Route every reader through singleflight (this falls out of R6), or build the Status counts from the panel loads. |
| P2 | M | `gui.go:1286-1290` | Every refresh, including automatic ones after each task, clears `KeyDoctorResult`/`KeyConfig` and re-runs `brew doctor` and `brew config`, the two slowest reads. | Clear them only on an explicit `R` press. |
| P3 | L | `commands.go:1789-1790`, `1818-1828` | The Outdated panel reads formula and cask outdated lists one after the other, and Services runs `List` then `Info` the same way. The worst case is twice `brewReadTimeout` (60 s). | Run them concurrently, as `statusDashboard` already does. |

### Render path (`View()` runs on every message)

| # | Sev | Where | Finding | Fix |
|---|---|---|---|---|
| P4 | H | `render.go:572` (`renderSearchInfo`) | `json.Unmarshal` of the raw `brew info` output runs on **every frame**. | Parse once when `SearchDoneMsg`/the info result lands, and store `*pkgInfo` plus an error. |
| P5 | H | `operation.go:182,233`, `motion.go:217`, `task.go:83`, `task/manager.go:150-200` | `parseSteps(op.Lines)` re-parses the entire operation output 3–4 times per frame (`renderOperation`, `operationBar`, `operationBody`, `motionProgress`). Every output line is its own `TaskOutputMsg` and so its own Update+View, which makes a long `brew upgrade` cost O(n²). `Lines` also grows without limit. | Keep `steps/total/begun` on `Operation` and update them in `AppendLine`. Cap `Lines`. Have `readChunk` drain up to N buffered lines into one `TaskOutputMsg{Lines []string}`. |
| P6 | H | `depviews.go:186` (`renderFiles`), `scroll.go:24,50` | The Files and Used By tabs re-render a table of every line (two lipgloss Renders plus a width measure per cell), re-wrap the whole string, and only then cut it to h rows. `contentLast` repeats this on every PgDn, wheel notch and log poll. Thousands of files means visible lag. | Cache the rendered rows per (`contentKey`, width) and invalidate when `tabContent` changes. |
| P7 | M | `depviews.go:48` (`installedList`) | Builds a map of all installed formulae or casks on every render of Deps, Used By, Tap Formulae and Search info. | Build it once on `DataLoadedMsg` and keep it on the panel. |
| P8 | M | `render.go:156`, `command_log.go:99` | `logHeight` calls `commandLog.Entries()`, which locks a mutex and **copies the whole ring**, only to test for zero length. `screenLayout()` runs about 4 times per View and also on each Update via `motionSidebar`. `CommandLog.View` copies everything and then slices the last h entries. | Add `Len()` and `Last(n)`. Compute `screenLayout` once per View and pass it down. |
| P9 | M | 9 files, ~28 sites | `lipgloss.NewStyle().Foreground(style.X)` is built per render, often per row: Muted ×10, Success ×5, Warning ×4. | Add `MutedText`, `SuccessText`, `WarningText`, `ErrorText` and `TextStyle` to `style.ApplyTheme` and reuse them. `OutdatedBadge`/`InstalledBadge` already match two of them. |
| P10 | M | `sidebar_rows.go:64`, `box.go:52` | Per sidebar row, every frame: up to 3× `cell(" ", lipgloss.NewStyle())`, a `.Background()` style copy per cell, `edge.Render("│")` twice, and repeated `lipgloss.Width` calls. | Render the edge glyphs and separator once per call, and build the highlight styles once. |
| P11 | L | `gui.go:1394-1400` (`needsTabFetch`) | Allocates a nested `map[PanelID]map[int]bool` on every call (every `j`/`k` press and every data landing). | Use a package-level var or a `switch`. |
| P12 | L | `help.go:221`, `presentation/detail.go:54`, `motion.go:122` | `helpRows` rebuilds every help page per frame and per scroll key. `Info` renders `Detail(left)` twice, once only to measure. `toastName` uses `fmt.Sprintf("%p")` plus string joins several times per toast per View. | Cache help rows per width. Measure without the second render. Key toast motion on the `*Toast` itself. |
| P13 | L | `gui/formulae.go:12` | `sortFormulae` calls `strings.ToLower` twice per comparison. | Lowercase the keys once, then `slices.SortStableFunc`. |

### brew package

| # | Sev | Where | Finding | Fix |
|---|---|---|---|---|
| P14 | M | `brew/runner.go:192-204` (`lineWriter.Write`) | Each write does `w.partial + string(p)` followed by `strings.Split`. Long output without newlines (brew's `\r` progress bars) is copied again on every write, so the cost is quadratic, and the progress bars pile up into one huge line. | Use a `bytes.Buffer` and `bytes.IndexByte`. Treat `\r` as a line break too. |
| P15 | L | `casks.go:50-104`, `formulae.go:94-104` | `flexStrings.UnmarshalJSON` decodes into `[]interface{}`, which builds a `map[string]interface{}` for every dependency of every installed formula. `bottleJSON` decodes the whole per-platform URL map only to test `len > 0`. | Decode into `[]json.RawMessage` and branch on the first byte. Decode `files` as `map[string]json.RawMessage`. |
| P16 | L | `runner.go:72-148` | `strings.Join(args, " ")` is computed up to 4 times per call. | Compute it once (see R10). |

---

## 2. Code reuse and duplication

### `internal/gui`

| # | Sev | Where | Finding | Fix |
|---|---|---|---|---|
| R1 | H | `commands.go` (17 functions) | The same tail is copy-pasted 17 times: `Enqueue`, the "Queue full" toast, the "Queued behind running task" toast, then `RunNext`. It appears in `executeTrustAction`, `trustItemCmd`, `executeUntap`, `executeRepair`, `executeBrewfileAction`, `doMutation`, `executeTapAdd`, `runServiceTask`, `togglePin`, `executeServicesCleanup`, `runVulns`, `runMissing`, `runDoctor`, `brewCleanup`, `runAutoremove`, `executeCleanup` and `executeAutoremove`. | Add `func (m *Model) enqueue(t *task.Task) tea.Cmd`. |
| R2 | H | `commands.go` | Two task `Run` bodies are repeated everywhere: the single-error body (`errCh := make(chan error,1); errCh <- f(ctx); return closedCh(), errCh, nil`, ~9 copies) and the stream body (~8 copies). | Add `errTask(id, title, func(ctx) error)` and `streamTask(id, title, func(ctx) (<-chan string, <-chan error))`. With R1, `executeCleanup`, `executeAutoremove`, `brewCleanup` and `runAutoremove` (4 × ~25 lines) each become one line. |
| R3 | M | `gui.go:426-436` and `833-843` | The `brew update` task is built twice, identically, including the reset of `isUpdating` when the enqueue fails. It also calls `Runner.ExecuteStream(ctx, "update")` directly, which skips `DiagnosticsWrite.Update` and its cache invalidation. | Add one `m.startBrewUpdate()` that calls `DiagnosticsWrite.Update`. |
| R4 | M | `commands.go:1199-1229`, `1583-1616` | `runDoctor`, `runMissing` and `runVulns` each hand-roll a goroutine with `select { case out <- x: case <-ctx.Done(): }`, and `runVulns` does it with a different buffering strategy. | Add one `linesTask(ctx, fetch, emptyMsg)` helper. |
| R5 | L | many `commands.go` sites | The `panel.selected >= len(panel.items)` check, `extractPackageName(...)` and the `name == ""` guard are repeated about 7 times, although `selectedItemName(p)` exists. In `doMutation`, four arms repeat `if casks { CasksWrite.X } else { FormulaeWrite.X }`. | Use a single `p.selectedName()` (this also fixes C1). Pick the writer once through a small `packageWriter` interface. |
| R6 | M | `panel.go:236`, `panel.go` select helpers | `selectedOutdatedFormula` has the same body as `selectedFormula`. `selectServiceNamed` and `selectTapNamed` are copies of each other. | Delete the duplicate getter. Add one generic `selectNamed`. |
| R7 | M | `gui/icons.go:31` vs `presentation/formatters.go:156`; `render.go:616` vs `formatters.go:163` | `withIcon` is defined twice, and so is `boolStr`/`yesNo`. | Export each once from `presentation` (or `style`). |
| R8 | M | `box.go:99` `clipLine`, `table.go:199` `cutCell`, 6× raw `ansi.Truncate(…,"…")`, `gui.fitLine` vs `modal.fitLine` | At least three truncation helpers do the same job. Two functions named `fitLine` do different things: one pads, the other cuts. | Use one `Clip`/`Fit` pair in a small shared text-helper package. Rename one of the `fitLine`s. |
| R9 | M | `help.go` `buildHelp`, `footer.go:52` `modalHints`/`panelHints`, the key handlers, `modal/*` `keyHint` | Key bindings are kept in sync by hand in three or four places, with two identical structs (`helpItem`, `keyHint`). | Use one keymap table that drives help, the footer and dispatch. Each modal exposes `Hints()`. |
| R10 | L | `mouse.go:194` | `handleTabBarClick` re-renders each tab to measure it, although `tabWidth()` (`render.go:263`) exists and a test keeps it equal to the rendered tab. `handleMouseWheel` recomputes `sidebarWidth` when `l.sideW` is already in scope. | Use `tabWidth` and `l.sideW`. |
| R11 | L | `modal/menu.go:75` | Re-implements `leftBlock` (`modal/frame.go:48`) inline. | Call `leftBlock`. |
| R12 | L | `gui.go:893-913`, `955-960` | Keys `"1"`…`"7"` are seven near-identical switch arms, and `[`/`left` and `]`/`right` are duplicated. | `case "1",…,"7": switchPanel(PanelID(s[0]-'1'))`, and combine the case labels. |
| R13 | L | `gui.go:552-600` | The leaves filter is inline logic in the `Update` switch. Three `MutationResultMsg` branches differ only in the toast text. | Move the filter to a `p.toggleLeaves()` method. Pick the toast first, then return once. |
| R14 | L | `gui.go:1264` `invalidatePanelCache` | Re-encodes knowledge of which cache keys belong to which panel, which `brew.InvalidateGroups` already holds. | Look the keys up from the brew groups. |

### `internal/brew`, `internal/config`

| # | Sev | Where | Finding | Fix |
|---|---|---|---|---|
| R15 | M | `formulae.go:118-149`, `casks.go:127-164`, and the services, taps, trust and doctor readers | Every reader repeats get from cache, type-assert, run, set with TTL. The formulae and casks `Outdated`/`fetchOutdated` methods match line for line. | Add a generic `cached[T](c, key, ttl, sf, fetch)`. It replaces the unused `TypedCache` (D1) and gives every read singleflight protection (P1). Add a `streamWrite(op, args...)` for the writers. |
| R16 | M | `brew/cache.go` `InvalidateGroups` | Half the writers use the groups and half call `Invalidate(...)` by hand. `"cleanup"` and `"doctor"` are never looked up. The map is an exported mutable global. | Use the groups everywhere (this fixes C8). Unexport the map. |
| R17 | L | `services.go:50-102` | `List` and `Info` share the same decode loop: `[]json.RawMessage`, decode each entry, skip bad ones, require a name. | Add `decodeEach[T](data, ok func(T) bool)`. |
| R18 | L | `sudo.go:56-69` vs `services.go:160-179` | `needsTerminal` and `refusedForRoot` are the same function with different marker lists. | Add `errorMentions(err, markers)`. |
| R19 | L | `formulae.go:226-255`, `doctor.go:131-148` | "Split output into trimmed, non-empty lines" is written three times (`Leaves`, `Uses`, `parseMissingOutput`). | Add `nonEmptyLines([]byte) []string`. |
| R20 | L | `runner.go:72-148` | `Execute` and `ExecuteStream` duplicate the command setup (env, stdin, process group). | Add `newCmd(ctx, args)`. |
| R21 | L | `config.go:121-138` and the TTL defaults in `loadFile` | The env-switch parsing for Animations and Preload is copy-pasted, and so are six TTL default blocks. TTL `<=0` guards also happen three times (config, each reader's `SetCacheTTLs`, and `Client.SetCacheTTLs` with a type assertion, because `TapsReader` lacks `SetCacheTTLs`). | Add `boolEnv(name, *bool)` and a defaults loop. Add `SetCacheTTLs` to `TapsReader` and call it directly. Guard once. |

---

## 3. Dead code

Reproduce with the tools installed in the scratchpad, or with:

```bash
go run golang.org/x/tools/cmd/deadcode@latest ./cmd/...   # unreachable from main
go run honnef.co/go/tools/cmd/staticcheck@latest ./...
```

### D1 — unreachable from `main` (deadcode) or unused (staticcheck) ✅

| Symbol | Where | Notes |
|---|---|---|
| `dependencyJSON` | `brew/formulae.go:90` | Unused type (U1000). |
| `TypedCache`, `NewTypedCache`, `.Get`, `.Set` | `brew/cache.go:123-149` | Used only by `cache_test`; `Set` is never called even there. Replace with R15, or delete. |
| `ParseTrustStatus`, `ParseServiceStatus` | `brew/types.go:124`, `:206` | Used only by tests. |
| `InvalidateAll` | `brew/cache.go` | Used only by tests. |
| `Model.newestToast` | `gui/toasts.go:54` | Unreachable. |
| `panelData.icon` | `gui/panel.go:71` | Never set or read (U1000). |
| `panelData.width` | `gui/panel.go` | Written at `gui.go:629`, never read. |
| `panelData.active` | `gui/panel.go` | Written at `gui.go:1358`, read only by a test; duplicates `m.activePanel`. |
| `sfCall.waiters` | `brew/singleflight.go` | Read only by `singleflight_test`. |

### D2 — interface methods with no production caller

`Formulae.Get`, `Casks.Get`, `Taps.Get`, `Trust.GetTapTrustStatus`, `Search.SearchDesc`, `TapsWrite.TapWithURL`, `CasksWrite.Pin/Unpin`, `DiagnosticsWrite.Update` (production bypasses it, see R3), `Diagnostics.Version`.

- **`CasksWrite.Pin/Unpin` should go regardless.** `brew pin` works only on formulae, so these always fail.
- Revive `Formulae.Get`/`Casks.Get` for C6 rather than deleting them.

### D3 — production code that exists only for the broken name round-trip

`presentation.FormatFormula`, `FormatCask`, `FormatTap`, `FormatService` and `FormatOutdated*` build the strings stored in `panel.items`. The sidebar never draws them, because `typedRow` covers every row. They exist only to be parsed back by `extractPackageName` (C1), and their `width` parameter is unused. After C1 is fixed, `items` can hold bare names, or be removed in favour of the typed slices.

### D4 — unused fields and types

| Symbol | Where |
|---|---|
| `ProgressLineMsg` and its `Update` case | `gui/task.go:71`, `gui.go:379`. Left over from the removed ProgressModal; only `gui_test.go:762` builds it. |
| `MutationResultMsg.Type`, `DepCheckMsg.Name` | `gui/messages.go`. Set but never read. |
| `thirdPartyTaps` | `commands.go:2011-2017`. Counted, never read. |
| `statusDashboard`'s `[]error` return | `commands.go:1848,1880`. The only production caller discards it; the errors are already in `summary.Problems`. |
| `Cask.ConflictsWith`, `Tap.LastCommit`, `BrewConfig.Arch`, `Formula.Size` | `brew/types.go`. Never populated. |
| `task.Status.String()`, `Task.Status` | `gui/task/types.go`. Read only by tests. `RunNext` also writes `Status` outside the lock. |
| Theme styles `ActiveBorder`, `InactiveBorder`, `HintKey`, `HintDesc`, `DocStyle` | `gui/style/theme.go:25-46`. No references at all (`ActiveBorder` appears only in a stale comment at `mouse.go:170`). |
| `PanelTitle`, `ActivePanelBg`, `LogBorder`, `CurrentTheme` | `gui/style/theme.go`. Used only by tests. |

### D5 — unreachable branches

| Where | Why it can't run |
|---|---|
| `commands.go:216-225` | The Taps arms of `executeMenuAction`. Every Taps menu sets `pendingAction = "trust-menu"` and is handled earlier; only the Brewfile menu reaches here. |
| `gui.go:822-827`, `967-971` | `searchInput.Focused()` checks inside `esc`/`enter`. A focused search input has already returned at lines 777-800. |
| `commands.go:1273-1296` | `renderCachedTabContent` always returns `""` and only stores a "ready" flag. If `formulaDepsTab` is dropped from `needsTabFetch` (it's drawn at render time), the matching arm in `fetchTabContentCmd` (1328-1333) becomes dead too. |
| `casks.go:299-318` | `_ = key` is a leftover. Ranging over a map also gives random artifact order between runs. |
| `runner.go:45-66` | The `c != "/bin/brew"` sentinel for an unset `HOMEBREW_PREFIX`, which also leaks into `BrewNotFoundError.Searched`. Add the candidate only when the variable is set. |

### D6 — test support in production files

`brew.MockRunner`/`NewMockRunner` (`runner.go:240-285`) ships in the binary. It is used by tests in five packages (brew, gui, gui/flows, gui/testutil, app), so renaming it to `_test.go` won't work. Move it to an `internal/brew/brewtest` package.

---

## 4. Code quality (other)

| # | Sev | Where | Finding | Fix |
|---|---|---|---|---|
| Q1 | M | `gui/commands.go` (2062 lines), `gui/gui.go` (1467 lines) | Two oversized files. `Update` is one huge switch mixing key dispatch, data landing and task events. R1, R2, R9 and R12 would shrink them a lot. | After the dedup, split by concern: `keys.go`, `land.go`, `tasks_*.go`. |
| Q2 | L | `Model` methods | Value receivers on a ~60-field struct. Task `Run` closures capture all of `m` only to use `m.client`. | Capture `client := m.client` and the flags needed, not `m`. |
| Q3 | L | `commands.go:151-160` | The `"mutation"` confirm-label map is rebuilt per confirm and silently falls back to "Install". | Store the label in `pendingLabel` when `confirmMutation` runs. |
| Q4 | L | `commands.go:1450` | `confirmUninstall` runs `brew uses` for casks too, which wastes a call and gives a meaningless result. | Skip it on `PanelCasks`. |
| Q5 | L | `commands.go:1340,1454` | Strings are built with `+=` in loops. | Use `strings.Join`. |
| Q6 | L | Casks tabs (`panel.go:61`, `render.go:409-440`) | Bare `0/1/2` tab indices where every other panel uses named constants. The constants are also split between `panel.go` and `details.go`. `initPanels` hard-codes `make(…, 7)` and keeps a leftover `loading := false`. | Name the constants and keep them in one file. |
| Q7 | L | `cmd/lazybrew/main.go:13,21` | `"1.0.0-rc1"` is repeated as a sentinel. | `var version = ""` with a build-info fallback. |

---

## 5. Tests (light pass)

- **File naming by review round:** `review_r1_test.go` … `review_r4_test.go`, `runner_stream_r1/r2_test.go` and `regression_test.go` are named after the review round, not the behaviour, which makes coverage hard to find. Over time, fold them into the test file for the feature they cover.
- **`gui_test.go` is 2.5k lines.** Splitting it by panel or feature would match the source layout.
- **staticcheck noise in tests:** 8× SA4006 (an assigned `m`/`updated` value is never used: `gui_test.go:1752,2189,2212`, `help_esc_test.go:43`, `preload_test.go:172,321`, `review_r4_test.go:77`, `services_logs_follow_test.go:128`, `sidebar_rows_test.go:146`). These are usually assertions someone meant to write and forgot, so check each one. There is also an empty branch at `style/theme_test.go:222` (SA9003).
- **Duplicated helper:** `assertAnError` is defined in both `gui_test.go:20` and `task/manager_test.go:10`, and is misnamed for an error var (ST1012). It is trivial, but would move naturally into the `brewtest` package (D6).
- **Gaps that matter:** `LoggingRunner` has no tests (C2). `parsePackageInfo`'s fixture only has `installed`, so C6 slipped through. No test uses a package name longer than 22 characters (C1).

---

## Suggested order

1. **C1, C2, C6.** User-visible wrong behaviour. Each needs a regression test.
2. **C3/C4, C5, C7–C9.** Latent bugs, all small.
3. **P4, P5, P6.** The render-path hot spots. This is where lazybrew will feel slow on big installs or long upgrades.
4. **R1 + R2 + R3.** The biggest dedup win (several hundred lines out of `commands.go`), and low risk.
5. **R15/R16 + P1/P2.** Brew-side `cached[T]` with singleflight everywhere. This removes `TypedCache` and halves the brew calls per refresh.
6. **Dead-code sweep (D1–D6).** Mechanical, and best done as one commit after the refactors above, so it doesn't reopen the same files twice.
