package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/lthiagol/lazybrew/internal/config"
)

func TestConfigSamplePrintsToStdout(t *testing.T) {
	var out, errOut bytes.Buffer
	if code := runConfigFlags(true, false, "", &out, &errOut); code != 0 {
		t.Fatalf("exit code = %d, want 0", code)
	}
	if out.String() != config.Sample() {
		t.Error("stdout is not config.Sample()")
	}
	if errOut.Len() != 0 {
		t.Errorf("stderr = %q, want nothing", errOut.String())
	}
}

func TestConfigCreateWritesSampleAndMakesDir(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "dir", "c.yml")
	var out, errOut bytes.Buffer
	if code := runConfigFlags(false, true, path, &out, &errOut); code != 0 {
		t.Fatalf("exit code = %d, stderr %q, want 0", code, errOut.String())
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != config.Sample() {
		t.Error("the written file is not config.Sample()")
	}
	if out.String() != path+"\n" {
		t.Errorf("stdout = %q, want the path written", out.String())
	}
}

func TestConfigCreateRefusesExistingFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "c.yml")
	mine := []byte("gui:\n  theme: light\n")
	if err := os.WriteFile(path, mine, 0o644); err != nil {
		t.Fatal(err)
	}
	var out, errOut bytes.Buffer
	if code := runConfigFlags(false, true, path, &out, &errOut); code == 0 {
		t.Fatal("exit code = 0 over an existing file, want non-zero")
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, mine) {
		t.Errorf("existing file changed to %q", got)
	}
	if !strings.Contains(errOut.String(), path) || !strings.Contains(errOut.String(), "not touched") {
		t.Errorf("stderr = %q, want the path and \"not touched\"", errOut.String())
	}
}

func TestConfigCreateDefaultPath(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	var out, errOut bytes.Buffer
	if code := runConfigFlags(false, true, "", &out, &errOut); code != 0 {
		t.Fatalf("exit code = %d, stderr %q, want 0", code, errOut.String())
	}
	want := filepath.Join(home, ".config", "lazybrew", "config.yml")
	got, err := os.ReadFile(want)
	if err != nil {
		t.Fatalf("no file at the default path %s: %v", want, err)
	}
	if string(got) != config.Sample() {
		t.Error("the default-path file is not config.Sample()")
	}
}
