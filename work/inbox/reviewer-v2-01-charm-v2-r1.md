verdict: changes-required

# reviewer: v2-01-charm-v2 — round r1

Reviewed `5f9c31e..53b5faa` on `wip`. AC commits: 1861585 (AC-01), 0a89f8e (AC-02), 0f65ba6 (AC-03), 10aade4 (AC-04).

| AC | result | evidence (run by the reviewer) |
|---|---|---|
| AC-01 | pass | V2 greps: no output. go.mod has only the four v2 modules plus transitive deps; v1 deps are gone. Test count is 407 (≥ 406). The only `t.Skip` calls are the 3 integration skips that were already there. Each of the 3 removed assertion lines has a ported replacement of equal strength (`View().Content`, `color.Color` maps). Space still reaches the batch toggle: uv `Key.String()` maps Text `" "` to `"space"`, and `TestSpaceTogglesOutdatedBatchSelection` passes. |
| AC-02 | pass | `gui.go` `View()` sets AltScreen, WindowTitle "lazybrew", and MouseModeCellMotion only when `cfg.GUI.Mouse` is true. `main.go` is now `tea.NewProgram(m)` with no options. `TestViewDeclaresTerminalSetup` checks mouse on and off. The runner put the View plumbing in AC-01, which the brief allows. |
| AC-03 | pass | Reproduced: `DUMP_DIR=$D go test ./internal/gui/dump/` gives ok, and `diff -r --exclude='*.ansi' baseline $D` exits 0 over 18 frames. Comparing the v1 and v2 harnesses shows only API changes (import, `KeyPressMsg`, `View().Content`). The screens, keys and sizes `{140,40},{80,24}` are the same. The baseline has not changed since 207be26. The harness does catch size changes: run on 1861585 (before the size fixes), all 18 frames differ. `internal/gui/dump` was removed afterwards. |
| AC-04 | pass | DESIGN.md has the stack line and the 2026-11-19 decision row. The STATE.md row matches the brief exactly. AGENTS.md names no v1-only API (`program.Send` still exists in v2), so leaving it unchanged is correct. |

Gate (V1), on HEAD: `make build`, `make lint`, `CGO_ENABLED=1 go test ./... -race -count=1` and `make cover-check` all exit 0. Coverage: brew 71.8, presentation 96.2, gui 70.3, modal 46.9, all floors unchanged. `gofmt -l .` prints nothing.
Every commit builds: 1861585 and 0a89f8e were each extracted with `git archive`, and `go build`, `go vet` and `go test ./...` passed on both.
Invariants (V7) hold: nothing changed under `internal/brew/**` or `scripts/`, no file was moved or renamed, `program.Send` appears only in a doc comment, and the worktree is clean.

## Findings

F-r1-01 (major) — `.github/workflows/wip-ci.yml:19`, `.github/workflows/stable-ci.yml:23` — CI still pins
`go-version: '1.24'`, but `go.mod` now declares `go 1.26.0`. The orchestrator ruling at the end of
`work/questions.md` says both workflows switch to `go-version-file: go.mod` inside this slice, because
"a slice whose CI cannot build it is not done". Evidence: `grep -n go-version .github/workflows/*.yml`
→ `wip-ci.yml:19: go-version: '1.24'` and `stable-ci.yml:23: go-version: '1.24'`. What the fix must
achieve: both `setup-go` steps use `go-version-file: go.mod` and no pinned `go-version` remains.
`cache: true` and the other steps stay as they are.

F-r1-02 (minor) — `internal/gui/gui.go:582-602` — the new `tea.PasteMsg` case has no test. This case is
what keeps v1 paste behaviour (AC-01.4): v1 delivered a paste as a `KeyRunes` KeyMsg to the active
modal or the search box. Nothing would catch it if this path broke. AGENTS.md change guideline 2 asks
for tests in the same change. Evidence: `grep -rn PasteMsg --include=*_test.go internal` prints nothing,
and the runner receipt says "No unit test covers it yet". What the fix must achieve: unit tests in
`internal/gui` showing that a `tea.PasteMsg{Content: ...}`:
(a) is inserted into the focused search input;
(b) on the Search panel with the input unfocused, focuses the input and inserts the text;
(c) reaches an open `InputModal`'s text field.

## Notes (not findings)

- The runner receipt was committed three times (bab4073, then two trim commits, 547658e and 1221e55).
  The same thing happened before and was accepted as F-03 in lb01. The history is not rewritten, so
  nothing needs doing.
- Real-terminal mouse, keys and paste were not exercised live. This is outside the reviewer gate and is
  covered by the human smoke gate (LB-05).
