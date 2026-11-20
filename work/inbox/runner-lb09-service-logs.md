# lb09-service-logs — runner receipt

Brief: `work/briefs/lb09-service-logs.md`. Lane: runner (builder).

## Commits, in order

- AC-01 `952fde9` feat(logtail): read the last lines of a log file, bounded
- AC-02 `0530f68` feat(gui): the Services Logs tab shows the last lines
- AC-03 `9967789` feat(gui): the Logs tab follows its log every second
- fix `5327df8` the shared-file note stays on screen (AC-02.3), found by the frame capture
- AC-04 `ec7f5d6` tests, help entry, frames, result, STATE row
- this receipt, with the `work/questions.md` entry (one commit)

## Evidence per AC

- **AC-01** `internal/logtail` (new package, not `internal/brew`). 14 tests in
  `logtail_test.go` pass: bounded tail with the partial first line dropped,
  `MaxBytes` never fully read, exact bound, no trailing newline, empty, CRLF,
  missing file is `fs.ErrNotExist`, empty path is `ErrNoPath`, escape and
  invalid-UTF-8 made safe.
- **AC-02** `services_logs_test.go`: tabs Info, Logs; `[ ]`; `e` toggle with
  the header; reset on service change; all four empty states; scroll keys;
  long lines cut at 80x24 and 140x40.
- **AC-03** `services_logs_follow_test.go`: ticks deliver appended lines; pinned
  at the bottom; scrolled-up view keeps its offset; shrunk file re-read; no read
  after leaving the tab (counted through the injected reader); stale tick and
  read dropped; unchanged file not re-applied.
- **AC-04** `internal/gui/flows/logs_test.go` `TestLogsFlow` (teatest, through
  the program) passes in about 1.1 s. Frames: 8 plain-text frames in
  `work/verification/lb09-service-logs/`, each checked to be exactly its size.
  The real moshi-hook log is read-only; its identifiers are redacted.

## Gate (final run, tree at `ec7f5d6`)

- `make ci`: exit 0. gofmt clean; `go vet` clean; `go test ./... -race -count=1`
  all `ok`; cover-check "All coverage floors met" (brew 80.0 ≥ 62, presentation
  98.5 ≥ 91, gui 86.2 ≥ 36, modal 90.2 ≥ 41); `bin/lazybrew` built.
- `GOOS=darwin go build ./...`: exit 0.

## Changed expectations (named)

- `TestServicesFooterAndInfoAgreeOnActions` (services_view_test.go) skips `[ ]`,
  as it already skips `j/k` and `c`. `[ ] tabs` is a panel key, not a service action.
- Services footer gains `[ ] tabs`; `e stream` on Logs only. Help gains `e`.
- Nothing deleted, skipped or weakened.

## For the next session

- **Open question `Q-lb09-slide`** (in `work/questions.md`): past 200 lines the
  window slides under a scrolled-up view. The offset rule holds as the brief
  reads; the lines under it shift. Needs a ruling before the review, or accept.
- The fix `5327df8` sits after AC-03 in history: the AC-04 capture showed the
  note scrolled off the screen.
- One AC-02 per-commit `make test-fast` call was mistyped (two package patterns
  as make goals) and its exit code was hidden by a pipe. Re-run correctly: exit 0.
- Each poll reads the tail (up to 256 KiB) with no stat first. Unchanged files
  skip the apply, not the read.
- No brew write command was run. No service was started, stopped or restarted.
- Frames come from the `logsModel` fixtures (temp or real log paths), not
  `brew.MockRunner`. Only `TestLogsFlow` goes through the mock runner and the program.
- The capture is `capture_frames_test.go.txt` (not compiled); run steps are in
  `result.md`. The unredacted real log is not in the repo.
- Worktree clean after this commit.
