# Testing

| Tier | Where | Use for |
|---|---|---|
| Unit | `internal/brew`, `internal/gui/task`, … | parsers, cache, services, logic |
| Snapshot | `internal/gui/presentation` | stable formatted output |
| E2E | `internal/gui/flows` (`teatest`) | user flows; assert on `View()` |
| Integration | `//go:build integration`, `make test-integration` | real brew round-trip (read-only on this machine) |

## Rules

- `make ci` is the gate (race detector on; the Makefile sets `CGO_ENABLED=1`).
  Per commit, `make test-fast PKG=<touched packages>` is enough.
- A fix gets a test that fails without it. Prove it once: revert the fix (or
  mutate the code), see the test fail, restore.
- Coverage floors are in `scripts/check-coverage.sh`. A new package gets a floor
  in the slice that adds it. Never lower one without a ruling in
  `work/questions.md`.
- New brew command support: see [brew.md](brew.md#adding-a-command).

## Fakes and helpers

- `brewtest.NewMockRunner` (`internal/brew/brewtest`) is the fake runner: set
  `ExecuteFn` / `ExecuteJSONFn` / `ExecuteStreamFn` to switch on the argv. A brew
  failure is a `&brew.BrewExitError{Command, ExitCode, Stderr}`.
- Fixtures from real brew go in `internal/brew/testdata/` — capture them with a
  read-only brew command, never invent the shape. See [brew.md](brew.md).
- GUI unit tests: `newTestModel()` (mock runner), `updateModel`, `sendKey`,
  `sendSpecial`, `drainCmds` in `internal/gui/gui_test.go`.
- Flows: `testutil.NewTestModelWithRunner(t, runner)` (`internal/gui/testutil`);
  wait with `waitForScreen` / `waitForScreenAll` / `waitForCall` (`flows/wait_test.go`; each
  waits up to `flowWaitCeiling`, 10 s, and returns as soon as its text shows). The flows'
  `TestMain` sets Unicode icons and animations off, so frames are plain resting text.
- To look at a frame, build a harness on the mock runner and dump
  `ansi.Strip(m.View().Content)`. Never on `brew.NewDefaultRunner()` (AGENTS.md rule 1).

## teatest flows (lessons)

- **Wait for a screen only the action under test can produce.** A toast such as
  "Data refreshed" may come from an earlier action; waiting on it let
  `TestTapsTrustedItemFlow` quit before `R`'s read landed (LB-31). Make the fake
  return something new after the action and wait for that.
- Text near the right edge can be cut by an overlay: wait on a prefix
  ("Data refresh"), not the whole word.
- A held fake read (to test "while loading") must be released with a deferred
  `close`, so no goroutine stays blocked.
- LB-37: `TestCleanupTabFlow` missed "Nothing to clean up" under load with a 3 s wait.
  The wait spans the 2 s auto-close of a successful Operation (`operationAutoClose`),
  so the ceiling is 10 s (`flowWaitCeiling`). A flow that still fails twice is yours to fix.
