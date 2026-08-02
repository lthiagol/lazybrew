package brew

import (
	"context"
	"strings"
	"time"
)

type TapsReader interface {
	List(ctx context.Context) ([]Tap, error)
	Get(ctx context.Context, name string) (*Tap, error)
}

type TapsWriter interface {
	Tap(ctx context.Context, name string) error
	TapWithURL(ctx context.Context, name, url string) error
	Untap(ctx context.Context, name string) error
	Repair(ctx context.Context, name string) error
}

type tapsReader struct {
	runner  Runner
	cache   *Cache
	tapsTTL time.Duration
}

type tapsWriter struct {
	runner Runner
	cache  *Cache
}

func NewTapsReader(runner Runner, cache *Cache) TapsReader {
	return &tapsReader{runner: runner, cache: cache}
}

func NewTapsWriter(runner Runner, cache *Cache) TapsWriter {
	return &tapsWriter{runner: runner, cache: cache}
}

type tapInfoJSON struct {
	Name         string `json:"name"`
	Remote       string `json:"remote"`
	FormulaCount int    `json:"formula_count"`
	CaskCount    int    `json:"cask_count"`
	CommandCount int    `json:"command_count"`
	Private      bool   `json:"private"`
	Installed    bool   `json:"installed"`
	Manifest     bool   `json:"manifest"`
	API          bool   `json:"api"`
	AutoPublish  bool   `json:"auto_publish"`

	Trusted      bool     `json:"trusted,omitempty"`
	FormulaNames []string `json:"formula_names,omitempty"`
	CaskNames    []string `json:"cask_names,omitempty"`
}

func (s *tapsReader) List(ctx context.Context) ([]Tap, error) {
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

	// M11: build the full list from names alone. We populate rich detail
	// (formula/cask names, counts, trust) in a single batch tap-info call
	// below instead of one shell invocation per tap.
	for _, name := range names {
		taps = append(taps, Tap{
			Name:       name,
			IsOfficial: strings.HasPrefix(name, "homebrew/"),
			Installed:  true,
		})
	}

	if len(names) > 0 {
		infos, err := s.fetchTapInfoBatch(ctx, names)
		if err != nil {
			// Keep name-only list for this call so the GUI still renders
			// taps, but do NOT cache — a transient batch failure must not
			// poison KeyTapsList and block retries for the whole TTL.
			return taps, nil
		}
		byName := make(map[string]tapInfoJSON, len(infos))
		for _, info := range infos {
			byName[info.Name] = info
		}
		for i, tap := range taps {
			info, ok := byName[tap.Name]
			if !ok {
				continue
			}
			// Reuse infoToTap so List and Get stay field-aligned.
			enriched := infoToTap(&info)
			enriched.IsOfficial = tap.IsOfficial
			taps[i] = *enriched
		}
	}

	s.cache.SetWithTTL(KeyTapsList, taps, s.tapsTTL)
	return taps, nil
}

func (s *tapsReader) Get(ctx context.Context, name string) (*Tap, error) {
	info, err := s.fetchTapInfo(ctx, name)
	if err != nil {
		return nil, err
	}
	if info == nil {
		return nil, nil
	}

	return infoToTap(info), nil
}

// infoToTap converts a parsed tapInfoJSON into a domain Tap. Shared by Get
// and the batch path in List so the two stay in lockstep.
func infoToTap(info *tapInfoJSON) *Tap {
	return &Tap{
		Name:         info.Name,
		Remote:       info.Remote,
		IsOfficial:   strings.HasPrefix(info.Name, "homebrew/"),
		FormulaCount: info.FormulaCount,
		CaskCount:    info.CaskCount,
		CommandCount: info.CommandCount,
		Installed:    true,
		IsAPI:        info.API,
		Trusted:      info.Trusted,
		FormulaNames: info.FormulaNames,
		CaskNames:    info.CaskNames,
	}
}

// SetCacheTTLs configures the cache TTL for `brew tap` results.
// Implements TTLSetter.
func (s *tapsReader) SetCacheTTLs(ttls CacheTTLs) {
	if ttls.Taps > 0 {
		s.tapsTTL = ttls.Taps
	}
}

func (s *tapsReader) fetchTapInfo(ctx context.Context, name string) (*tapInfoJSON, error) {
	infos, err := s.fetchTapInfoBatch(ctx, []string{name})
	if err != nil {
		return nil, err
	}
	if len(infos) == 0 {
		return nil, nil
	}
	return &infos[0], nil
}

// fetchTapInfoBatch issues a single `brew tap-info --json name1 name2 …`
// invocation and returns the parsed entries. M11 replaces the per-tap
// loop with this so listing N taps is O(1) shell calls instead of N+1.
func (s *tapsReader) fetchTapInfoBatch(ctx context.Context, names []string) ([]tapInfoJSON, error) {
	args := make([]string, 0, 2+len(names))
	args = append(args, "tap-info", "--json")
	args = append(args, names...)

	var data []tapInfoJSON
	if err := s.runner.ExecuteJSON(ctx, &data, args...); err != nil {
		return nil, err
	}
	return data, nil
}

func (s *tapsWriter) Tap(ctx context.Context, name string) error {
	s.cache.InvalidateFor("tap")
	_, err := s.runner.Execute(ctx, "tap", name)
	return err
}

func (s *tapsWriter) TapWithURL(ctx context.Context, name, url string) error {
	s.cache.InvalidateFor("tap")
	_, err := s.runner.Execute(ctx, "tap", name, url)
	return err
}

func (s *tapsWriter) Untap(ctx context.Context, name string) error {
	s.cache.InvalidateFor("untap")
	_, err := s.runner.Execute(ctx, "untap", name)
	return err
}

func (s *tapsWriter) Repair(ctx context.Context, name string) error {
	_, err := s.runner.Execute(ctx, "tap", "--repair", name)
	return err
}
