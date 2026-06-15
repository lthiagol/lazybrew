package brew

import (
	"context"
	"encoding/json"
)

type CasksReader interface {
	List(ctx context.Context) ([]Cask, error)
	Get(ctx context.Context, name string) (*Cask, error)
	Outdated(ctx context.Context) ([]Cask, error)
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
	runner Runner
	cache  *Cache
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

func (f *flexStrings) UnmarshalJSON(data []byte) error {
	var a []interface{}
	if err := json.Unmarshal(data, &a); err == nil {
		for _, v := range a {
			switch val := v.(type) {
			case string:
				*f = append(*f, val)
			case map[string]interface{}:
				if name, ok := val["name"].(string); ok {
					*f = append(*f, name)
				}
			}
		}
		if *f == nil {
			*f = flexStrings{}
		}
		return nil
	}
	var obj map[string]interface{}
	if err := json.Unmarshal(data, &obj); err == nil {
		if name, ok := obj["name"].(string); ok {
			*f = append(*f, name)
		}
		if *f == nil {
			*f = flexStrings{}
		}
		return nil
	}
	return json.Unmarshal(data, (*[]string)(f))
}

type casksJSON struct {
	Casks []caskJSON `json:"casks"`
}

type caskJSON struct {
	Name        flexString     `json:"name"`
	FullName    string        `json:"full_name"`
	Tap         string        `json:"tap"`
	Version     string        `json:"version"`
	Desc        string        `json:"desc"`
	Homepage    string        `json:"homepage"`
	Outdated    bool          `json:"outdated"`
	AutoUpdates bool          `json:"auto_updates"`
	Pinned      bool          `json:"pinned"`
	Sha256      string        `json:"sha256"`
	URL         string        `json:"url"`
	Installed   interface{}    `json:"installed"`
	Artifacts   []interface{} `json:"artifacts"`
	DependsOn   flexStrings    `json:"depends_on"`
	Namespace   string        `json:"_namespace"`
}

func (s *casksReader) List(ctx context.Context) ([]Cask, error) {
	if cached, ok := s.cache.Get(KeyCasksList); ok {
		if casks, ok := cached.([]Cask); ok {
			return casks, nil
		}
	}

	var data casksJSON
	if err := s.runner.ExecuteJSON(ctx, &data, "info", "--json=v2", "--installed", "--cask"); err != nil {
		return nil, err
	}

	casks := make([]Cask, 0, len(data.Casks))
	for _, c := range data.Casks {
		casks = append(casks, parseCask(c))
	}

	s.cache.Set(KeyCasksList, casks)
	return casks, nil
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
	if cached, ok := s.cache.Get(KeyOutdatedCasks); ok {
		if casks, ok := cached.([]Cask); ok {
			return casks, nil
		}
	}

	output, err := s.runner.Execute(ctx, "outdated", "--json=v2", "--cask")
	if err != nil {
		if IsExitCode(err, 1) {
			return []Cask{}, nil
		}
		return nil, err
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
		return nil, err
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

	s.cache.Set(KeyOutdatedCasks, outdated)
	return outdated, nil
}

func (s *casksWriter) Install(ctx context.Context, name string) (<-chan string, <-chan error) {
	s.cache.InvalidateFor("install")
	return s.runner.ExecuteStream(ctx, "install", "--cask", name)
}

func (s *casksWriter) Uninstall(ctx context.Context, name string) (<-chan string, <-chan error) {
	s.cache.InvalidateFor("uninstall")
	return s.runner.ExecuteStream(ctx, "uninstall", "--cask", name)
}

func (s *casksWriter) Reinstall(ctx context.Context, name string) (<-chan string, <-chan error) {
	s.cache.InvalidateFor("reinstall")
	return s.runner.ExecuteStream(ctx, "reinstall", "--cask", name)
}

func (s *casksWriter) Zap(ctx context.Context, name string) (<-chan string, <-chan error) {
	s.cache.InvalidateFor("uninstall")
	return s.runner.ExecuteStream(ctx, "uninstall", "--zap", "--cask", name)
}

func (s *casksWriter) Upgrade(ctx context.Context, name string) (<-chan string, <-chan error) {
	s.cache.InvalidateFor("upgrade")
	args := []string{"upgrade", "--cask"}
	if name != "" {
		args = append(args, name)
	}
	return s.runner.ExecuteStream(ctx, args...)
}

func (s *casksWriter) Pin(ctx context.Context, name string) error {
	s.cache.InvalidateFor("pin")
	_, err := s.runner.Execute(ctx, "pin", name)
	return err
}

func (s *casksWriter) Unpin(ctx context.Context, name string) error {
	s.cache.InvalidateFor("unpin")
	_, err := s.runner.Execute(ctx, "unpin", name)
	return err
}

func parseCask(c caskJSON) Cask {
	version := c.Version
	if installed, ok := c.Installed.([]interface{}); ok && len(installed) > 0 {
		if m, ok := installed[0].(map[string]interface{}); ok {
			if v, ok := m["version"].(string); ok {
				version = v
			}
		}
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

	return Cask{
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
		DependsOn:   []string(c.DependsOn),
	}
}
