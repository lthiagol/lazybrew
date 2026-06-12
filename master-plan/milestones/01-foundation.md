# Milestone 1 — Project Foundation & Core Types

> **Status:** 🔲 Not Started  
> **Depends on:** Nothing  
> **Enables:** Milestone 2 (TUI Shell), Milestone 3 (Brew Data Layer)

---

## Goal

Set up the Go project with the full directory structure, core domain types, the brew command runner abstraction, and the test infrastructure. After this milestone, we have a compilable Go project with a solid foundation — no UI yet, but all the "plumbing" is in place and tested.

---

## Why This Comes First

Everything depends on the domain types and the command runner. By building and thoroughly testing these in isolation, we avoid cascading bugs when the UI and data layers come together. This also lets us validate our JSON parsing against real brew output samples early.

---

## Steps

### 1.1 — Initialize Go Module & Directory Structure

**What:** Create the Go module, set up the full directory tree, add initial dependencies.

**Actions:**
- `go mod init github.com/<user>/lazybrew`
- Create the directory structure:
  ```
  lazybrew/
  ├── cmd/
  │   └── lazybrew/
  │       └── main.go              # Entry point (just prints "lazybrew" for now)
  ├── internal/
  │   ├── app/                     # App lifecycle (empty for now)
  │   ├── brew/                    # Brew command abstraction
  │   │   ├── runner.go
  │   │   ├── errors.go            # Typed error definitions
  │   │   ├── logger.go            # slog-based logging
  │   │   ├── types.go
  │   │   ├── formulae.go
  │   │   ├── casks.go
  │   │   ├── taps.go
  │   │   ├── services.go
  │   │   ├── search.go
  │   │   ├── doctor.go
  │   │   ├── trust.go
  │   │   └── cache.go
  │   ├── gui/                     # TUI layer (empty for now)
  │   │   ├── controllers/
  │   │   ├── presentation/
  │   │   └── style/
  │   ├── config/                  # Config loading
  │   └── utils/                   # Shared helpers
  ├── testdata/                    # Synthetic JSON fixtures for tests
  ├── master-plan/
  ├── go.mod
  ├── go.sum
  ├── Makefile
  └── README.md
  ```
- Add initial dependencies:
  - `github.com/charmbracelet/bubbletea`
  - `github.com/charmbracelet/lipgloss`
  - `github.com/charmbracelet/bubbles`
  - `github.com/stretchr/testify` (testing)

**Acceptance criteria:**
- [ ] `go build ./...` succeeds
- [ ] `go test ./...` succeeds (no tests yet, but no errors)
- [ ] Running the binary prints `lazybrew v0.0.1-dev`
- [ ] All directories exist and are tracked in git

---

### 1.2 — Define Core Domain Types

**What:** Create the Go structs that represent Homebrew's domain objects. These are the data types that flow through the entire app.

**File:** `internal/brew/types.go`

