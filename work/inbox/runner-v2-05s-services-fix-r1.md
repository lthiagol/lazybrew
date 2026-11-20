# Fix receipt r1 — runner lane, slice v2-05s-services

**Commits:** F-r1-01 befbf7e · F-r1-02 a97a13a · F-r1-03 190ff4c · this receipt.
Not pushed. Worktree clean.

**F-r1-01 (major):** `serviceStoppable` = running, `error` or `scheduled` (cites
`formula_wrapper.rb` to_hash and `cli.rb` stop). S and r confirm and run there.
`s start` applies to every not-running service: brew re-loads a loaded unit that is
not running. Info key line and footer share one list. Tests: five, named
`*ErrorAndScheduled*`, `TestServiceStoppable*`, `TestServiceHints*`, `TestServiceInfoKeyLine*`.
Open for the orchestrator: brew also reports `stopped` as loaded; AC-03.1 keeps S/r refused there.

**F-r1-02 (minor):** failed or cancelled services tasks send `servicesReloadMsg`
(Services only, no refresh counter, no "Data refreshed"). Success still sends
`RefreshMsg`. Tests: four, `TestService*RefreshToast`, `TestServiceSuccessStillRefreshes`,
`TestServicesReloadAddsNoToast`.

**F-r1-03 (minor):** harness waits 7 s after frame 18, then sends `ToastTickMsg`.
Frames 19 and 20 re-captured at both sizes; 01–18 unchanged. Logged in `result.md`.

**Gate (final tree, `CGO_ENABLED=1`):** build, lint, race tests (nine packages),
cover-check all pass. 40 frames, each W×H. 39 test functions added, none removed.

**Safety:** no brew services write ran. moshi-hook: MainPID 800, NRestarts 0,
ActiveEnterTimestamp 2026-11-18 11:37:48.
