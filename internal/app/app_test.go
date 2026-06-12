package app

import (
	"testing"
)

func TestNewWithDefaults(t *testing.T) {
	m, err := New(Options{})
	if err != nil {
		t.Fatal("New() returned error:", err)
	}
	if m == nil {
		t.Fatal("New() returned nil model")
	}
}

func TestNewWithDebug(t *testing.T) {
	m, err := New(Options{Debug: true})
	if err != nil {
		t.Fatal("New() with debug returned error:", err)
	}
	if m == nil {
		t.Fatal("New() returned nil model")
	}
}

func TestNewWithConfigPath(t *testing.T) {
	m, err := New(Options{ConfigPath: "/nonexistent/config.yml"})
	if err != nil {
		t.Fatal("New() with config path returned error:", err)
	}
	if m == nil {
		t.Fatal("New() returned nil model")
	}
}

func TestOptionsDefaults(t *testing.T) {
	opts := Options{}
	if opts.Debug {
		t.Error("Debug should default to false")
	}
	if opts.ConfigPath != "" {
		t.Error("ConfigPath should default to empty")
	}
}