**Types to define:**
```go
// Formula represents an installed Homebrew formula
type Formula struct {
    Name            string    `json:"name"`
    FullName        string    `json:"full_name"` // e.g., "homebrew/core/neovim"
    Tap             string    `json:"tap"`
    Version         string    `json:"version"`
    Description     string    `json:"desc"`
    Homepage        string    `json:"homepage"`
    License         string    `json:"license"`
    Pinned          bool      `json:"pinned"`
    Outdated        bool      `json:"outdated"`
    NewVersion      string    `json:"new_version"` // populated when outdated
    InstalledOn     time.Time `json:"installed_on"`
    Dependencies    []string  `json:"dependencies"`
    BuildDeps       []string  `json:"build_dependencies"`
    Caveats         string    `json:"caveats"`
    KegOnly         bool      `json:"keg_only"`
    Bottled         bool      `json:"bottled"`
    InstalledOnReq  bool      `json:"installed_on_request"` // installed on request vs as dep
    InstalledAsDep  bool      `json:"installed_as_dependency"`
    InstallPath     string    `json:"install_path"`
    Size            int64     `json:"size"` // bytes
    // 6.0.0 additions:
    Aliases             []string `json:"aliases"`
    Binaries            []string `json:"binaries"`              // executables installed by this formula
    InstalledDependents []string `json:"installed_dependents"`  // reverse deps (installed only)
    ListVersions        []string `json:"list_versions"`         // all installed version history
    Revision            string   `json:"revision"`              // formula revision
    Shadowed            bool     `json:"shadowed"`              // PATH shadowing warning
}

// Cask represents an installed Homebrew cask
type Cask struct {
    Name        string   `json:"name"`
    FullName    string   `json:"full_name"`
    Tap         string   `json:"tap"`
    Version     string   `json:"version"`
    Description string   `json:"desc"`
    Homepage    string   `json:"homepage"`
    Outdated    bool     `json:"outdated"`
    NewVersion  string   `json:"new_version"`
    Artifacts   []string `json:"artifacts"` // app names
    AutoUpdates bool     `json:"auto_updates"`
    // 6.0.0 additions:
    Pinned        bool     `json:"pinned"`          // cask pinning (6.0.0)
    Sha256        string   `json:"sha256"`          // checksum
    URL           string   `json:"url"`             // download URL
    DependsOn     []string `json:"depends_on"`      // cask dependencies
    ConflictsWith []string `json:"conflicts_with"`  // conflicting casks
}

// Tap represents a Homebrew tap repository
type Tap struct {
    Name         string `json:"name"`
    Remote       string `json:"remote"`
    IsOfficial   bool   `json:"is_official"`
    FormulaCount int    `json:"formula_count"`
    CaskCount    int    `json:"cask_count"`
    CommandCount int    `json:"command_count"`
    LastCommit   string `json:"last_commit"`
    Installed    bool   `json:"installed"`
    IsAPI        bool   `json:"is_api"` // API-sourced vs git clone
    // 6.0.0 additions:
    Trusted      bool     `json:"trusted"`       // trust status from tap-info
    FormulaNames []string `json:"formula_names"` // formulae in this tap
    CaskNames    []string `json:"cask_names"`    // casks in this tap
}

// TrustStatus represents a tap/formula/cask trust state
type TrustStatus int
const (
    TrustUnknown TrustStatus = iota
    TrustOfficial             // homebrew/* — always trusted
    TrustTrusted              // explicitly trusted via `brew trust`
    TrustUntrusted            // default for third-party
)

// TrustEntry represents a single trusted item
type TrustEntry struct {
    Name    string              // tap, formula, or cask name
    Type    TrustType           // tap | formula | cask | command
    Tap     string              // parent tap
}

type TrustType string
const (
    TrustTypeTap     TrustType = "tap"
    TrustTypeFormula TrustType = "formula"
    TrustTypeCask    TrustType = "cask"
    TrustTypeCommand TrustType = "command"
)

// Service represents a brew-managed service
type Service struct {
    Name     string
    Status   ServiceStatus
    User     string
    File     string
    ExitCode int
}

type ServiceStatus string
const (
    ServiceStarted ServiceStatus = "started"
    ServiceStopped ServiceStatus = "stopped"
    ServiceError   ServiceStatus = "error"
    ServiceNone    ServiceStatus = "none"
)

// SearchResult represents a search match
type SearchResult struct {
    Name        string
    IsFormula   bool
    IsCask      bool
    Installed   bool
    Version     string
    Description string
}

// DoctorWarning represents a single brew doctor warning
type DoctorWarning struct {
    Title   string
    Details string
}

// BrewConfig holds system-level brew configuration
type BrewConfig struct {
    HomebrewVersion string
    Prefix          string
    Cellar          string
    Repository      string
    CoreTap         string
    OS              string
    Arch            string
}
```

**Acceptance criteria:**
- [ ] All types compile
- [ ] Types have JSON struct tags matching brew's `--json=v2` output
- [ ] `String()` methods on enum types (TrustStatus, ServiceStatus)
- [ ] Unit tests for enum string conversions
- [ ] 6.0.0 fields present and tagged correctly

> **Implementation note — Bottled detection:** Do NOT hardcode macOS version bottle tags (e.g., `arm64_sonoma`). Instead, check if `bottle.stable.files` has any key matching the current `runtime.GOOS` + `runtime.GOARCH` combination. Use a mapping: `darwin/arm64` → prefix `arm64_`, `darwin/amd64` → prefix `big_sur`/`monterey`/etc., `linux/amd64` → `x86_64_linux`, `linux/arm64` → `arm64_linux`. If any key matches, `Bottled = true`. This avoids breaking on every new macOS release.

