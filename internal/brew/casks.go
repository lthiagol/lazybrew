package brew

import (
	"context"
	"encoding/json"
	"sort"
	"strings"
	"time"
)

type CasksReader interface {
	List(ctx context.Context) ([]Cask, error)
	Get(ctx context.Context, name string) (*Cask, error)
	Outdated(ctx context.Context) ([]Cask, error)
	SetCacheTTLs(ttls CacheTTLs)
}

type CasksWriter interface {
	Install(ctx context.Context, name string) (<-chan string, <-chan error)
	Uninstall(ctx context.Context, name string) (<-chan string, <-chan error)
	Reinstall(ctx context.Context, name string) (<-chan string, <-chan error)
	Zap(ctx context.Context, name string) (<-chan string, <-chan error)
	Upgrade(ctx context.Context, name string) (<-chan string, <-chan error)
	Pin(ctx context.Context, name string) error
	Unpin(ctx context.Context, name string) error
}

type casksReader struct {
	runner      Runner
	cache       *Cache
	casksTTL    time.Duration
	outdatedTTL time.Duration
}

type casksWriter struct {
	runner Runner
	cache  *Cache
}

func NewCasksReader(runner Runner, cache *Cache) CasksReader {
	return &casksReader{runner: runner, cache: cache}
}

func NewCasksWriter(runner Runner, cache *Cache) CasksWriter {
	return &casksWriter{runner: runner, cache: cache}
}

type flexString string

func (f *flexString) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		*f = flexString(s)
		return nil
	}
	var a []string
	if err := json.Unmarshal(data, &a); err == nil {
		if len(a) > 0 {
			*f = flexString(a[0])
		}
		return nil
	}
	return json.Unmarshal(data, (*string)(f))
}

type flexStrings []string

// dependencyName reads one dependency: a name string, or an object. Brew
// writes full_name for tap-qualified dependencies (and for installed runtime
// dependencies, which are objects with full_name, version, ...), so it is
// preferred; name is the fallback. Any other value reads as no name.
func dependencyName(raw json.RawMessage) (string, bool) {
	switch jsonKind(raw) {
	case '"':
		var name string
		if err := json.Unmarshal(raw, &name); err == nil {
			return name, true
		}
	case '{':
		var dep struct {
			Name     *string `json:"name"`
			FullName *string `json:"full_name"`
		}
		if err := json.Unmarshal(raw, &dep); err == nil {
			if dep.FullName != nil {
				return *dep.FullName, true
			}
			if dep.Name != nil {
				return *dep.Name, true
			}
		}
	}
	return "", false
}

// jsonKind is the first byte of a JSON value: '"' for a string, '{' for an
// object, '[' for an array, 'n' for null, and 0 for no bytes.
func jsonKind(raw []byte) byte {
	if len(raw) == 0 {
		return 0
	}
	return raw[0]
}

// UnmarshalJSON reads a list of dependencies: a JSON array whose items are
// names or dependency objects, or a single dependency object. null and an
// empty array both read as an empty, non-nil list.
func (f *flexStrings) UnmarshalJSON(data []byte) error {
	*f = flexStrings{}
	switch jsonKind(data) {
	case '[':
		var items []json.RawMessage
		if err := json.Unmarshal(data, &items); err != nil {
			return err
		}
		for _, item := range items {
			if name, ok := dependencyName(item); ok {
				*f = append(*f, name)
			}
		}
		return nil
	case '{':
		if name, ok := dependencyName(data); ok {
			*f = append(*f, name)
		}
		return nil
	case 'n':
		return nil
	}
	return json.Unmarshal(data, (*[]string)(f))
}

type casksJSON struct {
	Casks []caskJSON `json:"casks"`
}

type caskJSON struct {
	Name        flexString      `json:"name"`
	FullName    string          `json:"full_name"`
	Tap         string          `json:"tap"`
	Version     string          `json:"version"`
	Desc        string          `json:"desc"`
	Homepage    string          `json:"homepage"`
	Outdated    bool            `json:"outdated"`
	AutoUpdates bool            `json:"auto_updates"`
	Pinned      bool            `json:"pinned"`
	Sha256      string          `json:"sha256"`
	URL         string          `json:"url"`
	Installed   interface{}     `json:"installed"`
	Artifacts   []interface{}   `json:"artifacts"`
	DependsOn   json.RawMessage `json:"depends_on"`
	Namespace   string          `json:"_namespace"`
}

func (s *casksReader) List(ctx context.Context) ([]Cask, error) {
	return cached(s.cache, KeyCasksList, s.casksTTL, func() ([]Cask, bool, error) {
		var data casksJSON
		if err := s.runner.ExecuteJSON(ctx, &data, "info", "--json=v2", "--installed", "--cask"); err != nil {
			return nil, false, err
		}

		casks := make([]Cask, 0, len(data.Casks))
		for _, c := range data.Casks {
			casks = append(casks, parseCask(c))
		}
		return casks, true, nil
	})
}

