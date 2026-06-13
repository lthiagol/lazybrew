# Code & Architecture Review Template

> **Purpose:** Reusable checklist for humans and coding agents performing project reviews.  
> **Scope:** Project-agnostic — adapt section names to your stack, but keep the review dimensions.  
> **Output:** Findings report + milestone/plan updates (not drive-by code changes unless requested).

---

## How to Use This Template

1. Copy this file or reference it at the start of a review session.
2. Fill in the **Review Metadata** section.
3. Work through each dimension; record findings with severity.
4. Cross-check **Plan vs Code** and **Docs vs Reality**.
5. Produce a findings report; propose or update milestones — do not mark work complete without evidence.
6. Update the project's master plan / status tracker when the review concludes.

---

## Review Metadata

| Field | Value |
|---|---|
| **Project name** | |
| **Review date** | |
| **Reviewer(s)** | |
| **Branch / commit reviewed** | |
| **Review type** | ☐ Initial audit ☐ Milestone gate ☐ Pre-release ☐ Post-incident ☐ Periodic |
| **Scope** | ☐ Full repo ☐ Subsystem: ___ ☐ Plan only ☐ Diff only |
| **Prior review** | Link or "none" |

---

## Review Principles

Agents and reviewers should optimize for **truth over optimism**:

- Verify claims in README, plan docs, and milestone headers against the codebase.
- Prefer reproducible evidence: test output, coverage numbers, file paths, line references.
- Separate **"works in happy path"** from **"safe under concurrency/errors/edge cases"**.
- Flag **plan drift** as a first-class finding — it causes duplicate work and false confidence.
- Do not recommend rewrites unless the cost/benefit is explicit.
- Every critical/high finding should map to a tracked remediation item (milestone step or issue).

---

## Dimension 1 — Architecture & Design

### Layering & boundaries

| Question | Look for |
|---|---|
| Are layer responsibilities clear? | UI / domain / infra / entrypoint separation |
| Do dependencies point inward? | Domain should not import UI; shared kernel minimal |
| Is there a god module? | Single file/package doing orchestration + rendering + IO |
| Are interfaces at the right boundaries? | Testability without over-abstraction |

**Red flags:** circular imports, `internal` bypassed from wrong layers, duplicated domain logic in UI.

### Concurrency & state (adapt to your runtime)

| Question | Look for |
|---|---|
| Is there a single source of truth for mutable state? | One update loop, one lock strategy, one event bus |
| Are async patterns consistent? | Mixing callbacks, raw threads, and framework messages |
| Can state get stuck? | Flags set but not cleared on early return / error paths |
| Are reads blocked by writes unnecessarily? | Global "busy" locks when only writes conflict |

**Red flags:** background goroutines mutating shared model; fire-and-forget without cancellation; race-prone caches.

### Data flow & caching

| Question | Look for |
|---|---|
| Is cache keying correct? | Keys must include all dimensions that invalidate identity |
| Is invalidation complete? | Mutations clear all derived views |
| Are errors propagated or swallowed? | `_ = err`, empty catch blocks |
| Is startup work proportional? | N+1 subprocess/query storms on launch |

### Configuration

| Question | Look for |
|---|---|
| Does every config field do something? | Schema fields with no reader |
| Are defaults safe? | Fail-open vs fail-closed for security/reliability |
| Is config validated? | Invalid values clamped, rejected, or documented |

---

## Dimension 2 — Correctness & Functional Completeness

### Plan vs implementation

| Check | Method |
|---|---|
| Milestone Definition of Done | Walk each checkbox; cite code or test proving it |
| Feature inventory vs product spec | Command/API/UI action audit |
| Missing wiring | Handler exists but never called; modal result ignored |
| Stub behavior | Hardcoded strings, wrong command invoked |

### UX correctness (user-visible truth)

| Question | Look for |
|---|---|
| Does the UI show data for the current selection? | Stale content after navigation |
| Do tabs/views match their labels? | "Info" tab showing a list duplicate |
| Are empty/error/loading states honest? | Generic messages hiding failures |
| Are destructive actions confirmed? | Uninstall, delete, untap, cleanup |

### Edge cases checklist

- [ ] Empty inputs / empty collections
- [ ] Invalid selection index after data refresh
- [ ] Operation fails mid-stream
- [ ] User cancels long operation
- [ ] Duplicate rapid keypress / double submit
- [ ] Resource not found (missing binary, missing file)
- [ ] Partial failure in batch operations
- [ ] Unicode / wide characters in display truncation
- [ ] Small viewport / mobile / minimum terminal size

---

## Dimension 3 — Testing Strategy

### Coverage quality (not just percentage)

| Layer | Ask |
|---|---|
| **Unit** | Do tests assert behavior or implementation trivia? |
| **Integration** | Do tagged tests exist and run via documented command? |
| **E2E / UI** | Do tests exercise user flows and output, not just internal state flags? |
| **Regression** | Is each fixed bug backed by a test? |
| **Concurrency** | Race detector / stress tests for caches and task queues |

### Test infrastructure honesty

| Check | Evidence required |
|---|---|
| Test count in plan matches `go test -list` / equivalent | Command output |
| Coverage targets met per package | `go test -cover` or equivalent |
| CI runs the same commands as local dev | Workflow file vs Makefile |
| Flaky tests identified | Retries, sleeps, timing assumptions |

**Red flags:** 0 E2E for interactive app; integration Makefile target with no tagged tests; tests that never call render/output function.

### Minimum test matrix for a feature

When reviewing a feature, expect at least:

