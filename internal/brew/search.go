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

func (s *searchService) Search(ctx context.Context, query string) ([]SearchResult, error) {
	return s.search(ctx, query, false)
}

func (s *searchService) SearchDesc(ctx context.Context, query string) ([]SearchResult, error) {
	return s.search(ctx, query, true)
}

func (s *searchService) search(ctx context.Context, query string, desc bool) ([]SearchResult, error) {
	args := []string{"search"}
	if desc {
		args = append(args, "--desc")
	}
	args = append(args, query)

	output, err := s.runner.Execute(ctx, args...)
	if err != nil {
		if IsExitCode(err, 1) {
			return nil, nil
		}
		return nil, err
	}

	return parseSearchOutput(string(output)), nil
}

func parseSearchOutput(raw string) []SearchResult {
	var results []SearchResult
	var isFormula bool
	seenFormulae := false

	lines := strings.Split(raw, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "==> Formulae") {
			isFormula = true
			seenFormulae = true
			continue
		}
		if strings.HasPrefix(line, "==> Casks") {
			isFormula = false
			continue
		}
		results = append(results, SearchResult{
			Name:      line,
			IsFormula: isFormula || !seenFormulae,
			IsCask:    !isFormula && seenFormulae,
		})
	}
	return results
}
