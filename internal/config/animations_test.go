package config

import "testing"

func TestAnimationsDefaultIsOn(t *testing.T) {
	t.Setenv(AnimationsEnv, "")
	if !Default().GUI.Animations {
		t.Error("production default gui.animations = false, want true")
	}
	cfg, err := Load(writeConfig(t, "gui:\n  mouse: true\n"))
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.GUI.Animations {
		t.Error("gui.animations with no key = false, want true")
	}
}

func TestAnimationsFromFile(t *testing.T) {
	t.Setenv(AnimationsEnv, "")
	cfg, err := Load(writeConfig(t, "gui:\n  animations: false\n"))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.GUI.Animations {
		t.Error("gui.animations: false was not read")
	}
}

func TestAnimationsEnvOverridesFile(t *testing.T) {
	t.Setenv(AnimationsEnv, "1")
	cfg, err := Load(writeConfig(t, "gui:\n  animations: false\n"))
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.GUI.Animations {
		t.Error("LAZYBREW_ANIMATIONS=1 should override the file")
	}
	t.Setenv(AnimationsEnv, "0")
	cfg, err = Load(writeConfig(t, "gui:\n  animations: true\n"))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.GUI.Animations {
		t.Error("LAZYBREW_ANIMATIONS=0 should override the file")
	}
}

func TestAnimationsEnvRejectsOtherValues(t *testing.T) {
	t.Setenv(AnimationsEnv, "yes")
	if _, err := Load(writeConfig(t, "gui:\n  mouse: true\n")); err == nil {
		t.Error("LAZYBREW_ANIMATIONS=yes was accepted, want an error")
	}
}
