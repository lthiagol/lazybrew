package brew

import (
	"context"
	"strings"
)

type TapsService interface {
	List(ctx context.Context) ([]Tap, error)
	Get(ctx context.Context, name string) (*Tap, error)
	Tap(ctx context.Context, name string) error
	TapWithURL(ctx context.Context, name, url string) error
	Untap(ctx context.Context, name string) error
	Repair(ctx context.Context, name string) error
}

type tapsService struct {
	runner Runner
	cache  *Cache
}

func NewTapsService(runner Runner, cache *Cache) TapsService {
	return &tapsService{runner: runner, cache: cache}
}

type tapInfoJSON struct {
	Name         string   `json:"name"`
	Remote       string   `json:"remote"`
	FormulaCount int      `json:"formula_count"`
	CaskCount    int      `json:"cask_count"`
	CommandCount int      `json:"command_count"`
	Private      bool     `json:"private"`
	Installed    bool     `json:"installed"`
	Manifest     bool     `json:"manifest"`
	API          bool     `json:"api"`
	AutoPublish  bool     `json:"auto_publish"`

	Trusted      bool     `json:"trusted,omitempty"`
	FormulaNames []string `json:"formula_names,omitempty"`
	CaskNames    []string `json:"cask_names,omitempty"`
}

func (s *tapsService) List(ctx context.Context) ([]Tap, error) {
	if cached, ok := s.cache.Get(KeyTapsList); ok {
		if taps, ok := cached.([]Tap); ok {
			return taps, nil
		}
	}

	tapNamesOutput, err := s.runner.Execute(ctx, "tap")
	if err != nil {
		return nil, err
	}

	names := strings.Fields(string(tapNamesOutput))
	taps := make([]Tap, 0, len(names))

	for _, name := range names {
		tap := Tap{
			Name:       name,
			IsOfficial: strings.HasPrefix(name, "homebrew/"),
			Installed:  true,
		}
		taps = append(taps, tap)
	}

	for i, tap := range taps {
		info, err := s.fetchTapInfo(ctx, tap.Name)
		if err != nil {
			continue
		}
		taps[i].Remote = info.Remote
		taps[i].FormulaCount = info.FormulaCount
		taps[i].CaskCount = info.CaskCount
		taps[i].CommandCount = info.CommandCount
		taps[i].IsAPI = info.API
		taps[i].Trusted = info.Trusted
		taps[i].FormulaNames = info.FormulaNames
		taps[i].CaskNames = info.CaskNames
	}

	s.cache.Set(KeyTapsList, taps)
	return taps, nil
}

func (s *tapsService) Get(ctx context.Context, name string) (*Tap, error) {
	info, err := s.fetchTapInfo(ctx, name)
	if err != nil {
		return nil, err
	}

	tap := Tap{
		Name:         name,
		Remote:       info.Remote,
		IsOfficial:   strings.HasPrefix(name, "homebrew/"),
		FormulaCount: info.FormulaCount,
		CaskCount:    info.CaskCount,
		CommandCount: info.CommandCount,
		Installed:    true,
		IsAPI:        info.API,
		Trusted:      info.Trusted,
		FormulaNames: info.FormulaNames,
		CaskNames:    info.CaskNames,
	}
	return &tap, nil
}

func (s *tapsService) fetchTapInfo(ctx context.Context, name string) (*tapInfoJSON, error) {
	var data []tapInfoJSON
	if err := s.runner.ExecuteJSON(ctx, &data, "tap-info", "--json", name); err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return nil, nil
	}
	return &data[0], nil
}

func (s *tapsService) Tap(ctx context.Context, name string) error {
	s.cache.InvalidateFor("tap")
	_, err := s.runner.Execute(ctx, "tap", name)
	return err
}

func (s *tapsService) TapWithURL(ctx context.Context, name, url string) error {
	s.cache.InvalidateFor("tap")
	_, err := s.runner.Execute(ctx, "tap", name, url)
	return err
}

func (s *tapsService) Untap(ctx context.Context, name string) error {
	s.cache.InvalidateFor("untap")
	_, err := s.runner.Execute(ctx, "untap", name)
	return err
}

func (s *tapsService) Repair(ctx context.Context, name string) error {
	_, err := s.runner.Execute(ctx, "tap", "--repair", name)
	return err
}