func (s *casksReader) Get(ctx context.Context, name string) (*Cask, error) {
	var data casksJSON
	if err := s.runner.ExecuteJSON(ctx, &data, "info", "--json=v2", name); err != nil {
		return nil, err
	}

	if len(data.Casks) == 0 {
		return nil, nil
	}

	cask := parseCask(data.Casks[0])
	return &cask, nil
}

func (s *casksReader) Outdated(ctx context.Context) ([]Cask, error) {
	return cached(s.cache, KeyOutdatedCasks, s.outdatedTTL, func() ([]Cask, bool, error) {
		return s.fetchOutdated(ctx)
	})
}

func (s *casksReader) fetchOutdated(ctx context.Context) ([]Cask, bool, error) {
	output, err := s.runner.Execute(ctx, "outdated", "--json=v2", "--cask")
	if err != nil {
		if IsExitCode(err, 1) {
			return []Cask{}, true, nil
		}
		return nil, false, err
	}

	var data struct {
		Casks []struct {
			Name              string   `json:"name"`
			FullName          string   `json:"full_name"`
			InstalledVersions []string `json:"installed_versions"`
			CurrentVersion    string   `json:"current_version"`
		} `json:"casks"`
	}
	if err := json.Unmarshal(output, &data); err != nil {
		return nil, false, err
	}

	outdated := make([]Cask, 0, len(data.Casks))
	for _, oc := range data.Casks {
		cask := Cask{
			Name:       oc.Name,
			FullName:   oc.FullName,
			Outdated:   true,
			NewVersion: oc.CurrentVersion,
		}
		if len(oc.InstalledVersions) > 0 {
			cask.Version = oc.InstalledVersions[0]
		}
		outdated = append(outdated, cask)
	}
	return outdated, true, nil
}

// SetCacheTTLs configures how long `brew list --cask` and
// `brew outdated --cask` results are reused. Implements TTLSetter.
// Values <= 0 preserve the cache default for that class.
func (s *casksReader) SetCacheTTLs(ttls CacheTTLs) {
	if ttls.Casks > 0 {
		s.casksTTL = ttls.Casks
	}
	if ttls.Outdated > 0 {
		s.outdatedTTL = ttls.Outdated
	}
}

func (s *casksWriter) Install(ctx context.Context, name string) (<-chan string, <-chan error) {
	return streamWrite(ctx, s.runner, s.cache, "install", "install", "--cask", name)
}

func (s *casksWriter) Uninstall(ctx context.Context, name string) (<-chan string, <-chan error) {
	return streamWrite(ctx, s.runner, s.cache, "uninstall", "uninstall", "--cask", name)
}

func (s *casksWriter) Reinstall(ctx context.Context, name string) (<-chan string, <-chan error) {
	return streamWrite(ctx, s.runner, s.cache, "reinstall", "reinstall", "--cask", name)
}

func (s *casksWriter) Zap(ctx context.Context, name string) (<-chan string, <-chan error) {
	return streamWrite(ctx, s.runner, s.cache, "uninstall", "uninstall", "--zap", "--cask", name)
}

func (s *casksWriter) Upgrade(ctx context.Context, name string) (<-chan string, <-chan error) {
	args := []string{"upgrade", "--cask"}
	if name != "" {
		args = append(args, name)
	}
	return streamWrite(ctx, s.runner, s.cache, "upgrade", args...)
}

func (s *casksWriter) Pin(ctx context.Context, name string) error {
	defer s.cache.InvalidateFor("pin")
	_, err := s.runner.Execute(ctx, "pin", name)
	return err
}

func (s *casksWriter) Unpin(ctx context.Context, name string) error {
	defer s.cache.InvalidateFor("unpin")
	_, err := s.runner.Execute(ctx, "unpin", name)
	return err
}

func parseCask(c caskJSON) Cask {
	installed := installedVersions(c.Installed)
	version := c.Version
	if len(installed) > 0 {
		version = installed[0].Version
	}

	artifactNames := make([]string, 0)
	for _, a := range c.Artifacts {
		if m, ok := a.(map[string]interface{}); ok {
			for key, val := range m {
				switch v := val.(type) {
				case string:
					artifactNames = append(artifactNames, v)
				case []interface{}:
					for _, item := range v {
						if s, ok := item.(string); ok {
							artifactNames = append(artifactNames, s)
						}
					}
				}
				_ = key
			}
		}
	}

	cask := Cask{
		Name:        string(c.Name),
		FullName:    c.FullName,
		Tap:         c.Tap,
		Version:     version,
		Description: c.Desc,
		Homepage:    c.Homepage,
		AutoUpdates: c.AutoUpdates,
		Pinned:      c.Pinned,
		Sha256:      c.Sha256,
		URL:         c.URL,
		Artifacts:   artifactNames,
		Installed:   installed,
	}
	cask.setDepends(parseCaskDepends(c.DependsOn))
	return cask
}

