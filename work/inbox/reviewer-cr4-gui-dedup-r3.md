verdict: ok

# reviewer: cr4-gui-dedup r3 (base db7d464, HEAD adc42fc; fix 508df05)
## Gate
- `make ci` exit 0 (gofmt, vet, race, all 10 floors OK, build); `GOOS=darwin go build ./...` exit 0.
- 508df05 is test-only: `git diff c53e3e4 HEAD --stat -- ':!work'` → keymap_test.go only.
- deadcode/staticcheck not run (fetch denied).

## Frames and brew args
Harness re-run at HEAD and fresh at db7d464. The Services Logs tab tails the live moshi-hook log,
which grew between the two runs (39 frames differed only in log text, including where a confirm sits
over it). Re-run with the cached services JSON pointed at a frozen copy of the log: 832 frames
(dark+light × 80x24+140x40), 79 differ raw, all on header line 0 ("updated Ns ago"); 0 after
normalising it. Brew write args: 164 calls per version, identical.

## F-r2-01 — fixed
Mutations in a scratch worktree, each restored (`git status` clean):
| # | mutation | result |
|---|---|---|
| M1 | stDoctor off Status footer | FAIL :357 Status "d" |
| M2 | fFetch off Formulae help | FAIL :125, :357 Formulae "F" |
| M3 | sFetch off Search help | FAIL :125, :357 Search "F" |
| M4 | `n` off ConfirmModal.Hints | FAIL :524 |
| M5 | `enter` off InputModal.Hints | FAIL :524 |
| M6 | keyNavigate off Outdated footer | FAIL :357 Outdated "j", "down" |
| M7 | keyPanelTabs off Formulae footer | FAIL :357 Formulae "left", "]" |
| M8 | keyNavigate off tapsHints | FAIL :357 Taps "j", "down" |

## footerExempt
- Matches today's footers: Status (panelHints: no keyNavigate, no tabs) exempts j/k/down/up and
  [ ]/left/right; Outdated and Taps (keyNavigate shown, no tabs) exempt only [ ]/left/right; Search
  (no tabs) exempts [ ]/left/right. Formulae, Casks, Services show keyPanelTabs and have no entry.
  Each entry carries a reason comment.
- Minimal: dropping each panel's entry fails (Status 8 errors, Outdated 4, Taps 4, Search 4);
  dropping only Status j/k fails ("k").
- Stale entries fail: S1 exempt `d` on Outdated → :362 "which the panel does not answer";
  S2 exempt `u` on Outdated → :365 "but the footer shows it"; S3 keyPanelTabs added to the
  Outdated footer → :365 for left/right.
- By design an exemption can silence a real omission (exempt `d` on Status + drop stDoctor passes),
  but it is an explicit, reviewable entry in the test. The omission itself is LB-13 (ruled out of
  scope), not a finding.

## New problems
None. Earlier findings: F-r1-01 fixed (r2), F-r1-02 fixed via F-r2-01, F-r2-01 fixed.
