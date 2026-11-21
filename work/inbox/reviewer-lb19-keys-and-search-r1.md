verdict: changes-required

# Review r1 — lb19-keys-and-search

Scope: b1a5c97..28aa5a4 (cee6eb5, e00a573, 3b0677a, 28aa5a4). The review brief was
committed after the slice (a7c1783), so "after the commit that added this brief" would
review nothing; I reviewed the slice commits instead.

| AC | Verdict | Evidence |
|---|---|---|
| AC-01 | not met | Fix right; unit tests red on revert (`search_tab_test.go:19 activePanel = Search, want PanelStatus`, `:56 tab x7 … want Status`). The required View() flow passes with the fix reverted (F-r1-01). |
| AC-02 | met | Real `brew search --formula claude` / `--cask claude` (and `claude-code`, `jq`, `wget`) through `Search`+`MarkInstalled` with real lists: no `unknown`; `claude-code` cask and `jq` formula `installed`. Exit 1 on either read → no rows, no error; timeout → `context deadline exceeded` at 50ms; `user/tap/bar` matches `bar`; `foo` formula+cask marked per type. Reverting flag typing, marks, tap strip, or list-error tolerance each turns tests red. |
| AC-03 | met | Restoring `[` or `pgdown` in keys.go: `TestRetiredTabAndPageKeysDoNothing` fails (`tab 3->2`, `offset 0->5`), plus the keymap tests. h/l switch tabs on panels 1–6; they type into Search (focused and unfocused with rows) and the Add-tap input modal. Grep of README/DESIGN/SMOKE/AGENTS/code: no old keys left (code-review.md is history). |
| AC-04 | met (tests weak) | Args: `uninstall --cask claude-code`, `uninstall jq`, `uninstall --zap --cask claude-code`. The write goes through `doMutation` → `m.tasks.Enqueue`. `x` on a row that is not installed, and `i` on an installed row: toast, no write. Reverting the cask rule, the installed check, the i guard or X-on-Search turns tests red. Two pieces of wiring survive a revert (F-r1-02, F-r1-03). |

Gate: `make ci` exit 0 (gui 90.6% ≥ 83, brew 87.2% ≥ 78, all floors met);
`GOOS=darwin go build ./...` ok.

V6 frames (own harness, 7 panels × every tab + help pages + Search rows, 80x24 and 140x40, dark and light,
b1a5c97 vs 28aa5a4): 112 frames. Outside help and Search, only the footer row differs (`[ ] tabs`→`h/l tabs`,
`x uninstall` on Search). The help changes are the dropped PgUp/PgDn line, `h/l`, and the Search x/X lines (44 rows, was 42).
V7: one commit per AC, in order; floors untouched; no test deleted. The PgDn→End/wheel rewrites and
`BenchmarkFilesTabWheelDown` keep the tests' intent (same Files-tab scroll+draw path). No new goroutines;
`remarkSearch` is a `tea.Cmd` in `boundedRead`.

## Findings

F-r1-01 (major) — internal/gui/flows/navigation_test.go:75 `TestSearchTabCycleFlow` — passes with the
AC-01 fix reverted (`ok …/gui/flows` with the tab/shift+tab cases removed from keys.go). It only checks
"Services", which the sidebar always shows, and it does not cycle all 7 panels both ways. — Fix: a
teatest flow that cycles all 7 panels both ways through a focused Search, checks the active panel in
`View()` after each step (e.g. the main breadcrumb or footer), and fails without the fix.

F-r1-02 (major) — internal/gui/search_uninstall_test.go:190 `TestSearchMarkClearsAfterUninstall` calls
`m.remarkSearch()` itself. Delete `m.remarkSearch()` from the TaskCompletedMsg batch (gui.go:450) and
the whole gui suite stays green. — Fix: run the cmd that the TaskCompletedMsg update returns (drain the batch) and
check the mark clears; it must fail with the wiring removed.

F-r1-03 (major) — internal/gui/gui.go:262 the `SearchMarksMsg` name guard is untested. With
`&& m.searchResults[i].Name == r.Name` dropped, every test passes. V4 needs a stale re-mark
not to land on a newer search's rows. — Fix: a test where marks read for an older result set arrive
after a newer `SearchDoneMsg`; no row of the new set changes its mark.

F-r1-04 (minor) — internal/gui/tasks_packages.go:84 `markInstalled` + internal/brew/search.go:71 — an
unreadable list counts as empty, and `MarkInstalled` then sets `Installed=false`. In the re-mark after a
write, a failed `Formulae.List`/`Casks.List` wipes the `installed` mark from rows that are still installed.
The doc comment ("A nil list marks nothing of its type") says the opposite. — Fix: when a list cannot be read,
leave the marks of that type as they are on re-mark, with a test.

F-r1-05 (minor) — internal/brew/search.go:25 — if one typed read fails with anything but exit 1, the other
read's rows are thrown away: a mock where `--cask` fails with "boom" and `--formula` returns `jq` gives
`res=[] err=boom`. No test pins this behaviour, and the brief does not decide it. — Fix: file a question,
or keep the rows that succeeded; either way, add a test that pins the choice.

F-r1-06 (minor) — internal/gui/help.go:235 `helpKey` — put a `pgdown` case back in the help screen and the
gui suite stays green. `TestRetiredTabAndPageKeysDoNothing` covers only the main content, and AC-03 and V5
name the help screen. — Fix: extend that test to the open help screen (PgUp/PgDn leave `helpScroll` unchanged).

## Notes (not findings)
- Search → info still runs `brew info --json=v2 <name>` with no `--cask` (this predates the slice). For a name
  that is both a formula and a cask, brew may show the formula. Not checked live (there is no such name here).
