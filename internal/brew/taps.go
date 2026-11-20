package brew

import (
	"context"
	"strings"
	"time"
)

type TapsReader interface {
	List(ctx context.Context) (TapList, error)
	Get(ctx context.Context, name string) (*Tap, error)
}

// TapList is the Taps list. TrustUnknown is set when `brew tap-info` failed or
// timed out: Taps then holds the tap names only, so each tap's Trusted and its
// formula and cask names were not read. That is not the same as untrusted, and
// the GUI must not show it as such. A list with TrustUnknown set is never
// cached, so the next read tries tap-info again.
type TapList struct {
	Taps         []Tap
	TrustUnknown bool
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

func (s *tapsReader) List(ctx context.Context) (TapList, error) {
	if cached, ok := s.cache.Get(KeyTapsList); ok {
		if taps, ok := cached.([]Tap); ok {
			return TapList{Taps: taps}, nil
		}
	}

	tapNamesOutput, err := s.runner.Execute(ctx, "tap")
	if err != nil {
		return TapList{}, err
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
			// taps, flagged TrustUnknown so it does not show them as
			// untrusted, and do NOT cache — a transient batch failure must
			// not poison KeyTapsList and block retries for the whole TTL.
			return TapList{Taps: taps, TrustUnknown: true}, nil
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
	return TapList{Taps: taps}, nil
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

// tapInfoChunkMax and tapInfoChunkBytes bound one `brew tap-info` call: a name
// count, and a byte budget for the names, both well under the OS argument
// limit (ARG_MAX). Tests lower them to exercise the split.
var (
	tapInfoChunkMax   = 100
	tapInfoChunkBytes = 64 << 10
)

// fetchTapInfoBatch reads the names with `brew tap-info --json`, one call per
// chunk (see chunkTapNames), and returns the entries in input order. A typical
// install fits one chunk, so it is one call. M11 replaces the per-tap loop
// with this so listing N taps is O(1) shell calls instead of N+1.
func (s *tapsReader) fetchTapInfoBatch(ctx context.Context, names []string) ([]tapInfoJSON, error) {
	var all []tapInfoJSON
	for _, chunk := range chunkTapNames(names) {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		args := make([]string, 0, 2+len(chunk))
		args = append(args, "tap-info", "--json")
		args = append(args, chunk...)

		var data []tapInfoJSON
		if err := s.runner.ExecuteJSON(ctx, &data, args...); err != nil {
			return nil, err
		}
		all = append(all, data...)
	}
	return all, nil
}

// chunkTapNames splits names into runs of at most tapInfoChunkMax names and at
// most tapInfoChunkBytes bytes, each name counted with its terminating NUL. A
// name over the byte budget on its own still gets a run of its own.
func chunkTapNames(names []string) [][]string {
	var chunks [][]string
	start, size := 0, 0
	for i, name := range names {
		n := len(name) + 1
		if i > start && (i-start >= tapInfoChunkMax || size+n > tapInfoChunkBytes) {
			chunks = append(chunks, names[start:i])
			start, size = i, 0
		}
		size += n
	}
	if start < len(names) {
		chunks = append(chunks, names[start:])
	}
	return chunks
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
	// Invalidate like every other taps write: a repaired tap can change
	// its remote/manifest state (r1 finding 20).
	s.cache.InvalidateFor("tap")
	_, err := s.runner.Execute(ctx, "tap", "--repair", name)
	return err
}
