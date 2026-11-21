package config

import (
	"reflect"
	"strings"
	"testing"
	"time"
)

// clearEnvOverrides unsets the three environment overrides for one test, so a
// value in the shell cannot change what the sample loads to.
func clearEnvOverrides(t *testing.T) {
	t.Helper()
	t.Setenv(IconsEnv, "")
	t.Setenv(AnimationsEnv, "")
	t.Setenv(PreloadEnv, "")
}

// configKeyPaths returns the dotted path of every yaml key in struct type t,
// sections and leaves alike, in field order.
func configKeyPaths(t *testing.T, typ reflect.Type, prefix string) []string {
	t.Helper()
	var paths []string
	for i := range typ.NumField() {
		key := typ.Field(i).Tag.Get("yaml")
		if key == "" {
			t.Fatalf("%s.%s has no yaml tag", typ.Name(), typ.Field(i).Name)
		}
		path := key
		if prefix != "" {
			path = prefix + "." + key
		}
		paths = append(paths, path)
		if typ.Field(i).Type.Kind() == reflect.Struct {
			paths = append(paths, configKeyPaths(t, typ.Field(i).Type, path)...)
		}
	}
	return paths
}

// sampleLine returns the uncommented line that sets key in sample, or "" if
// there is none.
func sampleLine(sample, key string) string {
	for _, line := range strings.Split(sample, "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), key+":") {
			return line
		}
	}
	return ""
}

func TestSampleLoadsToDefault(t *testing.T) {
	clearEnvOverrides(t)
	cfg, err := Load(writeConfig(t, Sample()))
	if err != nil {
		t.Fatalf("the sample does not load: %v", err)
	}
	if !reflect.DeepEqual(cfg, Default()) {
		t.Errorf("the sample loads to %+v, want Default() %+v", cfg, Default())
	}
}

func TestSampleHasEveryKeyOnce(t *testing.T) {
	sample := Sample()
	for _, path := range configKeyPaths(t, reflect.TypeOf(Config{}), "") {
		key := path[strings.LastIndex(path, ".")+1:]
		n := 0
		for _, line := range strings.Split(sample, "\n") {
			if strings.HasPrefix(strings.TrimSpace(line), key+":") {
				n++
			}
		}
		if n != 1 {
			t.Errorf("key %s (%s) is set on %d uncommented lines in the sample, want 1", key, path, n)
		}
	}
}

func TestSampleCommentsCoverEveryKey(t *testing.T) {
	paths := configKeyPaths(t, reflect.TypeOf(Config{}), "")
	known := map[string]bool{}
	for _, path := range paths {
		known[path] = true
		if sampleComments[path] == "" {
			t.Errorf("key %s has no prose in sampleComments", path)
		}
	}
	for path := range sampleComments {
		if !known[path] {
			t.Errorf("sampleComments has %s, which is not a key of Config", path)
		}
	}
}

func TestSampleHeaderNamesEnvOverrides(t *testing.T) {
	header, _, _ := strings.Cut(Sample(), "\ngui:")
	for _, env := range []string{IconsEnv, AnimationsEnv, PreloadEnv} {
		if !strings.Contains(header, env) {
			t.Errorf("sample header does not name %s", env)
		}
	}
	if !strings.Contains(header, "delete a line") {
		t.Error("sample header does not say a deleted line follows the default")
	}
}

// TestSampleStatesTheRules checks that the prose names the rule loadFile
// enforces: a TTL of 0 or less keeps the default, and update_stale_hours must be
// 1 or more.
func TestSampleStatesTheRules(t *testing.T) {
	for path, prose := range sampleComments {
		if strings.HasSuffix(path, "_ttl") && !strings.Contains(prose, "0 or less keeps the default") {
			t.Errorf("%s comment %q does not say 0 or less keeps the default", path, prose)
		}
	}
	if !strings.Contains(sampleComments["brew.update_stale_hours"], "1 or more") {
		t.Errorf("brew.update_stale_hours comment %q does not say 1 or more", sampleComments["brew.update_stale_hours"])
	}
}

func TestSampleFollowsDefaults(t *testing.T) {
	cfg := Default()
	cfg.Brew.CasksTTL = 7 * time.Minute
	cfg.GUI.AutoRefreshPause = 0

	got := sample(cfg)
	if line := sampleLine(got, "casks_ttl"); !strings.Contains(line, "casks_ttl: 7m ") {
		t.Errorf("casks_ttl line = %q, want the changed default 7m", line)
	}
	if line := sampleLine(got, "auto_refresh_pause"); !strings.Contains(line, "auto_refresh_pause: 0s ") {
		t.Errorf("auto_refresh_pause line = %q, want the changed default 0s", line)
	}
	if line := sampleLine(Sample(), "casks_ttl"); !strings.Contains(line, "casks_ttl: 5m ") {
		t.Errorf("Sample() casks_ttl line = %q, want the real default 5m", line)
	}
}

func TestSampleDuration(t *testing.T) {
	tests := map[time.Duration]string{
		5 * time.Minute:        "5m",
		time.Hour:              "1h",
		90 * time.Minute:       "1h30m",
		30 * time.Second:       "30s",
		500 * time.Millisecond: "500ms",
		10 * time.Minute:       "10m",
		0:                      "0s",
	}
	for d, want := range tests {
		if got := sampleDuration(d); got != want {
			t.Errorf("sampleDuration(%s) = %q, want %q", d, got, want)
		}
	}
}
