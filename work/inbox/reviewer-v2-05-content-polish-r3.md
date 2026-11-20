verdict: ok

# reviewer — v2-05-content-polish r3

Reviewed bcf274a..f8fdaa1: fix commits 0831cd8 (F-r2-01), da7d80e (F-r2-02), c9acd20 (F-r2-03), receipt f8fdaa1. One commit per finding, subjects name the ids.

| AC | verdict | evidence |
|---|---|---|
| AC-01 | met | as r1 |
| AC-02 | met | as r1; cask fixture is real `brew info` output |
| AC-03 | met | cards, Health second, `doctor unavailable: <reason>`; mutations M1–M4 caught |
| AC-04 | met | as r1; Esc order tested (F-r1-03) |
| AC-05 | met | upgrade-all n/total right for formulae+cask and dependents runs (below) |
| AC-06 | met | harness re-run: 54 `after/` files byte-identical, 36 frames exactly W×H; result.md current |

Gate (my run, HEAD f8fdaa1): `make build` ok · `make lint` ok · `CGO_ENABLED=1 go test ./... -race -count=1` all ok · `make cover-check` brew 76.1 ≥ 62, presentation 98.2 ≥ 91, gui 81.9 ≥ 36, modal 88.3 ≥ 41. Worktree clean; `internal/gui/dump` and scratch worktree removed.

## Earlier findings

- **F-r2-01 — fixed.** Total is now the sum of every `outdated package(s):` header and every dependents header; headers and `<x> dependency: <y>` lines are not package steps. My r2 probes, re-run unchanged against HEAD in a throwaway worktree: formulae+cask, firefox running → `2/3` (was `1/1`), success `3/3`; formula+dependent, zsh running → `1/2` (was `1/1`), success `2/2`. Test transcripts checked against Homebrew's source on this machine (read only): `cmd/upgrade.rb:458` and `cask/upgrade.rb:107` (`Upgrading N outdated package(s):`), `upgrade.rb:246` and `cask/upgrade.rb:434` (`Upgrading <name>`, version on the next line), `upgrade.rb:400` (`Upgrading N dependent(s) of upgraded formula(e):`), `upgrade.rb:424` (`Checking for dependents of upgraded formulae...`), `formula_installer.rb:1013` (`Upgrading <f> dependency: <d> (…)`). All match. The receipt cites the cask header as `cask/upgrade.rb:102` (the method line); the string itself is on :107. Cosmetic.
- **F-r2-02 — fixed.** `oneOutdated` = `Outdated: +1 package\b` in five tests. M4 (`"packages"` for 1) re-run: FAIL in `TestStatusDashboardReportsCachedOutdatedCount`, `…ExplicitLoadShellsOutdatedOnce`, `…ExplicitLoadShellsOneCallPerOutdatedClass`, `TestRefreshMsgWithLoadOutdatedShellsOutdated`, `TestStatusRefreshKeepsOutdatedCache`.
- **F-r2-03 — fixed.** result.md lines 17 and 42 give the current order (Homebrew, Health, Packages, Taps, Services) and the 80×24 fold (Health rows 9–11, Services below); they match the frame.

## New problems in the fix commits

None found. One limit, not a defect: when brew upgrades an outdated dependency inside another package's step (`Upgrading A dependency: B`, formula_installer.rb:1013), B is in the header count but has no step of its own. The bar then reads low mid-run and jumps to total/total on success. It never reads full early. The receipt notes a similar "one short" case.

## Notes (no fix asked)

- The fix-r2 receipt is 31 lines; the fix brief capped it at 25 and asked for the count to be checked before the commit. It was committed once, as asked. I am leaving it to the orchestrator; it is not a product finding.
