verdict: ok

Reviewed fix r3 `22a1a87..b85a0ba` (brief `work/briefs/fix-lb24-copy-and-wording-r3.md`) on `wip`, against the r1 findings.

| AC / finding | Verdict | Evidence |
|---|---|---|
| F-r1-01 80x24 retry row | ok | Real `View()` after `WindowSizeMsg{80,24}`: `✗ Failed: needs sudo · y copy` / `Enter  run in this terminal · Esc  close` / `brew install --cask wireshark-app`, all uncut. 140x40: `✗ Failed: needs sudo · Enter  run in this terminal · Esc  close · y copy`, command below. Width sweep 80…200 at height 30: row, keys and command present, no `…` on either line (< 80 is the "Terminal too small" screen). Revert `operation.go` to `22a1a87^` → `TestFailedRetryRowInTheFrame`, `TestFailedRetryRowOffersYCopyWhereItFits` fail |
| F-r1-02 Logs/Config whole | ok | `fitLines` skips `cutLines` only on the copy's Model (`copying` set on copyText's value copy; the view path is unchanged). Forcing `fitLines` to cut → `TestYCopiesALongLogLineWhole` (613-cell line), `TestYCopiesALongConfigRowWhole` fail |
| F-r1-03 help open | ok | `y` with help open copies nothing, no toast. Dropping `!m.showHelp` → `TestYWithHelpOpenCopiesNothing` fails |
| F-r1-04 README | ok | README.md:74: "On Search `y` types into the box unless an Operation is open" |
| AC-01 / AC-02 | ok | r1 evidence stands; fix r3 changes only the narrow retry row and the Logs/Config copy path |

Gate: `make ci` exit 0 (gui 91.0 % ≥ 83 %, all floors met); `GOOS=darwin go build ./...` ok. staticcheck: the 5 baseline findings (2× SA1019 keymap_test, 3× SA4006 launch_check_test), none new. deadcode: 10 lines on HEAD = 10 on `22a1a87^` (testutil helpers), `-test` none. No `program.Send`, no new goroutines. One commit per finding plus the SA4006 cleanup; the receipt's "deadcode no output" holds only with `-test`.
View: the non-retry error row is the same code as before (only the retry branch moved to `retryFailedRow`); `fitLines` = `cutLines` outside a copy, so only the retry overlay below ~110 columns changes in the frames. The 140x40 retry row is the same as in r1.

## Findings

None blocking.

Notes (no fix needed for this slice):
- With help open on Search, `y` types into the hidden box, as every other letter already does there (also with an overlay open: `a` types). Existing pass-through, not new.
- At 80x24 the footer has no `y copy` (it ends `? more`); the overlay row carries it. The Logs copy leaves out the stream/file header line. Both called out in the receipt; backlog if wanted.
- `TestRefusedCommandKeepsItsMessage` still renders at panel width, not a frame; it guards the "text unchanged" case, which the 80x24 frame shows cut as before (`✗ Failed: needs sudo — run "brew… · Esc to close`).

Not checked: OSC 52 in a real macOS terminal (no TUI launch). No brew write commands run.
