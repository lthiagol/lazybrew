# Receipt — runner lane, slice v2-05s-services

**Commits (AC order):** AC-01 e8add34 · AC-02 a4bc483 · AC-03 421c0ba ·
AC-04 7d8b51c · AC-05 3cc1058 · this receipt. Not pushed. Worktree clean.

**Gate (final tree, `CGO_ENABLED=1`):** build, lint, `go test ./... -race -count=1`
(all nine packages), `make cover-check` (brew 76.5 ≥ 62, presentation 98.2 ≥ 91,
gui 82.9 ≥ 36, modal 90.1 ≥ 41). All pass.

**Safety:** no brew services write ran. The harness sent `S` on moshi-hook, captured
frame 20, then `esc`. Brew calls were reads only. moshi-hook is active, pid 800,
`ActiveEnterTimestamp` unchanged. Tests use the mock runner only.

**Evidence per AC:**
- AC-01: `ServicesReader.Info`, cached under `KeyServicesInfo`, invalidated by all five
  writes. Fixture `testdata/services_info_all.json` is real output.
- AC-02: Info tab (was Status), key line, state-aware footer, list fallback. Frames 19–20.
- AC-03: s on running and S/r on stopped toast and do nothing; S and r confirm with
  the briefed kind, chip and consequence; reload after any services task; cursor by name.
- AC-04: needs-sudo text from brew's `sudo` hint, "requires root" or "permission
  denied". Other errors pass unchanged.
- AC-05: `result.md`; 40 frames, each exactly W×H; the 18 old frames byte-identical.

**Tests:** 31 functions added, none removed. Four bodies changed, named:
`TestServicesRestartKeyHandler`, `TestServicesRestartHintKeyHandled` (they enqueued
on r; now they confirm first), `TestServiceActionUnknownResolvesWithError` and the
`review_r2` check (they call `runServiceTask`).

**Decisions, all in `work/questions.md` for redirect:**
- Q-v2-services-toast: the 82-cell sudo message cannot fit the 60-cell toast cut, so
  toasts wrap to three rows (extends Q-v2-04-wrap). Short toasts are unchanged.
- Q-v2-services-icons: Info glyphs come from the icon set (▶ ■ ✗ ⏲), not ● ○; the
  word is `running`. The sidebar keeps `started`.
- Cron is a hash in brew, shown as `hour 3, minute 0`. `c cleanup` is always shown.

**Next session must know:** no system service exists here, so the needs-sudo path is
checked against brew's source and tests, not a live refusal. AGENTS.md's "coverage
tracking file" does not exist, so no row was added. LB-09 (Logs tab) stays parked.
The manual smoke gate (LB-05) was not run.
