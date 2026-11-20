# Review brief — slice `v2-05s-services`

> The slice contract is `work/briefs/v2-05s-services.md`; it wins where
> they differ. You review; you do **not** edit code, tests, or commits.

Round **rN** (1..5) reviews everything on `wip` after the commit that added this
brief (`git log --oneline -- work/briefs/review-v2-05s-services.md`),
including earlier fix commits. Rounds ≥ 2: confirm each earlier finding is
fixed, then look for new problems. Orchestrator commits touching only
`work/questions.md`, `work/backlog.md`, `work/STATE.md`, `work/briefs/` are not
findings. A receipt is a claim, not evidence. `CGO_ENABLED=1` for the race
gate. Never run brew write commands.

## Verify

- **V1 gate:** `make build && make lint && CGO_ENABLED=1 go test ./... -race -count=1 && make cover-check`.
- **V2 frames:** run the committed harness against real Homebrew, diff vs
  `after/` (must match), every frame exactly W×H cells. Remove
  `internal/gui/dump` after.
- **V0 safety first:** confirm no brew services write ran: `systemctl --user
  status sh.brew.moshi-hook` must still show it active (read-only), and
  `journalctl --user -u sh.brew.moshi-hook --since "<slice start time>"` shows
  no stop/start in the slice window (read-only). Grep the harness and tests
  for any path that could confirm a services dialog (`y`/`enter` after `S`/`r`)
  or call the real runner for a write. Any such path is a **blocker**.
- **V3 brew layer (AC-01):** fixture shaped like real `brew services info
  --all --json` (run it — read only); nullable fields; own cache key with the
  services TTL; invalidated by every services write (test each); GUI read via
  `boundedRead`.
- **V4 Info view (AC-02):** fields, icons/roles, empty omitted; state-aware
  key line and footer; fallback before info loads.
- **V5 actions (AC-03):** guard toasts for already-running/stopped; stop and
  restart confirm with the right kind/chip/consequence; start/run without
  confirm; reload of list + info after completion; selection kept by name
  (test with reordering).
- **V6 errors (AC-04):** sudo/root detection from brew's text; plain message;
  lazybrew never runs sudo; other errors unchanged.
- **V7 every frame change explained (`result.md`)**; frame 20 shows the stop
  confirm and the harness escapes it.
- **V8 invariants:** no new keybinding, no other panel changed, no restyle,
  test count, removed assertions replaced, one commit per AC.

Also `cat after/140x40/*.ansi` in your terminal and flag anything that reads
badly as a **minor** finding with the frame name.

## Severity

blocker (gate fails, frame size wrong, test deleted/weakened without
replacement, regression) · major (AC not met, scope breach, unexplained
change) · minor (docs/process, visual polish).

## Report

`work/inbox/reviewer-v2-05s-services-rN.md` (≤ 60 lines): first line
`verdict: ok` or `verdict: changes-required`; AC table with evidence; findings
`F-rN-01 (severity) — file:line — what — evidence — what the fix must
achieve`; rounds ≥ 2: status of earlier findings. `ok` only with no findings
left. Commit once (`inbox: reviewer reports v2-05s-services rN`), worktree
clean, reply with exactly one line:
`done: work/inbox/reviewer-v2-05s-services-rN.md`. Do not touch other panes.
