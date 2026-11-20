# runner fix receipt r1 — v2-03b-sidebar-lists

Findings answered, in order (commit, subject id):
- F-orch-01 `01fb663` the batch marks show on the Outdated list only. `TestBatchMarksOnlyTheOutdatedList`; mutation (draw on every list) makes it fail.
- F-orch-02 `9ba0f80` the Trust tab shows the trust story. Official / trusted / untrusted, what each means for the tap's formulae and casks (wording from `brew trust --help`), the `t` key. Tap Info keeps the details. `TestTapTrustTabShowsTheTrustStory`.
- F-orch-03 `d68d66f` PgUp/PgDn move the content by a page; Home/End and g/G go to the top and bottom. Each key is clamped at both ends. `g` and `G` were unbound, so none is remapped. Help (Navigation) and the footer's global hints name them. `TestScrollKeysMoveAndClamp` (each key, both ends, the rows drawn), `TestScrollIsPerItem`, `TestScrollKeysAreInTheHelpAndFooter`.
- F-orch-03 follow-up `752583b` the page-key hint is the last global hint. Before, it cut `? help` at 140 columns.
- F-r1-01 no code change. Commit `0fa39ff` is AC-01 of v2-03b (its subject lacks the tag; history not rewritten).
- F-r1-02 `3c38d91` the Status Homebrew row is a ten-column label and its value. `TestStatusRowsAreCompactSummary` now expects `Homebrew  4.5.2`.
- F-r1-03 `8d2c171` `gui_test.go` says Info, not List; the `panel.go` comment sits on `unfilteredFormulae`.

Gate on `008d344` (HEAD), all four green:
- `make build` exit 0; `make lint` (go vet) exit 0.
- `CGO_ENABLED=1 go test ./... -race -count=1`: every package ok.
- `make cover-check`: brew 72.9, presentation 96.2, gui 77.9, modal 50.5; all floors met.
- Tests: 630 PASS (round r0 was 625; +1 F-orch-01, +1 F-orch-02, +3 F-orch-03).

Frames: `work/verification/v2-03b-sidebar-lists/after/` re-captured at `752583b` (real Homebrew 7.0.15, `.ansi` at 140×40). Every `.txt` frame is W×H. `result.md` has a fix-round section: the Status Homebrew row in every sidebar frame; the footer's `? more` in 06 and 09 at 140; the new help rows in 08. The batch and Trust changes are not in the harness and are covered by the tests above.

Deviations and notes for the reviewer:
- The footer: the page-key hint was first placed before `? help`, which hid `? help` at 140. Moved last (`752583b`), so the cut drops it first.
- F-orch-02 wording is from Homebrew's `brew trust --help` ("so Homebrew may load them") and the code's TrustOfficial / TrustTrusted / TrustUntrusted comments. No claim beyond those.
- The model viewport stays removed (`contentScroll` per item, see Q-v2-03b-d).

Worktree clean at report time. Branch `wip`, not pushed. No brew write command was run.
