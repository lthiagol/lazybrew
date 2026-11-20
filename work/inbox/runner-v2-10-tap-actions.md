# Receipt — runner, v2-10-tap-actions

## Commits (one per AC, in order)

- AC-00 5f91936 — trust unknown on a name-only Taps list (TapList flag, GUI, t/T refused)
- AC-01 1e91fad — state-aware key line in Tap Info and the footer
- AC-02 a4eaf97 — `t` toggles trust (untrust confirmed), `T` opens the menu, official refused
- AC-03 6d5c46a — untap confirm names the consequence and installed packages; tap errors say brew's line
- AC-04 d1ee781 — failed or cancelled tap write reloads; cursor kept by name
- AC-05 f05b443 — harness, before/after frames, result.md, STATE row, questions

## Evidence per AC

- AC-00: `TestTapsTrustUnknownOnTapInfoTimeout`, `TestTapsTrustKnownWhenTapInfoReads`; `tap_trust_state_test.go` (label and Tap Info say "trust unknown — press R", no trust word on rows, `t` refused with the toast, `x`/`r`/`a` still open).
- AC-01: `tap_actions_test.go` (each state's exact set; key line in Tap Info; footer and key line agree).
- AC-02: `tap_trust_actions_test.go` (t trusts at once; t on trusted asks; cancel runs nothing; t and T refused on official; T opens the menu; menu untrust confirms).
- AC-03: `tap_untap_test.go` (consequence; installed list from loaded lists, leaves filter included, no brew call; failure text is brew's `Error:` line).
- AC-04: `tap_reload_test.go` (reload on each failure and cancel; cursor by name; cursor after untap); `tap_state_matrix_test.go` (untrusted, trusted, official, unread × t T x r a, on the mock runner).
- AC-05: 24 frames at 140x40 (.ansi and .txt) and 24 at 80x24; every .txt is exactly W×H. Changed vs before: 06, 08, 13, 16 (both sizes), 14 (140x40). Frames 21–24 are new.

## Gate (run once, before this receipt, on the AC-04 tree)

- `make build` — ok
- `make lint` — ok
- `CGO_ENABLED=1 go test ./... -race -count=1` — every package ok
- `make cover-check` — All coverage floors met (brew 79.6%, gui 85.7%, presentation 98.5%, modal 90.1%)

## Safety

No brew write command was run. The harness only reads. It sent `t` once, to `lthiagol/tap`, after a guard confirmed a trusted, non-official tap; the confirm was cancelled with `esc`. `x` opened the untap confirm, cancelled with `esc`. No `y` or `enter` was sent. `danielfoehrkn/switch` (untrusted) never received `t`.

## For the next session

- Live trust, untrust, tap, untap and repair are not verified against real brew. They need the human smoke gate.
- Q-v2-10-a: an official tap's key line and footer offer `x untap` (brief AC-01), but `x` is refused there. Kept per brief; one line to change in `tapActionHints`.
- Q-v2-10-b: the T menu's "Untrust tap" entry confirms too; `T` is not in the footer (exact set). Both are open in `work/questions.md`.
- Interface change: `TapsReader.List` returns `TapList{Taps, TrustUnknown}`, not `[]Tap`. Three callers updated.
- Changed expectations (named in the commits): `review_r4_test.go` and `modal_first_frame_test.go` open the menu with `T`; `confirm_detail_test.go` looks for the new untap text; `sidebar_list_test.go` passes `trustKnown`; `taps_test.go` reads `.Taps`. None weakened.
- Worktree clean after this receipt. `bin/lazybrew` is gitignored.
