# runner lb26-list-filter — receipt

Commits (AC order):
- AC-01 `5abd469` filter state on panelData, leaves and query narrow one list
- AC-02 `44d7475` the `/` input on Formulae and Casks
- AC-02 fix `381ac43` drop an unused assignment in the filter test (SA4006)
- AC-03 `c298ca6` `/` filter in the key tables, help, footer, README, flows

## AC-01 — filter state on panelData
- Red first: `go test ./internal/gui/` → `p.setQuery undefined (type *panelData has no field or method setQuery)`.
- Green: `filter_test.go`, 9 tests (narrowing in step, indexing, no match, leaves + query, cursor by name, landData reapply/clamp).
- `filter.go`: `unfiltered*` hold the full list while narrowed; `narrow`/`restore`/`reapply`; `toggleLeaves` moved here.

## AC-02 — the `/` input
- Red: `go vet` → `filter_ui_test.go:39:42: m.filterInput undefined`.
- Green: `filter_ui_test.go`, 12 tests: `/` on Formulae/Casks, `/` elsewhere and `7` then `/`, typing, keys typed while focused (ctrl+c quits), up/down, Enter, Esc (in input and active), no-match, panel switch, click mapping, and frames at 80x24 and 140x40 in both themes (`assertFrameSize`).

## AC-03 — keymap, help, footer, README, flows
- Keymap test passes (`keyFilter` in `keyTable`). Footer shows `/ filter` on Formulae and Casks.
- `internal/gui/flows/filter_test.go`: `TestFilterUninstallsTheFilteredRow` (asserts "Uninstall fzf?" on View and the `uninstall fzf` stream call on the mock), `TestFilterEscClearsTheList` (asserts the final View).
- Golden or expectation changes: none. No existing test expectation was edited.

## Gate
- `GOFLAGS=-p=1 make ci` → exit 0 (fmt-check, vet, `-race` tests, cover-check, build). gui 91.2% (floor 83), floors not lowered.
- `GOOS=darwin go build ./...` → OK.
- `deadcode ./cmd/...` → no output.
- `staticcheck ./...` → only pre-existing hits: `keymap_test.go:39,439` (ParseDir, SA1019), `launch_check_test.go:180,220,453` (SA4006).
- Per-commit `make test-fast PKG=./internal/gui/...` passed at each commit.
- No brew command was run; the TUI was not launched.

## Behaviour changes beyond the brief's wording
- A reload now re-applies the leaves filter. Before, the list went full while `leavesActive` stayed on, and `L` off restored a stale list.
- The installed index and the untap count read the full list, not the narrowed one. This also fixes dep marks under leaves.

## Seen, not done
- `7` then `/` on Search with an empty box: the box already has focus after `7`, so `/` is typed into it. This is existing Search behaviour. The brief's path reaches the Search box only when results exist.
- The leaves set is a snapshot. A package installed while `L` is on is not in it until `L` is toggled again.
- Real-terminal feel and mouse in a real terminal are not verified; the click mapping is unit-tested through `handleMouseClick`.

Questions filed: none.

## Fix r2 — F-r1-01, F-r1-02
- `2596040` F-r1-01: `7` then `/` no longer types a slash into Search; `/` on an empty focused box only keeps it focused.
- `7489dd6` F-r1-02: the header badge counts `fullFormulae()`/`fullCasks()` while a query or leaves narrows the list.
- Revert evidence (red before each fix): `filter_ui_test.go:61` `box holds "/"`; `:71` `value "/"`; `:263` `got "… 2 formulae"` (expected `5 formulae`).
- Gate: `GOFLAGS=-p=1 make ci` exit 0, floors met (gui 91.2% ≥ 83). `GOOS=darwin go build ./...` ok. `deadcode ./cmd/...` empty. `staticcheck ./...`: only the 5 pre-existing hits.
- No brew command run; TUI not launched. F-r1-03 (info) not addressed.
