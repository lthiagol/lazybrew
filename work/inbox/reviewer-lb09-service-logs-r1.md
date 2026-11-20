verdict: changes-required

# Review lb09-service-logs — round r1

Scope: `65662ee..409ee9d` (AC-01 `952fde9`, AC-02 `0530f68`, AC-03 `9967789`,
fix `5327df8`, AC-04 `ec7f5d6`, Q-lb09-slide `409ee9d`, receipt commits).
Ruling Q-lb09-slide read as part of the contract.

## Gate (run by the reviewer)

- `make ci` at `409ee9d`: exit 0. gofmt, vet, `go test ./... -race -count=1` all ok
  (`internal/gui` 21.3s, `gui/flows` 8.9s, `logtail` 1.0s), "All coverage floors met", build ok.
- `GOOS=darwin go build ./...`: ok.

## AC table

| AC | Verdict | Evidence |
|---|---|---|
| AC-01 tail | met | `internal/logtail/logtail.go`: `MaxBytes = 256<<10`, seek to `size-maxBytes`, partial first line dropped; `ErrNoPath`, `fs.ErrNotExist`, other errors distinct; `Size`+`ModTime` returned. 14 tests cover every listed case, incl. `TestReadNeverReadsWholeFile`, CRLF, ESC/C1/invalid UTF-8 → U+FFFD. |
| AC-02 tab | met | Ran `./bin/lazybrew` in a pty (pyte), 80x24 and 140x40, read-only: `6 j ] e k esc pgup`. Tabs `Info  Logs`; header `stdout · /…/.linuxbrew/var/log/moshi-hook.log`; `e` → `stderr` + "stderr goes to the same file as stdout" with lines; dbus → `no stdout log for this service`, stream reset to stdout; footer `[ ] tabs`, `e stream` only on Logs; help lists `e`; lines cut with `…`; frames exactly 24 / 40 rows, matching the saved frames (newer lines only). Info tab unchanged. pgup scrolls under a fixed header. |
| AC-03 follow | met (see F-r1-01) | `tea.Tick` → `logsTickMsg{gen}` → `readLogsCmd` → `logsReadMsg{gen}`; stale gen dropped in both cases; leaving the tab clears `started`, a late read/tick ends the loop; re-visit bumps gen so only one loop. No goroutine, no `program.Send`, no watcher. Pin/scrolled-up/re-anchor (`shiftedBy`) and clamp per Q-lb09-slide, tests `…ReanchorsOnSlide`, `…ClampsWhenTopLineLeaves`; shrunk file test; unchanged size+mtime early-return test. |
| AC-04 tests/evidence | met with a gap (F-r1-02) | flow `TestLogsFlow` asserts on the screen: open, append → "tick", `e` → stderr. Unit tests for read counting, empty states, help, footer. Eight frames + `result.md`; STATE row present. |

Invariants: other panels/actions unchanged (diff limited to Services tabs,
footer/help entries, shared scroll helpers which are no-ops off Logs —
`contentHeadRows` returns 0). Only changed expectation:
`TestServicesFooterAndInfoAgreeOnActions` skips `[ ]` (named in receipt).
No test deleted/weakened. No new deps; floors unchanged. One commit per AC
plus one named fix commit and the ruled Q-lb09-slide commit.

## Findings

F-r1-01 (major) — `internal/gui/gui.go:280` (also :278) — `return nm, tea.Batch(cmd, nm.syncLogs())`
relies on an evaluation order the Go spec leaves unspecified: `syncLogs` mutates
`nm.logs` (`service`, `started`, `gen`), but the spec only orders calls among
themselves, not the read of the variable `nm` relative to the call (spec,
"Order of evaluation": `x := []int{a, f()}` may be `[1 2]` or `[2 2]`).
The `Model` (value) branch is the one `update` normally returns. gc happens to
read `nm` after the call today, so tests pass; another compiler or version may
return the pre-sync copy, which would lose `started`/`gen` — a new read on every
message and stale-gen confusion (AC-03.2). — Fix must achieve: call `syncLogs`
into a local first (`c := nm.syncLogs(); return nm, tea.Batch(cmd, c)`) in both
branches, so the returned model is certainly the synced one.

F-r1-02 (minor) — `internal/gui/services_logs_test.go:181` — AC-04.1 asks
"switch service → stdout of the new one, no stale lines"; the test switches to
`backup`, which has no log path, so it never shows the new service's stdout
lines, and no test delivers a stale read of service A after moving to service
B (only stream toggle and tab leave are covered). — Fix must achieve: a test
where two services each have a stdout file; after `j` the view shows B's lines,
none of A's, and an old-gen `logsReadMsg` carrying A's lines is dropped.

## Not findings (noted)

- Each tick still reads up to 256 KiB even when the file is unchanged; the brief
  only requires no re-render work (met). Possible later optimisation: stat first.
- Safety: no brew write command run; moshi-hook not touched; its log only read.
