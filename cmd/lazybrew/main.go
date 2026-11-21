package main

import (
	"flag"
	"fmt"
	"os"
	"runtime/debug"

	tea "charm.land/bubbletea/v2"
	"github.com/lthiagol/lazybrew/internal/app"
)

// version is set at link time with -ldflags "-X main.version=...". Left empty,
// it falls back to the module version in the build info, then to "dev".
var version = ""

// buildVersion is the module version the binary was built from. A build from a
// checkout records "(devel)", which is not a version, so it reads as "dev".
func buildVersion() string {
	if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" && info.Main.Version != "(devel)" {
		return info.Main.Version
	}
	return "dev"
}

func main() {
	showVersion := flag.Bool("version", false, "Print version and exit")
	configPath := flag.String("config", "", "Path to config file")
	configSample := flag.Bool("config-sample", false, "Print a documented sample config and exit")
	configCreate := flag.Bool("config-create", false, "Write the sample config to the config path and exit; never overwrites")
	enableDebug := flag.Bool("debug", false, "Enable debug logging")
	flag.Parse()

	if version == "" {
		version = buildVersion()
	}

	if *showVersion {
		fmt.Printf("lazybrew %s\n", version)
		return
	}

	// The config flags exit before the app is built, so they never touch brew.
	if *configSample || *configCreate {
		os.Exit(runConfigFlags(*configSample, *configCreate, *configPath, os.Stdout, os.Stderr))
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
