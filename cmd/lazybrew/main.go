package main

import (
	"flag"
	"fmt"
	"os"
	"runtime/debug"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/thiago/lazybrew/internal/app"
	"github.com/thiago/lazybrew/internal/gui"
)

var version = "v0.1.0-dev"

func main() {
	showVersion := flag.Bool("version", false, "Print version and exit")
	configPath := flag.String("config", "", "Path to config file")
	enableDebug := flag.Bool("debug", false, "Enable debug logging")
	flag.Parse()

	if info, ok := debug.ReadBuildInfo(); ok && version == "v0.1.0-dev" {
		if info.Main.Version != "(devel)" {
			version = info.Main.Version
		}
	}

	if *showVersion {
		fmt.Printf("lazybrew %s\n", version)
		return
	}

	m, err := app.New(app.Options{
		ConfigPath: *configPath,
		Debug:      *enableDebug,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	opts := []tea.ProgramOption{tea.WithAltScreen()}
	if gm, ok := m.(*gui.Model); ok && gm.Cfg().GUI.Mouse {
		opts = append(opts, tea.WithMouseCellMotion())
	}

	p := tea.NewProgram(m, opts...)

	if gm, ok := m.(*gui.Model); ok {
		gm.SetProgram(p)
	}

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
