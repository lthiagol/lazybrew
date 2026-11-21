package brew

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type DiagnosticsReader interface {
	Doctor(ctx context.Context) ([]DoctorWarning, error)
	Missing(ctx context.Context) ([]MissingDep, error)
	Vulns(ctx context.Context) (string, error)
	Config(ctx context.Context) (*BrewConfig, error)
	LastUpdate(ctx context.Context) (time.Time, error)
	SetCacheTTLs(ttls CacheTTLs)
}

type DiagnosticsWriter interface {
	Update(ctx context.Context) (<-chan string, <-chan error)
	Cleanup(ctx context.Context, dryRun bool) (<-chan string, <-chan error)
	Autoremove(ctx context.Context, dryRun bool) (<-chan string, <-chan error)
}

type MissingDep struct {
	Formula string
	Missing string
}

type diagnosticsReader struct {
	runner    Runner
	cache     *Cache
	doctorTTL time.Duration

	repoMu sync.Mutex
	repo   string // brew --repository, read once by repository()
}

type diagnosticsWriter struct {
	runner Runner
	cache  *Cache
}

func NewDiagnosticsReader(runner Runner, cache *Cache) DiagnosticsReader {
	return &diagnosticsReader{runner: runner, cache: cache}
}

func NewDiagnosticsWriter(runner Runner, cache *Cache) DiagnosticsWriter {
	return &diagnosticsWriter{runner: runner, cache: cache}
}

func (s *diagnosticsReader) Doctor(ctx context.Context) ([]DoctorWarning, error) {
	return cached(s.cache, KeyDoctorResult, s.doctorTTL, func() ([]DoctorWarning, bool, error) {
		output, err := s.runner.Execute(ctx, "doctor")
		if err != nil {
			// brew doctor exits 1 when warnings are found — still parse output
			if !IsExitCode(err, 1) {
				return nil, false, err
			}
		}

		text := string(output)
		if strings.Contains(text, "Your system is ready to brew") {
			return []DoctorWarning{}, true, nil
		}
		return parseDoctorWarnings(text), true, nil
	})
}

// LastUpdate is when brew last updated its taps: the mtime of
// $(brew --repository)/.git/FETCH_HEAD. A brew with no FETCH_HEAD yet gives the
// zero time. The file is read directly; git is not run.
func (s *diagnosticsReader) LastUpdate(ctx context.Context) (time.Time, error) {
	repo, err := s.repository(ctx)
	if err != nil {
		return time.Time{}, err
	}
	info, err := os.Stat(filepath.Join(repo, ".git", "FETCH_HEAD"))
	if errors.Is(err, fs.ErrNotExist) {
		return time.Time{}, nil
	}
	if err != nil {
		return time.Time{}, err
	}
	return info.ModTime(), nil
}

// repository is brew --repository. It is read once; a failed read is not kept,
// so the next call tries again.
func (s *diagnosticsReader) repository(ctx context.Context) (string, error) {
	s.repoMu.Lock()
	defer s.repoMu.Unlock()
	if s.repo != "" {
		return s.repo, nil
	}
	output, err := s.runner.Execute(ctx, "--repository")
	if err != nil {
		return "", err
	}
	repo := strings.TrimSpace(string(output))
	if repo == "" {
		return "", errors.New("brew --repository printed no path")
	}
	s.repo = repo
	return repo, nil
}

// SetCacheTTLs configures the cache TTL for `brew doctor` results. A zero TTL
// is the cache default.
func (s *diagnosticsReader) SetCacheTTLs(ttls CacheTTLs) {
	s.doctorTTL = ttls.Doctor
}

func parseDoctorWarnings(text string) []DoctorWarning {
	var warnings []DoctorWarning
	lines := strings.Split(text, "\n")
	var current DoctorWarning

	for _, line := range lines {
		if strings.HasPrefix(line, "Warning:") || strings.HasPrefix(line, "Error:") {
			if current.Title != "" {
				warnings = append(warnings, current)
			}
			current = DoctorWarning{
				Title:   strings.TrimSpace(strings.TrimPrefix(strings.TrimPrefix(line, "Warning:"), "Error:")),
				Details: "",
			}
		} else if current.Title != "" {
			trimmed := strings.TrimSpace(line)
			if trimmed != "" {
				if current.Details != "" {
					current.Details += "\n"
				}
				current.Details += trimmed
			}
		}
	}
	if current.Title != "" {
		warnings = append(warnings, current)
	}

	return warnings
}

