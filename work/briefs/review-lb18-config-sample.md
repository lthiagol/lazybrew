# Review brief — slice `lb18-config-sample`

> The slice contract is `work/briefs/lb18-config-sample.md` (with
> `work/briefs/cr-common.md`); it wins where they differ. You review; you do
> **not** edit code, tests, or commits.

Round **rN** (1..5) reviews everything on `wip` after the commit that added this
brief (`git log --oneline -- work/briefs/review-lb18-config-sample.md`),
including fix commits. Orchestrator commits touching only `work/` are not
findings. A receipt is a claim, not evidence. `CGO_ENABLED=1` for the race gate.
Never run brew write commands and never launch the TUI against the real brew;
scratch output under /home/agent/.cache, not /tmp. Do not touch other panes or
the user's running `make run`.

## Verify

- **V1 gate:** `make ci`; `GOOS=darwin go build ./...`.
- **V2 regression tests:** mutate and confirm failure: drop one key from the
  comment table (completeness test must fail); change a default in
  `Default()` (sample must follow; loads-to-default test must still pass);
  make the uncommented default differ from `Default()`; make `--config-create`
  overwrite (the refuse test must fail).
- **V3 by hand:** `./bin/lazybrew --config-sample` and `--config-create
  --config <tmp>/a/b/c.yml` (missing parent dirs), twice (second refuses,
  exit non-zero, file byte-identical), and `--config` on a read-only dir
  (clear error, non-zero). Neither flag may start the TUI or run brew: run with
  `env PATH=/nonexistent`.
- **V4 sample quality:** read the whole sample. Every key has a correct,
  useful comment (values, units, `auto`/`true`/`false`, durations, the env
  overrides and what they override); defaults uncommented, alternatives
  commented; it loads through the real loader and any uncommented line is
  valid. Durations round-trip (`5m`, `30s`, `500ms`). Check it against the
  README block and `config.Config` for stale wording.
- **V5 code:** `DefaultPath()` extraction changes no load behaviour (tests for
  `~` and `--config` paths); the flag branch exits before `app.New`; no
  new dependency; no goroutine or global state added; error messages go to
  stderr.
- **V6 invariants:** one commit per AC (the extra AC-01 alignment commit is
  fine); floors not lowered; no golden changes.

## Severity

blocker (gate fails, overwrite of an existing file, TUI/brew started by a flag)
· major (AC not met, sample wrong or does not load, test passes without the
fix) · minor.

## Report

`work/inbox/reviewer-lb18-config-sample-rN.md` (≤ 50 lines): first line
`verdict: ok` or `verdict: changes-required`; AC table with evidence; findings
`F-rN-01 (severity) — file:line — what — evidence — what the fix must achieve`.
Commit once (`inbox: reviewer reports lb18-config-sample rN`), worktree clean,
reply with exactly one line: `done: work/inbox/reviewer-lb18-config-sample-rN.md`.
