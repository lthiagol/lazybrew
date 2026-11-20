# runner fix r1: v2-01-charm-v2

Commits, in finding order:
- F-r1-01 6eec052 fix(ci): build with the Go version go.mod declares (F-r1-01)
- F-r1-02 7f32f2f test(gui): pin paste routing to search and the input modal (F-r1-02)

F-r1-01 (major): both setup-go steps now read `go-version-file: go.mod`.
- `grep -n go-version .github/workflows/*.yml` → wip-ci.yml:19 `go-version-file: go.mod`, stable-ci.yml:23 `go-version-file: go.mod`; no pinned go-version remains.
- `cache: true` and all other steps unchanged. Only those two lines differ.

F-r1-02 (minor): three tests in internal/gui/paste_test.go, all PASS under `-race`.
- (a) TestPasteIntoFocusedSearchInserts: focused search input value becomes "lazy".
- (b) TestPasteOnSearchPanelStartsSearchInput: unfocused search on the Search panel is focused and value becomes "brew".
- (c) TestPasteReachesOpenInputModal: open InputModal's Result().Value becomes "homebrew/cask".
- Mutation check: with the PasteMsg case disabled, (a) and (b) FAIL. (c) still passes, because an unmatched PasteMsg falls through to the active-modal forward at the end of Update. The test pins the behaviour either way, but does not isolate the new case.
- No production code changed; no paste bug found.

Gate (real output):
- `make build` → exit 0.
- `make lint` → exit 0.
- `CGO_ENABLED=1 go test ./... -race -count=1` → exit 0, every package ok (internal/gui 9.876s, internal/gui/modal ok).
- `make cover-check` → exit 0: brew 71.8≥62, presentation 96.2≥91, gui 70.9≥36, modal 46.9≥41.
- `grep -rho "^func Test\w*" --include=*_test.go . | wc -l` → 410 (was 407; +3 paste tests, none removed).
- `gofmt -l .` → no output.

AC-03 frame equivalence re-run (no production file touched, re-run anyway):
- `diff -r work/verification/v2-01-charm-v2/baseline $SCR/frames-r1 --exclude='*.ansi'` → no output, `frame diff exit 0`. 18 of 18 frames identical; TestDump PASS.

Worktree clean at report.
