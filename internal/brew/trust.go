package brew

import (
	"context"
	"strings"
)

type TrustReader interface {
	ListTrusted(ctx context.Context) ([]TrustEntry, error)
}

type TrustWriter interface {
	TrustTap(ctx context.Context, tapName string) error
	TrustFormula(ctx context.Context, fullName string) error
	TrustCask(ctx context.Context, fullName string) error
	UntrustTap(ctx context.Context, tapName string) error
	UntrustFormula(ctx context.Context, fullName string) error
	UntrustCask(ctx context.Context, fullName string) error
	UntrustCommand(ctx context.Context, name string) error
}

type trustReader struct {
	runner Runner
	cache  *Cache
}

type trustWriter struct {
	runner Runner
	cache  *Cache
}

func NewTrustReader(runner Runner, cache *Cache) TrustReader {
	return &trustReader{runner: runner, cache: cache}
}

func NewTrustWriter(runner Runner, cache *Cache) TrustWriter {
	return &trustWriter{runner: runner, cache: cache}
}

// trustJSON is the output of `brew trust --json=v1`: one list of names per kind.
type trustJSON struct {
	Taps     []string `json:"taps"`
	Formulae []string `json:"formulae"`
	Casks    []string `json:"casks"`
	Commands []string `json:"commands"`
}

func (s *trustReader) ListTrusted(ctx context.Context) ([]TrustEntry, error) {
	return cached(s.cache, KeyTrustList, 0, func() ([]TrustEntry, bool, error) {
		var data trustJSON
		if err := s.runner.ExecuteJSON(ctx, &data, "trust", "--json=v1"); err != nil {
			return nil, false, err
		}

		entries := make([]TrustEntry, 0, len(data.Taps)+len(data.Formulae)+len(data.Casks)+len(data.Commands))
		entries = appendTrusted(entries, TrustTypeTap, data.Taps)
		entries = appendTrusted(entries, TrustTypeFormula, data.Formulae)
		entries = appendTrusted(entries, TrustTypeCask, data.Casks)
		entries = appendTrusted(entries, TrustTypeCommand, data.Commands)
		return entries, true, nil
	})
}

func appendTrusted(entries []TrustEntry, typ TrustType, names []string) []TrustEntry {
	for _, name := range names {
		entries = append(entries, TrustEntry{Name: name, Type: typ, Tap: tapOf(typ, name)})
	}
	return entries
}

// tapOf is the tap a trusted name belongs to. A tap is its own tap; a full
// name owner/repo/name belongs to owner/repo; a bare name has none.
func tapOf(typ TrustType, name string) string {
	if typ == TrustTypeTap {
		return name
	}
	parts := strings.Split(name, "/")
	if len(parts) != 3 {
		return ""
	}
	return parts[0] + "/" + parts[1]
}

func (s *trustWriter) TrustTap(ctx context.Context, tapName string) error {
	defer s.cache.InvalidateFor("trust")
	_, err := s.runner.Execute(ctx, "trust", tapName)
	return err
}

func (s *trustWriter) TrustFormula(ctx context.Context, fullName string) error {
	defer s.cache.InvalidateFor("trust")
	_, err := s.runner.Execute(ctx, "trust", "--formula", fullName)
	return err
}

func (s *trustWriter) TrustCask(ctx context.Context, fullName string) error {
	defer s.cache.InvalidateFor("trust")
	_, err := s.runner.Execute(ctx, "trust", "--cask", fullName)
	return err
}

func (s *trustWriter) UntrustTap(ctx context.Context, tapName string) error {
	defer s.cache.InvalidateFor("untrust")
	_, err := s.runner.Execute(ctx, "untrust", tapName)
	return err
}

func (s *trustWriter) UntrustFormula(ctx context.Context, fullName string) error {
	defer s.cache.InvalidateFor("untrust")
	_, err := s.runner.Execute(ctx, "untrust", "--formula", fullName)
	return err
}

func (s *trustWriter) UntrustCask(ctx context.Context, fullName string) error {
	defer s.cache.InvalidateFor("untrust")
	_, err := s.runner.Execute(ctx, "untrust", "--cask", fullName)
	return err
}

func (s *trustWriter) UntrustCommand(ctx context.Context, name string) error {
	defer s.cache.InvalidateFor("untrust")
	_, err := s.runner.Execute(ctx, "untrust", "--command", name)
	return err
}
