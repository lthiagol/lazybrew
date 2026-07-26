package brew

import (
	"context"
	"strings"
)

type DiagnosticsReader interface {
	Doctor(ctx context.Context) ([]DoctorWarning, error)
	Missing(ctx context.Context) ([]MissingDep, error)
	Vulns(ctx context.Context) (string, error)
	Config(ctx context.Context) (*BrewConfig, error)
	Version(ctx context.Context) (string, error)
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
	runner Runner
	cache  *Cache
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
	if cached, ok := s.cache.Get(KeyDoctorResult); ok {
		if warnings, ok := cached.([]DoctorWarning); ok {
			return warnings, nil
		}
	}

	output, err := s.runner.Execute(ctx, "doctor")
	if err != nil {
		// brew doctor exits 1 when warnings are found — still parse output
		if !IsExitCode(err, 1) {
			return nil, err
		}
	}

	text := string(output)
	if strings.Contains(text, "Your system is ready to brew") {
		s.cache.Set(KeyDoctorResult, []DoctorWarning{})
		return []DoctorWarning{}, nil
	}

	warnings := parseDoctorWarnings(text)
	s.cache.Set(KeyDoctorResult, warnings)
	return warnings, nil
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
		if strings.Contains(err.Error(), "missing") || strings.Contains(err.Error(), "exited with code 1") {
			return s.parseMissingOutput(string(output)), nil
		}
		return nil, err
	}

	return s.parseMissingOutput(string(output)), nil
}

func (s *diagnosticsReader) parseMissingOutput(text string) []MissingDep {
	var missing []MissingDep
	lines := strings.Split(strings.TrimSpace(text), "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
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

func (s *diagnosticsReader) Vulns(ctx context.Context) (string, error) {
	output, err := s.runner.Execute(ctx, "vulns")
	if err != nil {
		if !IsExitCode(err, 1) {
			return "", err
		}
	}
	return string(output), nil
}

func (s *diagnosticsReader) Config(ctx context.Context) (*BrewConfig, error) {
	if cached, ok := s.cache.Get(KeyConfig); ok {
		if cfg, ok := cached.(*BrewConfig); ok {
			return cfg, nil
		}
	}

	output, err := s.runner.Execute(ctx, "config")
	if err != nil {
		return nil, err
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

	s.cache.Set(KeyConfig, cfg)
	return cfg, nil
}

func (s *diagnosticsReader) Version(ctx context.Context) (string, error) {
	if cached, ok := s.cache.Get(KeyConfig); ok {
		if cfg, ok := cached.(*BrewConfig); ok {
			return cfg.HomebrewVersion, nil
		}
	}

	output, err := s.runner.Execute(ctx, "--version")
	if err != nil {
		return "", err
	}

	version := strings.TrimSpace(string(output))
	parts := strings.SplitN(version, " ", 3)
	if len(parts) >= 2 {
		return parts[1], nil
	}
	return version, nil
}

func (s *diagnosticsWriter) Update(ctx context.Context) (<-chan string, <-chan error) {
	s.cache.InvalidateFor("update")
	return s.runner.ExecuteStream(ctx, "update")
}

func (s *diagnosticsWriter) Cleanup(ctx context.Context, dryRun bool) (<-chan string, <-chan error) {
	args := []string{"cleanup"}
	if dryRun {
		args = append(args, "-n")
	}
	return s.runner.ExecuteStream(ctx, args...)
}

func (s *diagnosticsWriter) Autoremove(ctx context.Context, dryRun bool) (<-chan string, <-chan error) {
	args := []string{"autoremove"}
	if dryRun {
		args = append(args, "-n")
	}
	return s.runner.ExecuteStream(ctx, args...)
}
