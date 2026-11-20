package gui

import "fmt"

// statusSummary is what the status dashboard read, typed. The dashboard load
// carries it into the model, so the Status sidebar and the header show counts
// before the Formulae, Casks and other panels are opened (v2-03b AC-03.4).
type statusSummary struct {
	Version  string
	Prefix   string
	Formulae knownCount
	Casks    knownCount
	Outdated knownCount
	Taps     knownCount
	Services knownCount
	// OfficialTaps and ServicesStarted split the Taps and Services counts.
	OfficialTaps    int
	ServicesStarted int
	// Doctor counts the warnings of the last brew doctor run; it is unknown
	// until that run has answered.
	Doctor knownCount
	// Problems are the sources that failed or timed out, as ⚠ lines.
	Problems []string
}

// knownCount is a count and whether its source answered. An unknown count is
// left out of the sidebar and the header; it is never shown as 0.
type knownCount struct {
	N     int
	Known bool
}

// statusItemName keys the Status tab content. The Status tabs read the whole
// dashboard, so they do not depend on which summary row is selected.
const statusItemName = "dashboard"

// statusRows is the Status sidebar: one compact row per known fact, the
// version and prefix first, then the counts. A row whose count is not known is
// left out.
func statusRows(s statusSummary) []string {
	var rows []string
	if s.Version != "" {
		rows = append(rows, fmt.Sprintf("%-10s%s", "Homebrew", s.Version))
	}
	if s.Prefix != "" {
		rows = append(rows, fmt.Sprintf("%-10s%s", "Prefix", s.Prefix))
	}
	counts := []struct {
		label string
		c     knownCount
	}{
		{"Formulae", s.Formulae},
		{"Casks", s.Casks},
		{"Outdated", s.Outdated},
		{"Taps", s.Taps},
		{"Services", s.Services},
	}
	for _, row := range counts {
		if row.c.Known {
			rows = append(rows, fmt.Sprintf("%-10s%d", row.label, row.c.N))
		}
	}
	return rows
}
