package brew

import (
	"context"
	"encoding/json"
	"strings"
	"time"
)

type FormulaeReader interface {
	List(ctx context.Context) ([]Formula, error)
	Get(ctx context.Context, name string) (*Formula, error)
	Outdated(ctx context.Context) ([]Formula, error)
	Leaves(ctx context.Context) ([]string, error)
	Deps(ctx context.Context, name string) (string, error)
	Uses(ctx context.Context, name string) ([]string, error)
}

type FormulaeWriter interface {
	Install(ctx context.Context, name string) (<-chan string, <-chan error)
	Uninstall(ctx context.Context, name string) (<-chan string, <-chan error)
	Reinstall(ctx context.Context, name string) (<-chan string, <-chan error)
	Upgrade(ctx context.Context, name string) (<-chan string, <-chan error)
	Pin(ctx context.Context, name string) error
	Unpin(ctx context.Context, name string) error
}

type formulaeReader struct {
	runner Runner
	cache  *Cache
}

type formulaeWriter struct {
	runner Runner
	cache  *Cache
}

func NewFormulaeReader(runner Runner, cache *Cache) FormulaeReader {
	return &formulaeReader{runner: runner, cache: cache}
}

func NewFormulaeWriter(runner Runner, cache *Cache) FormulaeWriter {
	return &formulaeWriter{runner: runner, cache: cache}
}

type formulaeJSON struct {
	Formulae []formulaJSON `json:"formulae"`
}

type formulaJSON struct {
	Name                string           `json:"name"`
	FullName            string           `json:"full_name"`
	Tap                 string           `json:"tap"`
	Versions            versionsJSON     `json:"versions"`
	Desc                string           `json:"desc"`
	Homepage            string           `json:"homepage"`
	License             string           `json:"license"`
	Installed           []installedJSON  `json:"installed"`
	Dependencies        flexStrings      `json:"dependencies"`
	BuildDependencies   []string         `json:"build_dependencies"`
	Caveats             string           `json:"caveats"`
	KegOnly             bool             `json:"keg_only"`
	Bottle              bottleJSON       `json:"bottle"`
	Pinned              bool             `json:"pinned"`
	Aliases             []string         `json:"aliases,omitempty"`
	Binaries            []string         `json:"binaries,omitempty"`
	InstalledDependents []string         `json:"installed_dependents,omitempty"`
	ListVersions        []string         `json:"list_versions,omitempty"`
	Revision            int              `json:"revision,omitempty"`
	Shadowed            bool             `json:"shadowed,omitempty"`
}

type versionsJSON struct {
	Stable string `json:"stable"`
	Head   string `json:"head"`
}

type installedJSON struct {
	Version             string   `json:"version"`
	InstalledOnRequest  bool     `json:"installed_on_request"`
	InstalledAsDep      bool     `json:"installed_as_dependency"`
	Time                int64    `json:"time"`
	RuntimeDependencies flexStrings `json:"runtime_dependencies"`
}

type dependencyJSON struct {
	Name string `json:"name"`
}

type bottleJSON struct {
	Stable stableBottleJSON `json:"stable"`
}

type stableBottleJSON struct {
	Files map[string]bottleFileJSON `json:"files"`
}

type bottleFileJSON struct {
	URL string `json:"url"`
}

type outdatedJSON struct {
	Formulae []outdatedFormulaJSON `json:"formulae"`
}

type outdatedFormulaJSON struct {
	Name              string   `json:"name"`
	FullName          string   `json:"full_name"`
	InstalledVersions []string `json:"installed_versions"`
	CurrentVersion    string   `json:"current_version"`
	Pinned            bool     `json:"pinned"`
}

func (s *formulaeReader) List(ctx context.Context) ([]Formula, error) {
	if cached, ok := s.cache.Get(KeyFormulaeList); ok {
		if formulae, ok := cached.([]Formula); ok {
			return formulae, nil
		}
	}

	var data formulaeJSON
	if err := s.runner.ExecuteJSON(ctx, &data, "info", "--json=v2", "--installed"); err != nil {
		return nil, err
	}

	formulae := make([]Formula, 0, len(data.Formulae))
	for _, f := range data.Formulae {
		formula := parseFormula(f)
		formulae = append(formulae, formula)
	}

	s.cache.Set(KeyFormulaeList, formulae)
	return formulae, nil
}

func (s *formulaeReader) Get(ctx context.Context, name string) (*Formula, error) {
	var data formulaeJSON
	if err := s.runner.ExecuteJSON(ctx, &data, "info", "--json=v2", name); err != nil {
		return nil, err
	}

	if len(data.Formulae) == 0 {
		return nil, nil
	}

	formula := parseFormula(data.Formulae[0])
	return &formula, nil
}

