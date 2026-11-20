# v2-01-charm-v2 — frame-equivalence evidence

`baseline/` holds the **v1** frames: `Model.View()` with ANSI stripped, captured
on `wip` at `5f9c31e` (before the migration) against the real Homebrew on this
machine, at 140×40 and 80×24, for nine screens (status, formulae list/info/deps,
outdated, taps, the uninstall confirm modal, help, search). Two consecutive runs
were byte-identical, so the capture is deterministic.

`frame_dump_test.go.txt` is the harness that produced them (v1 API). It is kept
as `.txt` so it never compiles in the repo.

## How to reproduce against the migrated code

The harness is **not** committed as a test. To run it:

```bash
D=$(mktemp -d)
mkdir -p internal/gui/dump
cp work/verification/v2-01-charm-v2/frame_dump_test.go.txt internal/gui/dump/dump_test.go
# port the harness to the v2 API (key messages, View().Content, …) — keep the
# same screens, the same key sequence and the same sizes
DUMP_DIR=$D go test ./internal/gui/dump/ -count=1
diff -r work/verification/v2-01-charm-v2/baseline "$D" --exclude='*.ansi'
rm -rf internal/gui/dump   # never commit it
```

An empty `diff` is the pass condition. The ported harness used for the receipt
is saved as `frame_dump_v2_test.go.txt` next to this file.
