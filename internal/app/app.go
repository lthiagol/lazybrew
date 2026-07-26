package app

import (
	"fmt"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
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

func New(opts Options) (tea.Model, error) {
	cfg, err := config.Load(opts.ConfigPath)
	if err != nil {
		return nil, fmt.Errorf("loading config: %w", err)
	}

	if opts.Debug {
		brew.SetDebug(true)
		home, err := os.UserHomeDir()
		if err == nil {
			logPath := filepath.Join(home, ".config", "lazybrew", "debug.log")
			if err := brew.EnableFileLogging(logPath); err != nil {
				fmt.Fprintf(os.Stderr, "warning: could not enable file logging: %v\n", err)
			}
		}
	}

	switch cfg.GUI.Theme {
	case "light":
		style.ApplyTheme(style.LightTheme())
	default:
		style.ApplyTheme(style.DarkTheme())
	}

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
	model := gui.New(client, cfg)
	logRunner.OnStart = model.CommandLogStartCallback()
	logRunner.OnExec = model.CommandLogCallback()

	return model, nil
}
