# lazybrew

A TUI for managing Homebrew on macOS and Linux.

## Status

Early development. Not ready for daily use.

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

Prints the current version and exits. The interactive TUI is not yet built.

## Development

### Prerequisites

- Go 1.24+
- Homebrew (for integration tests and runtime)
- `git` — for version control
- `make` — to run build targets

Optional but recommended:
- `gofmt` — bundled with Go, needed for `make fmt`
- `goreleaser` — for building releases

Run `./check_dependencies.sh` to verify all dependencies are met.

### Commands

| Command | Description |
|---|---|
| `make build` | Build binary to `bin/lazybrew` |
| `make run` | Run directly |
| `make test` | Run unit tests |
| `make test-integration` | Run integration tests (requires Homebrew) |
| `make lint` | Run `go vet` |
| `make fmt` | Format code |
| `make clean` | Remove build artifacts |
| `make cover` | Test coverage report |

### Project Structure

```
cmd/lazybrew/          Entry point
internal/
  brew/                Brew CLI abstraction and domain types
    runner.go          Brew command execution
    formulae.go        Formula read/write services
    cache.go           In-memory cache with TTL
    types.go           Domain types (Formula, Cask, Tap, Service, ...)
    errors.go          Typed error definitions
    logger.go          Structured logging (slog)
  gui/                 TUI layer (not yet built)
  config/              Configuration (not yet built)
master-plan/           Planning documents and milestones
testdata/              Test fixtures
```

### Target

Homebrew 6.0.0+ (requires `brew trust`/`brew untrust` support).

## License

MIT
