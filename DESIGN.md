# Lazybrew Design

## Overview

lazybrew is a terminal UI for managing Homebrew packages. It wraps `brew` CLI commands, caches outputs, and presents them in a Bubble Tea TUI with panels for formulae, casks, taps, services, search, and diagnostics.

**Stack:** Go + [Bubble Tea v2](https://github.com/charmbracelet/bubbletea) (`charm.land/bubbletea/v2`) + [Lip Gloss v2](https://github.com/charmbracelet/lipgloss) (`charm.land/lipgloss/v2`) + [Bubbles v2](https://github.com/charmbracelet/bubbles) (`charm.land/bubbles/v2`)  
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
  │  View() → tea.View         │
  │  Init() → tea.Cmd          │
  └────────────────────────────┘
          │
     ┌────┴─────────┬──────────────┐
     v              v              v
  gui/panel     gui/modal     gui/task
  (panelData,   (confirm,     (Manager,
   renderList)   input,        Task,
                 menu,         messages)
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
| D19-4 | Operation overlay on Model, not Manager (the progress modal was removed) | Manager owns UI — coupling |
| D19-5 | Queue max 10 tasks | Unbounded — memory risk |

### Rules

1. **All write operations** (install, uninstall, reinstall, upgrade, update, pin/unpin, tap/untap/repair, trust/untrust, services start/stop/restart/run/cleanup, cleanup, autoremove) go through `task.Manager`
2. **Read operations** (list, get, search, doctor, config, status) use direct `tea.Cmd` — never queued
3. **No raw goroutines** — every concurrent operation returns a `tea.Cmd` that Bubble Tea schedules
4. **No `program.Send`** — all state updates happen through `Update()` message handlers; `program.Send` is forbidden in production code
5. **TaskManager serializes** — one task runs at a time; queue is FIFO with max 10
6. **brew runs in its own session** (`Setsid`, Unix): a cancel kills the whole process group (brew and the curl/git/ruby it started), with a 2 s `WaitDelay`; brew has no controlling terminal, so a step that asks for a password (macOS `.pkg` casks via sudo) fails at once with a needs-sudo message instead of hanging
7. **Startup preload** — with `gui.preload` on, Status loads first, then Formulae → Outdated → Casks → Taps → Services as ordinary bounded reads, at most two in flight, driven by messages in `Update`
8. **The Operation view's lifecycle** — a successful run closes itself after 2 s (a tick that names its operation; never mid-batch); failed and cancelled runs stay until Esc or a panel switch

## Configuration Schema

See the config ADR topic (configuration schema decisions).

| Field | Status | Notes |
|---|---|---|
| `gui.theme` | Wired | `dark` or `light` |
| `gui.sidebar_width` | Wired | Integer |
| `gui.show_icons` | Wired | `false` draws no icons at all |
| `gui.icons` | Wired | `nerd` (default) or `unicode`; `LAZYBREW_ICONS` overrides |
| `gui.animations` | Wired | `true` (default) springs the motion; `false` draws every change at once; `LAZYBREW_ANIMATIONS` overrides |
| `gui.preload` | Wired | `true` (default) loads every panel in the background after Status, two reads at a time; `false` loads each panel on first visit; `LAZYBREW_PRELOAD` overrides |
| `gui.mouse` | Wired | `View().MouseMode = tea.MouseModeCellMotion` |
| `gui.auto_refresh_seconds` | Planned | Tick in M20.8 |
| `brew.path` | Planned | Custom runner path in M20.8 |
| `brew.update_on_start` | Wired | `true` (default): `brew update` at every launch; `auto`: update on launch only when brew's last update is older than `brew.update_stale_hours`; `false`: never |

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
| 2026-08-01 | Manual `stable-ci.yml` tap-update path is canonical; `.goreleaser.yml` removed | M07. Goreleaser was considered but rejected: the manual path (sed + tap PR) was just used successfully for v1.0.0-rc1, release cadence is low (rc1 then occasional bugfixes), and goreleaser-brews would have replaced the working flow with a more complex secret+formula integration. If release cadence grows, revisit. |
| 2026-10-31 | Detached the mp/master-plan planning workflow; planning now runs through the herdr lane flow in work/ (briefs, inbox, state row). master-plan/ is kept locally as a frozen archive and is no longer tracked by git. | Slice detach-master-plan |
| 2026-11-19 | Moved to the Charm v2 stack (charm.land bubbletea/lipgloss/bubbles v2) — prerequisite for overlay compositing and animation in the visual overhaul. | Slice v2-01-charm-v2 |
| 2026-11-20 | Visual overhaul closed (LB-08: v2-01 … v2-06, v2-05s) and user follow-ups v2-07 … v2-11: finished operations auto-close, startup preload, main-area layout (padding, tables, two-column Info), tap actions, process-group cancel, fast tests, CI with `-race`. | Rulings `Q-v2-*` in `work/questions.md` |
| 2026-11-20 | brew runs with `Setsid`; tty prompts (sudo, git credentials) fail fast with a needs-sudo message. Suspending the TUI for them (`tea.ExecProcess`) is parked as LB-11 — the user ruled the message is enough. | Slice v2-11, `Q-v2-11-b` |
| 2026-11-22 | Tap bumps are manual (`docs/agents/release.md`); `stable-ci.yml`'s `tap-update` job is a dead leftover (LB-41). Supersedes the 2026-08-01 row. | Docs rewrite review |