> **Test fixture strategy:** Do NOT capture real `brew` output as fixtures — it changes between brew versions and rots. Instead, write **minimal synthetic JSON** fixtures that cover:
> - Happy path (all fields populated)
> - Missing optional fields (null, empty arrays, empty strings)
> - Edge cases (unicode names, very long descriptions, zero timestamps)
> - Empty lists (`{"formulae": []}`)
> 
> Keep fixtures in `testdata/` as `.json` files. Each fixture should be <50 lines — just enough to exercise the parser.

---

### 1.3 — Build the Brew Command Runner

**What:** Create the abstraction that executes `brew` CLI commands and returns parsed output.

**File:** `internal/brew/runner.go`

**Interface:**
```go
type Runner interface {
    // Execute runs a brew command and returns stdout bytes.
    // Stderr is captured separately and included in errors.
    Execute(ctx context.Context, args ...string) ([]byte, error)
    
    // ExecuteJSON runs a brew command expecting JSON output and unmarshals it.
    // Callers must include --json=v2 in args if needed (not auto-prepended).
    ExecuteJSON(ctx context.Context, result interface{}, args ...string) error
    
    // ExecuteStream runs a brew command and streams stdout line by line.
    // Stderr is captured and returned via the error channel on failure.
    ExecuteStream(ctx context.Context, args ...string) (<-chan string, <-chan error)
    
    // BrewPath returns the resolved path to the brew binary
    BrewPath() string
}
```

> **Design note:** `ExecuteJSON` no longer auto-prepends `--json=v2`. Some brew commands don't support JSON, and some callers need `--json=v1` (e.g., `brew trust --json=v1`). The caller is responsible for passing the correct flags.

**Implementation details:**
- Auto-detect brew path: check `$HOMEBREW_PREFIX/bin/brew`, `/opt/homebrew/bin/brew` (macOS ARM), `/usr/local/bin/brew` (macOS Intel), `/home/linuxbrew/.linuxbrew/bin/brew` (Linux)
- **Capture stdout and stderr separately** — `Execute()` must use `cmd.StdoutPipe()` + `cmd.StderrPipe()` (not `CombinedOutput()`), because `ExecuteJSON()` needs clean stdout for JSON parsing. Stderr is captured for error messages.
- **Set `HOMEBREW_NO_ASK=1` and `HOMEBREW_NO_AUTO_UPDATE=1`** in the command environment from day one. Homebrew 6.0.0 defaults to interactive "ask mode" which prompts for confirmation if stdin is a TTY. Since lazybrew runs brew as a subprocess, we must suppress this.
- **Pipe a non-TTY stdin** (`bytes.NewReader(nil)`) as a belt-and-suspenders approach to prevent any interactive prompts.
- Set reasonable timeouts (30s for queries, 5min for installs)
- `ExecuteStream` uses a goroutine to read stdout line-by-line and send to a channel; stderr is captured separately and included in error messages
- Wrap errors with context (`fmt.Errorf("brew %s: %w", cmd, err)`)

**Error types (define in `internal/brew/errors.go`):**
```go
// BrewNotFoundError — brew binary not found on PATH
type BrewNotFoundError struct { Searched []string }

// BrewExitError — brew exited with non-zero status
type BrewExitError struct { Command string; ExitCode int; Stderr string }

// JSONParseError — failed to parse brew JSON output
type JSONParseError struct { Command string; Cause error; RawOutput []byte }

// TimeoutError — brew command exceeded context deadline
type TimeoutError struct { Command string; Timeout time.Duration }
```

**Logging:**
- Use `log/slog` (stdlib) from day one
- Create `internal/brew/logger.go` with a package-level `slog.Logger`
- Log all brew command invocations (args, duration, exit code) at DEBUG level
- Log errors at WARN level
- `--debug` flag (M9) sets log level to DEBUG; default is WARN

**Testing approach:**
- Define a `MockRunner` that implements `Runner` with canned responses
- Unit tests use `MockRunner`
- Integration tests (build-tagged) use the real runner

