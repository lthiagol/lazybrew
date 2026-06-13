//go:build integration

package brew

import (
	"context"
	"os"
	"os/exec"
	"strings"
	"testing"
)

func requireBrew(t *testing.T) *DefaultRunner {
	t.Helper()
	path := os.Getenv("HOMEBREW_PREFIX")
	if path == "" {
		candidates := []string{"/opt/homebrew/bin/brew", "/usr/local/bin/brew", "/home/linuxbrew/.linuxbrew/bin/brew"}
		for _, c := range candidates {
			if _, err := os.Stat(c); err == nil {
				path = c
				break
			}
		}
	}
	if path == "" {
		if _, err := exec.LookPath("brew"); err != nil {
			t.Skip("brew not found")
		}
	}
	r, err := NewDefaultRunner()
	if err != nil {
		t.Skip("brew not available:", err)
	}
	return r
}

func TestIntegrationBrewVersion(t *testing.T) {
	r := requireBrew(t)
	ctx := context.Background()
	out, err := r.Execute(ctx, "--version")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(out), "Homebrew") {
		t.Errorf("expected Homebrew in version output, got: %s", out)
	}
}

func TestIntegrationBrewListFormulaJSON(t *testing.T) {
	r := requireBrew(t)
	ctx := context.Background()
	var result interface{}
	err := r.ExecuteJSON(ctx, &result, "list", "--formula", "--json=v2")
	if err != nil {
		t.Fatal(err)
	}
}

func TestIntegrationBrewSearchJSON(t *testing.T) {
	r := requireBrew(t)
	ctx := context.Background()
	var result interface{}
	err := r.ExecuteJSON(ctx, &result, "search", "--json=v2", "git")
	if err != nil {
		t.Skip("search failed (network?):", err)
	}
}

func TestIntegrationBrewDoctor(t *testing.T) {
	r := requireBrew(t)
	ctx := context.Background()
	_, err := r.Execute(ctx, "doctor")
	if err != nil {
		t.Log("doctor reported issues:", err)
	}
}

func TestIntegrationBrewConfig(t *testing.T) {
	r := requireBrew(t)
	ctx := context.Background()
	out, err := r.Execute(ctx, "config")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(out), "HOMEBREW_PREFIX") {
		t.Errorf("expected HOMEBREW_PREFIX in config output, got: %s", out)
	}
}
