# runner fix receipt r2 — v2-03b-sidebar-lists

- F-r2-01 (major) `3fefcf8`: g/G top/bottom only outside Search; on an unfocused Search panel they type into the search box as before. Tests `TestLetterKeysOnUnfocusedSearchTypeToSearch` (g and G focus the input with that letter) and `TestScrollLettersOutsideSearch` (g and G still scroll elsewhere). Mutation dropping the guard fails the first.
- F-r2-01 follow-up `b501d1f`: help row is `Home/End` / `Top/bottom (g/G outside Search)`. The first wording wrapped in a 45-cell help column.
- F-r2-02 (minor) `9c91b5d`: `Its 1 formula and 1 cask follow …`, plurals otherwise. `TestTrustStoryPluralises` (one formula, two formulae, zero casks, one cask).
- Frames `b501d1f` re-captured (commit `(docs)` after it): only 08-help changed, at both sizes; `result.md` has a fix-round r2 section.

Gate on HEAD: `make build` ok; `make lint` ok; `CGO_ENABLED=1 go test ./... -race -count=1` every package ok; `make cover-check` brew 72.9, presentation 96.2, gui 78.1, modal 50.5, all floors met. Tests: 633 PASS.

Worktree clean at report time. Branch `wip`, not pushed. No brew write command run.
