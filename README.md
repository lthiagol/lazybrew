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
| `--config-sample` | Print a documented sample config to stdout and exit |
| `--config-create` | Write the sample config to the config path (`--config PATH` or the default) and exit; never overwrites an existing file |
| `--debug` | Enable debug logging |

### Keybindings

| Key | Action |
|-----|--------|
| `Tab` / `Shift+Tab` | Next / previous panel |
| `1`–`7` | Jump to panel |
| `↑`/`↓` or `j`/`k` | Move the cursor / scroll the list |
| `←`/`→` or `h`/`l` | Switch tabs |
| `R` | Refresh all data |
| `/` | Search packages. On Formulae and Casks, filter the list instead |
| `q` | Quit |
| `?` | Help |
| `y` | Copy the text on screen (the Operation, or the panel's content) to the clipboard. On Search `y` types into the box unless an Operation is open |

The bottom bar lists the active panel's actions, a dim `│`, then the global keys. When
the terminal is too narrow it drops keys from the right and ends with `? more`; `?`
lists every key. The bottom bar leaves out `Shift+Tab`, `j`/`k` and `h`/`l`; the `?` help
screen lists them.

On Status, `d`, `m` and `v` open the Doctor, Missing and Vulns tabs. `d` and `m` load
their tab afresh. `v` runs `brew vulns`, which queries OSV.dev over the network;
opening the Vulns tab alone runs nothing.

On Status, `c` and `A` open the Cleanup and Unused tabs. They show what `brew cleanup -n`
and `brew autoremove -n` would do, and read nothing else. Press `c` or `A` again on its
tab to confirm and run the real command; the tab then reads brew again.

On Formulae and Casks, `/` opens a filter on the loaded list: type to narrow it by
name (case-insensitive), `↑`/`↓` move the cursor, `Enter` keeps the filter and `Esc`
clears it. Actions act on the row shown. To search brew from there, press `7`, then
`/`.

Mouse capture is on by default and blocks the terminal's own text selection. Hold
Option (macOS) or Shift to select, or set `gui.mouse: false`.

When an install, reinstall, upgrade or uninstall fails because sudo needs a
password, the failed operation shows `Enter  run in this terminal`. Enter suspends
lazybrew and runs the same brew command in your terminal, so sudo can ask for
the password; then lazybrew shows the result and refreshes. `Esc` closes it.

On Search, `i` installs a result that is not installed, `x` uninstalls one that is,
and `X` zaps an installed cask. Installed results are marked ✔, and each row ends
with its type, `formula` or `cask`.

## Configuration

Config file at `~/.config/lazybrew/config.yml`. The file is optional: without it
lazybrew runs on its defaults.

**Create a config.** `lazybrew --config-create` writes a complete, commented
file to the default path (or to `--config PATH`). It never overwrites: if the
file exists, it says so and exits non-zero. `lazybrew --config-sample` prints
the same file to stdout, so `lazybrew --config-sample > ~/.config/lazybrew/config.yml`
also works for a fresh file. Every option is in it: the default is uncommented,
and the other values it accepts are named in the comment beside it. Delete a
line to follow lazybrew's default for that key. Here is the sample:

```yaml
# lazybrew config. The file is optional: delete a line to follow
# lazybrew's default for that key.
#
# For one run, these environment variables override the file:
#   LAZYBREW_ICONS=nerd|unicode  gui.icons
#   LAZYBREW_ANIMATIONS=0|1      gui.animations
#   LAZYBREW_PRELOAD=0|1         gui.preload

gui:                             # the interface
  theme: dark                    # dark, light, catppuccin-mocha, catppuccin-macchiato, catppuccin-frappe, catppuccin-latte, dracula or nord
  sidebar_width: 35              # percent of the terminal width, 15-50
  show_icons: true               # true or false; false draws no icons
  icons: nerd                    # nerd or unicode (no Nerd Font needed)
  mouse: true                    # true or false; true blocks terminal text selection: hold Option/Shift, or set false
  animations: true               # true or false; motion, short and never delays a key
  preload: true                  # true or false; load every panel in the background
  auto_refresh_seconds: 0        # 0 is off; N refreshes the data every N seconds
  auto_refresh_pause: 500ms      # a duration, e.g. 500ms; 0 turns the pause off
brew:                            # how lazybrew runs brew
  path: ""                       # the brew binary; empty auto-detects it
  update_on_start: true          # true: brew update at every launch; the other two values are below
  # update_on_start: auto        # brew update at launch only when brew's last update is older than update_stale_hours
  # update_on_start: false       # never run brew update at launch
  update_stale_hours: 6          # 1 or more hours; auto updates when brew's last update is older
  formulae_ttl: 5m               # a duration, e.g. 5m; how long formula results are reused; 0 or less keeps the default
  casks_ttl: 5m                  # a duration, e.g. 5m; how long cask results are reused; 0 or less keeps the default
  outdated_ttl: 30m              # a duration, e.g. 30m; how long the outdated list is reused; 0 or less keeps the default
  taps_ttl: 1h                   # a duration, e.g. 1h; how long tap results are reused; 0 or less keeps the default
  services_ttl: 30s              # a duration, e.g. 30s; how long service results are reused; 0 or less keeps the default
  doctor_ttl: 1h                 # a duration, e.g. 1h; how long doctor output is reused; 0 or less keeps the default
```

`brew.update_on_start` says whether lazybrew runs `brew update` when it starts.
`true` (the default) runs `brew update` at every launch. The header shows
`updating…` while it runs and `updated 2h ago` after. If the update fails
(offline, for example) lazybrew shows the error in a toast and loads the panels
from brew's old metadata. To keep the old behaviour, set `auto`: it reads when
brew last updated, and runs `brew update` only if that was more than
`update_stale_hours` ago. `false` never runs `brew update` on its own: lazybrew
does not update behind your back, and the outdated list is only as fresh as your
last `brew update`. The `R` key runs `brew update` with `auto` and `true`; with
`false` it only reloads.

**Themes.** `gui.theme` takes one of these names, in any case. Set it in the
file and restart lazybrew; an unknown name is a config error that lists them.

- `dark`: Catppuccin Mocha, the default.
- `light`: Catppuccin Latte.
- `catppuccin-mocha`, `catppuccin-macchiato`, `catppuccin-frappe`, `catppuccin-latte`: the four Catppuccin flavours.
- `dracula`: Dracula.
- `nord`: Nord.

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
