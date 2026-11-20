# Review brief — slice `v2-04-overlays-styled`

> The slice contract is `work/briefs/v2-04-overlays-styled.md`; it wins where
> they differ. You review; you do **not** edit code, tests, or commits.

Round **rN** (1..5) reviews everything on `wip` after the commit that added this
brief (`git log --oneline -- work/briefs/review-v2-04-overlays-styled.md`),
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
- **V3 action kinds (AC-01):** every `NewConfirmModal` call site passes a
  kind; compare against the contract's mapping and the receipt's list; no
  kind inferred from title text.
- **V4 confirm face (AC-02):** thick border in the action accent, chip,
  question, consequence line (uninstall uses the computed dependents), key
  line, visible Yes/No selection; width `min(W-8, 72)`; no line wraps past the
  box at 80×24. Input/menu share the frame.
- **V5 dim backdrop (AC-03):** done at compositing, not per-panel; the test
  compares the same cell with/without a modal (style differs, text equal);
  modal and footer not dimmed. Look at `after/140x40/07-confirm-modal.ansi`.
- **V6 toasts (AC-04):** stack of 3, eviction, per-type lifetime (3 s / 6 s
  errors), click dismiss, no goroutines, never over header/footer, 80×24
  border overlap fixed (check the frame).
- **V7 every frame change explained (`result.md`);** sidebar/main/header/
  footer restyle or animation = **major**.
- **V8 invariants:** confirm semantics unchanged (keys, default selection,
  `handleModalResult`); toast/modal triggers and texts unchanged; test count;
  removed assertions replaced; no keybinding change; `internal/brew/**`
  untouched; one commit per AC.

Also `cat after/140x40/*.ansi` in your terminal and flag anything that reads
badly as a **minor** finding with the frame name.

## Severity

blocker (gate fails, frame size wrong, test deleted/weakened without
replacement, regression) · major (AC not met, scope breach, unexplained
change) · minor (docs/process, visual polish).

## Report

`work/inbox/reviewer-v2-04-overlays-styled-rN.md` (≤ 60 lines): first line
`verdict: ok` or `verdict: changes-required`; AC table with evidence; findings
`F-rN-01 (severity) — file:line — what — evidence — what the fix must
achieve`; rounds ≥ 2: status of earlier findings. `ok` only with no findings
left. Commit once (`inbox: reviewer reports v2-04-overlays-styled rN`), worktree
clean, reply with exactly one line:
`done: work/inbox/reviewer-v2-04-overlays-styled-rN.md`. Do not touch other panes.
