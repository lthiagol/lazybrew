package app

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/thiago/lazybrew/internal/brew"
	"github.com/thiago/lazybrew/internal/config"
	"github.com/thiago/lazybrew/internal/gui"
	"github.com/thiago/lazybrew/internal/gui/style"
)

type Options struct {
	ConfigPath string
	Debug      bool
}

func New(opts Options) (tea.Model, error) {
	cfg, err := config.Load(opts.ConfigPath)
	if err != nil {
		return nil, fmt.Errorf("loading config: %w", err)
	}

	if opts.Debug {
		brew.SetDebug(true)
	}

	switch cfg.GUI.Theme {
	case "light":
		style.ApplyTheme(style.LightTheme())
	default:
		style.ApplyTheme(style.DarkTheme())
	}

	var runner *brew.DefaultRunner
	if cfg.Brew.Path != "" {
		runner, err = brew.NewDefaultRunnerWithPath(cfg.Brew.Path)
	} else {
		runner, err = brew.NewDefaultRunner()
	}
	if err != nil {
		return nil, fmt.Errorf("cannot start lazybrew: %w", err)
	}
	client := brew.NewClient(runner)
	return gui.New(client, cfg), nil
}
