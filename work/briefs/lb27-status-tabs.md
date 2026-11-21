# Slice: lb27-status-tabs — Missing and Vulns tabs on Status; `d`, `m`, `v` open them

> **Contract.** This brief wins over any plan, note, or prior convention where
> they differ. `work/briefs/cr-common.md` applies. If the repo contradicts the
> brief, follow the brief and report the contradiction. Don't stop to ask.

**Depends on:** `lb26-list-filter` closed. User request (2026-11-21), backlog
LB-24. Never launch the TUI against the real brew; no brew write commands
(read-only `brew missing`/`brew vulns` are fine for a one-off check of the real
output shape: `brew vulns` queries the OSV.dev database over the network, so run
it once at most).

## Today

Status has tabs Dashboard, Config, Doctor. `d`, `m`, `v` run `brew doctor`,
`brew missing`, `brew vulns` as Operation-overlay tasks (`tasks_diagnostics.go`:
`runDoctor`, `runMissing`, `runVulns`): the result shows in the overlay and
vanishes when it closes. The Doctor **tab** loads its own content
(`fetchTabContentCmd`, `statusDoctorTab`, `doctorView`), cached in `tabContent`.

## Behaviour

Status gets tabs **Dashboard | Config | Doctor | Missing | Vulns**.

- **Missing** loads like Doctor when the tab is opened (cached in `tabContent`,
  kept across panel switches): `formula: missing dep` rows, or `All dependencies
  satisfied` (✔ style as the doctor "ready" line).
- **Vulns** never runs by opening the tab (it is slow and uses the network). The
  tab shows `Press v to check installed formulae for known vulnerabilities
  (queries OSV.dev)` until run; then it shows the report, kept until refreshed;
  while running it shows the spinner and `Checking…`. An error (command missing,
  network) shows brew's own first error line in the error style plus the same
  hint to retry with `v`. Empty report: `No known vulnerabilities`.
- **Keys** (Status panel): `d`, `m`, `v` switch to the Doctor, Missing, Vulns tab
  and (re)load it: `d` and `m` refresh their data, `v` runs the check. They no
  longer open the Operation overlay. `R` refreshes Doctor and Missing when they
  were opened (as `ReloadDiagnostics` does for Doctor today) and never runs Vulns.
  The tab keys (`h/l`, arrows) reach all five tabs.
- Output is wrapped to the main width without cutting words; `y` copies the
  whole tab text (lb24) ANSI-free and uncut.

## Acceptance criteria

To-do (one commit each, in this order):

- [ ] AC-01 — Missing tab
- [ ] AC-02 — Vulns tab (on demand)
- [ ] AC-03 — `d`/`m`/`v` open the tabs; remove the overlay runners

### AC-01
Tab entry (`panelTabs`, `tabs.go` constant), `fetchTabContentCmd` case, a
`missingView` formatter in `presentation` or beside `doctorView`, render case
beside Config/Doctor (`render.go`, `panels.go` tab-fetch gate `needsTabFetch`),
`R` reload. Tests first: parser rows → view; empty → satisfied line; exit-1 with
output (brew exits 1 when something is missing: `Diagnostics.Missing` already
handles it); error → error body; tab content cached across a panel switch;
frames at 80x24 and 140x40, both themes.

### AC-02
Tab entry, a run path that is **not** triggered by opening the tab, per-tab
state (not-run / running / done / error) kept in the model (a typed state, not a
magic string in `tabContent`), a `tea.Cmd` that calls `Diagnostics.Vulns` through
a bounded read (`boundedRead`, timeout like the other reads; long: say what you
chose) and returns a message. No write, so no `task.Manager`. Tests first: opening
the tab runs nothing (mock runner records zero calls); `v` runs it once, a second
`v` while running does nothing; result, empty result, exit-1 report, error;
`R` does not run it; frames.

### AC-03
`d`/`m`/`v` in `keys.go` switch tab + trigger the loads above; delete
`runDoctor`/`runMissing`/`runVulns` and their overlay tests if now unused
(deadcode must stay clean); keymap/footer/help (`stDoctor`, `stMissing`,
`stVulns` text: `doctor`, `missing`, `vulns` keep their keys; help lines say they
open the tab); README Status section. A teatest flow: Status → `m` → tab shows the
rows (mock runner), `v` → hint, run, report, `h`/`l` cycle through all five tabs.
Name each golden that changes.

## Done means

`make ci` green (`GOFLAGS=-p=1` if /tmp is tight); `GOOS=darwin go build ./...`;
floors not lowered; deadcode and staticcheck (installed at
`/tmp/claude-1000/-home-agent-coding-lazybrew/020ffb7e-4ae6-44f3-a324-1eba91f46e86/scratchpad/tools-bin/`)
show nothing new. Disk is tight: no large scratch. Receipt
`work/inbox/runner-lb27-status-tabs.md` (≤ 40 lines, committed, not edited
afterwards), one commit per AC, tests first.
