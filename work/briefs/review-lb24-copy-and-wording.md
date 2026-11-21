# Review brief — slice `lb24-copy-and-wording`

> Contract: `work/briefs/lb24-copy-and-wording.md` (with `cr-common.md`). You
> review; you do **not** edit code, tests or commits. Review everything on `wip`
> after the commit that added the slice brief
> (`git log --oneline -- work/briefs/lb24-copy-and-wording.md`) up to HEAD. A
> receipt is a claim, not evidence. `CGO_ENABLED=1` for the race gate. No brew
> write commands; never launch the TUI against the real brew; scratch under
> /home/agent/.cache; do not touch other panes.

## Verify
- **V1:** `make ci`; `GOOS=darwin go build ./...`.
- **V2 copy payload:** drive the Model (mock runner) and capture the
  `SetClipboard` payload: failed, running and done overlays; Formulae/Casks Info,
  Deps, a Services Logs tab, Status; the payload has no ESC sequences, is not
  clipped (long lines and long outputs intact, no `…` from the view), and
  matches what the user sees. Real OSC 52 output: run a tiny harness that
  executes the returned cmd and check the escape written is `ESC ] 52 ; c ;
  <base64> BEL/ST` with the payload base64-decoding to the text.
- **V3 keys:** `y` while the search input or the add-tap input is focused types
  `y`; with a confirm/modal open it does not copy behind it; no collision in the
  key table; keymap test (footer, help, dispatch) passes and fails if the entry
  is dropped (mutate it).
- **V4 wording (AC-02):** retry offered → new line and the command below; not
  offered (e.g. failed `update`, services needs-sudo, a non-sudo failure) → text
  unchanged. Own `View()` diff base vs HEAD at 80x24 and 140x40: only those
  frames and footers/help differ.
- **V5:** README and config sample comment are right and consistent; one commit
  per AC; floors not lowered; no new goroutines; the clipboard write is a
  `tea.Cmd`, not `program.Send`.

## Report
`work/inbox/reviewer-lb24-copy-and-wording-rN.md` (≤ 45 lines): first line
`verdict: ok` or `verdict: changes-required`; AC table; findings `F-rN-01
(severity) — file:line — what — evidence — fix must achieve`. Commit once
(`inbox: reviewer reports lb24-copy-and-wording rN`), worktree clean, reply with
one line: `done: work/inbox/reviewer-lb24-copy-and-wording-rN.md`.