func (s *formulaeReader) Outdated(ctx context.Context) ([]Formula, error) {
	if cached, ok := s.cache.Get(KeyOutdatedFormulae); ok {
		if formulae, ok := cached.([]Formula); ok {
			return formulae, nil
		}
	}

	output, err := s.runner.Execute(ctx, "outdated", "--json=v2", "--formula")
	if err != nil {
		if IsExitCode(err, 1) {
			return []Formula{}, nil
		}
		return nil, err
	}

	var data outdatedJSON
	if err := json.Unmarshal(output, &data); err != nil {
		return nil, err
	}

	outdated := make([]Formula, 0, len(data.Formulae))
	for _, of := range data.Formulae {
		formula := Formula{
			Name:       of.Name,
			FullName:   of.FullName,
			Pinned:     of.Pinned,
			Outdated:   true,
			NewVersion: of.CurrentVersion,
		}
		if len(of.InstalledVersions) > 0 {
			formula.Version = of.InstalledVersions[0]
		}
		outdated = append(outdated, formula)
	}

	s.cache.Set(KeyOutdatedFormulae, outdated)
	return outdated, nil
}

func (s *formulaeReader) Leaves(ctx context.Context) ([]string, error) {
	output, err := s.runner.Execute(ctx, "leaves")
	if err != nil {
		return nil, err
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	leaves := make([]string, 0)
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			leaves = append(leaves, line)
		}
	}
	return leaves, nil
}

func (s *formulaeReader) Deps(ctx context.Context, name string) (string, error) {
	output, err := s.runner.Execute(ctx, "deps", "--tree", name)
	if err != nil {
		return "", err
	}
	return string(output), nil
}

func (s *formulaeReader) Uses(ctx context.Context, name string) ([]string, error) {
	output, err := s.runner.Execute(ctx, "uses", "--installed", name)
	if err != nil {
		return nil, err
	}
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	result := make([]string, 0)
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			result = append(result, line)
		}
	}
	return result, nil
}

func (s *formulaeWriter) Install(ctx context.Context, name string) (<-chan string, <-chan error) {
	s.cache.InvalidateFor("install")
	return s.runner.ExecuteStream(ctx, "install", name)
}

func (s *formulaeWriter) Uninstall(ctx context.Context, name string) (<-chan string, <-chan error) {
	s.cache.InvalidateFor("uninstall")
	return s.runner.ExecuteStream(ctx, "uninstall", name)
}

func (s *formulaeWriter) Reinstall(ctx context.Context, name string) (<-chan string, <-chan error) {
	s.cache.InvalidateFor("reinstall")
	return s.runner.ExecuteStream(ctx, "reinstall", name)
}

func (s *formulaeWriter) Upgrade(ctx context.Context, name string) (<-chan string, <-chan error) {
	s.cache.InvalidateFor("upgrade")
	args := []string{"upgrade"}
	if name != "" {
		args = append(args, name)
	}
	return s.runner.ExecuteStream(ctx, args...)
}

func (s *formulaeWriter) Pin(ctx context.Context, name string) error {
	s.cache.InvalidateFor("pin")
	_, err := s.runner.Execute(ctx, "pin", name)
	return err
}

func (s *formulaeWriter) Unpin(ctx context.Context, name string) error {
	s.cache.InvalidateFor("unpin")
	_, err := s.runner.Execute(ctx, "unpin", name)
	return err
}

func parseFormula(f formulaJSON) Formula {
	var installedOn time.Time
	var version string
	var installedOnReq, installedAsDep bool
	var deps []string

	if len(f.Installed) > 0 {
		inst := f.Installed[0]
		version = inst.Version
		installedOn = time.Unix(inst.Time, 0)
		installedOnReq = inst.InstalledOnRequest
		installedAsDep = inst.InstalledAsDep
		deps = []string(inst.RuntimeDependencies)
	} else if f.Versions.Stable != "" {
		version = f.Versions.Stable
	}

	buildDeps := []string(f.Dependencies)

	bottled := len(f.Bottle.Stable.Files) > 0

	return Formula{
		Name:                f.Name,
		FullName:            f.FullName,
		Tap:                 f.Tap,
		Version:             version,
		Description:         f.Desc,
		Homepage:            f.Homepage,
		License:             f.License,
		Pinned:              f.Pinned,
		Dependencies:        deps,
		BuildDeps:           buildDeps,
		Caveats:             f.Caveats,
		KegOnly:             f.KegOnly,
		Bottled:             bottled,
		InstalledOnReq:      installedOnReq,
		InstalledAsDep:      installedAsDep,
		InstalledOn:         installedOn,
		Aliases:             f.Aliases,
		Binaries:            f.Binaries,
		InstalledDependents: f.InstalledDependents,
		ListVersions:        f.ListVersions,
		Revision:            f.Revision,
		Shadowed:            f.Shadowed,
	}
}
