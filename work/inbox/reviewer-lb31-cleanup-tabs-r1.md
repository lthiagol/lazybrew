verdict: changes-required

# Review — lb31-cleanup-tabs r1 (baf820b..521537b, incl. fix r1 9170eab)

| AC | Verdict | Evidence |
|---|---|---|
| AC-01 brew layer | ok, one test gap (F-r1-02) | Regexes match brew source here (`cleanup.rb` `cleanup_path`/`prune_content_addressed_cache`/broken link/empty dir, `cmd/cleanup.rb` total, `Cleanup.autoremove` `oh1`, `Pathname#abv`, `Formatter.number_readable` commas). Real `brew cleanup -n` / `autoremove -n` here: empty, exit 0 → `✔ Nothing…`. Live-shaped macOS listing (1,700 / 7,512 files, B/KB/MB/GB, prune, broken link, empty dir, Warning line, autoremove block, total) parses right. Mutation `previewOutput` exit-1 rule → `TestCleanupPreviewExitCodes` fails. |
| AC-02 tabs | ok, one defect (F-r1-01) | Own harness: biggest first, sizes right-aligned, warning colour on top third, header `Would free ~1.7GB`, Unused list + header, ✔ empty lines, one read on open, none on panel switch, `R` before opening reads nothing, `R` on tab re-reads; `y` copy whole, ANSI-free. |
| AC-03 keys | ok | `c`/`A` off tab open+read; on tab confirm (`Run brew cleanup, freeing about 1.7GB?` + `Also removes 2 unused formulae.`; `Uninstall 3 unused formulae?`); `n` runs nothing; `y` streams exactly `cleanup` via task.Manager; flow re-reads after run. Services `c` unchanged. Old preview names gone. Keymap mutation (drop `stCleanup`) → `TestPanelKeysAreOnFooterAndHelp` fails. |
| Fix r1 window | ok | See below. |

**Window (ruling Q-lb31-tabs-80), own harness, real `View()` + real `MouseClickMsg`:**
every panel × every tab × 80x24/100x30/140x40 × both themes: active tab visible, whole names
only, `‹`/`›` exactly on cut sides, no line wider than the terminal; a click on each shown tab
selects it and a click on each marker moves one tab, all through `handleMouse`. Only Status
at 80x24 windows (e.g. Vulns: `‹ Doctor  Missing  Vulns  Cleanup  Unused`). 60x24 is below
`minTerminalWidth` (80): the too-small screen, no tab row (so `TestTabWindowAt60Columns` tests
content width 60, not a 60-column terminal; it's harmless but the name says otherwise).
Frames base vs HEAD (80x24, 140x40, both themes, every panel/tab + help): only Status tab-row
lines and the two Status help lines (`c`, `A`) differ; every other panel byte-identical.
Cleanup note (ruling Q-lb31-cleanup-autoremove): in tab and confirm; run stays plain `cleanup`.

## Findings

F-r1-01 (medium) — internal/gui/cleanup_tabs.go `cutMiddle` — the cut snaps to the last `/`
in the head and the first `/` in the tail however far away they are, so at the default 80x24
it throws away the part that identifies the row. The receipt says "at a `/` where one is near";
the code has no "near". Evidence (`cleanupView` at width 48, prefix `/opt/homebrew`, home
`/Users/alice`): `<prefix>/Cellar/python@3.12/3.12.4_1` → `<prefix>/…/3.12.4_1` (19 of 28 cells
used, formula gone); two different bottles both → `~/Library/….bottle.tar.gz` (rows identical).
With a `broken link`/`empty directory` row (Files column 15 wide), `<prefix>/Cellar/llvm/17.0.6`
→ `<prefix>/…/17.0.6`. Fix must achieve: the cut uses the room it has (snap to `/` only within a
few cells, else cut at the width); distinct paths stay distinct at 80x24 and the formula /
file name survives where it fits; a test at 80x24 width with the paths above.

F-r1-02 (low) — internal/brew/cleanup_preview_test.go — no fixture has a thousands-separated
count, although the review brief names `1,700`. Evidence: mutation `atoiComma` without
`ReplaceAll` → `go test ./internal/brew/ -run 'Cleanup|Autoremove|ParseSize'` still `ok`.
Fix must achieve: a fixture line `Would remove: … (1,700 files, 64.4MB)` asserting
`Files: 1700`; the mutation then fails. (Prune counts come from `Utils.pluralize`: no commas.)

## Gate
- `GOFLAGS=-p=1 make ci`: exit 0 (fmt, vet, race tests, `All coverage floors met`, build).
  Note: setting GOTMPDIR/TMPDIR to a long path makes 3 Services Logs tests fail (path shortened
  in the header); environmental, they pass with the default temp dir.
- `GOOS=darwin go build ./...`: ok. deadcode: none (base: none). staticcheck: same 5 as base.
- One commit per AC + one fix commit; floors/Makefile/scripts untouched; no `program.Send`,
  no new `go func` in handlers.
