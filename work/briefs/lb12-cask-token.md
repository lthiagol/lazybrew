# Slice: lb12-cask-token — cask rows carry the token, not the display name

> **Contract.** This brief wins over any plan, note, or prior convention where
> they differ. `work/briefs/cr-common.md` applies (bug fixes start with a
> failing test). If the repo contradicts the brief, follow the brief and
> report the contradiction. Don't stop to ask.

**Depends on:** `cr5-dead-code` closed. Backlog item LB-12 (`Q-cr1-cask-name`).

## Problem

`brew info --json=v2 --installed --cask` returns, per cask, `"token":
"claude-code"` and `"name": ["Claude Code"]`. `parseCask` reads only `name`,
so `Cask.Name` is the display name. The Casks panel shows it, and every cask
action (info, upgrade, uninstall, reinstall, zap…) sends it to brew, which
rejects `Claude Code`. `brew outdated --json=v2 --cask` already returns the
token in `name`, so the Outdated rows are right and the two lists disagree.

## Acceptance criteria

To-do (one commit each, in this order):

- [ ] AC-01 — `parseCask` reads `token` [LB-12]
- [ ] AC-02 — every consumer of `Cask.Name` is right with a token

### AC-01 — `parseCask` reads `token`

1. Add `Token string \`json:"token"\`` to `caskJSON`. `Cask.Name` becomes
   `token`, falling back to the first display name when `token` is empty
   (older brew shapes, the existing fixtures). Keep the display name in
   `Cask.FullName`'s sibling if one exists, or add `Cask.DisplayName`; use it
   where the UI wants a human title (details header), and say which you chose.
2. **Test first**, in the current brew shape: a fixture
   `internal/brew/testdata/casks_installed_token.json` with a cask whose
   `token` is `claude-code` and `name` is `["Claude Code"]`. Assert
   `Name == "claude-code"`. Show it failing before the fix in the receipt.
3. Add a second case with no `token` that falls back to `name`.

### AC-02 — consumers

1. Grep every reader of `Cask.Name` (sidebar rows, details, actions, outdated
   merge, search, preload, `selectedName`). The Casks panel rows and every
   action must use the token. Where the UI shows the display name, show it
   next to the token or in the details only; do not change a frame unless the
   token replaces the name. Name each golden that changes.
2. A gui test: select the Casks row built from the new fixture, press the
   upgrade/info key, and assert the brew args carry `claude-code`.
3. Check the Outdated ↔ Casks join (if any) matches on the token.

## Done means

- `make ci` green. Receipt shows the failing run, the golden list, and
  `git diff --shortstat`.
- Invariants: no write goes around `task.Manager`; floors not lowered.
