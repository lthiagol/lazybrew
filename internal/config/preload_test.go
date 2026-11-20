package config

import "testing"

func TestPreloadDefaultIsOn(t *testing.T) {
	t.Setenv(PreloadEnv, "")
	if !Default().GUI.Preload {
		t.Error("production default gui.preload = false, want true")
	}
	cfg, err := Load(writeConfig(t, "gui:\n  mouse: true\n"))
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.GUI.Preload {
		t.Error("gui.preload with no key = false, want true")
	}
}

func TestPreloadFromFile(t *testing.T) {
	t.Setenv(PreloadEnv, "")
	cfg, err := Load(writeConfig(t, "gui:\n  preload: false\n"))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.GUI.Preload {
		t.Error("gui.preload: false was not read")
	}
}

func TestPreloadEnvOverridesFile(t *testing.T) {
	t.Setenv(PreloadEnv, "1")
	cfg, err := Load(writeConfig(t, "gui:\n  preload: false\n"))
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.GUI.Preload {
		t.Error("LAZYBREW_PRELOAD=1 should override the file")
	}
	t.Setenv(PreloadEnv, "0")
	cfg, err = Load(writeConfig(t, "gui:\n  preload: true\n"))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.GUI.Preload {
		t.Error("LAZYBREW_PRELOAD=0 should override the file")
	}
}

func TestPreloadEnvRejectsOtherValues(t *testing.T) {
	t.Setenv(PreloadEnv, "yes")
	if _, err := Load(writeConfig(t, "gui:\n  mouse: true\n")); err == nil {
		t.Error("LAZYBREW_PRELOAD=yes was accepted, want an error")
	}
}
