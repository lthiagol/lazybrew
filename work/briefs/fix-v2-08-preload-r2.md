# Fix brief — v2-08-preload, round r2

Source: `work/inbox/reviewer-v2-08-preload-r2.md` (verdict changes-required).
Fix every finding, one commit each (subject names the id), gate, receipt
`work/inbox/runner-v2-08-preload-fix-r2.md` (≤ 30 lines, committed once).
The slice brief `work/briefs/v2-08-preload.md` still applies (Safety included).

- **F-r2-01 (minor)** — frames 15 and 18 show a dashboard production cannot
  reach (`Taps 0` beside a Taps box of real taps): the harness injects a Taps
  landing with no `Taps` slice. Fix the harness so the injected landing has a
  production shape (`Taps` consistent with `Items`), or re-land the real Taps
  read after frame 13; regenerate the frames; `result.md` names any change
  that remains in 15 and 18.
- **F-r2-02 (minor)** — `m.landed` keeps a count after its source later fails,
  so a Status landing shows a stale count as known (Outdated panel `timeout`,
  dashboard `1 package`). Fix: a failed landing forgets that source's count;
  an unknown count is never filled from a read older than the source's last
  failure. Add a test on `View()`.

Reply with exactly one line: `done: work/inbox/runner-v2-08-preload-fix-r2.md`.
