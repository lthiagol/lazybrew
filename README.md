# lazybrew

![CI](https://github.com/lthiagol/lazybrew/actions/workflows/ci.yml/badge.svg)

A TUI for managing Homebrew on macOS and Linux.

## Status

Early development — TUI functional, not daily-driver reliable.

## Installation

```bash
go install github.com/thiago/lazybrew@latest
```

Or build from source:

```bash
git clone https://github.com/thiago/lazybrew
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
  sidebar_width: 30
  mouse: true
  auto_refresh_seconds: 60
brew:
  path: ""                # auto-detect if empty
  update_on_start: false
```

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
| `make run` | Run directly |
| `make test` | Run unit tests (with `-race`) |
| `make test-integration` | Run integration tests (requires Homebrew) |
| `make lint` | Run `go vet` |
| `make fmt` | Format code |
| `make clean` | Remove build artifacts |
| `make cover` | Test coverage report |

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
master-plan/             Planning documents and milestones
testdata/                Test fixtures
```

### Target

Homebrew 6.0.0+ (requires `brew trust`/`brew untrust` support).

## License

MIT
