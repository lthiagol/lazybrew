# Lazybrew — Master Plan Status

> **Project:** lazybrew — A TUI for managing Homebrew  
> **Stack:** Go + Bubble Tea + Lip Gloss + Bubbles  
> **Platforms:** macOS + Linux  
> **Created:** 2026-06-11  
> **Last Updated:** 2026-06-12  
> **Target Homebrew:** 6.0.0+

---

## Overall Progress

```
[X] Milestone 1  — Project Foundation & Core Types
[X] Milestone 2  — TUI Shell & Layout                        (small terminal not done)
[X] Milestone 3  — Brew Data Layer
[X] Milestone 4  — Read-Only Panels                          (empty states generic)
[X] Milestone 5  — Modals, Input & Search
[X] Milestone 6  — Package Mutations
[X] Milestone 7  — Taps & Trust Management
[X] Milestone 8  — P1 Features
[X] Milestone 9  — Polish, Config & Release
[X] Milestone 10 — GUI Architecture Payoff
[X] Milestone 11 — Feature Completion
[X] Milestone 12 — Test Infrastructure & QA
[X] Milestone 13 — Critical Bug Fixes
[X] Milestone 14 — Wire Dead Code & Fix Broken Functionality
[X] Milestone 15 — Architecture Cleanup
[X] Milestone 16 — Test Coverage Completion                  (186 tests, all 8 packages covered)
```

**Current Phase:** v0.2.0 — Done  
**Current Milestone:** —  
**Blockers:** None

---

## Milestone Index

| # | Milestone | Status | Description |
|---|---|---|---|---|---|
| 1 | [Foundation](milestones/01-foundation.md) | ✅ Complete | Go module, directory structure, domain types, brew runner, test infra |
| 2 | [TUI Shell](milestones/02-tui-shell.md) | ✅ Done | Bubble Tea app, layout, panel navigation, bottom bar, **missing: small terminal detection** |
| 3 | [Brew Data Layer](milestones/03-brew-data-layer.md) | ✅ Complete | JSON parsing, cache, R/W service splits (Formulae/Casks/Diagnostics), search JSON |
| 4 | [Read-Only Panels](milestones/04-read-only-panels.md) | ✅ Done | Formulae, Casks, Outdated, Status, Taps panels with real data, **empty states are generic** |
| 5 | [Modals & Search](milestones/05-modals-and-search.md) | ✅ Complete | Confirmation, text input, menu, progress modals + search flow + toast system |
| 6 | [Package Mutations](milestones/06-package-mutations.md) | ⚠️ Partial | Install/upgrade work. **Missing: TaskManager (no queuing/retry), uninstall lacks confirmation modal, no reinstall flow** |
| 7 | [Taps & Trust](milestones/07-taps-and-trust.md) | ⚠️ Partial | Tap add works. **Missing: untap keybinding + official protection, trust formula/cask granularity from GUI, repair keybinding** |
| 8 | [P1 Features](milestones/08-p1-features.md) | ⚠️ Partial | Pin/unpin works (formulae+casks), services start/stop, brew missing. **Missing: cleanup keybinding, doctor keybinding, leaves toggle, autoremove keybinding; vulns is a broken stub (calls Doctor instead of brew vulns)** |
| 9 | [Polish & Release](milestones/09-polish-and-release.md) | ✅ Complete | Config system (YAML), theming (dark/light), help overlay, CLI flags, goreleaser |
| 10 | [GUI Architecture](milestones/10-gui-architecture.md) | ✅ Complete | Tab content switching, progress streaming, config consumption, Esc closes help, gui.go decomposed |
| 11 | [Feature Completion](milestones/11-feature-completion.md) | ⚠️ Partial | Used-by tab works, missing wired. **Broken: services run (`f`) action silently ignored, serviceCleanup modal only (no execute), Brewfile menu is a stub (calls Doctor), vulns calls wrong command** |
| 12 | [Test Infrastructure](milestones/12-test-infrastructure.md) | ⚠️ Partial | Modal tests (14), fuzz tests (2), snapshots (5). **Missing: E2E tests (0), integration tests (0), gui/ package has zero tests** |
| 13 | [Critical Bug Fixes](milestones/13-critical-bug-fixes.md) | ✅ Complete | Cache RLock→Lock upgrade, separate KeyOutdated keys, ConfirmModal defaults No, padRight UTF-8 |
| 14 | [Wire Dead Code](milestones/14-wire-dead-code.md) | ⚠️ Partial | Brewfile menu wiring, type assertions, runVulns/runMissing output, ModalDoneMsg removed, batchState.mu removed. **Missing: serviceCleanup doesn't execute, Formulae Files tab placeholder, Cask Caveats hardcoded, Doctor tab shows wrong data** |
| 15 | [Architecture Cleanup](milestones/15-architecture-cleanup.md) | ⚠️ In Progress | 5/10 done: itoa→strconv, jsonUnmarshal removed, atomic Logger, unexported Program, NewMockClient removed |
| 16 | [Test Coverage](milestones/16-test-coverage.md) | ⚠️ In Progress | Config/style/logger tests done. Missing: app_test.go, gui_test.go, E2E, integration; 120/147 tests |

