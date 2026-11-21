verdict: ok

Review of `c05ebf0..6ba4360` against `work/briefs/lb20-cask-followups.md` + `cr-common.md`.

| AC | Commit | Verdict | Evidence |
|---|---|---|---|
| AC-01 LB-17 | c05ebf0 | ok (trust-menu part pending `Q-lb20-trust-menu`) | `parseCask` FullName = full_token, else full_name. Fix reverted in a temp worktree: `casks_token_test.go:112: FullName = "", want the full_token "user/tap/bar"`, `tap_untap_test.go:155: confirm lacks "gcloud"` |
| AC-02 LB-19 | 179adf2 | ok | Fix reverted: `search_info_test.go:226: row 0 … [info --json=v2 same], want [info --json=v2 --formula same]`, same for row 1 with `--cask`. The unknown row keeps the no-flag call |

## Verify
- **V1:** `make ci` exit 0, all floors met (gui 90.7 ≥ 83, brew 87.8 ≥ 78, rest unchanged). `GOOS=darwin go build ./...` ok.
- **V2:** each fix reverted alone at 179adf2; its tests fail (lines above). They pass again with the fix back.
- **V3:** live read-only `brew info --json=v2 --installed --cask` and `--cask claude-code` run through `ParseInfo`:
  claude-code → `FullName="claude-code"`, gcloud-cli → `"gcloud-cli"`, tap homebrew/cask, no `full_name` key. The fixture matches.
  Default tap: the FullName has no `homebrew/cask/` prefix, so `installedFromTap` doesn't match it. That's harmless:
  `confirmUntap` refuses official taps before it reads the list. Third-party tap: no such cask is installed here;
  covered by the `user/tap/gcloud` payload test. Formulae code is untouched.
  Trust menu: confirmed that `startTrustMenu` reads `tap.FormulaNames/CaskNames` from tap-info and no installed package.
  The brief's premise doesn't hold, and the runner filed it correctly.
- **V4:** the no-flag, `--formula` and `--cask` payloads are byte-identical live (alsa-lib, claude-code), and `--cask` parses.
  Own full `View()` test for a formula row and a cask row at 80x24 and 140x40, at a04af85 vs 179adf2: frames identical.
- **V5:** one commit per AC, in order, subjects per cr-common. No floor, Makefile, scripts or go.mod changes. No new goroutines.

## Findings
- F-r1-01 (low, non-blocking) — `work/inbox/runner-lb20-cask-followups.md` — the committed receipt is 58 lines,
  over the 40-line limit (cr-common report protocol). A trimmed version sits uncommitted in the worktree. I didn't
  stage it (not mine). Fix must achieve: a committed receipt of ≤ 40 lines.

## For the orchestrator
- Rule on `Q-lb20-trust-menu`. Recommended: accept the runner's reading, untap confirm only. The trust menu
  lists the tap's own formulae/casks from tap-info, which is correct.
- LB-17 and LB-19 are still marked parked in `work/backlog.md`.
