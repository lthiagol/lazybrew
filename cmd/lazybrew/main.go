package main

import (
	"flag"
	"fmt"
	"os"
	"runtime/debug"

	tea "charm.land/bubbletea/v2"
	"github.com/lthiagol/lazybrew/internal/app"
)

var version = "1.0.0-rc1"

func main() {
	showVersion := flag.Bool("version", false, "Print version and exit")
	configPath := flag.String("config", "", "Path to config file")
	enableDebug := flag.Bool("debug", false, "Enable debug logging")
	flag.Parse()

	if info, ok := debug.ReadBuildInfo(); ok && version == "1.0.0-rc1" {
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

	p := tea.NewProgram(m)

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
