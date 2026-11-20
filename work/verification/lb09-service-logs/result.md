# lb09-service-logs — result

Brief: `work/briefs/lb09-service-logs.md`. Commits, one per AC, in order:

| AC | Commit | Subject |
|---|---|---|
| AC-01 | 952fde9 | feat(logtail): read the last lines of a log file, bounded (lb09 AC-01) |
| AC-02 | 0530f68 | feat(gui): the Services Logs tab shows the last lines of a service's log (lb09 AC-02) |
| AC-03 | 9967789 | feat(gui): the Logs tab follows its log every second (lb09 AC-03) |
| AC-02.3 fix | 5327df8 | fix(gui): the shared-file note stays on screen while the Logs lines scroll (lb09 AC-02.3) |
| AC-04 | (this commit) | tests, frames, this file, and the STATE row |

The fix commit comes after AC-03: the frame capture for AC-04 showed that the
shared-file note scrolled off the screen, so it was fixed as its own commit.

## Decisions the brief left open

- **Package:** the tail reader is a new package, `internal/logtail` (not
  `internal/brew`). It is added to the AGENTS.md repository map.
- **Fixed head:** the Logs header (stream and path) and, when it applies, the
  shared-file note are rows above the scrolled lines. The scroll area is one
  row shorter (two when the note shows), so the header stays visible while the
  view is pinned to the newest line.
- **Loop:** one tail loop at a time. A generation number is bumped on every
  start, so a tick or read from an older view is dropped. Changing service
  resets the view to stdout; toggling `e` drops the other stream's lines.

## Tests and what each proves

**AC-01** — `internal/logtail/logtail_test.go` (14 tests)
- Short file; longer than n keeps the last n; n = 0 gives no lines.
- A file bigger than the byte bound reads only its end: the partial first line is
  dropped, and every kept line is whole (`TestReadBoundedByBytes`).
- A file of 3 × `MaxBytes` drops its first line (`TestReadNeverReadsWholeFile`).
- A file exactly at the bound drops nothing; no trailing newline keeps the last
  line; an empty file has no lines; a lone newline is one empty line.
- CRLF loses its `\r`. Escape bytes, C0/DEL controls and invalid UTF-8 come out
  as U+FFFD, a tab as a space (`TestReadMakesLinesSafeToDraw`).
- Errors: missing file is `fs.ErrNotExist`; empty path is `ErrNoPath`; a
  directory is some other error (not either of the two).

**AC-02** — `internal/gui/services_logs_test.go`
- Tabs are Info, Logs; `[ ]` reaches Logs; the footer lists `[ ] tabs`, and
  `e stream` only on Logs (`TestLogsFooterListsStreamOnLogsOnly`); help lists
  `e` under Services (`TestLogsHelpListsStreamKeyUnderServices`).
- The view holds the last 200 lines of a 250-line file, and opens on the newest
  (`TestLogsTabShowsLastLinesOfStdout`).
- `e` toggles stdout and stderr with each stream's path in the header; the
  stream resets to stdout on a service change, and no old line is kept
  (`TestLogsEToggleStreamAndBack`, `TestLogsStreamResetsOnServiceChange`).
- Empty states, each a muted line and no toast: path unset (stdout and stderr),
  file missing, read error (a directory: `is a directory`)
  (`TestLogsEmptyStates`).
- Shared file: the note and the lines show; no note while the file is missing
  (`TestLogsStderrSameFileNote`, `TestLogsNoNoteWithoutLines`).
- Header elision keeps the file name; long lines are cut, not wrapped, and the
  frame is 80x24 / 140x40 (`TestLogsHeaderCutsPathToFit`,
  `TestLogsCutsLongLinesAtEveryFrame`).
- Scroll keys `g` and `G` work on the tab (`TestLogsScrollKeys`).
- A visit starts one read, counted through the injected reader; leaving the tab
  starts none (`TestLogsReadStartsOnEachVisit`).

**AC-03** — `internal/gui/services_logs_follow_test.go`
- A tick shows an appended line; a view at the bottom stays on the newest line
  over ten appends (`TestLogsFollowShowsAppendedLines`,
  `TestLogsFollowStaysPinnedAtBottom`).
- A view scrolled up keeps its offset and its top row
  (`TestLogsFollowScrolledUpDoesNotMove`).
- A shrunk file is read again from its new end (`TestLogsFollowRereadsShrunkFile`).
- Leaving the tab: a tick starts no read, and the reads counted stay at 0
  (`TestLogsLeavingTabStopsReads`); a read that lands after leaving schedules no
  tick (`TestLogsLateReadAfterLeavingSchedulesNothing`).
- A stale tick or read after a stream toggle is dropped
  (`TestLogsStaleTickAfterStreamToggleIsDropped`).
- A read of an unchanged file (same path, size, mod time, error) keeps the lines
  (`TestLogsUnchangedFileIsNotReapplied`).
