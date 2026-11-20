# v2-01-charm-v2 — AC-03 frame-equivalence result

Run on the migrated tree (AC-01, AC-02 and the AC-03 fixes), harness
`frame_dump_v2_test.go.txt` copied to `internal/gui/dump/dump_test.go` for the
run and removed afterwards.

## Frame diff (empty = pass)

```
$ cp work/verification/v2-01-charm-v2/frame_dump_v2_test.go.txt internal/gui/dump/dump_test.go
$ DUMP_DIR=$SCR/frames-v2 CGO_ENABLED=1 go test ./internal/gui/dump/ -count=1 -v
=== RUN   TestDump
ok
--- PASS: TestDump (8.08s)
PASS
ok  	github.com/lthiagol/lazybrew/internal/gui/dump	8.084s
$ diff -r work/verification/v2-01-charm-v2/baseline $SCR/frames-v2 --exclude='*.ansi'
$ echo "diff exit $?"
diff exit 0
$ rm -rf internal/gui/dump
```

The diff is empty: 18 frames (9 screens × 140x40 and 80x24) are byte-identical
to the v1 baseline after ANSI stripping. `$SCR` is the session scratchpad
directory; the frames are not committed.

## Test count (brief: ≥ 406)

```
$ grep -rho "^func Test\w*" --include=*_test.go . | wc -l
407
```

Before the port: 406. The one addition is `TestViewDeclaresTerminalSetup`
(internal/gui/gui_test.go). No test was deleted, skipped or weakened. The three
`t.Skip` calls in `internal/brew/runner_integration_test.go` predate this slice.

## go.mod and imports (brief AC-01.1, AC-01.2)

```
$ grep -n "charmbracelet\|charm.land" go.mod
6:	charm.land/bubbles/v2 v2.2.1
7:	charm.land/bubbletea/v2 v2.1.2
8:	charm.land/lipgloss/v2 v2.1.1
9:	github.com/charmbracelet/x/exp/teatest/v2 v2.0.0-20261118134922-ee8f5034ce23
(indirect lines: colorprofile, ultraviolet, x/ansi, x/exp/golden, x/term, x/termios, x/windows)

$ grep -rn '"github.com/charmbracelet/\(bubbletea\|lipgloss\|bubbles\)' --include=*.go .
(no output, exit 1)

$ grep -c 'github.com/charmbracelet/bubbletea \|github.com/charmbracelet/lipgloss \|github.com/charmbracelet/bubbles \|github.com/charmbracelet/x/exp/teatest ' go.mod go.sum
go.mod:0
go.sum:0
```

No v1 module remains in go.mod, including as `// indirect`.
