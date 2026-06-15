package brew

import (
	"context"
	"encoding/json"
	"testing"
	"time"
)

var sampleFormulaeJSON = `{
  "formulae": [
    {
      "name": "ripgrep",
      "full_name": "ripgrep",
      "tap": "homebrew/core",
      "versions": { "stable": "14.1.2" },
      "desc": "Search tool",
      "homepage": "https://github.com/BurntSushi/ripgrep",
      "license": "MIT",
      "installed": [
        {
          "version": "14.1.1",
          "installed_on_request": true,
          "installed_as_dependency": false,
          "time": 1700000000,
          "runtime_dependencies": ["pcre2"]
        }
      ],
      "dependencies": [{ "name": "pcre2" }],
      "build_dependencies": [],
      "caveats": "",
      "keg_only": false,
      "bottle": { "stable": { "files": { "arm64_sonoma": { "url": "..." } } } },
      "pinned": false,
      "binaries": ["rg"]
    },
    {
      "name": "python@3.12",
      "full_name": "python@3.12",
      "tap": "homebrew/core",
      "versions": { "stable": "3.12.8" },
      "desc": "Python 3.12",
      "homepage": "https://www.python.org",
      "license": "Python-2.0",
      "installed": [
        {
          "version": "3.12.8",
          "installed_on_request": true,
          "installed_as_dependency": false,
          "time": 1702000000,
          "runtime_dependencies": ["openssl", "xz", "zlib"]
        }
      ],
      "dependencies": [{ "name": "openssl" }, { "name": "xz" }, { "name": "zlib" }],
      "build_dependencies": ["pkg-config"],
      "caveats": "",
      "keg_only": true,
      "bottle": { "stable": { "files": { "arm64_sonoma": { "url": "..." } } } },
      "pinned": true
    },
    {
      "name": "fish",
      "full_name": "fish",
      "tap": "homebrew/core",
      "versions": { "stable": "3.7.1" },
      "desc": "Friendly interactive shell",
      "homepage": "https://fishshell.com",
      "license": "GPL-3.0-only",
      "installed": [],
      "dependencies": [{ "name": "pcre2" }],
      "build_dependencies": ["cmake"],
      "caveats": "",
      "keg_only": false,
      "bottle": { "stable": { "files": {} } },
      "pinned": false
    }
  ]
}`

var sampleOutdatedJSON = `{
  "formulae": [
    {
      "name": "ripgrep",
      "full_name": "ripgrep",
      "installed_versions": ["14.1.1"],
      "current_version": "14.1.2",
      "pinned": false
    }
  ]
}`

func newMockFormulaeRunner() *MockRunner {
	return NewMockRunner()
}

