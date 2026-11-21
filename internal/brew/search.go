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

// Search asks brew for the formulae and the casks that match the query, in two
// reads: `brew search --formula` and `brew search --cask`. Without a TTY brew
// prints no section headers, so a single read cannot tell the two apart; each
// read's flag says the type of its rows instead.
func (s *searchService) Search(ctx context.Context, query string) ([]SearchResult, error) {
	formulae, err := s.searchType(ctx, "--formula", query)
	if err != nil {
		return nil, err
	}
	casks, err := s.searchType(ctx, "--cask", query)
	if err != nil {
		return nil, err
	}
	return append(formulae, casks...), nil
}

// searchType runs one typed read and types every row it returns by the flag.
func (s *searchService) searchType(ctx context.Context, flag, query string) ([]SearchResult, error) {
	output, err := s.runner.Execute(ctx, "search", flag, query)
	if err != nil {
		if IsExitCode(err, 1) {
			return nil, nil
		}
		return nil, err
	}

	results := parseSearchOutput(string(output))
	for i := range results {
		results[i].IsFormula = flag == "--formula"
		results[i].IsCask = flag == "--cask"
	}
	return results, nil
}

// MarkInstalled sets Installed on each result that matches an installed
// formula (for a formula row) or an installed cask (for a cask row). A
// tap-qualified name, such as homebrew/cask/claude-code, matches on its last
// part. A nil list leaves no row of its type installed; a caller whose list
// could not be read must keep those rows' marks itself.
func MarkInstalled(results []SearchResult, formulae []Formula, casks []Cask) []SearchResult {
	installedFormulae := make(map[string]bool, len(formulae))
	for _, f := range formulae {
		installedFormulae[f.Name] = true
	}
	installedCasks := make(map[string]bool, len(casks))
	for _, c := range casks {
		installedCasks[c.Name] = true
	}
	for i, r := range results {
		name := r.Name[strings.LastIndex(r.Name, "/")+1:]
		switch {
		case r.IsFormula:
			results[i].Installed = installedFormulae[name]
		case r.IsCask:
			results[i].Installed = installedCasks[name]
		}
	}
	return results
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