- A tick reads once and schedules the next one (`TestLogsTickReadsAndSchedulesNext`).
- The frame stays 80x24 and 140x40 while lines arrive, and the note stays on
  screen while following (`TestLogsFollowFrameSizeAtEachSize`,
  `TestLogsStderrSameFileNoteStaysWhileFollowing`).

**AC-04** — `internal/gui/flows/logs_test.go`
- `TestLogsFlow` runs the program: `6`, then `]` shows the stdout lines; a line
  appended to the file shows on the next poll (1 s, no sleep in the test); `e`
  shows the stderr log; `q` quits. It asserts on the rendered output and runs
  against temp files and the mock runner only.

## Frames

Under `work/verification/lb09-service-logs/`, plain text, each checked to be
exactly its size (the capture fails otherwise):

- `temp-stdout-80x24.txt`, `temp-stdout-140x40.txt`: a 250-line temp log.
- `temp-stderr-80x24.txt`, `temp-stderr-140x40.txt`: the temp stderr log.
- `moshi-hook-stdout-80x24.txt`, `moshi-hook-stdout-140x40.txt`: the real
  moshi-hook log, read-only.
- `moshi-hook-stderr-80x24.txt`, `moshi-hook-stderr-140x40.txt`: the same file
  as stderr; the frame shows the shared-file note.

The real log holds host IDs, session UUIDs and API hosts. The capture redacts
the first three kinds before writing (`host_<redacted>`, `<uuid>`, `<ip>`;
`lb09Redact`), so the unredacted values are not in the repository. Each 140x40
real frame had two host IDs, both redacted. No UUID or IP address appeared in
any frame. API hosts were left as they are.

`capture_frames_test.go.txt` makes the frames. It is not part of the suite:
copy it to `internal/gui/zz_lb09_capture_test.go`, run
`LB09_DIR=<this directory> LB09_REAL_LOG=/home/linuxbrew/.linuxbrew/var/log/moshi-hook.log go test ./internal/gui/ -run TestLB09CaptureFrames -v`,
then delete the copy.

The real log path came from `brew services info --all --json` (a query, run
with `HOMEBREW_NO_AUTO_UPDATE=1`). No brew write command was run, and no
service was started, stopped or restarted.

The temp-log and moshi-hook frames are built from `logsModel` (in the gui test
package), which loads the fixtures in `services_view_test.go` and points the log
paths at the temp file or the real one. They do not go through
`brew.MockRunner`. `TestLogsFlow` is the one that goes through the mock runner
and the program.

## Gate

`make ci` (exit 0):
- gofmt check: clean. `go vet ./...`: clean.
- `CGO_ENABLED=1 go test ./... -race -count=1`: every package `ok`, including
  `internal/logtail` and `internal/gui/flows`.
- `make cover-check`: All coverage floors met — brew 80.0% ≥ 62%,
  presentation 98.5% ≥ 91%, gui 86.2% ≥ 36%, modal 90.2% ≥ 41%.
- `go build -o bin/lazybrew ./cmd/lazybrew`: ok.

`GOOS=darwin go build ./...`: exit 0.

Per commit, `make test-fast PKG=<touched packages>` passed after each AC and after
the fix. One AC-02 run was mistyped (two package patterns as make goals) and
was re-run correctly.

## Changed expectations

Named, as the brief asks:

- `TestServicesFooterAndInfoAgreeOnActions` (`internal/gui/services_view_test.go`)
  now skips `[ ]`, as it already skips `j/k` and `c`: `[ ] tabs` is a panel key,
  not a service action. The footer does list `[ ] tabs` now.
- `TestServicesFooterLeadsWithStateActions` is unchanged and still passes:
  the state actions still lead the Services footer.
- The Services help page gained an `e` entry. The footer gains `e stream` on
  Logs only.

Nothing was deleted, skipped or weakened.

## Notes for the reviewer

1. **The 200-line window slides.** Once a file holds more than 200 lines, each
   new line pushes one out of the top. A view scrolled up keeps its offset, so
   the rows under it shift by the lines that left. The offset itself does not
   move, and `TestLogsFollowScrolledUpDoesNotMove` covers a file under 200 lines,
   where nothing slides. Keeping the same lines in view would mean adjusting the
   offset by the lines that left. Is that wanted?
2. **Each poll reads the tail**, up to 256 KiB, even when the file is unchanged:
   there is no stat before the read. The unchanged check skips applying the
   read (no rebuild of lines or scroll), but the I/O still happens once a second
   while Logs is shown.
3. **No content cache.** The framework draws the frame after each message, as it
   did before. The Logs content is wrapped on each draw, as the other tabs are.
4. **The stream and the scroll.** Scroll position is per panel, tab and service,
   not per stream, so `e` keeps the offset. A toggle restarts the view as loading,
   which pins it to the bottom.
5. **The stream resets on a service change only.** Leaving Logs and coming back
   keeps the stream.
