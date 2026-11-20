# runner receipt — cr3-render-perf, fix round r1

Commits: F-r1-01 b014689 · F-r1-02 d19018e (on a82bdd5).

## F-r1-01 (blocker) — b014689
- Red at a82bdd5 (`-run 'OperationOverCached|OperationOverUncached|LoadingOverCached|MainRowsDuringOperation'`):
  - `operation_main_cache_test.go:28: the overlay should show while the operation runs`
  - `operation_main_cache_test.go:52: the Files rows should be back once the operation closes`
  - `operation_main_cache_test.go:63: the loading state should show over the cached tab`
  - `operation_main_cache_test.go:81: the rows during the operation should be the overlay's`
- Fix: `cachedTab` gives no rows when `opState != nil`, the panel is loading, or no item is selected. Green after.
- Other early returns in `renderContent`: `opState` and `loading` were the hole (now gated); "No package selected" is gated by `tabItemName`; the spinner for missing content is never cached; Deps, Info, Caveats, Casks and Status are not cached tabs. `renderFiles`/`renderUses` read only width, theme (keyed) and installed list (keyed by `installedGen`).

## F-r1-02 (minor) — d19018e
- `toastOffset` built `toast:<id>:dx|dy` per call. `toastMotionNames`, keyed by toast id, is now built in `motionToasts` on each stack layout; a departed toast drops out.
- Red at a82bdd5: `hot_path_test.go:154: reading a toast's offsets allocated 2 times, want none`. Green after.
- `anim.Set` keys unchanged: no `Q-cr3-toastname` needed.

## Frame comparison
96 frames = 24 states × 2 sizes × 2 themes (new: files-op-cached running/failed/closed, files-op-uncached running/closed, files-loading, uses-op-cached-running, toasts).
- dc4df13 vs a82bdd5: **20 differ**: 16 F-r1-01 (4 states × 4 frames) + 4 search-info-error (C11).
- dc4df13 vs d19018e: **4 differ**, all search-info-error (C11): `Parse error: invalid character 'E'…` → `Error: No available formula with the name "zz".`
- Harness, not committed: `/tmp/claude-1000/-home-agent-coding-lazybrew/882b3bc0-dda2-4db4-b3b4-af1965e5cbb3/scratchpad/frames/frames_dump_test.go`.

## Gate
- `make test-fast PKG=./internal/gui/...` ok per commit.
- `make ci` exit 0: fmt-check, vet, `go test ./... -race -count=1` (12 ok), coverage floors met (gui 87.6% ≥ 83%), build. `GOOS=darwin go build ./...` ok.

## Questions and seen, not done
- No questions filed. Seen, not done: the cr3 AC-05 leftovers listed in `runner-cr3-render-perf.md`.
