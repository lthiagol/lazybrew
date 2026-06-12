package brew

import (
	"context"
	"strings"
)

type TrustService interface {
	ListTrusted(ctx context.Context) ([]TrustEntry, error)
	TrustTap(ctx context.Context, tapName string) error
	TrustFormula(ctx context.Context, fullName string) error
	TrustCask(ctx context.Context, fullName string) error
	UntrustTap(ctx context.Context, tapName string) error
	UntrustFormula(ctx context.Context, fullName string) error
	UntrustCask(ctx context.Context, fullName string) error
	GetTapTrustStatus(ctx context.Context, tapName string) (TrustStatus, error)
}

type trustService struct {
	runner Runner
	cache  *Cache
}

func NewTrustService(runner Runner, cache *Cache) TrustService {
	return &trustService{runner: runner, cache: cache}
}

type trustJSON struct {
	Entries []trustEntryJSON `json:"entries"`
}

type trustEntryJSON struct {
	Name string `json:"name"`
	Type string `json:"type"`
	Tap  string `json:"tap"`
}

func (s *trustService) ListTrusted(ctx context.Context) ([]TrustEntry, error) {
	if cached, ok := s.cache.Get(KeyTrustList); ok {
		if entries, ok := cached.([]TrustEntry); ok {
			return entries, nil
		}
	}

	var data trustJSON
	if err := s.runner.ExecuteJSON(ctx, &data, "trust", "--json=v1"); err != nil {
		return nil, err
	}

	entries := make([]TrustEntry, 0, len(data.Entries))
	for _, e := range data.Entries {
		entries = append(entries, TrustEntry{
			Name: e.Name,
			Type: TrustType(e.Type),
			Tap:  e.Tap,
		})
	}

	s.cache.Set(KeyTrustList, entries)
	return entries, nil
}

func (s *trustService) TrustTap(ctx context.Context, tapName string) error {
	s.cache.InvalidateFor("trust")
	_, err := s.runner.Execute(ctx, "trust", tapName)
	return err
}

func (s *trustService) TrustFormula(ctx context.Context, fullName string) error {
	s.cache.InvalidateFor("trust")
	_, err := s.runner.Execute(ctx, "trust", "--formula", fullName)
	return err
}

func (s *trustService) TrustCask(ctx context.Context, fullName string) error {
	s.cache.InvalidateFor("trust")
	_, err := s.runner.Execute(ctx, "trust", "--cask", fullName)
	return err
}

func (s *trustService) UntrustTap(ctx context.Context, tapName string) error {
	s.cache.InvalidateFor("untrust")
	_, err := s.runner.Execute(ctx, "untrust", tapName)
	return err
}

func (s *trustService) UntrustFormula(ctx context.Context, fullName string) error {
	s.cache.InvalidateFor("untrust")
	_, err := s.runner.Execute(ctx, "untrust", "--formula", fullName)
	return err
}

func (s *trustService) UntrustCask(ctx context.Context, fullName string) error {
	s.cache.InvalidateFor("untrust")
	_, err := s.runner.Execute(ctx, "untrust", "--cask", fullName)
	return err
}

func (s *trustService) GetTapTrustStatus(ctx context.Context, tapName string) (TrustStatus, error) {
	if strings.HasPrefix(tapName, "homebrew/") {
		return TrustOfficial, nil
	}

	entries, err := s.ListTrusted(ctx)
	if err != nil {
		return TrustUnknown, err
	}

	for _, e := range entries {
		if e.Type == TrustTypeTap && e.Name == tapName {
			return TrustTrusted, nil
		}
	}

	return TrustUntrusted, nil
}
