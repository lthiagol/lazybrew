# Fix brief — v2-08-preload, round r1

Source: `work/inbox/reviewer-v2-08-preload-r1.md` (verdict changes-required).
Fix every finding, one commit each (subject names the id), gate, receipt
`work/inbox/runner-v2-08-preload-fix-r1.md` (≤ 30 lines, committed once).
The slice brief `work/briefs/v2-08-preload.md` still applies (Safety included).

- **F-r1-01 (minor)** — with preload on, a panel opened before Status lands is
  loaded through `lazyLoadPanel` and its count never reaches the dashboard
  (Outdated then says `not loaded — press R` although it loaded). Fix: every
  successful panel landing, lazy or preloaded, fills the dashboard count when
  preload is on. Add a test for the early-switch case (press `4` before Status
  lands; assert on `View()` that the dashboard shows the count).
- **F-r1-02 (minor)** — `work/STATE.md` and the runner receipt still describe
  the first, slow-brew capture. Update the STATE paragraph to match the
  committed frames (after `a8a1c38`); mark the slow capture as superseded.
  Q-v2-08-b stays open. Do not rewrite the earlier receipt commit — say it in
  the fix receipt.

Reply with exactly one line: `done: work/inbox/runner-v2-08-preload-fix-r1.md`.
