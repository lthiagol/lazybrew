package brew

import (
	"context"
	"strings"
)

type SearchService interface {
	Search(ctx context.Context, query string) ([]SearchResult, error)
	SearchDesc(ctx context.Context, query string) ([]SearchResult, error)
}

type searchService struct {
	runner Runner
}

func NewSearchService(runner Runner) SearchService {
	return &searchService{runner: runner}
}

type searchJSON struct {
	Formulae []searchItemJSON `json:"formulae"`
	Casks    []searchItemJSON `json:"casks"`
}

type searchItemJSON struct {
	Name        string        `json:"name"`
	FullName    string        `json:"full_name"`
	Description string        `json:"desc"`
	Tap         string        `json:"tap"`
	Homepage    string        `json:"homepage"`
	Installed   []interface{} `json:"installed"`
}

func (s *searchService) Search(ctx context.Context, query string) ([]SearchResult, error) {
	return s.search(ctx, query, false)
}

func (s *searchService) SearchDesc(ctx context.Context, query string) ([]SearchResult, error) {
	return s.search(ctx, query, true)
}

func (s *searchService) search(ctx context.Context, query string, desc bool) ([]SearchResult, error) {
	args := []string{"search", "--json=v2"}
	if desc {
		args = append(args, "--desc")
	}
	args = append(args, query)

	var data searchJSON
	if err := s.runner.ExecuteJSON(ctx, &data, args...); err != nil {
		return nil, err
	}

	results := make([]SearchResult, 0, len(data.Formulae)+len(data.Casks))
	for _, f := range data.Formulae {
		results = append(results, searchResultFromItem(f, true))
	}
	for _, c := range data.Casks {
		results = append(results, searchResultFromItem(c, false))
	}

	return results, nil
}

func searchResultFromItem(item searchItemJSON, isFormula bool) SearchResult {
	installed := len(item.Installed) > 0

	version := ""
	if installed {
		if m, ok := item.Installed[0].(map[string]interface{}); ok {
			if v, ok := m["version"].(string); ok {
				version = v
			}
		}
	}

	return SearchResult{
		Name:        item.Name,
		IsFormula:   isFormula,
		IsCask:      !isFormula,
		Installed:   installed,
		Version:     version,
		Description: strings.TrimSpace(item.Description),
	}
}
