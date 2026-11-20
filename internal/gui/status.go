package gui

import (
	"fmt"

	"github.com/lthiagol/lazybrew/internal/brew"
)

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
	// until that run has answered. DoctorErr is the reason its run failed.
	Doctor    knownCount
	DoctorErr string
	// Problems are the sources other than doctor that failed or timed out,
	// as ⚠ lines.
	Problems []string
}

// absorbPanel carries one panel's read into the summary, so its count shows on
// the dashboard as the panel lands, with no second run of the dashboard's reads.
// It counts what statusDashboard counts, from the panel's own data.
func (s *statusSummary) absorbPanel(id PanelID, msg DataLoadedMsg) {
	switch id {
	case PanelFormulae:
		s.Formulae = knownCount{N: len(msg.Formulae), Known: true}
	case PanelCasks:
		s.Casks = knownCount{N: len(msg.Casks), Known: true}
	case PanelOutdated:
		s.Outdated = knownCount{N: len(msg.Formulae) + len(msg.Casks), Known: true}
	case PanelTaps:
		s.Taps = knownCount{N: len(msg.Taps), Known: true}
		s.OfficialTaps = 0
		for _, t := range msg.Taps {
			if t.IsOfficial {
				s.OfficialTaps++
			}
		}
	case PanelServices:
		s.Services = knownCount{N: len(msg.Services), Known: true}
		s.ServicesStarted = 0
		for _, svc := range msg.Services {
			if svc.Status == brew.ServiceStarted {
				s.ServicesStarted++
			}
		}
	}
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
