# M25 — Diagnostics Error Handling

> **Status:** 🔜 Planned  
> **Size estimate:** S (<1 day)  
> **Depends on:** M22 ✅ (release tag)  
> **Enables:** M27 (clean diagnostics tab / status dashboard behavior)  
> **Parallel track:** —  
> **Gate criteria:** `brew doctor` warnings are displayed without treating exit code 1 as a fatal error, and no false "brew failed" log spam remains.

<!-- See [archive/architecture-review-2026-06-13.md](../archive/architecture-review-2026-06-13.md) correctness findings -->

---

## Goal

`brew doctor` returns exit code 1 whenever it has warnings, but lazybrew treats that as a hard failure and logs a warning every refresh cycle. This milestone fixes that semantics bug so diagnostics are shown correctly and the debug log stops filling with false failures.

---

## Readiness to Start

Before executing this milestone, confirm:

- [ ] M22 release tag is done.
- [ ] A sample of `brew doctor` output that produces exit=1 is available for test fixtures.
- [ ] `BrewExitError` exposes `ExitCode` (already true in `internal/brew/errors.go`).

---

## Why Now

- The debug log shows `brew doctor` failing every ~60 s with exit=1. This is the only recurring error in the log.
- It is a pure correctness fix with no dependency on larger refactoring (M24–M27).
- Fixing it first makes later performance work easier to verify — the log will no longer hide real errors behind doctor noise.

---

## Challenged Assumptions

| Assumption | Challenge | Decision |
|---|---|---|
| Non-zero exit from `brew doctor` means failure. | Homebrew uses exit=1 to signal warnings; stdout still contains the warnings we want to show. | Treat exit=1 as warnings; only propagate as error on exit>1 or if stdout is unreadable. |

---

## Out of Scope

- **Reducing how often doctor runs** — covered by **M27** tiered refresh strategy.
- **Changing doctor parsing format** — keep existing `parseDoctorWarnings`; only exit-code handling changes.
- **Other diagnostics commands** — unless audit in 25.2 reveals a similar pattern.

---

## Architecture Decisions (ADRs)

| ID | Decision | Alternatives rejected | Rationale |
|---|---|---|---|
| D25-1 | `Doctor()` accepts exit code 1 and parses stdout for warnings. | Treat all non-zero exits as errors; skip stderr parsing. | Matches Homebrew CLI contract and user expectation that warnings are visible, not failures. |

Copy stable ADRs to `DESIGN.md` decision log when merged.

---

## Phases

| Phase | Steps | Theme | Phase gate |
|---|---|---|---|
| **A — Fix** | 25.1 | Correct doctor exit-code handling | Unit tests prove warnings are parsed despite exit=1 |
| **B — Audit** | 25.2 | Verify no other command has the same pattern | Test/log review signed off |

---

## Step Index

| Step | Title | Size | Depends | Deliverable |
|---|---|---|---|---|
| 25.1 | Accept exit=1 from `brew doctor` | S | — | `Doctor()` returns warnings on exit=1 |
| 25.2 | Audit non-zero exit semantics | S | 25.1 | Document or fix other commands |

---

## Steps

### 25.1 — Accept Exit=1 from `brew doctor`

**Size:** S  
**Phase:** A  
**Depends on:** —  
**Blocks:** 25.2

**Context:** `internal/brew/doctor.go:52-55` returns `nil, err` on any `Execute` error. Because `brew doctor` exits 1 when warnings exist, the Status dashboard and Doctor tab never see the warnings; instead they get an error path.

**Implementation checklist:**
1. In `Doctor()`, inspect the error: if it is a `BrewExitError` with `ExitCode == 1`, continue to parse stdout.
2. If exit code > 1 or the error is not an exit error, return the error.
3. Ensure `parseDoctorWarnings` handles the case where "Your system is ready to brew" is absent.
4. Cache the parsed warnings as before.

**Files:**

| File | Action |
|---|---|
| `internal/brew/doctor.go` | Modify `Doctor()` exit-code handling |
| `internal/brew/doctor_test.go` | Add test for exit=1 with warnings |

**Acceptance criteria:**
- [ ] Mock runner returning exit=1 with warning text → `Doctor()` returns warnings, no error.
- [ ] Mock runner returning exit=2 → `Doctor()` returns error.
- [ ] Real `brew doctor` with warnings no longer logs `level=WARN brew failed`.

**Tests (same change set):**
- [ ] `TestDoctorExitOneReturnsWarnings` — exit=1 yields parsed warnings.
- [ ] `TestDoctorExitTwoReturnsError` — exit=2 yields error.

**Risks & mitigations:**

| Risk | Mitigation |
|---|---|
| Older Homebrew versions return different exit codes. | Target is 6.0.0+; document assumption. |

**Rollback:** Revert `doctor.go` to return error on any non-zero exit.

---

### 25.2 — Audit Non-Zero Exit Semantics

**Size:** S  
**Phase:** B  
**Depends on:** 25.1  
**Blocks:** —

**Context:** `brew missing` already accepts exit=1 in `parseMissingOutput`. We should confirm no other reader silently misclassifies warnings as errors.

**Implementation checklist:**
1. Review all `runner.Execute` callers in `internal/brew/` for exit-code assumptions.
2. Document any command that legitimately uses non-zero exit for non-fatal conditions.
3. If another bug is found, either fix in this step or route to a new milestone/backlog item.

**Files:**

| File | Action |
|---|---|
| `internal/brew/*.go` | Review only |
| `DESIGN.md` or milestone ADR | Document non-zero-exit contracts |

**Acceptance criteria:**
- [ ] Audit list recorded in milestone file.
- [ ] No other recurring `level=WARN` in debug log from expected non-zero exits.

**Tests (same change set):**
- [ ] None unless a new bug is fixed.

**Risks & mitigations:**

| Risk | Mitigation |
|---|---|
| Audit expands scope. | Strictly review; fixes go to backlog unless trivial and safe. |

**Rollback:** N/A.

---

## Test Plan (milestone-level)

| Test | Tier | Step | Proves |
|---|---|---|---|
| `TestDoctorExitOneReturnsWarnings` | unit | 25.1 | Warnings visible despite exit=1 |
| `TestDoctorExitTwoReturnsError` | unit | 25.1 | Real errors still propagate |

**Verification commands:**

```bash
make test
go test -race ./...
```

---

## Definition of Done

- [ ] All steps 25.1–25.2 complete; acceptance criteria checked
- [ ] Every Test Plan row has an existing passing test
- [ ] Verification commands pass
- [ ] `DESIGN.md` updated if ADR is stable
- [ ] [status.md](../status.md) updated; **this file header Status matches**
- [ ] No open **critical/high** findings in this milestone's scope
- [ ] **Remaining** section empty or removed (if marking ✅ Complete)

---

## Post-Milestone Gate

Before starting **Enables** milestones, confirm:

- [ ] Header **Gate criteria** satisfied
- [ ] [review-template.md](../review-template.md) Dimension 8 (plan hygiene) for this milestone

---

## Rollback Plan

If the change causes regressions:

1. Revert `internal/brew/doctor.go`.
2. Keep tests; mark new tests as skipped with a note.

---

## Version History

| Date | Change |
|---|---|
| 2026-06-15 | Created from [templates/milestone.md](../templates/milestone.md) |
