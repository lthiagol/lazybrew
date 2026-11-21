# lazybrew

A TUI for managing Homebrew on macOS and Linux.

## Status

Early development — TUI functional, not daily-driver reliable.

## Installation

### Homebrew (recommended)

The tap is hosted on Codeberg, so add it once with the full URL:

```bash
brew tap lthiagol/homebrew-tap https://github.com/lthiagol/homebrew-tap
brew install lthiagol/homebrew-tap/lazybrew
```

Updates:

```bash
brew update && brew upgrade lthiagol/homebrew-tap/lazybrew
```

For the rolling development build (tracks the `wip` branch — unstable, use at your own risk):

```bash
brew install lthiagol/homebrew-tap/lazybrew-dev
```

### `go install`

```bash
go install github.com/lthiagol/lazybrew@latest
```

### Build from source

```bash
git clone https://github.com/lthiagol/lazybrew
cd lazybrew
make build
```

## Usage

```bash
lazybrew
```

Launch the interactive TUI. Available flags:

| Flag | Description |
|------|-------------|
| `--version` | Print version and exit |
| `--config PATH` | Path to config file (default `~/.config/lazybrew/config.yml`) |
| `--debug` | Enable debug logging |

### Keybindings

| Key | Action |
|-----|--------|
| `Tab` / `Shift+Tab` | Next / previous panel |
| `1`–`7` | Jump to panel |
| `j`/`k` | Scroll list |
| `[`/`]` | Switch tabs |
| `R` | Refresh all data |
| `/` | Search packages |
| `q` | Quit |
| `?` | Help |

Per-panel actions appear in the bottom bar and `?` help screen.

## Configuration

Config file at `~/.config/lazybrew/config.yml`:

```yaml
gui:
  theme: dark             # dark or light
  sidebar_width: 35       # percent of the terminal width, 15-50; default 35
  mouse: true
  auto_refresh_seconds: 60
  icons: nerd             # nerd (Nerd Font glyphs) or unicode (no Nerd Font needed)
  show_icons: true        # false draws no icons at all
  animations: true        # false draws every change at once, with no motion
  preload: true           # false loads each panel only when it is first opened
brew:
  path: ""                # auto-detect if empty
  update_on_start: auto   # auto, true or false; see below
  update_stale_hours: 6   # auto updates when brew's last update is older than this
```

`brew.update_on_start` says whether lazybrew runs `brew update` when it starts.
`auto` (the default) reads when brew last updated, and runs `brew update` only
if that was more than `update_stale_hours` ago. The header shows `updating…`
while it runs and `updated 2h ago` after. `true` runs `brew update` at every
launch. `false` never runs `brew update` on its own: lazybrew does not update
behind your back, and the outdated list is only as fresh as your last
`brew update`. The `R` key runs `brew update` with `auto` and `true`; with
`false` it only reloads.

`LAZYBREW_ICONS=unicode` (or `nerd`) overrides `gui.icons` for one run.

`LAZYBREW_ANIMATIONS=0` (or `1`) overrides `gui.animations` for one run. The
motion is short and never delays a key: a closing overlay disappears at once.

`LAZYBREW_PRELOAD=0` (or `1`) overrides `gui.preload` for one run. With the
preload on, every panel loads in the background once the Status dashboard has
loaded, two reads at a time, so the sidebar fills in without a visit; keys work
while it runs.

## Development

### Prerequisites

- Go 1.24+
- Homebrew (for integration tests and runtime)
- `git` — for version control
- `make` — to run build targets

Run `./check_dependencies.sh` to verify all dependencies are met.

### Commands

| Command | Description |
|---------|-------------|
| `make build` | Build binary to `bin/lazybrew` |
| `make run` | Build, then launch lazybrew |
| `make test` | All unit tests with the race detector (sets `CGO_ENABLED=1`) |
| `make test-fast` | Unit tests without `-race`; `PKG=./internal/gui/...` narrows it |
| `make ci` | Everything CI runs — gofmt check, `go vet`, `make test`, build — plus the coverage floors |
| `make test-integration` | Run integration tests (requires Homebrew) |
| `make lint` | Run `go vet` |
| `make fmt` | Format code (`make fmt-check` only reports) |
| `make clean` | Remove build artifacts |
| `make cover` | Test coverage report (`make cover-check` enforces the floors) |

### Project Structure

```
cmd/lazybrew/            Entry point
internal/
  app/                   Application bootstrap and options
  brew/                  Brew CLI abstraction and domain types
    runner.go            Brew command execution
    formulae.go          Formula read/write services
    cache.go             In-memory cache with TTL
    types.go             Domain types (Formula, Cask, Tap, Service, ...)
    errors.go            Typed error definitions
    logger.go            Structured logging (slog)
  config/                Configuration loading and defaults
  gui/                   TUI layer (Bubble Tea model, views, keybindings)
    modal/               Modal widgets (confirm, input, menu, progress, toast)
    presentation/        Formatting helpers for display strings
    style/               Lip Gloss styles and themes
work/                    Lane flow: briefs, receipts (inbox/), state, backlog, questions
testdata/                Test fixtures
```

### Target

Homebrew 6.0.0+ (requires `brew trust`/`brew untrust` support).

## Releasing

Cut a release from `stable`:

```bash
git checkout stable
git tag vX.Y.Z
git push origin vX.Y.Z
```

Then update the formula in [`lthiagol/homebrew-tap`](https://github.com/lthiagol/homebrew-tap) → `Formula/lazybrew.rb`: set the new `url` (Codeberg archive for the tag) and recompute the `sha256`. Users get the new version on `brew upgrade`.

## License

MIT
