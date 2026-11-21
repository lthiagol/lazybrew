# Fix brief — lb18-config-sample, round r1 (minor findings, orchestrator-verified)

Findings: `work/inbox/reviewer-lb18-config-sample-r1.md`, F-r1-01 … F-r1-04. One
commit per finding that changes code, staged by name, tests first where the
finding is a test gap.

- F-r1-01: the sample's TTL comments say "0 or less keeps the default";
  `update_stale_hours` says "1 or more". Add a test that every `*_ttl` comment
  carries the rule.
- F-r1-02: `TestSampleFollowsDefaults` compares to
  `sampleDuration(DefaultCasksTTL)`, not a hard-coded `5m`.
- F-r1-03: `--config-sample` with `--config-create` is rejected: one stderr
  line naming both flags, non-zero exit, nothing printed to stdout, no file
  written. Test it.
- F-r1-04: add a HOME-redirected test that `Load("")` reads the file at
  `DefaultPath()`.

Gate: `make test-fast PKG=./internal/config/... ./cmd/lazybrew/...` per commit,
`make ci` before the receipt. Never launch the TUI or `go run ./cmd/lazybrew`
(the receipt of round 0 says it happened once): use `./bin/lazybrew` with
`env PATH=/nonexistent` for the flag checks. Append "Fix r1" (≤ 8 lines) to
`work/inbox/runner-lb18-config-sample.md`, commit it
(`inbox: runner fix r1 lb18-config-sample`), reply with one done line.