---

## Testing Strategy (Cross-Cutting)

| Test Type | Framework | What It Covers | When Written |
|---|---|---|---|
| **Unit** | Go `testing` + `testify` | Types, parsers, formatters, cache, config | Every milestone |
| **Snapshot** | `go-snaps` or custom | Presentation output (formatted panel strings) | Milestones 4+ |
| **TUI / E2E** | `teatest` (Bubble Tea) | Full app interaction flows (key presses → UI state) | Milestones 2+ |
| **Integration** | Go `testing` + build tag | Actual `brew` CLI calls (requires brew installed) | Milestone 3+ |
| **Fuzz** | Go `testing.F` | JSON parsing edge cases | Milestone 3 |

> Integration tests that call real `brew` commands use a build tag (`//go:build integration`) so they don't run in CI without brew installed.

---

## Architecture Reference

See the design document for the full architecture, feature inventory, and UI layout specifications.

---

## Decision Log

| Date | Decision | Context |
|---|---|---|
| 2026-06-11 | Go + Bubble Tea | Modern API, Elm architecture, Lip Gloss styling |
| 2026-06-11 | macOS + Linux | Support both Homebrew and Linuxbrew |
| 2026-06-11 | Native `brew trust`/`brew untrust` only | Latest brew only, no backward compat |
| 2026-06-11 | Separate Formulae & Casks panels | Consistent with lazygit/lazydocker pattern |
| 2026-06-11 | Name: lazybrew | Matches lazy* convention |
| 2026-06-11 | P0 feature set for MVP | Confirmed as-is from design doc |
| 2026-06-11 | Target Homebrew 6.0.0+ | Reviewed brew 6.0.0 changelog; plan is compatible. Trust is now mandatory (was opt-in in 5.x), validating our M7 plans. |
| 2026-06-11 | Pin casks too (was only formulae) | Homebrew 6.0.0 added `brew pin` for casks; extended M8 to cover both |
| 2026-06-11 | Run brew non-interactively | 6.0.0 made "ask mode" default — brew may prompt for confirmation if stdin is a TTY. M3 runner will detect non-TTY or set `HOMEBREW_NO_ASK` env var |
| 2026-06-11 | No BrewUI panic | Homebrew announced BrewUI (official GUI). lazybrew is a TUI — complementary, not competing directly. Still worth mentioning in docs |
| 2026-06-11 | Plan review: 34 issues fixed | Comprehensive review of all milestones. Key fixes: Runner stdout/stderr split, HOMEBREW_NO_ASK from day one, typed errors, split read/write service interfaces, generics cache, JSON search, graceful shutdown, concurrent read safety |
| 2026-06-11 | Read/write service split | FormulaeService split into FormulaeReader + FormulaeWriter. Same for Casks and Diagnostics. Reads are cacheable/concurrent; writes go through task manager |
| 2026-06-11 | Synthetic test fixtures | Don't capture real brew output as fixtures (rots between versions). Use minimal synthetic JSON covering edge cases |
| 2026-06-11 | No hardcoded bottle tags | Bottled detection must not hardcode macOS version tags (arm64_sonoma etc.). Use generic platform-matching logic |
| 2026-06-11 | Services: `f` for run, not `R` | Changed "Run service" keybinding from `R` to `f` to avoid case-sensitivity confusion with `r` (restart) |
| 2026-06-11 | Modal results are typed | No `interface{}` results. Each modal type returns typed results via ModalResult struct |
| 2026-06-11 | Progress cancel via context | Esc cancels running brew via context.CancelFunc -> SIGINT -> 5s -> SIGKILL |
| 2026-06-12 | YAML config system | `~/.config/lazybrew/config.yml` with defaults for theme, sidebar width, mouse, keybindings |
| 2026-06-12 | Help overlay `?` | Full-screen keyboard reference organized by panel, toggle with `?`/`Esc` |
| 2026-06-12 | CLI flags with `flag` stdlib | `--version`, `--config`, `--debug` parsed via standard library |
| 2026-06-12 | `.goreleaser.yml` | Cross-platform builds for darwin/linux, amd64/arm64, tar.gz with checksums |
| 2026-06-12 | All 9 milestones complete | v0.1.0-dev — 5,650 lines across 42 Go files, 71 tests, 65%+ coverage |
| 2026-06-12 | Code audit: 4 critical bugs found | Cache RLock race (will crash), KeyOutdated collision, ConfirmModal defaults to Yes, padRight splits UTF-8 |
| 2026-06-12 | Code audit: 5 broken features found | Brewfile menu unhandled, serviceCleanup skips confirm, vulns/missing discard output, hard type assertions, context cancel discarded |
| 2026-06-12 | Code audit: 10 design issues found | Dead batch code, unused mutex, inconsistent interfaces, interface{} overuse, duplicate itoa, mixed json patterns, global state races |
| 2026-06-12 | M13-M16 created | 4 new milestones to address audit findings: critical bugs, dead code, architecture cleanup, test coverage |
| 2026-06-12 | Second code audit: 120 tests, not 147 | All 16 milestones reviewed against actual code. 10 fully complete, 6 partially complete. Gaps documented in each milestone file.

