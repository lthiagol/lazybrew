package brew

import (
	"context"
	"strings"
	"time"
)

type TapsReader interface {
	List(ctx context.Context) (TapList, error)
	SetCacheTTLs(ttls CacheTTLs)
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

// tapInfoJSON is one entry of `brew tap-info --json`. The keys are the ones
// brew's Tap#to_hash emits; brew has no formula, cask or command count keys,
// so the counts are len() of the name lists (see infoToTap).
type tapInfoJSON struct {
	Name         string   `json:"name"`
	Remote       string   `json:"remote"`
	Private      bool     `json:"private"`
	Installed    bool     `json:"installed"`
	Trusted      bool     `json:"trusted"`
	FormulaNames []string `json:"formula_names"`
	CaskTokens   []string `json:"cask_tokens"`
	CommandFiles []string `json:"command_files"`
}

func (s *tapsReader) List(ctx context.Context) (TapList, error) {
	return cached(ctx, s.cache, KeyTapsList, s.tapsTTL, func(ctx context.Context) (TapList, bool, error) {
		tapNamesOutput, err := s.runner.Execute(ctx, "tap")
		if err != nil {
			return TapList{}, false, err
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
				return TapList{Taps: taps, TrustUnknown: true}, false, nil
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
		return TapList{Taps: taps}, true, nil
	})
}

// infoToTap converts a parsed tapInfoJSON into a domain Tap.
func infoToTap(info *tapInfoJSON) *Tap {
	return &Tap{
		Name:         info.Name,
		Remote:       info.Remote,
		IsOfficial:   strings.HasPrefix(info.Name, "homebrew/"),
		FormulaCount: len(info.FormulaNames),
		CaskCount:    len(info.CaskTokens),
		CommandCount: len(info.CommandFiles),
		Installed:    true,
		Trusted:      info.Trusted,
		FormulaNames: info.FormulaNames,
		CaskNames:    info.CaskTokens,
	}
}

// SetCacheTTLs configures the cache TTL for `brew tap` results. A zero TTL is
// the cache default.
func (s *tapsReader) SetCacheTTLs(ttls CacheTTLs) {
	s.tapsTTL = ttls.Taps
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
	defer s.cache.InvalidateFor("tap")
	_, err := s.runner.Execute(ctx, "tap", name)
	return err
}

func (s *tapsWriter) Untap(ctx context.Context, name string) error {
	defer s.cache.InvalidateFor("untap")
	_, err := s.runner.Execute(ctx, "untap", name)
	return err
}

func (s *tapsWriter) Repair(ctx context.Context, name string) error {
	// Invalidate like every other taps write: a repaired tap can change
	// its remote/manifest state (r1 finding 20).
	defer s.cache.InvalidateFor("tap")
	_, err := s.runner.Execute(ctx, "tap", "--repair", name)
	return err
}
