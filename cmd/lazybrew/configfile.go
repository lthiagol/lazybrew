package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/lthiagol/lazybrew/internal/config"
)

// runConfigFlags carries out --config-sample or --config-create and returns the
// exit code. It is the whole of what those flags do: it never builds the app.
func runConfigFlags(sample, create bool, path string, stdout, stderr io.Writer) int {
	if sample && create {
		fmt.Fprintln(stderr, "Error: --config-sample and --config-create cannot be used together")
		return 1
	}
	if sample {
		fmt.Fprint(stdout, config.Sample())
		return 0
	}
	if !create {
		return 0
	}
	if path == "" {
		p, err := config.DefaultPath()
		if err != nil {
			fmt.Fprintf(stderr, "Error: %v\n", err)
			return 1
		}
		path = p
	}
	if err := createConfig(path); err != nil {
		fmt.Fprintf(stderr, "Error: %v\n", err)
		return 1
	}
	fmt.Fprintln(stdout, path)
	return 0
}

// createConfig writes the sample config to path, creating its directory. An
// existing file is refused and left as it is.
func createConfig(path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if os.IsExist(err) {
		return fmt.Errorf("%s already exists; not touched", path)
	}
	if err != nil {
		return err
	}
	_, err = f.WriteString(config.Sample())
	if cerr := f.Close(); err == nil {
		err = cerr
	}
	return err
}
