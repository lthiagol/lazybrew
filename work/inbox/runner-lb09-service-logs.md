# lb09-service-logs — runner receipt

Brief: `work/briefs/lb09-service-logs.md`. Lane: runner (builder).

## Commits, in order

- AC-01 `952fde9` feat(logtail): read the last lines of a log file, bounded
- AC-02 `0530f68` feat(gui): the Services Logs tab shows the last lines
- AC-03 `9967789` feat(gui): the Logs tab follows its log every second
- fix `5327df8` shared-file note stays on screen (AC-02.3), found by the frame capture
- AC-04 `ec7f5d6` tests, help entry, frames, result, STATE row
- `9ea7419` receipt and `Q-lb09-slide`; this trim is the next commit

## Evidence per AC

- **AC-01** `internal/logtail` (new package). 14 tests pass: bounded tail, the
  partial first line dropped, `MaxBytes` never fully read, CRLF, missing file
  (`fs.ErrNotExist`), empty path (`ErrNoPath`), escape and invalid UTF-8 made safe.
- **AC-02** `services_logs_test.go`: tabs, `[ ]`, `e` toggle, reset on service
  change, all four empty states, scroll keys, long lines cut at 80x24 and 140x40.
- **AC-03** `services_logs_follow_test.go`: ticks show appended lines; pinned at
  the bottom; scrolled-up view keeps its offset; shrunk file re-read; no read after
  leaving the tab (counted); stale tick and read dropped; unchanged file not re-applied.
- **AC-04** `flows/logs_test.go` `TestLogsFlow` (teatest, through the program)
  passes. Frames: 8 plain-text frames in `work/verification/lb09-service-logs/`,
  each checked to its exact size. The real moshi-hook log is read-only and redacted.

## Gate (final run, tree at `ec7f5d6`)

- `make ci`: exit 0. gofmt clean; `go vet` clean; `go test ./... -race -count=1`
  all `ok`; "All coverage floors met" (brew 80.0 ≥ 62, presentation 98.5 ≥ 91,
  gui 86.2 ≥ 36, modal 90.2 ≥ 41); build ok.
- `GOOS=darwin go build ./...`: exit 0.

## Changed expectations (named)

- `TestServicesFooterAndInfoAgreeOnActions` skips `[ ]`, as it skips `j/k` and `c`.
- Services footer gains `[ ] tabs`; `e stream` on Logs only. Help gains `e`.
- Nothing deleted, skipped or weakened.

## For the next session

- Open question `Q-lb09-slide` (in `work/questions.md`): past 200 lines the window
  slides under a scrolled-up view. Needs a ruling before review, or accept.
- One AC-02 per-commit `make test-fast` call was mistyped and its exit hidden by
  a pipe; re-run correctly, exit 0.
- Each poll reads the tail with no stat first; unchanged files skip the apply only.
- No brew write command run; no service started, stopped or restarted.
- Frames come from `logsModel` fixtures, not `brew.MockRunner`. The unredacted
  real log is not in the repo. Capture steps are in `result.md`.
