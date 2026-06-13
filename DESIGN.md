# Lazybrew Design

## Overview

lazybrew is a terminal UI for managing Homebrew packages. It wraps `brew` CLI commands, caches outputs, and presents them in a Bubble Tea TUI with panels for formulae, casks, taps, services, search, and diagnostics.

**Stack:** Go + [Bubble Tea](https://github.com/charmbracelet/bubbletea) + [Lip Gloss](https://github.com/charmbracelet/lipgloss) + [Bubbles](https://github.com/charmbracelet/bubbles)  
**Platforms:** macOS + Linux  
**Target Homebrew:** 6.0.0+

## Goals

- Provide a keyboard-driven TUI for everyday Homebrew operations
- Cache brew command outputs with TTL to avoid redundant calls
- Serialize all write operations through a single TaskManager
- Keep the UI responsive during long-running brew operations
- Testable: unit tests for pure logic, teatest for E2E flows

## Non-Goals

- Replace Homebrew — all mutations delegate to `brew` CLI
- Visual screenshot testing
- Performance benchmarks

## System Architecture

```
cmd/lazybrew/main.go
       |
       v
   app.New()
       |
       v
  gui.Model  ←─────  brew.Client
       |                     |
       |            ┌────────┼────────┐
       |        Reader    Writer   Diagnostics
       |        (List,    (Install, (Doctor,
       |         Get,      Uninstall, Config,
       |         Search)   Pin, ...)  Update, ...)
       |
  tea.NewProgram()
       |
  ┌────┴───────────────────────┐
  │  Update(msg) → tea.Cmd     │
  │  View() → string           │
  │  Init() → tea.Cmd          │
  └────────────────────────────┘
          │
     ┌────┴─────────┬──────────────┐
     v              v              v
  gui/panel     gui/modal     gui/task
  (panelData,   (confirm,     (Manager,
   renderList)   input,        Task,
                 progress,     messages)
                 toast)
```

### Layer Rules

- `cmd/` depends on `internal/app/` and `internal/gui/`
- `internal/gui/` depends on `internal/brew/`, `internal/config/`, and its own sub-packages (`modal/`, `presentation/`, `style/`, `task/`)
- `internal/brew/` depends on nothing internal — pure domain logic + runner abstraction
- No circular dependencies; `internal/brew/` never imports `internal/gui/`

## Concurrency Architecture (ADR)

| ID | Decision | Alternatives Rejected |
|---|---|---|
| D19-1 | Package `internal/gui/task/` | Top-level package — too scattered |
| D19-2 | **Zero `program.Send` in handlers** | Keep pump — violates Bubble Tea purity |
| D19-3 | Reads outside TaskManager | Queue reads — unnecessary blocking |
| D19-4 | Progress modal on Model, not Manager | Manager owns UI — coupling |
| D19-5 | Queue max 10 tasks | Unbounded — memory risk |

### Rules

1. **All write operations** (install, uninstall, reinstall, upgrade, pin, tap, untap, trust, untrust, service start/stop, repair, cleanup, update) go through `task.Manager`
2. **Read operations** (list, get, search, doctor, config, status) use direct `tea.Cmd` — never queued
3. **No raw goroutines** — every concurrent operation returns a `tea.Cmd` that Bubble Tea schedules
4. **No `program.Send`** — all state updates happen through `Update()` message handlers; `program.Send` is forbidden in production code
5. **TaskManager serializes** — one task runs at a time; queue is FIFO with max 10

## Configuration Schema

See [config ADR in M18.9](master-plan/milestones/18-documentation-and-project-hygiene.md).

| Field | Status | Notes |
|---|---|---|
| `gui.theme` | Wired | `dark` or `light` |
| `gui.sidebar_width` | Wired | Integer |
| `gui.show_icons` | Deferred P2 | No-op until M17 |
| `gui.mouse` | Wired | `tea.WithMouseCellMotion` |
| `gui.auto_refresh_seconds` | Planned | Tick in M20.8 |
| `brew.path` | Planned | Custom runner path in M20.8 |
| `brew.update_on_start` | Wired | Non-blocking update on launch |

## Testing Tiers

| Tier | What | When | Tool |
|---|---|---|---|
| T0 | Safety net: TypedCache, cache race, regression stubs | During M19 | `go test` |
| T1 | Integration: brew CLI round-trip | M19.5+ | `go test -tags=integration` |
| T2 | E2E: teatest flows with View assertions | M20 phase A+ | `teatest` |
| T3 | Gates: coverage floors per package | Pre-release | `make cover-check` |

## Decision Log

| Date | Decision | Context |
|---|---|---|
| 2026-06-13 | TaskManager in `internal/gui/task/` | D19-1 |
| 2026-06-13 | Zero `program.Send` in handlers | D19-2 |
| 2026-06-13 | Tab cache key includes item name | D20-1 |
| 2026-06-13 | Viewport for tab content scroll | D20-4 |
| 2026-06-13 | teatest for E2E; unit for pure logic | D21-1 |
| 2026-06-13 | Integration tag; never in default CI | D21-2 |
| 2026-06-13 | TypedCache safe get before concurrency | Moved from M21 to M19.0 |
| 2026-06-13 | M17 deferred until M19–M22 done | Visual polish after correctness |
