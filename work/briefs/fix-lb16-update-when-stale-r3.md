# Fix brief — lb16-update-when-stale, round r3 (cosmetic, orchestrator-verified)

Findings: `work/inbox/reviewer-lb16-update-when-stale-r2.md`, F-r2-01 and
F-r2-02. One commit, staged by name, no behaviour change.

- F-r2-01: move the `internal/config` import into the module group in
  `internal/gui/brew_update_test.go`, `gui_misc_test.go`, `tick_chain_test.go`;
  then scan every `.go` file the slice touched (`git diff --name-only 75ecbba^
  HEAD -- '*.go'`) for the same slip: none may remain. `make fmt`.
- F-r2-02: reword the doc of `TestNonRRefreshEmittersStayCacheOnly`
  (`outdated_count_test.go`): install, uninstall and mutation emitters stay
  plain; R and the brew-update completion are the two that load the count.

Gate: `make test-fast PKG=./internal/gui/...`, then `make ci`. Append "Fix r3"
(≤ 6 lines) to `work/inbox/runner-lb16-update-when-stale.md`, commit
(`inbox: runner fix r3 lb16-update-when-stale`), reply with one done line.
