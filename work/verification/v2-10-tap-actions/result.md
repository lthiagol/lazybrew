# v2-10-tap-actions — result

The Taps list now works like Services. Tap Info and the footer show the actions
that apply to the selected tap. `t` toggles trust (an untrust is confirmed). `T`
opens the formula and cask trust menu. The untap confirm says what it does and
which installed packages block it. A failed tap write reloads the list and keeps
the cursor on the same tap. A Taps list read without trust state says so.

## Commits

One commit per acceptance criterion, in AC order, staged by name.

| AC | commit | subject |
|---|---|---|
| AC-00 | 5f91936 | feat(taps): a Taps list read without tap-info says trust is unknown (v2-10 AC-00) |
| AC-01 | 1e91fad | feat(gui): Tap Info and the footer show the tap actions that apply now (v2-10 AC-01) |
| AC-02 | a4eaf97 | feat(gui): t toggles tap trust, T opens the item menu (v2-10 AC-02) |
| AC-03 | 6d5c46a | feat(gui): the untap confirm says what it does, and a tap failure says brew's line (v2-10 AC-03) |
| AC-04 | d1ee781 | feat(gui): a tap write reloads the Taps list and keeps the cursor on the same tap (v2-10 AC-04) |
| AC-05 | (this commit) | docs(plan): v2-10 tap actions evidence, result, STATE row (v2-10 AC-05) |

Per-commit gate (Q-v2-builder-speed amendment): each commit built, vetted and
passed `go test -count=1` (no `-race`) for `internal/brew` and `internal/gui/...`.

## Gate (run once, before the receipt)

Run on the AC-04 tree (the evidence directory holds no Go code and is not in
the build; the harness was run and removed before the gate):

- `make build` — ok
- `make lint` (`go vet ./...`) — ok
- `CGO_ENABLED=1 go test ./... -race -count=1` — every package ok (brew, config,
  app, gui, gui/anim, gui/flows, gui/modal, gui/presentation, gui/style,
  gui/task, gui/testutil)
- `make cover-check` — "All coverage floors met" (brew 79.6% ≥ 62%, gui 85.7% ≥ 36%,
  presentation 98.5% ≥ 91%, modal 90.1% ≥ 41%)

## Safety

No brew write command was run on this machine. The frame harness only reads
(`brew tap`, `brew tap-info`, the status and services reads that the app makes
at startup). Every unit test uses the mock runner, and the write tests record
calls on it.

The harness sends `t` in one place only: frame 23, to `lthiagol/tap`. Before that
key, a guard checks the frame shows a trusted, non-official tap; otherwise the
harness stops (`tapFrameIs`). `t` on the untrusted tap `danielfoehrkn/switch` is
never sent. The opened confirm is cancelled with `esc`. Frame 24 opens the untap
confirm with `x` and is cancelled with `esc`. No `y` or `enter` is sent, and the
add-tap input is not submitted.

## Evidence

Harness: `work/verification/v2-10-tap-actions/frame_dump_test.go.txt` (the v2-09
harness plus frames 21–24 and the guard). `before/` is `v2-09-main-layout/after/`.
`after/` has 24 frames at 140×40 (`.ansi` and `.txt`) and 24 at 80×24 (`.txt`).

Frame sizes: every `.txt` frame in `after/` is exactly W×H (checked with a script
that counts east-asian width per line and the line count).

New frames:

- `21-tap-info-untrusted`: `danielfoehrkn/switch`, the key line `t trust  x untap  r repair`.
- `22-tap-info-trusted`: `lthiagol/tap`, the key line `t untrust  x untap  r repair`.
- `23-tap-untrust-confirm`: `t` on the trusted tap; the UNTRUST chip (upgrade kind)
  and "Homebrew will refuse to load its formulae and casks until you trust it
  again."; then `esc`.
- `24-tap-untap-confirm`: `x` on `lthiagol/tap`; the UNTAP chip; the installed
  list (`frond`, from this tap) and "brew will refuse to untap while they are
  installed"; then `esc`.

Frames that changed against `before/`:

- `06-taps`, `13-toast-stack` (both sizes), `14-input-modal` (140×40 only): the
  Tap Info key line is new, and the footer leads with the state's actions and
  `a add tap` (was `a add  x remove  t trust  r repair`).
- `08-help`, `16-help-scrolled` (both sizes): the Taps help section reads `x Untap`,
  `t Trust or untrust tap`, `T Trust a formula or cask`. The extra row moves the
  two-column layout, and at 140×40 the count reads "of 41" (was "of 40").

All other frames are byte-identical to `before/`.

## AC-00 — trust unknown (commit 5f91936)

- `internal/brew/taps.go`: `TapsReader.List` returns `TapList{Taps, TrustUnknown}`.
  A failed `brew tap-info` gives the name-only list with `TrustUnknown` set, and
  that list is not cached. A full read is unflagged and cached as before.
- GUI: `DataLoadedMsg.TrustUnknown` → `panelData.trustUnknown`. The Taps border
  label reads `trust unknown — press R` (warning role); rows carry no trusted or
  untrusted word; Tap Info's summary and Trusted row say the same; the Trust tab
  says "unknown". `t` and `T` enqueue nothing and show `trust state not loaded — press R`.
