# Fix brief — cr4-gui-dedup, round r2

Source: `work/inbox/reviewer-cr4-gui-dedup-r2.md`. One finding, one commit,
subject names the id. Briefs `work/briefs/cr4-gui-dedup.md`,
`work/briefs/fix-cr4-gui-dedup-r1.md` and `work/briefs/cr-common.md` still
apply. Everything else on `wip` for this slice is accepted; do not redo it.
Frames, goldens and production code stay unchanged (this is a test fix).

- **F-r2-01 (minor)** — `internal/gui/keymap_test.go:362-373`
  (`onFooterAndHelp`). A key is accepted if ANY table entry with that key
  satisfies "(no footer word or on the footer) and (no help line or on the
  help)". `j/k/down/up` have `keyNavigate` (footer only) and `keyScroll` (help
  only); `[ ] left right` have `keyPanelTabs` (footer) and `keyTabs`/`stTabs`
  (help). `keyScroll`/`keyTabs` sit on the Global help page and always satisfy
  the rule, so the footer entry is never required.
  Fix: check footer and help **separately**. For each key a panel answers: if
  some table entry with that key has a footer word, one such entry must be on
  that panel's footer; if some entry with that key has a help line, one such
  entry must be on that panel's help. Keep the reverse direction (every footer
  and help line names a key the panel answers).
  Mutations, each shown failing and then restored (`git status` clean after):
  M1 drop `stDoctor` from the Status footer; M2 drop `fFetch` from the
  Formulae help page; M3 drop `sFetch` from the Search help page; M4 drop `n`
  from `ConfirmModal.Hints()`; M5 drop `enter` from `InputModal.Hints()` — all
  must still fail; and the two new ones, which must now fail too: M6 drop
  `keyNavigate` (`j/k navigate`) from the Outdated footer; M7 drop
  `keyPanelTabs` (`[ ] tabs`) from the Formulae footer. If the stricter rule
  fails on the current tree for a legitimate reason (a footer that really
  omits a key on purpose), add an explicit, commented exemption with the
  reason, and list each one in the receipt. Do not weaken the rule to make the
  test pass.

Gate: `make test-fast PKG=./internal/gui/...`, then `make ci`.
Receipt `work/inbox/runner-cr4-gui-dedup-fix-r2.md` (at most 25 lines,
committed once), worktree clean, reply with exactly one line:
`done: work/inbox/runner-cr4-gui-dedup-fix-r2.md`.