1. **Happy path** unit or integration test  
2. **Error path** test  
3. **User flow** test (if UI-facing)  
4. **Regression** test (if fixing a bug)

---

## Dimension 4 — Performance & Reliability

| Area | Questions |
|---|---|
| Hot paths | Blocking I/O on UI thread? Unbounded loops? |
| Startup | Parallel fan-out reasonable? Lazy loading possible? |
| Memory | Unbounded maps/caches/logs in long sessions? |
| Subprocesses | Timeouts, cancellation, zombie processes? |
| Idempotency | Safe to retry refresh/mutation? |

**Red flags:** unbounded `tabContent`-style caches; no timeout on external commands; loading entire datasets when paginated view suffices.

---

## Dimension 5 — Security & Safety

Adapt to project type; for CLI/TUI tools focus on:

| Area | Look for |
|---|---|
| Command injection | User input passed to shell |
| Path traversal | File pickers, config paths |
| Secrets | Tokens in logs, config, test fixtures |
| Privilege assumptions | sudo, MDM, multi-user |
| Supply chain | Pin deps, verify release artifacts |

---

## Dimension 6 — Documentation & Developer Experience

| Document | Should answer |
|---|---|
| README | What it is, install, run, status (honest), dev commands |
| Design doc | Architecture, boundaries, key decisions |
| Agent/contributor guide | Where to change what; test tiers; completion rules |
| Plan/status | Single source of truth for milestone state |
| Changelog / release notes | User-visible changes |

**Red flags:** README contradicts entrypoint; referenced docs missing; Makefile targets that are no-ops.

---

## Dimension 7 — Build, CI & Release

| Check | Evidence |
|---|---|
| CI on push/PR | Workflow exists and matches local test command |
| Lint/static analysis | vet, golangci-lint, eslint, etc. |
| Release pipeline | Dry-run succeeds; LICENSE/manifest complete |
| Cross-platform | Matrix or documented platform support |
| Versioning | Tags, modules, artifacts consistent |

---

## Dimension 8 — Project Management Hygiene

| Check | Look for |
|---|---|
| Milestone status matches code | Header emoji vs status index vs DoD |
| Dependencies between milestones | Circular or missing deps |
| Steps sized for execution | Anything >3 days without sub-steps |
| Each step has acceptance criteria | Not just "implement X" |
| Tests specified per step | Not deferred to end of milestone |
| Explicit out-of-scope | Prevents scope creep |
| Decision log updated | ADRs for contested choices |
| Templates followed | Milestones/steps match project `templates/` conventions (if present) |

---

## Finding Severity Rubric

| Severity | Definition | Action |
|---|---|---|
| **Critical** | Data loss, crash, security issue, corruption, stuck state | Block release; fix before next milestone gate |
| **High** | Broken feature, wrong data shown, missing tests for core path | Schedule in current phase |
| **Medium** | Maintainability, partial feature, doc drift, perf concern | Next milestone or backlog with owner |
| **Low** | Style, minor UX, nice-to-have | Backlog; don't block |

### Finding format (use consistently)

```markdown
#### [SEVERITY] Title
- **ID:** REV-001
- **Category:** architecture | correctness | testing | performance | docs | process
- **Location:** path:line or module
- **Evidence:** test output, command, screenshot description
- **Impact:** what breaks for users or developers
- **Recommendation:** specific fix or milestone step
- **Remediation tracking:** M19.6 / issue #123
```

---

## Plan vs Code Audit Worksheet

For each active milestone:

| Milestone | Plan status | Code status | Tests exist? | Gap summary | Action |
|---|---|---|---|---|---|
| M__ | ☐ Done ☐ Partial ☐ Not started | Same | ☐ Yes ☐ Partial ☐ No | | |

Rules:

- **Done** requires all DoD checkboxes verified — not "mostly works."
- If plan says Done and code disagrees, **plan is wrong** until updated.
- New gaps discovered during review become new steps — don't hide in prose.

---

## Review Output Checklist

Before closing the review session:

- [ ] Findings report written (ordered by severity)
- [ ] Metrics captured (LOC, test count, coverage, CI status)
- [ ] Plan drift documented with specific milestones affected
- [ ] New/changed milestones have: goal, deps, sized steps, DoD, tests
- [ ] Challenged prior decisions explicitly (what changed and why)
- [ ] Recommended execution order with parallel tracks noted
- [ ] `status.md` or equivalent updated
- [ ] No false "complete" markers left in plan

---

## Agent-Specific Guidance

When an agent performs a review:

1. **Read before judging** — README, plan, design doc, then code; note absences.
2. **Run tests** — `make test`, coverage, race detector if applicable; record numbers.
3. **Grep for smells** — `TODO`, `FIXME`, `program.Send`, `go func`, `_ = err`, `panic`.
4. **Trace one user flow end-to-end** — e.g. install, search, refresh — through UI to subprocess.
5. **Do not expand scope** during execution unless finding is critical; log to milestone backlog.
6. **Update plan, not just chat** — durable artifacts in master-plan/; follow [templates/README.md](templates/README.md) conventions.
7. **Challenge previous review decisions** — ask what was wrong, over-scoped, or under-sequenced.

---

## Optional: Quick Review (30-minute pass)

For small PRs or milestone gates:

1. Metadata + scope  
2. Plan vs code for touched area only  
3. Tests for changed behavior  
4. One concurrency/error-path check  
5. Docs updated if user-visible  
6. Finding list + go/no-go recommendation  

---

## Version History

| Date | Change |
|---|---|
| 2026-06-13 | Initial template (lazybrew planning session) |