- Tests: `taps_test.go` `TestTapsTrustUnknownOnTapInfoTimeout` (a
  `context.DeadlineExceeded` from tap-info), `TestTapsTrustKnownWhenTapInfoReads`,
  and the flag asserted in `TestTapsBatchFailureDoesNotPoisonCache`.
  `tap_trust_state_test.go`: View asserts, rows, refused `t`, `x`/`r`/`a` still
  working, and a full read clearing the state.

## AC-01 — the actions that apply now (commit 1e91fad)

- `tapActionHints(tap, trustKnown)`: untrusted `t trust  x untap  r repair`;
  trusted `t untrust  x untap  r repair`; official `r repair` (no `x`: lazybrew refuses to untap an official tap, Q-v2-10-a); trust
  unknown `x untap  r repair` (no `t`).
- The Tap Info key line is drawn from it. The footer is `tapsHints`: those
  actions, then `j/k navigate`, then `a add tap`.
- Tests: `tap_actions_test.go` (each state's exact set; the key line in Tap Info;
  footer and key line agree; the footer's leading order).

## AC-02 — t toggles, T opens the menu (commit a4eaf97)

- `t` on untrusted → trust at once (`executeTrustAction(0)`, task `trust:<name>`).
- `t` on trusted → untrust confirm: `Untrust`, upgrade kind, chip `UNTRUST`, the
  consequence line. Confirmed → `executeTrustAction(1)` through the task manager.
- `t` or `T` on official → `homebrew/core is official — no trust needed`, nothing
  enqueued and no menu.
- `T` → the existing menu (formula and cask submenus). Its "Untrust tap" entry opens
  the same untrust confirm.
- `help.go`: `x Untap`, `t Trust or untrust tap`, `T Trust a formula or cask`.
- Tests: `tap_trust_actions_test.go` (see also the matrix under AC-04).
- Changed expectations: `review_r4_test.go` and `modal_first_frame_test.go` open the
  menu with `T` (were `t`).

## AC-03 — the untap confirm, and readable errors (commit 6d5c46a)

- Untap confirm: title `Untap` (chip `UNTAP`, danger kind), question `Untap <name>?`,
  detail `The tap's formulae and casks are no longer available.` If installed
  packages are loaded with full name under `<name>/`, the body lists them and
  says `brew will refuse to untap while they are installed`. `installedFromTap`
  reads the loaded Formulae list (the unfiltered one under the leaves filter) and
  the Casks list. No brew call.
- Failed tap tasks (`tap-`, `untap:`, `repair:`, `trust:`, `untrust:`) show
  `<title>: <brew's line>`. The line is the first `Error:` line of stderr, else the
  first non-blank line. With no stderr the error text is kept. Other failures are unchanged.
- Changed expectation: `confirm_detail_test.go` `TestUntapConfirmShowsWarning` looks
  for the new question and detail.
- Tests: `tap_untap_test.go`.

## AC-04 — reload and selection (commit d1ee781)

- A failed or cancelled tap write reloads the Taps list and its trust state
  through `tapsReloadMsg` (like `servicesReloadMsg`, no "Data refreshed" toast).
  A successful one reloads through `RefreshMsg`, as before.
- `landData` keeps the cursor on the same tap by name (`selectTapNamed`). After an
  untap the name is gone, the index is kept, so the cursor is on the next tap (or
  the last tap, when the removed one was last).
- Tests: `tap_reload_test.go` (reload on each kind of failure and cancel, no
  reload for a success or a non-tap failure, cursor by name, cursor after untap,
  cursor after untapping the last tap). `tap_state_matrix_test.go` is the
  state × key matrix on the mock runner: untrusted, trusted, official and unread
  against `t T x r a`. Each cell is checked for the confirm, menu, input, toast
  or task it opens, and only the trust-at-once task reaches brew.

## Decisions and contradictions

Recorded in `work/questions.md` as Q-v2-10-a and Q-v2-10-b for the orchestrator.

- Official tap (Q-v2-10-a, orchestrator ruling): the key line and footer show
  `r repair` only, with `a add tap` in the footer. The existing guard still
  refuses `x` on an official tap ("Cannot untap official taps"). Regenerated
  frame 13 (both sizes) for this: it is the only frame with an official tap on
  screen, and its key line and footer lost `x untap`.
- Trust unread: `t` is left out of the key line and footer (the brief does not
  say). `t` and `T` refuse with the toast.
- The T menu's "Untrust tap" entry confirms (the brief names the confirm only for `t`).
- `T` is in the help and the panel's key list, not the footer (the footer set is exact).
- Trust tab, unread: says "unknown" and "Press R to read the trust state again."
- `brew tap-info` timing out in the GUI: the brew layer returns the name-only list
  with no error, so the flag is what reaches the GUI, as the brief says. The
  30 s bound is unchanged.
- The live writes (trust, untrust, tap, untap, repair) are not run on this machine.
  They need the human smoke gate.
