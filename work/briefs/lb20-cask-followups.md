# Slice: lb20-cask-followups — installed casks know their tap; Search info asks for the right kind

> **Contract.** This brief wins over any plan, note, or prior convention where
> they differ. `work/briefs/cr-common.md` applies (bug fixes start with a
> failing test). If the repo contradicts the brief, follow the brief and
> report the contradiction. Don't stop to ask.

**Depends on:** `lb19-keys-and-search` closed. Backlog items LB-17 and LB-19.

## AC-01 — LB-17: installed casks carry `full_token`

`brew info --json=v2 --installed --cask` has `full_token` (and `tap`) but no
`full_name`, so `Cask.FullName` is empty for installed casks, `installedFromTap`
never lists them, and the untap warning misses them. Add `FullToken` to
`caskJSON` and fill `Cask.FullName` from it (fall back to `full_name`, then
empty). Failing test first with the real shape (reuse
`internal/brew/testdata/casks_installed_token.json`; add `full_token` if the
fixture lacks it, matching `brew info` on this machine: read-only
`brew info --json=v2 --cask claude-code` shows it). Then a gui test: a tap that
owns an installed cask names it in the untap confirm (find the current
wording in `confirmUntap`); and the tap-trust menu lists it where formulae
are listed. Name each golden that changes.

## AC-02 — LB-19: Search → info passes `--cask` for cask rows

`fetchSelectedSearchInfo` (`tasks_packages.go`) runs `brew info --json=v2
<name>` for every row. The row's type is certain since lb19: for a cask row run
`brew info --json=v2 --cask <name>`, for a formula row `--formula <name>`.
Failing test first (mock runner records the args, for a name that is both a
formula and a cask: the two rows send different args). A row of unknown type
(fixtures only) keeps today's no-flag call. The parsed info (`brew.ParseInfo`)
must still work for the `--cask` shape; check with the fixture. No frame
changes expected; name any that do.

## Done means

`make ci` green; `GOOS=darwin go build ./...` ok; floors not lowered. Invariants:
Bubble Tea rules; no write outside `task.Manager`; never launch the TUI; no
brew write commands (read-only `brew info` is fine). Receipt
`work/inbox/runner-lb20-cask-followups.md` (≤ 40 lines), one commit per AC.