func (s *diagnosticsReader) Missing(ctx context.Context) ([]MissingDep, error) {
	output, err := s.runner.Execute(ctx, "missing")
	if err != nil {
		// `brew missing` exits 1 when dependencies are missing — still
		// parse output. Classify strictly by exit code: the old substring
		// match on the error text treated ANY failure of the missing
		// command (the command name contains "missing") as success
		// (r1 finding 18).
		if IsExitCode(err, 1) {
			return s.parseMissingOutput(string(output)), nil
		}
		return nil, err
	}

	return s.parseMissingOutput(string(output)), nil
}

func (s *diagnosticsReader) parseMissingOutput(text string) []MissingDep {
	var missing []MissingDep
	for _, line := range nonEmptyLines([]byte(text)) {
		parts := strings.SplitN(line, ": ", 2)
		if len(parts) == 2 {
			missing = append(missing, MissingDep{
				Formula: strings.TrimSpace(parts[0]),
				Missing: strings.TrimSpace(parts[1]),
			})
		}
	}
	return missing
}

// Vulns is brew vulns' report. brew vulns exits 1 when it finds vulnerabilities
// and prints them on stdout, so exit 1 with a report is a result. Exit 1 with no
// report, or with an Error: line on stderr, is a failure: brew did not run the
// check (an unknown command, a network error).
func (s *diagnosticsReader) Vulns(ctx context.Context) (string, error) {
	output, err := s.runner.Execute(ctx, "vulns")
	if err == nil {
		return string(output), nil
	}
	var exit *BrewExitError
	if !errors.As(err, &exit) || exit.ExitCode != 1 {
		return "", err
	}
	if strings.TrimSpace(string(output)) == "" || hasErrorLine(exit.Stderr) {
		return "", err
	}
	return string(output), nil
}

// hasErrorLine reports whether a line of stderr starts with Error:.
func hasErrorLine(stderr string) bool {
	for _, line := range strings.Split(stderr, "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "Error:") {
			return true
		}
	}
	return false
}

func (s *diagnosticsReader) Config(ctx context.Context) (*BrewConfig, error) {
	return cached(s.cache, KeyConfig, 0, func() (*BrewConfig, bool, error) {
		output, err := s.runner.Execute(ctx, "config")
		if err != nil {
			return nil, false, err
		}

		cfg := &BrewConfig{}
		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			parts := strings.SplitN(line, ": ", 2)
			if len(parts) != 2 {
				continue
			}
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			switch {
			case key == "HOMEBREW_VERSION":
				cfg.HomebrewVersion = value
			case key == "HOMEBREW_PREFIX":
				cfg.Prefix = value
			case key == "HOMEBREW_CELLAR":
				cfg.Cellar = value
			case key == "HOMEBREW_REPOSITORY":
				cfg.Repository = value
			case key == "HOMEBREW_CORE_TAP":
				cfg.CoreTap = value
			case key == "HOMEBREW_SYSTEM":
				cfg.OS = value
			}
		}
		return cfg, true, nil
	})
}

func (s *diagnosticsWriter) Update(ctx context.Context) (<-chan string, <-chan error) {
	return streamWrite(ctx, s.runner, s.cache, "update", "update")
}

func (s *diagnosticsWriter) Cleanup(ctx context.Context, dryRun bool) (<-chan string, <-chan error) {
	args := []string{"cleanup"}
	if dryRun {
		args = append(args, "-n")
	}
	return streamWrite(ctx, s.runner, s.cache, "cleanup", args...)
}

func (s *diagnosticsWriter) Autoremove(ctx context.Context, dryRun bool) (<-chan string, <-chan error) {
	args := []string{"autoremove"}
	if dryRun {
		args = append(args, "-n")
	}
	return streamWrite(ctx, s.runner, s.cache, "autoremove", args...)
}
