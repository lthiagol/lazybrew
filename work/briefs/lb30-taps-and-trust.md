# Slice: lb30-taps-and-trust — one pane for taps and what you trust

> **Contract.** This brief wins over any plan, note, or prior convention where
> they differ. `work/briefs/cr-common.md` applies. If the repo contradicts the
> brief, follow the brief and report the contradiction. Don't stop to ask.

**Depends on:** `lb29-footer-redesign` closed. User rulings (2026-11-21), backlog
LB-28: the Taps pane is named **`Taps & Trust`** and also lists the individually
trusted **formulae, casks and commands**, **commands included**. Never launch the TUI
against the real brew; no brew write commands (read-only `brew trust --json=v1` is
fine to check the real shape).

## Facts

- `brew trust --json=v1` prints `{"taps":[…],"formulae":[…],"casks":[…],"commands":[…]}`
  (full names: `lthiagol/tap/lazybrew-dev`, `openai/tools/tart`).
- `brew untrust --tap|--formula|--cask|--command <name>` stops trusting one.
- `brew.TrustReader.ListTrusted` expects another shape (`{"entries":[{name,type,tap}]}`)
  and has **no production caller**: it never worked on the real output. The tap trust
  state in the pane comes from `brew tap-info` (`Tap.Trusted`), unchanged.
- Today `t` on a tap row trusts/untrusts the tap, `T` opens the per-formula/cask trust
  menu of that tap (`startTrustMenu`), `x` untaps; the writers have `TrustTap`,
  `TrustFormula`, `TrustCask`, `UntrustTap` only.

## Behaviour

- The sidebar title reads `Taps & Trust` (help, README, tests, `?` pages too; key `5`
  unchanged; fits in the title row at 80x24 with the count).
- The list is the taps (as today), then the trusted items: one row per formula, cask
  and command from `brew trust --json=v1` that is **not a tap** — in that order within
  the group (formulae, casks, commands), alphabetical in each. A trusted-item row:
  the ✔ mark, the full name on the left, `formula`/`cask`/`command` right-aligned in
  the subtle style (as Search rows). The title count is the number of rows.
- No separator rows: the cursor stays simple (taps first, items after; the right-hand
  column tells them apart).
- **Main area for a trusted row**: only an `Info` tab (the tab row shows just `Info`;
  `h/l`/arrows do nothing on tabs). Info shows: Name, Type, Tap (the tap part of the
  name), `Tap trusted: yes/no/unknown`, and one sentence: *trusted on its own; Homebrew
  may load it even if its tap is not trusted*. Taps rows keep their tabs.
- **Keys on a trusted row**: `t` untrusts it, with a confirm that names the consequence
  (*Homebrew will not load <name> until it is trusted again*), through `task.Manager`
  (`brew untrust --formula|--cask|--command <name>`); `x` (untap) and `r` (repair) are
  refused with an info toast naming `t`; `T` and `a` behave as on any row of the pane
  (`a` adds a tap; `T` needs a tap row and says so otherwise). Footer and help are
  contextual as `tapsHints(panel)` already is.
- After any trust/untrust (tap or item) the cache group is dropped, both lists reload,
  and the cursor **stays on the same row** (by tap name or item name; if the item is
  gone, clamp to the nearest). `R`, auto-refresh and the preload keep both groups in
  step. Empty states: no taps and no trusted items → today's empty message; taps but no
  items → no item rows (nothing says "none").
- A read failure of the trust list leaves the taps usable and shows no item rows plus
  an error toast once (not on every refresh); `trustUnknown` semantics for taps are
  unchanged.

## Acceptance criteria

To-do (one commit each, in this order):

- [ ] AC-01 — brew layer: the real trust-list shape, untrust writers
- [ ] AC-02 — the pane: rename, rows, selection, Info tab, reloads
- [ ] AC-03 — actions, footer/help/keymap, README, flows

### AC-01
Replace the stale `trustJSON` with the real shape; `ListTrusted` returns the entries
with `Name`, `Type` (formula/cask/command; taps become `TrustTypeTap` entries) and `Tap`
derived from the full name (`owner/repo/name` → `owner/repo`; a name without a tap part
keeps `Tap` empty). Add `UntrustFormula`, `UntrustCask`, `UntrustCommand` (and
`TrustCommand` only if a caller needs it: no) to `TrustWriter`, with the `untrust`
cache group invalidation the table already has. Fixture
`internal/brew/testdata/trust_v1.json` in the real shape (taken from the live
`brew trust --json=v1` output, names anonymised if you like). Tests first: parse, the
tap derivation (3-part, 2-part, 1-part names), empty lists, malformed JSON → error,
writers' args (`untrust --formula x`), cache invalidation. Remove anything now unused
(deadcode).

### AC-02
Load the trust list with the Taps panel data (one more bounded read, cached, in the
same `fetchPanelData` path and `ReloadDiagnostics`/refresh paths; a failure does not
fail the taps). `panelData` carries the trusted items; `itemCount`, `nameAt`,
`selectedTap()` (nil on an item row), a new `selectedTrusted()`, clamp/normalize,
mouse rows and the wheel all see taps+items. Row cells (`typedRow`/`tapCells`-style),
title `Taps & Trust` (grep every user-visible `Taps`: sidebar title, help, README,
footers; keep the key name `taps` in code where it is an identifier), the Info tab and
the single-tab row for item rows. Cursor kept by name across reloads. Tests first;
frames at 80x24 and 140x40, both themes: taps only, taps+items, item selected (Info),
long names truncated with `…` keeping the type visible.

### AC-03
`t`/`x`/`r`/`T`/`a` on item rows per the behaviour above; the untrust confirm and
task; keymap/footer/help (`tapsHints` gets a branch for an item row); README. Flows
(teatest): select a trusted item, `t`, confirm, brew args `untrust --formula <name>`
recorded by the mock, the row disappears and the cursor lands on a neighbour; `x` on an
item toasts and runs nothing. Name each golden that changes.

## Done means

`make ci` green (`GOFLAGS=-p=1` if /tmp is tight); `GOOS=darwin go build ./...`; floors
not lowered; deadcode/staticcheck (installed at
`/tmp/claude-1000/-home-agent-coding-lazybrew/020ffb7e-4ae6-44f3-a324-1eba91f46e86/scratchpad/tools-bin/`)
show nothing new. Disk is tight: no large scratch. Receipt
`work/inbox/runner-lb30-taps-and-trust.md` (≤ 45 lines, committed, not edited
afterwards), one commit per AC, tests first, subjects end `[LB-28]`.