// installedVersions reads a cask's "installed" field. Brew writes the installed
// version as a string; an older payload wrote a list of objects, each with a
// version. Nothing installed reads as nil.
func installedVersions(raw interface{}) []InstalledVersion {
	switch v := raw.(type) {
	case string:
		if v != "" {
			return []InstalledVersion{{Version: v}}
		}
	case []interface{}:
		var out []InstalledVersion
		for _, e := range v {
			if m, ok := e.(map[string]interface{}); ok {
				if ver, ok := m["version"].(string); ok {
					out = append(out, InstalledVersion{Version: ver})
				}
			}
		}
		return out
	}
	return nil
}

// UnmarshalJSON decodes a cask whose depends_on is brew's object shape as well
// as a plain list, so a payload that names a formula dependency still decodes.
func (c *Cask) UnmarshalJSON(data []byte) error {
	type plain Cask
	var aux struct {
		plain
		DependsOn json.RawMessage `json:"depends_on,omitempty"`
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	*c = Cask(aux.plain)
	c.setDepends(parseCaskDepends(aux.DependsOn))
	return nil
}

// caskDepends is a cask's depends_on, read into its parts.
type caskDepends struct {
	Formulae     []string
	Casks        []string
	Requirements []string
}

// setDepends fills the cask's dependency fields from its parts. DependsOn is
// the formulae and casks together, for callers that want one list.
func (c *Cask) setDepends(d caskDepends) {
	c.DependsOnFormulae = d.Formulae
	c.DependsOnCasks = d.Casks
	c.Requirements = d.Requirements
	c.DependsOn = append(append([]string{}, d.Formulae...), d.Casks...)
}

// parseCaskDepends reads brew's depends_on. The object shape is
// {"formula": [...], "cask": [...], "macos": {...}, "arch": [...]}, and a
// plain list of formula names is read as formulae. Any other shape reads as
// no dependencies, so one odd cask cannot fail the whole list.
func parseCaskDepends(raw json.RawMessage) caskDepends {
	var d caskDepends
	if len(raw) == 0 {
		return d
	}
	var obj struct {
		Formula []string        `json:"formula"`
		Cask    []string        `json:"cask"`
		Macos   json.RawMessage `json:"macos"`
		Arch    json.RawMessage `json:"arch"`
	}
	if err := json.Unmarshal(raw, &obj); err != nil {
		var names flexStrings
		if json.Unmarshal(raw, &names) == nil {
			d.Formulae = names
		}
		return d
	}
	d.Formulae = obj.Formula
	d.Casks = obj.Cask
	d.Requirements = append(macosRequirements(obj.Macos), archRequirements(obj.Arch)...)
	return d
}

// macosRequirements reads brew's macOS requirement: a map of comparisons such
// as {">=": ["13"]}, or a plain list of releases.
func macosRequirements(raw json.RawMessage) []string {
	var bounds map[string][]string
	if err := json.Unmarshal(raw, &bounds); err == nil {
		ops := make([]string, 0, len(bounds))
		for op := range bounds {
			ops = append(ops, op)
		}
		sort.Strings(ops)
		var out []string
		for _, op := range ops {
			for _, release := range bounds[op] {
				out = append(out, "macOS "+comparison(op)+" "+release)
			}
		}
		return out
	}
	var releases []string
	if err := json.Unmarshal(raw, &releases); err == nil && len(releases) > 0 {
		return []string{"macOS " + strings.Join(releases, " or ")}
	}
	return nil
}

// comparison is the symbol a macOS requirement's operator reads as.
func comparison(op string) string {
	switch op {
	case ">=":
		return "≥"
	case "<=":
		return "≤"
	case "==", "=":
		return "="
	default:
		return op
	}
}

// archRequirements reads brew's arch requirement: a list of strings, or of
// {"type": "arm", "bits": 64} objects, which read as arm64 and x86_64.
func archRequirements(raw json.RawMessage) []string {
	var items []json.RawMessage
	if err := json.Unmarshal(raw, &items); err != nil || len(items) == 0 {
		return nil
	}
	var names []string
	for _, item := range items {
		var name string
		if json.Unmarshal(item, &name) == nil {
			names = append(names, name)
			continue
		}
		var arch struct {
			Type string `json:"type"`
			Bits int    `json:"bits"`
		}
		if json.Unmarshal(item, &arch) != nil || arch.Type == "" {
			continue
		}
		switch {
		case arch.Type == "arm" && arch.Bits == 64:
			names = append(names, "arm64")
		case arch.Type == "intel" && arch.Bits == 64:
			names = append(names, "x86_64")
		default:
			names = append(names, arch.Type)
		}
	}
	if len(names) == 0 {
		return nil
	}
	return []string{"arch " + strings.Join(names, " or ")}
}
