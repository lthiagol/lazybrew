package brew

import (
	"context"
	"strings"
)

type SearchService interface {
	Search(ctx context.Context, query string) ([]SearchResult, error)
}

type searchService struct {
	runner Runner
}

func NewSearchService(runner Runner) SearchService {
	return &searchService{runner: runner}
}

func (s *searchService) Search(ctx context.Context, query string) ([]SearchResult, error) {
	output, err := s.runner.Execute(ctx, "search", query)
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
	seenSection := false

	lines := strings.Split(raw, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "==> Formulae") {
			isFormula = true
			seenSection = true
			continue
		}
		if strings.HasPrefix(line, "==> Casks") {
			isFormula = false
			seenSection = true
			continue
		}
		// Track the last seen section header; lines before any header are
		// UNKNOWN, not formulae — defaulting to formula miscategorized
		// entries in headerless/odd outputs (r1 finding 21).
		results = append(results, SearchResult{
			Name:      line,
			IsFormula: isFormula && seenSection,
			IsCask:    !isFormula && seenSection,
		})
	}
	return results
}
