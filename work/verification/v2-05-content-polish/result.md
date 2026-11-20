# v2-05-content-polish — frame result

`before/` is `work/verification/v2-04-overlays-styled/after/` (frames 01–15).
`after/` is this slice, dumped by `frame_dump_test.go.txt` (the v2-04 harness
with three new screens, 16–18). `140x40` has `.txt` and `.ansi`; `80x24` has
`.txt` only, as before.

Captured with the real Homebrew on this machine (3 taps, 0 official, 0
outdated), the Unicode icon set, and read commands only. No brew write ran.
Doctor ran once, in the dashboard load the harness makes at start; frame 18
reads that cached result and runs nothing new. Nothing was confirmed.

## Per screen

| screen | 140x40 | 80x24 | what changed |
|---|---|---|---|
| 01-status | changed | changed | The Dashboard tab is five cards (Homebrew, Packages, Taps, Services, Health) instead of one text block. At 140 wide the cards sit in pairs (Homebrew + Packages, Taps + Services, then Health). Outdated is `not loaded — press R`, as no `brew outdated` ran. Health shows `✓ ready to brew`. At 80 wide the cards stack; the Services and Health cards sit below the fold and need scrolling (no row is cut sideways). |
| 02-formulae | changed | changed | The Info tab of the selected formula uses the detail rows: one label column (Muted), values in Text, the homepage in Info. Status is `✓ installed` (Success). |
| 03-formulae-info | changed | changed | Same detail rows for bash. The label column is one width for the view (`Description:` sets it). |
| 04-formulae-deps | changed | changed | Deps shows `Runtime` (Accent heading, no colon) and each dependency with an installed icon: `✓ readline` (installed). |
| 05-outdated | same | same | No outdated packages on this machine, so the empty list is unchanged. |
| 06-taps | changed | changed | Tap Info uses the detail rows (Trusted, Official, Remote in Info, counts). |
| 07-confirm-modal | changed | changed | The base frame is the Formulae Info rows (detail style); the dialog is unchanged. |
| 08-help | changed | changed | Help is the new layout: two columns at 140 (whole pages, balanced by height), the key column width per section, headings and keys in Accent, descriptions Subtle, the position hint `1–37 of 40` on the top border. The `(g/G outside Search)` note is on one row. |
| 09-search | same | same | No search result open. |
| 10-toast | changed | changed | Base frame only: Formulae Info rows changed; the toast box is unchanged. |
| 11-deps-bash | changed | changed | Deps of bash: `✓ ncurses` (installed), `✓ readline` (installed), under Runtime. |
| 12-confirm-upgrade | changed | changed | Base frame only: Formulae Info rows changed; the upgrade dialog is unchanged. |
| 13-toast-stack | same | same | The base is Taps, whose main area reads `No custom taps` (the injected tap is official, so no detail row is drawn); it is unchanged. The three toasts are as in v2-04. |
| 14-input-modal | changed | changed | Base frame only: Tap Info rows changed; the input modal is unchanged. |
| 15-menu-brewfile | changed | changed | Base frame only: the Status dashboard is now cards; the Brewfile menu is unchanged. |
| 16-help-scrolled | new | new | Help after one page down: `4–40 of 40` (the last full page; the page-down clamps). The first rows are gone and the box is still framed, with `?/esc close` on the footer. |
| 17-cask-info | new | new | Casks → first cask Info (Claude Code): Name, Version, Tap, `✓ installed`, Homepage in Info, Description. The label column is set by `Description:`. |
| 18-doctor | new | new | Status → Doctor, from the cached doctor run: `✓ Your system is ready to brew.` in Success. Each warning would be a warning-icon heading and its details (no warning on this machine, so no block is drawn). |

## What the frames do not show

- The Operation view is covered by tests, not frames (no brew writes).
- The cask Deps tab for gcloud-cli (`python@3.14`) is covered by tests from the
  real `brew info --json=v2 --cask --installed` fixture; the frame set does
  not open it.
- At 80×24 the five dashboard cards need scrolling to reach Services and
  Health. Nothing overflows the width.

## Fix round r1 (F-r1-04)

Health is the second card, so stacked at 80×24 it is on the first screen; a
failed doctor reads `⚠ doctor unavailable: <reason>` instead of `press d to run
doctor`. Only the dashboard changed, so two frames moved, at both sizes where
they exist. Every other frame is byte-identical to the previous capture.

- **01-status** (140x40 and 80x24, `.ansi` at 140x40): the cards are Homebrew,
  then Health, Packages, Taps, Services. At 140 wide the first pair is Homebrew
  and Health. At 80 wide Health sits under Homebrew, with `✓ ready to brew`
  on row 5; Packages, Taps and Services follow and still need scrolling.
- **15-menu-brewfile** (140x40, `.ansi`; 80x24 is unchanged): the base frame is
  the same dashboard, so its cards are in the new order. The Brewfile menu is
  unchanged.
