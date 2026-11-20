verdict: changes-required

# Reviewer — v2-10-tap-actions, round r1

Scope: the v2-10 commits on `wip` after the brief (5f91936..62d3950: AC-00..AC-05,
9c4afce Q-v2-10-a fix, receipts). v2-09/v2-11 commits in the range are closed slices.
Rulings applied: Q-v2-10-a (official: no `x`), Q-v2-10-b (menu untrust confirms, `T` not in
footer), Q-v2-builder-speed (full gate once, before the receipt).

## Verification

- **V1 gate:** `make build && make lint && CGO_ENABLED=1 go test ./... -race -count=1 && make cover-check`
  → exit 0; every package ok; floors met (brew 79.6, gui 85.7, presentation 98.5, modal 90.1).
- **V2 frames:** I read the harness first: `t` goes only to a tap whose Tap Info matches
  `Trusted: trusted` / `Official: no` (it calls `t.Fatal` otherwise), and every confirm, menu and
  input it opens is closed with `esc` (07, 12, 14, 15, 20, 23, 24). No `y`/`enter` is sent anywhere, and the add-tap input is never
  submitted. Ran against real Homebrew (`DUMP_DIR=scratch`): PASS in 47.8s. 24 frames at 140x40
  (+24 `.ansi`) and 24 at 80x24, all exactly W×H by display width. My output is byte-identical to the
  committed `after/`. `before/` == `v2-09-main-layout/after/` (`diff -rq` clean). `internal/gui/dump`
  removed; worktree clean.
- **V6 safety:** I ran no brew write commands. The harness sent reads plus one `t` (it opened the confirm, then `esc`) and one `x` (same).

## AC table

| AC | verdict | evidence |
|---|---|---|
| AC-00 | met | `TapList{Taps,TrustUnknown}` (taps.go:19); the flagged list is not cached (`TestTapsBatchFailureDoesNotPoisonCache` now asserts the flag, and the 2nd call retries); `TestTapsTrustUnknownOnTapInfoTimeout`. GUI: border label and Tap Info say `trust unknown — press R` in Warning (`OutdatedBadge` = `t.Warning`); rows show no trust word; `t`/`T` show an info toast and enqueue nothing; `x`/`r`/`a` still open (matrix row "unread"). View asserts in tap_trust_state_test.go. |
| AC-01 | met | `tapActionHints` (keybindings.go:33): untrusted `t trust x untap r repair`, trusted `t untrust x untap r repair`, official `r repair` (Q-v2-10-a). The same set leads `tapsHints`; `a add tap` is always in the footer. Frames 21/22 footers match. |
| AC-02 | met | `toggleTrust`: an untrusted tap is trusted at once; a trusted one gets the `ActionUpgrade` confirm with chip `UNTRUST` and the brief's consequence (frame 23); an official tap gets the toast `<name> is official — no trust needed` and nothing runs. `T` opens the menu, and the menu's untrust confirms (Q-v2-10-b). Help and panelHints are updated. |
| AC-03 | met | Untap confirm: `ActionDanger`, chip `UNTAP`, the brief's consequence, and the installed packages from the loaded Formulae (incl. unfiltered leaves) and Casks lists. No brew call (`TestUntapConfirmMakesNoBrewCall`). Frame 24 lists `frond`. Tap failures toast brew's `Error:` line (`tapFailureText`). |
| AC-04 | met | Failure or cancel → `tapsReloadMsg`; success → `RefreshMsg`. `landData` → `selectTapNamed`; after an untap the cursor stays on the same index and is clamped (tap_reload_test.go). The state × key matrix runs on the mock runner, and only the task arm calls brew. |
| AC-05 | met | Harness, frames, result.md and the STATE row are present; the new frames are 21–24. |

Invariants: every write goes through `m.tasks.Enqueue`. There is no new brew command. The `internal/brew` change is `taps.go` only (AC-00). There are no new goroutines or `program.Send`. I found no deleted or weakened tests: the changed expectations (confirm_detail, review_r4, modal_first_frame, sidebar_list, taps_test) are named in the commits, and taps_test got stronger. One commit per AC, in order.

## Findings

F-r1-01 (minor) — internal/gui/details.go:116 — the Taps **Trust** tab still says
"Press t to trust or untrust the tap, or one of its formulae or casks." That was true before AC-02.2. Now `t`
acts on the tap only, a formula or cask is trusted with `T`, and on an official tap both `t` and `T` are
refused with a toast. So the tab names a key that no longer does what it says. That is a missed
part of "Update help" (AC-02.2). Evidence: `grep -n "Press t to" internal/gui/details.go` → line 116. The
string is shown for official, trusted and untrusted taps alike. Fix must achieve: the Trust tab's key
hint matches the AC-02 behaviour — `t` trusts/untrusts the tap, `T` trusts one of its formulae or casks — and an official
tap's Trust tab offers neither key. Add a View assert to the test.
