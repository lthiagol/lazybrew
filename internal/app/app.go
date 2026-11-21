package app

import (
	"fmt"
	"os"
	"path/filepath"

	tea "charm.land/bubbletea/v2"
	"github.com/lthiagol/lazybrew/internal/brew"
	"github.com/lthiagol/lazybrew/internal/config"
	"github.com/lthiagol/lazybrew/internal/gui"
	"github.com/lthiagol/lazybrew/internal/gui/style"
)

type Options struct {
	ConfigPath string
	Debug      bool
	// Runner, if set, overrides the default Homebrew-backed runner.
	// Mainly used in tests to avoid requiring a real brew binary on PATH.
	Runner brew.Runner
}

// enableDebugLog sends brew's debug logging to ~/.config/lazybrew/debug.log.
func enableDebugLog() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	return brew.EnableFileLogging(filepath.Join(home, ".config", "lazybrew", "debug.log"))
}

func New(opts Options) (tea.Model, error) {
	cfg, err := config.Load(opts.ConfigPath)
	if err != nil {
		return nil, fmt.Errorf("loading config: %w", err)
	}

	if opts.Debug {
		// The TUI owns the terminal, so debug output goes to the log file or
		// nowhere, never to stderr. The brew logger stays off when the file
		// cannot be opened; the warning is printed before the TUI starts.
		if err := enableDebugLog(); err != nil {
			fmt.Fprintf(os.Stderr, "warning: debug logging is off: %v\n", err)
		}
	}

	theme, ok := style.ThemeByName(cfg.GUI.Theme)
	if !ok {
		// Load already rejects unknown names; this catches the two lists drifting apart.
		return nil, fmt.Errorf("gui.theme: unknown value %q", cfg.GUI.Theme)
	}
	style.ApplyTheme(theme)

	var runner brew.Runner
	if opts.Runner != nil {
		runner = opts.Runner
	} else {
		var defaultRunner *brew.DefaultRunner
		if cfg.Brew.Path != "" {
			defaultRunner, err = brew.NewDefaultRunnerWithPath(cfg.Brew.Path)
		} else {
			defaultRunner, err = brew.NewDefaultRunner()
		}
		if err != nil {
			return nil, fmt.Errorf("cannot start lazybrew: %w", err)
		}
		runner = defaultRunner
	}

	logRunner := brew.NewLoggingRunner(runner, nil, nil)
	client := brew.NewClient(logRunner)
	client.SetCacheTTLs(brew.CacheTTLs{
		Formulae: cfg.Brew.FormulaeTTL,
		Casks:    cfg.Brew.CasksTTL,
		Outdated: cfg.Brew.OutdatedTTL,
		Taps:     cfg.Brew.TapsTTL,
		Services: cfg.Brew.ServicesTTL,
		Doctor:   cfg.Brew.DoctorTTL,
	})
	model := gui.New(client, cfg)
	logRunner.OnStart = model.CommandLogStartCallback()
	logRunner.OnExec = model.CommandLogCallback()

	return model, nil
}
