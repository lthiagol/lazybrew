# Slice: lb32-vulns-view — a readable, coloured Vulns tab with next steps

> **Contract.** This brief wins over any plan, note, or prior convention where
> they differ. `work/briefs/cr-common.md` applies. If the repo contradicts the
> brief, follow the brief and report the contradiction. Don't stop to ask.

**Depends on:** `lb31-cleanup-tabs` closed. User request (2026-11-21), backlog LB-30: the
Vulns tab shows brew's raw text; render it better (tables), with colours, and tell the user
what to do. **Do not run `brew vulns` yourself** (network); use the saved real output:
`/home/agent/.cache/vulns-real.json` (copy it into `internal/brew/testdata/vulns_v1.json`; it
has no secrets). Never launch the TUI against the real brew; no brew write commands.

## The real data (`brew vulns --json --max-summary 0`, exit 1 when findings exist)

```
{ "findings":[{"formula","version","tag","repo_url",
     "vulnerabilities":[{"id","severity","summary","aliases":[],"fixed_versions":[…]}],
     "patched":[{…same shape…}]}],
  "skipped_formulae":[names], "checked":90, "complete":false,
  "outdated_without_sbom":[], "unscanned_formulae":{} }
```
Severities seen: HIGH, MEDIUM, UNKNOWN (also CRITICAL, LOW exist). `fixed_versions` holds
40-hex commit SHAs for unreleased fixes; a real version string (no 40-hex) means a released fix.
`patched` are findings already resolved by a formula patch (not counted by brew).

## Today (lb27)

`Diagnostics.Vulns` runs `brew vulns` and returns the raw text; the tab (`gui/vulns.go`)
wraps it. States, the `v` key, the bounded read and the exit-1 rules stay as lb27 made them.

## Behaviour

Run `brew vulns --json --max-summary 0` (same states and exit-code rules: exit 1 **with** a
JSON report is Done; exit 1 with no/invalid JSON is the failure body with brew's first error
line). If the JSON cannot be parsed, fall back to showing brew's stdout as plain wrapped text
(today's view), with a one-line muted note `Could not read brew's JSON; showing its text.`

The report view, top to bottom:
1. **Summary strip:** `24 vulnerabilities in 7 of 90 packages` and, muted, `37 not scanned`
   (`skipped_formulae` length; plus `outdated_without_sbom`/`unscanned_formulae` counts when
   non-zero); then severity chips in a row, worst first, only those present:
   `CRITICAL 2  HIGH 9  MEDIUM 3  LOW 1  UNKNOWN 12`.
2. **One block per affected package**, sorted by worst severity (CRITICAL > HIGH > MEDIUM >
   LOW > UNKNOWN), then vulnerability count (desc), then name. Block header: `name version`,
   the worst-severity chip, `×N`, and the **next step** (below), right-aligned if it fits.
   Under it a table `ID | Severity | Summary` (use the existing `renderTable`/`pkgColumns`
   machinery if it fits; severity chip fixed width; summary **wrapped** to the width, not
   cut), vulnerabilities sorted worst first then id.
3. **Footer, muted:** `14 resolved by formula patches (libssh2 14)` (count of `patched` per
   package) and `37 packages not scanned (no supported source URL)`; the names of the skipped
   packages are only in the `y` copy.

Colours come from the theme roles, never raw colours: CRITICAL = error + bold, HIGH = error,
MEDIUM = warning, LOW = info, UNKNOWN = muted. Next-step colours: success / warning / muted
as below. Chips are text (e.g. ` HIGH `) in the role colour; no background fills unless the
theme has a role for it.

**Next step per package** (a short phrase, one of):
- the package is in the loaded **Outdated** list (`m.panels[PanelOutdated]` formulae, when
  that data is loaded): `upgrade → <new version>` in the success colour (press `4`, then
  `u`); if Outdated is not loaded, skip this rule (no wrong claim);
- else some vulnerability has a **released** fix version: `fix in <version>` (warning),
  with the highest such version string shown;
- else only commit SHAs: `fixed upstream, not released` (muted);
- else: `no fix yet` (warning).
Packages whose vulnerabilities are all UNKNOWN keep the same rule but the phrase stays muted.

`y` copies the whole tab as plain text: everything the view shows **plus** the full
`fixed_versions` lists, the skipped package names and the patched IDs (ANSI-free, uncut,
including wrapped summaries joined back to one line each).

Long reports must be reachable: if tab bodies cannot scroll today (check how Doctor, Logs
and Files do), say so in a question and do not invent new keys; otherwise use the existing
mechanism.

## Acceptance criteria

To-do (one commit each, in this order):

- [ ] AC-01 — brew layer: the typed JSON report and the fallback
- [ ] AC-02 — the report view (summary, blocks, tables, colours, footer)
- [ ] AC-03 — next steps from Outdated, `y` copy, docs, flows

### AC-01
`Diagnostics.VulnsReport(ctx)` (keep `Vulns` only if still used; deadcode must stay clean):
typed `VulnReport{Findings[], Skipped[], Checked, Complete, …}` parsed from the saved real
JSON (`testdata/vulns_v1.json`), exit-code rules unchanged (exit 1 + JSON = report; exit 1 +
no JSON = error; invalid JSON = typed fallback carrying brew's raw stdout). `SeverityRank`,
`ReleasedFix(v)` (not a 40-hex SHA). Tests first: the real fixture (counts: 24 findings in
7 packages, 14 patched), a CRITICAL/LOW synthetic, empty findings, exit 1 without JSON,
malformed JSON, fixed_versions SHA vs version.

### AC-02
The view in `gui/vulns.go` (or beside it), per the behaviour: summary, chips, blocks sorted
as stated, tables, wrapped summaries, muted footer, the fallback note. Frames at 80x24 and
140x40, both themes, from the real fixture and from a no-findings report (`✔ No known
vulnerabilities in N packages checked`, with the not-scanned count): nothing wider than the
width, no clipped word, colours by role (assert SGR per role in a test).

### AC-03
Next steps (rules above) with the Outdated cross-reference and its "not loaded" case; `y`
copy; README Status/Vulns text; keymap/help unchanged unless you add a hint; a teatest flow:
Status → `v` → report from the fixture via the mock runner → rows visible → `y` payload has the
full SHAs. Name each golden that changes.

## Done means

`make ci` green (`GOFLAGS=-p=1` if /tmp is tight); `GOOS=darwin go build ./...`; floors not
lowered; deadcode/staticcheck (installed at
`/tmp/claude-1000/-home-agent-coding-lazybrew/020ffb7e-4ae6-44f3-a324-1eba91f46e86/scratchpad/tools-bin/`)
show nothing new. Disk is tight: no large scratch. Receipt
`work/inbox/runner-lb32-vulns-view.md` (≤ 45 lines, committed, not edited afterwards), one
commit per AC, tests first, subjects end `[LB-30]`.
