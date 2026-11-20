# runner: v2-01-charm-v2

Commits, in AC order:
- AC-01 1861585 refactor(gui): port TUI to Charm v2 stack
- AC-02 0a89f8e feat(gui): declare window title and pin terminal setup in View
- AC-03 0f65ba6 fix(gui): keep v1 frame sizes under Lip Gloss v2 (harness .txt + result.md)
- AC-04 10aade4 docs: record the Charm v2 stack in DESIGN and STATE

AC-01 (port, go.mod, no v1 import, tests):
- go.mod: charm.land/bubbles/v2 v2.2.1, bubbletea/v2 v2.1.2, lipgloss/v2 v2.1.1, x/exp/teatest/v2 (@latest).
- `grep -rn '"github.com/charmbracelet/\(bubbletea\|lipgloss\|bubbles\)' --include=*.go .` → no output. `grep -c` of the v1 module paths in go.mod and go.sum → 0 and 0.
- Tests: `grep -rho "^func Test\w*" --include=*_test.go . | wc -l` → 406 before, 407 after. No test deleted, skipped or weakened; the 3 t.Skip calls predate this slice.
- Ports: key literals → tea.KeyPressMsg; sendSpecial takes a rune; shift+tab → KeyPressMsg{Code: KeyTab, Mod: ModShift}; theme_test compares color.Color values (same assertions).

AC-02 (View, WindowTitle, main.go, test):
- Substitution: removing WithAltScreen/WithMouseCellMotion needed the View change to compile, so the View plumbing, the main.go cleanup and the cfg.GUI.Mouse condition all landed in AC-01.
- AC-02 adds WindowTitle = "lazybrew" and TestViewDeclaresTerminalSetup (mouse on and off).
- `go test ./internal/gui/ -run 'TestViewDeclaresTerminalSetup|TestViewNoSmoke|TestMouse' -race` → all PASS.

AC-03 (frame equivalence):
- Harness ported, run from internal/gui/dump/ (then removed), committed as frame_dump_v2_test.go.txt.
- `diff -r work/verification/v2-01-charm-v2/baseline $SCR/frames-v2 --exclude='*.ansi'` → no output, `diff exit 0`. 18 of 18 frames identical. Full transcript in result.md.
- Code fixes, not baseline edits: v2 Width/Height count the border, v1 added it. Adjusted renderBox, the help overlay, the command-log box, and the confirm/input/menu modals.

AC-04 (docs):
- DESIGN.md: stack line names v2 modules; 2026-11-19 decision row; gui.mouse row names View().MouseMode.
- STATE.md: row exactly as the brief gives it. AGENTS.md names no v1 API, unchanged.

Gate (real output):
- `make build` → exit 0.
- `make lint` (go vet ./...) → exit 0.
- `CGO_ENABLED=1 go test ./... -race -count=1` → exit 0, every package ok. Plain `go test -race` fails here with "-race requires cgo" because go env CGO_ENABLED=0.
- `make cover-check` → exit 0 (brew 71.8≥62, presentation 96.2≥91, gui 70.3≥36, modal 46.9≥41). Floors unchanged. Pre-port 207be26 also passes both.

For the next session:
- CI: wip-ci.yml and stable-ci.yml pin go 1.24; v2 needs go 1.26 (go.mod now 1.26.0, plus transitive upgrades). Logged in work/questions.md.
- Added beyond the brief: a tea.PasteMsg case in gui.go Update. v1 sent paste as a KeyMsg, so without it a paste into search would be lost. No unit test covers it yet.
- Not verified live: real-terminal mouse, key feel and paste. Mouse and key paths are covered by the ported unit tests; the frame harness covers rendering only.
- Toast.View stays a string (not a tea.Model); confirm, input and menu return tea.View.
- Untouched: internal/brew/**, scripts/, Makefile, master-plan/ (not read).