func TestFormulaeReaderList(t *testing.T) {
	r := newMockFormulaeRunner()
	r.ExecuteFn = func(ctx context.Context, args ...string) ([]byte, error) {
		return []byte(sampleFormulaeJSON), nil
	}
	cache := NewCache(time.Minute)
	reader := NewFormulaeReader(r, cache)

	formulae, err := reader.List(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	if len(formulae) != 3 {
		t.Fatalf("expected 3 formulae, got %d", len(formulae))
	}

	ripgrep := formulae[0]
	if ripgrep.Name != "ripgrep" {
		t.Errorf("Name = %q, want ripgrep", ripgrep.Name)
	}
	if ripgrep.Version != "14.1.1" {
		t.Errorf("Version = %q, want 14.1.1", ripgrep.Version)
	}
	if ripgrep.Pinned {
		t.Error("ripgrep should not be pinned")
	}
	if !ripgrep.Bottled {
		t.Error("ripgrep should be bottled")
	}
	if !ripgrep.InstalledOnReq {
		t.Error("ripgrep should be installed on request")
	}
	if ripgrep.InstalledAsDep {
		t.Error("ripgrep should not be installed as dep")
	}
	if len(ripgrep.Dependencies) != 1 || ripgrep.Dependencies[0] != "pcre2" {
		t.Errorf("Dependencies = %v, want [pcre2]", ripgrep.Dependencies)
	}
	if len(ripgrep.Binaries) != 1 || ripgrep.Binaries[0] != "rg" {
		t.Errorf("Binaries = %v, want [rg]", ripgrep.Binaries)
	}

	python := formulae[1]
	if python.Name != "python@3.12" {
		t.Errorf("Name = %q, want python@3.12", python.Name)
	}
	if !python.KegOnly {
		t.Error("python@3.12 should be keg-only")
	}
	if !python.Pinned {
		t.Error("python@3.12 should be pinned")
	}

	fish := formulae[2]
	if fish.Version != "3.7.1" {
		t.Errorf("fish Version = %q, want 3.7.1 (from versions.stable, no installed entry)", fish.Version)
	}
	if fish.Bottled {
		t.Error("fish should not be bottled (empty files map)")
	}
}

func TestFormulaeReaderListCached(t *testing.T) {
	callCount := 0
	r := newMockFormulaeRunner()
	r.ExecuteFn = func(ctx context.Context, args ...string) ([]byte, error) {
		callCount++
		return []byte(sampleFormulaeJSON), nil
	}
	cache := NewCache(time.Minute)
	reader := NewFormulaeReader(r, cache)

	reader.List(context.Background())
	reader.List(context.Background())

	if callCount != 1 {
		t.Errorf("expected 1 brew call (cached), got %d", callCount)
	}
}

func TestFormulaeReaderListEmpty(t *testing.T) {
	r := newMockFormulaeRunner()
	r.ExecuteFn = func(ctx context.Context, args ...string) ([]byte, error) {
		return []byte(`{"formulae":[],"casks":[]}`), nil
	}
	cache := NewCache(time.Minute)
	reader := NewFormulaeReader(r, cache)

	formulae, err := reader.List(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(formulae) != 0 {
		t.Errorf("expected empty list, got %d", len(formulae))
	}
}

func TestFormulaeReaderOutdated(t *testing.T) {
	r := newMockFormulaeRunner()
	r.ExecuteFn = func(ctx context.Context, args ...string) ([]byte, error) {
		return []byte(sampleOutdatedJSON), nil
	}
	cache := NewCache(time.Minute)
	reader := NewFormulaeReader(r, cache)

	outdated, err := reader.Outdated(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	if len(outdated) != 1 {
		t.Fatalf("expected 1 outdated, got %d", len(outdated))
	}

	o := outdated[0]
	if o.Name != "ripgrep" {
		t.Errorf("Name = %q, want ripgrep", o.Name)
	}
	if !o.Outdated {
		t.Error("should be marked outdated")
	}
	if o.NewVersion != "14.1.2" {
		t.Errorf("NewVersion = %q, want 14.1.2", o.NewVersion)
	}
}

func TestFormulaeReaderOutdatedEmpty(t *testing.T) {
	r := newMockFormulaeRunner()
	r.ExecuteFn = func(ctx context.Context, args ...string) ([]byte, error) {
		return []byte(`{"formulae":[]}`), nil
	}
	cache := NewCache(time.Minute)
	reader := NewFormulaeReader(r, cache)

	outdated, err := reader.Outdated(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(outdated) != 0 {
		t.Errorf("expected empty, got %d", len(outdated))
	}
}

func TestFormulaeReaderLeaves(t *testing.T) {
	r := newMockFormulaeRunner()
	r.ExecuteFn = func(ctx context.Context, args ...string) ([]byte, error) {
		return []byte("ripgrep\nneovim\n"), nil
	}
	cache := NewCache(time.Minute)
	reader := NewFormulaeReader(r, cache)

	leaves, err := reader.Leaves(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(leaves) != 2 || leaves[0] != "ripgrep" || leaves[1] != "neovim" {
		t.Errorf("got %v, want [ripgrep neovim]", leaves)
	}
}

func TestFormulaeReaderDeps(t *testing.T) {
	r := newMockFormulaeRunner()
	r.ExecuteFn = func(ctx context.Context, args ...string) ([]byte, error) {
		return []byte("ripgrep\n└── pcre2\n"), nil
	}
	cache := NewCache(time.Minute)
	reader := NewFormulaeReader(r, cache)

	deps, err := reader.Deps(context.Background(), "ripgrep")
	if err != nil {
		t.Fatal(err)
	}
	if deps == "" {
		t.Error("expected non-empty deps output")
	}
}

func TestFormulaeReaderUses(t *testing.T) {
	r := newMockFormulaeRunner()
	r.ExecuteFn = func(ctx context.Context, args ...string) ([]byte, error) {
		return []byte("neovim\nripgrep\n"), nil
	}
	cache := NewCache(time.Minute)
	reader := NewFormulaeReader(r, cache)

	uses, err := reader.Uses(context.Background(), "pcre2")
	if err != nil {
		t.Fatal(err)
	}
	if len(uses) != 2 {
		t.Errorf("expected 2 dependents, got %d", len(uses))
	}
}

func TestFormulaeWriterInstall(t *testing.T) {
	r := newMockFormulaeRunner()
	r.ExecuteStreamFn = func(ctx context.Context, args ...string) (<-chan string, <-chan error) {
		ch := make(chan string, 1)
		errCh := make(chan error, 1)
		ch <- "installed"
		close(ch)
		close(errCh)
		return ch, errCh
	}
	cache := NewCache(time.Minute)
	writer := NewFormulaeWriter(r, cache)

	outChan, errChan := writer.Install(context.Background(), "ripgrep")
	line := <-outChan
	if line != "installed" {
		t.Errorf("got %q, want installed", line)
	}
	if err := <-errChan; err != nil {
		t.Fatal(err)
	}
}

func TestFormulaeWriterReinstall(t *testing.T) {
	r := newMockFormulaeRunner()
	r.ExecuteStreamFn = func(ctx context.Context, args ...string) (<-chan string, <-chan error) {
		ch := make(chan string, 1)
		errCh := make(chan error, 1)
		ch <- "reinstalled"
		close(ch)
		close(errCh)
		return ch, errCh
	}
	cache := NewCache(time.Minute)
	writer := NewFormulaeWriter(r, cache)

	outChan, errChan := writer.Reinstall(context.Background(), "ripgrep")
	line := <-outChan
	if line != "reinstalled" {
		t.Errorf("got %q, want reinstalled", line)
	}
	if err := <-errChan; err != nil {
		t.Fatal(err)
	}
}

func TestFormulaeWriterPin(t *testing.T) {
	r := newMockFormulaeRunner()
	r.ExecuteFn = func(ctx context.Context, args ...string) ([]byte, error) {
		return []byte(""), nil
	}
	cache := NewCache(time.Minute)
	writer := NewFormulaeWriter(r, cache)

	if err := writer.Pin(context.Background(), "python@3.12"); err != nil {
		t.Fatal(err)
	}
}

func TestFormulaeWriterUnpin(t *testing.T) {
	r := newMockFormulaeRunner()
	r.ExecuteFn = func(ctx context.Context, args ...string) ([]byte, error) {
		return []byte(""), nil
	}
	cache := NewCache(time.Minute)
	writer := NewFormulaeWriter(r, cache)

	if err := writer.Unpin(context.Background(), "python@3.12"); err != nil {
		t.Fatal(err)
	}
}

func TestFormulaeJSONParse6Point0Fields(t *testing.T) {
	raw := `{
		"formulae": [{
			"name": "ripgrep",
			"full_name": "ripgrep",
			"tap": "homebrew/core",
			"versions": { "stable": "14.1.2" },
			"installed": [{"version":"14.1.1","installed_on_request":true,"installed_as_dependency":false,"time":1700000000}],
			"binaries": ["rg"],
			"list_versions": ["14.1.0","14.1.1"],
			"revision": 1,
			"shadowed": false
		}]
	}`
	r := newMockFormulaeRunner()
	r.ExecuteFn = func(ctx context.Context, args ...string) ([]byte, error) {
		return []byte(raw), nil
	}
	cache := NewCache(time.Minute)
	reader := NewFormulaeReader(r, cache)

	formulae, err := reader.List(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(formulae) != 1 {
		t.Fatalf("expected 1, got %d", len(formulae))
	}
	f := formulae[0]
	if len(f.Binaries) != 1 || f.Binaries[0] != "rg" {
		t.Errorf("Binaries = %v, want [rg]", f.Binaries)
	}
	if len(f.ListVersions) != 2 {
		t.Errorf("ListVersions = %v, want [14.1.0 14.1.1]", f.ListVersions)
	}
	if f.Revision != 1 {
		t.Errorf("Revision = %d, want 1", f.Revision)
	}
}

func TestFormulaeOutdatedDoesNotCallList(t *testing.T) {
	listCalls := 0
	r := newMockFormulaeRunner()
	r.ExecuteFn = func(ctx context.Context, args ...string) ([]byte, error) {
		if len(args) >= 1 && args[0] == "info" {
			listCalls++
		}
		return []byte(sampleOutdatedJSON), nil
	}
	cache := NewCache(time.Minute)
	reader := NewFormulaeReader(r, cache)

	reader.Outdated(context.Background())
	if listCalls > 0 {
		t.Error("Outdated() called List() internally — should be independent")
	}
}

func TestFormulaeGetReturnsNilForMissing(t *testing.T) {
	r := newMockFormulaeRunner()
	r.ExecuteFn = func(ctx context.Context, args ...string) ([]byte, error) {
		data, _ := json.Marshal(formulaeJSON{Formulae: []formulaJSON{}})
		return data, nil
	}
	cache := NewCache(time.Minute)
	reader := NewFormulaeReader(r, cache)

	f, err := reader.Get(context.Background(), "nonexistent")
	if err != nil {
		t.Fatal(err)
	}
	if f != nil {
		t.Error("expected nil for missing formula")
	}
}
