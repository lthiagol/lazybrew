# Lazybrew — Agent & Contributor Guide

> **Purpose:** Operational conventions for humans and coding agents working on lazybrew.  
> **Audience:** Anyone opening a PR, running an agent session, or doing a code review.  
> **Architectural context:** See [DESIGN.md](DESIGN.md).  
> **Planning context:** See [master-plan/status.md](master-plan/status.md).

---

## Quick Start

```bash
git clone https://github.com/thiago/lazybrew
cd lazybrew
make build      # produces bin/lazybrew
make test       # unit tests + race detector
make run        # launch TUI with default config
```

Run integration tests (requires real Homebrew):

```bash
make test-integration
```

---

## Repository Map

| Path | What lives here | Test command |
|---|---|---|
| `cmd/lazybrew/` | `main.go`, CLI flags | — |
| `internal/app/` | Bootstrap: config, theme, runner, `gui.New` | `go test ./internal/app/...` |
| `internal/brew/` | Domain models, `Runner`, services, cache, parsers | `go test ./internal/brew/...` |
| `internal/config/` | YAML config loading/validation | `go test ./internal/config/...` |
| `internal/gui/` | Bubble Tea model, panels, rendering, keybindings | `go test ./internal/gui/...` |
| `internal/gui/modal/` | Reusable modals (confirm, input, progress, toast, menu) | `go test ./internal/gui/modal/...` |
| `internal/gui/presentation/` | Formatters and snapshot tests | `go test ./internal/gui/presentation/...` |
| `internal/gui/style/` | Lip Gloss styles and themes | `go test ./internal/gui/style/...` |
| `internal/gui/task/` | TaskManager for serializing write operations | `go test ./internal/gui/task/...` |
| `internal/gui/flows/` | `teatest` E2E flows | `go test ./internal/gui/flows/...` |
| `internal/gui/testutil/` | Test helper for E2E flows | — |
| `master-plan/` | Milestones, status, decision logs | — |
| `scripts/` | Build/test helpers (e.g. coverage floor check) | — |

---

## Change Guidelines

1. **Scope:** One milestone/step per change set. Avoid drive-by refactors.
2. **Tests:** Add or update tests in the same commit/PR as the code change.
3. **Style:** Run `make fmt` before committing.
4. **Verification:** Run `make test` and `go vet ./...` before considering work done.
5. **No secrets:** Never commit tokens, keys, or personal config files.
6. **No large binaries:** `bin/lazybrew` is gitignored; do not force-add it.

---

## Bubble Tea Rules

All GUI state mutation must go through the Bubble Tea `Update(msg tea.Msg)` loop.

1. **Zero `program.Send` in production code.** If you find `program.Send`, replace it with a `tea.Cmd` that returns a message.
2. **All write operations through `task.Manager`.** Install, uninstall, reinstall, upgrade, pin, tap, untap, trust, service start/stop, repair, cleanup, autoremove, and `brew update` must be enqueued via `m.tasks.Enqueue`.
3. **No raw goroutines in handlers.** Use `tea.Cmd` closures or the TaskManager streaming pattern.
4. **Progress modal owns streaming output.** `TaskOutputMsg` is appended to the active `ProgressModal`; `TaskCompletedMsg` closes it and shows a toast.
5. **Model fields are the source of truth.** Do not read from `m.program` to infer state.

---

## Testing Rules

| Tier | Tool | When to use | Example |
|---|---|---|---|
| Unit | `testing` | Pure logic, parsers, cache, TaskManager | `internal/brew`, `internal/gui/task` |
| Snapshot | `cmp` / string compare | Stable presentation output | `internal/gui/presentation` |
| Integration | `//go:build integration` | Real `brew` CLI round-trip | `internal/brew/runner_integration_test.go` |
| E2E | `teatest` | Full user flows with `View()` assertions | `internal/gui/flows` |

Rules:

- Unit tests must pass with `-race` (`make test`).
- E2E flows must assert on `View()` output, not just model fields.
- Coverage floors are enforced by `make cover-check`. Do not lower them without a decision log entry.
- New brew command support requires: service method, cache invalidation, unit test, and a row in `master-plan/coverage-audit.md`.

---

## Planning Rules

1. **Source of truth:** [master-plan/status.md](master-plan/status.md).
2. **New milestone:** Use [templates/milestone.md](master-plan/templates/milestone.md).
3. **Step size:** Every step must have size, dependencies, acceptance criteria, and tests.
4. **Out of scope:** If a finding is not in the current milestone, add it to [master-plan/backlog.md](master-plan/backlog.md) unless it is critical.
5. **Status updates:** Mark steps complete in the milestone file **and** `status.md`.
6. **Decision log:** Non-obvious architectural choices go in `DESIGN.md` decision log and the milestone ADR table.

---

## Git Rules

1. Do not commit unless explicitly asked.
2. Do not push, force-push, rebase, or amend commits unless explicitly asked.
3. Keep commits focused on a single milestone/step.
4. Commit messages should match repo style (concise, imperative).

---

## Verification Commands

Run these before declaring any milestone step done:

```bash
go test -race ./...
go vet ./...
make lint
make cover-check
```

For release readiness:

```bash
go test -race ./...
make cover-check
make test-integration   # requires Homebrew
./bin/lazybrew --version
goreleaser release --snapshot --clean
```

---

## Release Readiness Checklist (Agent View)

Before v0.2.0 tag:

- [x] M18.8 AGENTS.md done (this file)
- [ ] M17.3 update summary toast done
- [ ] M21.2 ≥8 teatest flows done
- [ ] M22.1b CI green on push/PR
- [x] M22.2 integration workflow file exists
- [ ] M22.3 goreleaser snapshot succeeds
- [ ] M22.4 release checklist signed off

---

## References

- [DESIGN.md](DESIGN.md) — architecture and concurrency ADR
- [master-plan/status.md](master-plan/status.md) — current milestone status
- [master-plan/review-template.md](master-plan/review-template.md) — future audit template
- [master-plan/backlog.md](master-plan/backlog.md) — deferred work
