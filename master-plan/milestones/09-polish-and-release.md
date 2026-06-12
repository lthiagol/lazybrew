# Milestone 9 — Polish, Config & Release

> **Status:** 🔲 Not Started  
> **Depends on:** Milestone 8 (P1 Features)  
> **Enables:** v1.0 Release

---

## Goal

Final polish pass: user configuration system, theming, keyboard help overlay, mouse support, comprehensive error handling, documentation, and release pipeline. After this milestone, lazybrew is ready for public release via `brew install lazybrew`.

---

## Steps

### 9.1 — Configuration System

**What:** YAML-based user config with sensible defaults.

**Files:** `internal/config/config.go`, `internal/config/defaults.go`

**Config file location:** `~/.config/lazybrew/config.yml`

**Implementation:**
- Load config on startup: check `~/.config/lazybrew/config.yml`
- If not found, use defaults (don't create the file — only create when user explicitly configures)
- Support `LAZYBREW_CONFIG` env var to override config path
- Config struct with YAML tags
- Hot-reload: watch config file for changes (optional — nice but not essential for v1)

**Config structure:**
```yaml
gui:
  theme: "dark"               # dark | light
  sidebar_width: 30           # percentage of terminal width
  show_icons: true            # unicode icons in panel titles
  mouse: true                 # mouse support
  auto_refresh_seconds: 60    # 0 to disable

brew:
  path: ""                    # auto-detect if empty
  update_on_start: false      # run brew update on launch

keybindings:
  universal:
    quit: "q"
    search: "/"
    help: "?"
    refresh: "R"
  formulae:
    install: "i"
    uninstall: "x"
    upgrade: "u"
    reinstall: "r"
    pin: "p"
    leaves: "L"
  casks:
    uninstall: "x"
    zap: "X"
    upgrade: "u"
    reinstall: "r"
    pin: "p"
  taps:
    add: "a"
    remove: "x"
    trust: "t"
    repair: "r"
  services:
    start: "s"
    stop: "S"
    restart: "r"
    run: "f"
    cleanup: "c"
  status:
    update: "u"
    upgrade_all: "U"
    doctor: "d"
    cleanup: "c"
    autoremove: "A"
    bundle: "B"
    vulns: "v"
    missing: "m"
```

**Acceptance criteria:**
- [ ] Config loads from YAML
- [ ] Defaults used when config file absent
- [ ] Invalid config shows helpful error (not crash)
- [ ] Keybinding overrides apply correctly
- [ ] `LAZYBREW_CONFIG` env var works
- [ ] Config validation (e.g., sidebar_width between 15–50)

---

### 9.2 — Theming

**What:** Dark and light themes with consistent color palettes.

**File:** `internal/gui/style/theme.go`

**Dark theme (default):**
```
Background:  #1a1b26 (tokyo night inspired)
Foreground:  #c0caf5
Accent:      #7aa2f7 (blue)
Secondary:   #bb9af7 (purple)
Success:     #9ece6a (green)
Warning:     #e0af68 (amber)
Error:       #f7768e (red)
Subtle:      #565f89 (gray)
Border:      #3b4261
ActiveBdr:   #7aa2f7 (accent)
```

**Light theme:**
```
Background:  #f5f5f5
Foreground:  #343b58
Accent:      #34548a
Secondary:   #5a4a78
Success:     #33635c
Warning:     #8f5e15
Error:       #8c4351
Subtle:      #9699a3
Border:      #c0c0c0
ActiveBdr:   #34548a
```

**Implementation:**
- Theme struct holds all color values
- Theme loaded from config (`gui.theme`)
- All Lip Gloss styles derived from the active theme
- Theme applies globally on startup (no hot-swap needed for v1)

**Acceptance criteria:**
- [ ] Dark theme looks polished
- [ ] Light theme looks polished
- [ ] Config switch works
- [ ] All panels use theme colors (no hardcoded colors)
- [ ] Contrast is good in both themes
- [ ] Works on 256-color terminals (fallback colors)

---

### 9.3 — Keyboard Help Overlay

**What:** Full-screen keybinding reference triggered by `?`.

**File:** `internal/gui/help.go`

**Visual:**
```
┌── Keyboard Shortcuts ─────────────────────────────────┐
│                                                       │
│  Global                          Formulae             │
│  ──────                          ────────             │
│  q      Quit                     i      Install       │
│  Tab    Next panel               x      Uninstall     │
│  S-Tab  Previous panel           r      Reinstall     │
│  1-7    Jump to panel            u      Upgrade       │
│  /      Search                   U      Upgrade all   │
│  ?      This help                p      Pin/Unpin     │
│  R      Refresh                  L      Toggle leaves │
│  [ ]    Switch tabs              o      Open homepage │
│  Ctrl+B Toggle sidebar           y      Copy name     │
│                                                       │
│  Casks                           Taps                 │
│  ─────                           ────                 │
│  i      Install                  a      Add tap       │
│  x      Uninstall                x      Remove tap    │
│  X      Zap uninstall            t      Trust config  │
│  u      Upgrade                  r      Repair tap    │
│  p      Pin/Unpin                o      Open in browser│
│  r      Reinstall                                       │
│                                                       │
│  Services                        Status               │
│  ────────                        ──────               │
│  s      Start                    u      Update brew   │
│  S      Stop                     U      Upgrade all   │
│  r      Restart                  d      Doctor        │
│  f      Run (foreground)         c      Cleanup       │
│  c      Cleanup stale files      A      Autoremove    │
│                                  B      Brewfile      │
│                                  v      Vulns check   │
│                                  m      Missing deps  │
│                                                       │
│                        Press ? or Esc to close         │
└───────────────────────────────────────────────────────┘
```

**Implementation:**
- Rendered as a full-screen overlay (like modals but larger)
- Keybindings populated dynamically from the actual registered keybindings
- Organized by panel/context
- Reflects any user customizations
- `?` or `Esc` closes the overlay

**Acceptance criteria:**
- [ ] `?` opens the help overlay from any panel
- [ ] All keybindings listed, organized by context
- [ ] Reflects user-customized keybindings
- [ ] `?` or `Esc` closes
- [ ] Scrollable if content exceeds terminal height

---

### 9.4 — Mouse Support

**What:** Basic mouse support for panel and item selection.

**Implementation:**
- Click on a sidebar panel → switch to that panel
- Click on a list item → select that item
- Click on a main panel tab → switch to that tab
- Scroll wheel in lists → scroll the list
- Scroll wheel in main panel → scroll the viewport
- Mouse support enabled by default, configurable via `gui.mouse`

**Acceptance criteria:**
- [ ] Click to switch panels
- [ ] Click to select items
- [ ] Click to switch tabs
- [ ] Scroll wheel works in lists and viewports
- [ ] Can be disabled via config

---

### 9.5 — Error Handling & Graceful Shutdown

**What:** Comprehensive error handling for all edge cases, plus graceful shutdown on Ctrl+C or `q`.

**Graceful shutdown:**
- On `q` key: cancel any running task (SIGINT → 5s → SIGKILL), clear cache, exit cleanly
- On Ctrl+C (SIGINT): same as `q` — cancel running task, clean up, exit
- On SIGTERM: same behavior
- If a brew subprocess is running when shutdown is triggered:
  1. Send SIGINT to the subprocess
  2. Wait up to 5 seconds for it to exit
  3. If still running, send SIGKILL
  4. Log the forced kill at WARN level
- The task manager's `context.CancelFunc` is called, which propagates to `exec.CommandContext`

**Edge cases to handle:**
| Scenario | Behavior |
|---|---|
| brew not installed | Full-screen error: "Homebrew not found. Install from https://brew.sh" |
| brew command fails | Error toast with stderr summary, full error in main panel |
| Network error during install | Progress modal shows error, suggests retry |
| Permission denied | Error modal with suggestion to fix permissions |
| Terminal too small | Warning bar: "Terminal too small (80x24 required)" |
| No packages installed | Friendly empty states per panel |
| JSON parse error | Log error, show "Failed to load data" with retry option |
| Interrupted operation | Task manager handles SIGINT gracefully |

**Implementation:**
- Centralized error handler in root model
- Error types: `BrewError`, `ParseError`, `PermissionError`, `NetworkError`
- Each error type has a user-friendly message template
- All errors logged to `~/.config/lazybrew/debug.log` (when `--debug` flag used)

**Acceptance criteria:**
- [ ] Every error shows a user-friendly message (not raw stderr)
- [ ] Debug log captures full details
- [ ] App never panics on brew errors
- [ ] All empty states have friendly messages
- [ ] Retry option where applicable

---

### 9.6 — CLI Flags & Version Info

**What:** Command-line flags for the lazybrew binary.

**File:** Update `cmd/lazybrew/main.go`

**Flags:**
```
lazybrew [flags]

Flags:
  -v, --version     Print version and exit
  -c, --config      Path to config file
  -d, --debug       Enable debug logging
  -h, --help        Show help
```

**Version info:**
- Embed version at build time via `ldflags`
- Show in Status panel: "lazybrew v1.0.0 (built 2026-06-11)"

**Acceptance criteria:**
- [ ] `lazybrew --version` prints version
- [ ] `lazybrew --config /path/to/config.yml` uses custom config
- [ ] `lazybrew --debug` enables debug logging
- [ ] Version shown in Status panel

---

### 9.7 — Release Pipeline

**What:** Automated build and release using GoReleaser.

**Files:** `.goreleaser.yml`, `.github/workflows/release.yml`

**GoReleaser config:**
- Build for: `darwin/arm64` (priority), `darwin/amd64`, `linux/amd64`, `linux/arm64`
- Produce: `.tar.gz` archives, checksums
- Homebrew tap: auto-publish to a `homebrew-lazybrew` tap repo
- Release to GitHub Releases

**Platform note (Homebrew 6.0.0):**
- macOS 27 (Golden Gate) drops Intel support; Homebrew Intel `x86_64` moves to Tier 3 (Sept 2026) and unsupported (Sept 2027)
- **`darwin/arm64` (Apple Silicon) is the primary target** for macOS releases
- `darwin/amd64` is retained but secondary; Intel macOS users should be encouraged to migrate

**GitHub Actions:**
- **CI workflow** (on PR/push):
  - `make lint`
  - `make test`
  - `make build`
- **Release workflow** (on tag push):
  - GoReleaser → build → GitHub Release → Homebrew tap update

**Homebrew formula:**
```ruby
class Lazybrew < Formula
  desc "A TUI for managing Homebrew"
  homepage "https://github.com/<user>/lazybrew"
  # ... auto-generated by goreleaser
end
```

**Acceptance criteria:**
- [ ] GoReleaser builds for all target platforms
- [ ] Binaries work on macOS (Intel + ARM) and Linux
- [ ] GitHub Release created with changelog
- [ ] Homebrew tap updated automatically
- [ ] `brew install <user>/lazybrew/lazybrew` works

---

### 9.8 — Documentation

**What:** README, contributing guide, and screenshots.

**Files:**
- `README.md` — project overview, installation, screenshots, keybindings, config reference
- `CONTRIBUTING.md` — development setup, architecture overview, PR guidelines
- `docs/keybindings.md` — auto-generated full keybinding reference
- `docs/configuration.md` — full config reference with examples

**README sections:**
1. Hero banner / screenshot
2. Features list
3. Installation (`brew install`, `go install`, binary download)
4. Quick start
5. Keybindings overview
6. Configuration
7. Contributing
8. License

**Acceptance criteria:**
- [ ] README has screenshots/GIFs of the app
- [ ] Installation instructions for all methods
- [ ] Keybinding reference complete
- [ ] Config reference with all options documented
- [ ] Contributing guide covers dev setup

---

## Tests for This Milestone

| Test | Type | File | What It Validates |
|---|---|---|---|
| `TestConfigLoad` | Unit | `internal/config/config_test.go` | YAML parsing, defaults |
| `TestConfigDefaults` | Unit | `internal/config/config_test.go` | Missing file → defaults |
| `TestConfigInvalid` | Unit | `internal/config/config_test.go` | Bad YAML → helpful error |
| `TestConfigValidation` | Unit | `internal/config/config_test.go` | Out-of-range values caught |
| `TestKeybindingOverride` | E2E | `internal/gui/gui_test.go` | Custom keybinding applies |
| `TestThemeDark` | Snapshot | `internal/gui/style/theme_test.go` | Dark theme renders correctly |
| `TestThemeLight` | Snapshot | `internal/gui/style/theme_test.go` | Light theme renders correctly |
| `TestHelpOverlay` | E2E | `internal/gui/help_test.go` | `?` opens, lists all keys, closes |
| `TestBrewNotFound` | E2E | `internal/gui/gui_test.go` | Shows error when brew missing |
| `TestTerminalTooSmall` | E2E | `internal/gui/gui_test.go` | Warning on small terminal |
| `TestVersionFlag` | Unit | `cmd/lazybrew/main_test.go` | --version prints version |
| `TestDebugFlag` | Unit | `cmd/lazybrew/main_test.go` | --debug enables logging |
| `TestCrossPlatformBuild` | CI | `.github/workflows/ci.yml` | Builds on macOS + Linux |

---

## Definition of Done

- [ ] Config system loads and validates YAML config
- [ ] Dark and light themes polished
- [ ] Help overlay shows all keybindings
- [ ] Mouse support works
- [ ] All error edge cases handled gracefully
- [ ] CLI flags work (--version, --config, --debug)
- [ ] GoReleaser builds for all platforms
- [ ] `brew install` works from custom tap
- [ ] README with screenshots and full documentation
- [ ] All tests pass, CI green
- [ ] **lazybrew v1.0.0 ready for release** 🚀