**Acceptance criteria:**
- [ ] `Runner` interface defined
- [ ] `DefaultRunner` implementation works on macOS and Linux
- [ ] `MockRunner` implementation for testing
- [ ] Brew path auto-detection with fallback
- [ ] `Execute()` captures stdout and stderr **separately** (not `CombinedOutput`)
- [ ] `Execute()` sets `HOMEBREW_NO_ASK=1` and `HOMEBREW_NO_AUTO_UPDATE=1` in env
- [ ] `Execute()` pipes non-TTY stdin to suppress any interactive prompts
- [ ] `ExecuteJSON()` does NOT auto-prepend `--json=v2` (caller's responsibility)
- [ ] `ExecuteStream` correctly streams line-by-line, captures stderr
- [ ] Unit tests with `MockRunner` for Execute, ExecuteJSON, ExecuteStream
- [ ] Error wrapping with typed errors (`BrewNotFoundError`, `BrewExitError`, `JSONParseError`, `TimeoutError`)
- [ ] `slog` logger integrated — all commands logged at DEBUG level

---

### 1.4 — Create Test Infrastructure & Sample Data

**What:** Set up the test fixtures (synthetic JSON samples) and test helpers.

**Actions:**
- Create **minimal synthetic** JSON fixtures in `testdata/` (NOT captured real brew output — see design note in §1.2):
  - `testdata/formulae_installed.json` — synthetic `brew info --json=v2 --installed` with 3-5 formulae covering: normal, pinned, outdated, keg-only, no caveats, multiple installed versions
  - `testdata/casks_installed.json` — synthetic cask output with: normal, outdated, auto_updates, pinned (6.0.0)
  - `testdata/outdated.json` — synthetic `brew outdated --json=v2` with formulae and casks
  - `testdata/tap_info.json` — synthetic `brew tap-info --json` with: official, trusted third-party, untrusted third-party
  - `testdata/services.json` — synthetic `brew services list --json` with: started, stopped, error states
  - `testdata/search_results.json` — synthetic `brew search --json=v2` (use JSON, not text)
  - `testdata/doctor_clean.txt` — "Your system is ready to brew."
  - `testdata/doctor_warnings.txt` — multiple warning blocks
  - `testdata/empty_formulae.json` — `{"formulae": []}` edge case
- Create test helpers:
  - `internal/brew/testutil/mock_runner.go` — configurable mock runner (already defined in §1.3)
  - `internal/brew/testutil/fixtures.go` — helpers to load test fixtures by name

**Acceptance criteria:**
- [ ] At least one fixture file per brew command we'll parse
- [ ] `MockRunner` can be configured per-test with specific responses
- [ ] Fixture loader helper reads from `testdata/`
- [ ] `go test ./...` passes with fixture-based tests

---

### 1.5 — Makefile & Dev Tooling

**What:** Create a Makefile with common development commands.

**File:** `Makefile`

**Targets:**
```makefile
.PHONY: build run test test-integration lint fmt clean

build:            # go build -o bin/lazybrew ./cmd/lazybrew
run:              # go run ./cmd/lazybrew
test:             # go test ./... -v -race
test-integration: # go test ./... -v -race -tags=integration
lint:             # golangci-lint run
fmt:              # gofmt -s -w .
clean:            # rm -rf bin/
cover:            # go test ./... -coverprofile=coverage.out && go tool cover -html=coverage.out
```

**Also create:**
- `.golangci.yml` — linter config
- `.gitignore` — Go binaries, coverage files, IDE files
- `README.md` — project description, build instructions (minimal for now)

**Acceptance criteria:**
- [ ] `make build` produces `bin/lazybrew`
- [ ] `make test` runs all unit tests
- [ ] `make lint` runs without errors
- [ ] `.gitignore` covers Go artifacts

---

## Tests for This Milestone

| Test | Type | File | What It Validates |
|---|---|---|---|
| `TestFormulaTypes` | Unit | `internal/brew/types_test.go` | Type construction, JSON tags, enum String() |
| `TestRunnerExecute` | Unit | `internal/brew/runner_test.go` | MockRunner returns canned output |
| `TestRunnerExecuteJSON` | Unit | `internal/brew/runner_test.go` | JSON unmarshaling into types |
| `TestRunnerExecuteStream` | Unit | `internal/brew/runner_test.go` | Channel receives lines in order |
| `TestRunnerBrewPath` | Unit | `internal/brew/runner_test.go` | Path detection logic |
| `TestRunnerRealBrew` | Integration | `internal/brew/runner_integration_test.go` | Real `brew --version` call |
| `TestFixtureLoading` | Unit | `internal/brew/testutil/fixtures_test.go` | Fixture files load correctly |

---

## Definition of Done

- [ ] All steps (1.1–1.5) completed
- [ ] All tests pass (`make test`)
- [ ] `make build` produces a working binary
- [ ] `make lint` passes
- [ ] Code committed with clean git history
- [ ] No UI code yet — this is all backend/plumbing
