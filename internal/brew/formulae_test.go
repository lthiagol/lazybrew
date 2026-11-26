package brew

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"testing/synctest"
	"time"

	"github.com/lthiagol/lazybrew/internal/brew/brewtest"
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
      "installed_versions": ["14.1.1"],
      "current_version": "14.1.2",
      "pinned": false,
      "pinned_version": null
    }
  ],
  "casks": []
}`

func newMockFormulaeRunner() *brewtest.MockRunner {
	return brewtest.NewMockRunner()
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
	if len(python.Dependencies) != 3 {
		t.Errorf("python Dependencies = %v, want [openssl xz zlib]", python.Dependencies)
	}
	if len(python.BuildDeps) != 1 || python.BuildDeps[0] != "pkg-config" {
		t.Errorf("python BuildDeps = %v, want [pkg-config] (from build_dependencies, not dependencies)", python.BuildDeps)
	}

	fish := formulae[2]
	if fish.Version != "3.7.1" {
		t.Errorf("fish Version = %q, want 3.7.1 (from versions.stable, no installed entry)", fish.Version)
	}
	if fish.Bottled {
		t.Error("fish should not be bottled (empty files map)")
	}
	if len(fish.Dependencies) != 1 || fish.Dependencies[0] != "pcre2" {
		t.Errorf("uninstalled fish Dependencies = %v, want [pcre2] (declared dependencies, no installed keg)", fish.Dependencies)
	}
	if len(fish.BuildDeps) != 1 || fish.BuildDeps[0] != "cmake" {
		t.Errorf("fish BuildDeps = %v, want [cmake]", fish.BuildDeps)
	}
}

func TestParseFormulaBuildDepsFromBuildDependencies(t *testing.T) {
	raw := `{
		"name":"tool","full_name":"tool","tap":"homebrew/core",
		"versions":{"stable":"1.0"},"desc":"","homepage":"","license":"",
		"installed":[{"version":"1.0","installed_on_request":true,
			"installed_as_dependency":false,"time":1700000000,
			"runtime_dependencies":["runtime-a"]}],
		"dependencies":[{"name":"runtime-a"}],
		"build_dependencies":["build-only"],
		"caveats":"","keg_only":false,
		"bottle":{"stable":{"files":{}}},
		"pinned":false
	}`
	var fj formulaJSON
	if err := json.Unmarshal([]byte(raw), &fj); err != nil {
		t.Fatal(err)
	}
	f := parseFormula(fj)
	if len(f.Dependencies) != 1 || f.Dependencies[0] != "runtime-a" {
		t.Errorf("Dependencies = %v, want [runtime-a]", f.Dependencies)
	}
	if len(f.BuildDeps) != 1 || f.BuildDeps[0] != "build-only" {
		t.Errorf("BuildDeps = %v, want [build-only]", f.BuildDeps)
	}
	for _, d := range f.BuildDeps {
		if d == "runtime-a" {
			t.Error("BuildDeps must not contain runtime dependency runtime-a")
		}
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

// TestFormulaeOutdatedNamesAreBrewsNames: brew writes an outdated formula's full
// name as name and no full_name (cmd/outdated.rb json_info). A tap formula reads
// with its short name, and its full name is the whole user/tap/foo.
//
// testdata/outdated_formulae_v2.json is BUILT, not captured: it follows brew's
// json_info (cmd/outdated.rb) because nothing was outdated on this machine when
// it was written (`brew outdated --json=v2` was empty). Replace it with a real
// capture once a tap formula is outdated.
func TestFormulaeOutdatedNamesAreBrewsNames(t *testing.T) {
	raw, err := os.ReadFile("testdata/outdated_formulae_v2.json")
	if err != nil {
		t.Fatal(err)
	}
	r := newMockFormulaeRunner()
	r.ExecuteFn = func(ctx context.Context, args ...string) ([]byte, error) {
		return raw, nil
	}
	outdated, err := NewFormulaeReader(r, NewCache(time.Minute)).Outdated(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(outdated) != 2 {
		t.Fatalf("expected 2 outdated, got %d", len(outdated))
	}
	tap := outdated[1]
	if tap.Name != "tool" || tap.FullName != "user/tap/tool" {
		t.Errorf("tap formula Name, FullName = %q, %q; want tool, user/tap/tool", tap.Name, tap.FullName)
	}
	if tap.Version != "0.4.0" || tap.NewVersion != "0.5.0" {
		t.Errorf("tap formula versions = %q -> %q, want 0.4.0 -> 0.5.0", tap.Version, tap.NewVersion)
	}
	core := outdated[0]
	if core.Name != "ripgrep" || core.FullName != "ripgrep" {
		t.Errorf("core formula Name, FullName = %q, %q; want ripgrep, ripgrep", core.Name, core.FullName)
	}
}

// TestFormulaeInstalledUsesNewestKeg: brew lists installed kegs oldest first
// (formula.rb sorts them by version), so the newest is the last entry. The
// fixture is the real bash entry with an older keg (no runtime dependencies)
// added before it; no formula on this machine has two kegs.
func TestFormulaeInstalledUsesNewestKeg(t *testing.T) {
	raw, err := os.ReadFile("testdata/info_bash_two_kegs.json")
	if err != nil {
		t.Fatal(err)
	}
	formulae, _, err := ParseInfo(raw)
	if err != nil {
		t.Fatal(err)
	}
	if len(formulae) != 1 {
		t.Fatalf("expected 1 formula, got %d", len(formulae))
	}
	f := formulae[0]
	if f.Version != "5.3.20" {
		t.Errorf("Version = %q, want the newest keg 5.3.20", f.Version)
	}
	if want := time.Unix(1793798897, 0); !f.InstalledOn.Equal(want) {
		t.Errorf("InstalledOn = %v, want %v", f.InstalledOn, want)
	}
	if got := strings.Join(f.Dependencies, ","); got != "ncurses,readline" {
		t.Errorf("Dependencies = %q, want the newest keg's ncurses,readline", got)
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

func TestFormulaeOutdatedCoalescesConcurrentCalls(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		r := newMockFormulaeRunner()
		var calls int32
		gate := make(chan struct{})
		r.ExecuteFn = func(ctx context.Context, args ...string) ([]byte, error) {
			atomic.AddInt32(&calls, 1)
			<-gate
			return []byte(sampleOutdatedJSON), nil
		}
		cache := NewCache(time.Minute)
		reader := NewFormulaeReader(r, cache)

		const goroutines = 8
		results := make([][]Formula, goroutines)
		errs := make([]error, goroutines)
		var wg sync.WaitGroup
		for i := 0; i < goroutines; i++ {
			i := i
			wg.Add(1)
			go func() {
				defer wg.Done()
				o, err := reader.Outdated(context.Background())
				results[i] = o
				errs[i] = err
			}()
		}

		// Release fn once every caller is durably blocked: the first inside fn, the rest behind it.
		synctest.Wait()
		close(gate)
		wg.Wait()

		if got := atomic.LoadInt32(&calls); got != 1 {
			t.Fatalf("expected 1 brew invocation for %d concurrent callers, got %d", goroutines, got)
		}
		for i, r := range results {
			if r == nil {
				t.Errorf("result[%d] = nil", i)
			}
			if errs[i] != nil {
				t.Errorf("errs[%d] = %v", i, errs[i])
			}
		}
	})
}

func TestFormulaeOutdatedCachesWithinTTL(t *testing.T) {
	r := newMockFormulaeRunner()
	var calls int32
	r.ExecuteFn = func(ctx context.Context, args ...string) ([]byte, error) {
		atomic.AddInt32(&calls, 1)
		return []byte(sampleOutdatedJSON), nil
	}
	cache := NewCache(time.Minute)
	reader := NewFormulaeReader(r, cache)
	reader.SetCacheTTLs(CacheTTLs{Outdated: time.Hour})

	for i := 0; i < 5; i++ {
		_, err := reader.Outdated(context.Background())
		if err != nil {
			t.Fatal(err)
		}
	}
	if got := atomic.LoadInt32(&calls); got != 1 {
		t.Errorf("expected 1 shell invocation within TTL across 5 calls, got %d", got)
	}
}

func TestFormulaeOutdatedRespectsOutdatedTTL(t *testing.T) {
	r := newMockFormulaeRunner()
	var calls int32
	r.ExecuteFn = func(ctx context.Context, args ...string) ([]byte, error) {
		atomic.AddInt32(&calls, 1)
		return []byte(sampleOutdatedJSON), nil
	}
	// Cache default is 1 minute but we want a 50ms Outdated TTL.
	cache := NewCache(time.Minute)
	reader := NewFormulaeReader(r, cache)
	reader.SetCacheTTLs(CacheTTLs{Outdated: 50 * time.Millisecond})

	_, _ = reader.Outdated(context.Background())
	time.Sleep(75 * time.Millisecond)
	_, _ = reader.Outdated(context.Background())

	if got := atomic.LoadInt32(&calls); got != 2 {
		t.Errorf("expected 2 shell invocations (one after TTL expires), got %d", got)
	}
}

// TestFormulaeListHonoursFormulaeTTL locks the M12 per-class TTL
// behaviour: SetCacheTTLs{Formulae: ...} wires the formulae list cache,
// independent of the Outdated TTL.
func TestFormulaeListHonoursFormulaeTTL(t *testing.T) {
	r := newMockFormulaeRunner()
	var calls int32
	r.ExecuteFn = func(ctx context.Context, args ...string) ([]byte, error) {
		atomic.AddInt32(&calls, 1)
		return []byte(sampleFormulaeJSON), nil
	}
	// Cache default 50ms, formulae TTL 1h.
	cache := NewCache(50 * time.Millisecond)
	reader := NewFormulaeReader(r, cache)
	reader.SetCacheTTLs(CacheTTLs{Formulae: time.Hour})

	for i := 0; i < 3; i++ {
		_, err := reader.List(context.Background())
		if err != nil {
			t.Fatal(err)
		}
	}
	if got := atomic.LoadInt32(&calls); got != 1 {
		t.Errorf("expected 1 list shell call across 3 reads within formulae TTL, got %d", got)
	}
}

// TestFormulaeTTLDontAffectOutdated: changing formulae TTL must not
// accidentally widen the outdated TTL.
func TestFormulaeTTLDontAffectOutdated(t *testing.T) {
	r := newMockFormulaeRunner()
	var formulaeCalls, outdatedCalls int32
	r.ExecuteFn = func(ctx context.Context, args ...string) ([]byte, error) {
		switch {
		case len(args) >= 1 && args[0] == "info":
			atomic.AddInt32(&formulaeCalls, 1)
			return []byte(sampleFormulaeJSON), nil
		case len(args) >= 1 && args[0] == "outdated":
			atomic.AddInt32(&outdatedCalls, 1)
			return []byte(sampleOutdatedJSON), nil
		}
		return []byte{}, nil
	}
	cache := NewCache(50 * time.Millisecond)
	reader := NewFormulaeReader(r, cache)
	reader.SetCacheTTLs(CacheTTLs{Formulae: time.Hour}) // outdated TTL stays 0 → default 50ms

	_, _ = reader.List(context.Background())
	_, _ = reader.Outdated(context.Background())
	time.Sleep(75 * time.Millisecond)
	_, _ = reader.List(context.Background())     // formulae TTL 1h → hit
	_, _ = reader.Outdated(context.Background()) // outdated TTL 50ms → miss

	if formulaeCalls != 1 {
		t.Errorf("expected 1 formulae list call, got %d", formulaeCalls)
	}
	if outdatedCalls != 2 {
		t.Errorf("expected 2 outdated calls (initial + after TTL expiry), got %d", outdatedCalls)
	}
}
