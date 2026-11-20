# lb09-service-logs — runner receipt (brief `work/briefs/lb09-service-logs.md`)

## Commits, in order

- AC-01 `952fde9` feat(logtail): bounded tail read
- AC-02 `0530f68` feat(gui): the Services Logs tab
- AC-03 `9967789` feat(gui): the Logs tab follows its log
- fix `5327df8` shared-file note stays on screen (AC-02.3)
- AC-04 `ec7f5d6` tests, help entry, frames, result, STATE row
- `9ea7419` receipt and `Q-lb09-slide`; `f67ef96` trims this receipt

## Evidence per AC

- **AC-01** `internal/logtail` (new package), 14 tests in `logtail_test.go`:
  bounded read, partial first line dropped, missing and empty-path errors, CRLF,
  escape and invalid-UTF-8 made safe.
- **AC-02** `services_logs_test.go`: tabs, `e` toggle, reset on service change,
  the four empty states, scroll keys, long lines cut at 80x24 and 140x40.
- **AC-03** `services_logs_follow_test.go`: ticks show appended lines; pinned
  at the bottom; scrolled-up view keeps its offset; no read after leaving the tab.
- **AC-04** `flows/logs_test.go` `TestLogsFlow` (teatest) passes. Eight frames
  in `work/verification/lb09-service-logs/`, each checked to its exact size.

## Gate (final run, tree at `ec7f5d6`)

- `make ci`: exit 0. gofmt, `go vet`, `go test ./... -race -count=1` all clean;
  cover-check "All coverage floors met"; build ok.
- `GOOS=darwin go build ./...`: exit 0.

## Changed expectations (named)

- `TestServicesFooterAndInfoAgreeOnActions` skips `[ ]`, as it skips `j/k` and `c`.
- Services footer gains `[ ] tabs`; `e stream` on Logs only; help gains `e`.

## For the next session

- Open question `Q-lb09-slide`: the window slides past 200 lines; rule before review.
- One AC-02 `make test-fast` call was mistyped; re-run correctly, exit 0.
- No brew write command run; no service touched.
- The unredacted real moshi-hook log is not in the repo.
