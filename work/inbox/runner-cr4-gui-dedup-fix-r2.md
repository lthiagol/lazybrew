# runner: cr4-gui-dedup fix r2

Commit: 508df05 cr4 fix F-r2-01: footer and help checked apart in the keymap test
(test file only; no production code, footer or help frame changed).

## F-r2-01 — footer and help checked apart

- Rule (`onFooterAndHelp`): for each key a panel answers, if some table entry for
  it has a footer word, one such entry must be on the footer; if some entry has a
  help line, one such entry must be on the help; the key needs one of the two.
- `footerExempt` (keymap_test.go), checked on the help only, reason in the file:
  - Status: `j k down up` (keyScroll, help-only); `[ ] left right` (stTabs, help-only).
  - Outdated, Taps: `[ ] left right` (tabs exist in `panelTabs`; Global help lists keyTabs).
  - Search: `[ ] left right` (no tabs; the keys do nothing there).
- Stale check: an exempt key the panel does not answer, or the footer shows, fails.
  Shown with a temporary Search `j` exemption: `Search: footerExempt lists "j",
  but the footer shows it; drop it from footerExempt`.

## Mutations (each run, restored, `git status` clean after)

| # | change | result |
|---|---|---|
| M1 | stDoctor off Status footer | FAIL keymap_test.go:357 `Status: the dispatch answers "d"` |
| M2 | fFetch off Formulae help | FAIL TestHelpIsBuiltFromKeyTable, :357 Formulae "F" |
| M3 | sFetch off Search help | FAIL TestHelpIsBuiltFromKeyTable, :357 Search "F" |
| M4 | n off ConfirmModal.Hints | FAIL TestModalHintsMatchUpdate :524 `Update answers "n"` |
| M5 | enter off InputModal.Hints | FAIL TestModalHintsMatchUpdate :524 `Update answers "enter"` |
| M6 | keyNavigate off Outdated footer | FAIL :357 `Outdated: the dispatch answers "j"` |
| M7 | keyPanelTabs off Formulae footer | FAIL :357 `Formulae: the dispatch answers "["` |
| M8 | keyNavigate off Taps footer | FAIL :357 `Taps: the dispatch answers "j"` |

## Gate

- `make test-fast PKG=./internal/gui/...`: pass.
- `make ci`: exit 0 (gofmt, vet, race, all coverage floors OK, build). Output:
  /tmp/claude-1000/-home-agent-coding-lazybrew/a26d2251-29ec-4655-a712-7a990809408f/scratchpad/ci-r2.txt
- `GOOS=darwin go build ./...`: exit 0.
- deadcode and staticcheck: not run (not installed; fetch denied, as instructed).

## Next session must know

- Judgement call: Status does list rows (`statusRows`), so `j/k` move something
  there. The exemption follows the table's help-only keyScroll. The reviewer may rule otherwise.
- Open question filed: `Q-cr4-gui-dedup-r2-footer-tabs` in work/questions.md (tab
  keys missing from the Status, Outdated and Taps footers).
- Seen, not done: none new.
