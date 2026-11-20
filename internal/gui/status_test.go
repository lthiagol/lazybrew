package gui

import (
	"regexp"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
)

var (
	statusRowFormulae = regexp.MustCompile(`Formulae +[0-9]`)
	statusRowCasks    = regexp.MustCompile(`Casks +[0-9]`)
)

// dashboardSummary is a dashboard load where every count was read, except the
// ones a test leaves unknown.
func dashboardSummary() statusSummary {
	return statusSummary{
		Version:  "4.5.2",
		Prefix:   "/opt/homebrew",
		Formulae: knownCount{N: 125, Known: true},
		Casks:    knownCount{N: 2, Known: true},
		Taps:     knownCount{N: 3, Known: true},
		Services: knownCount{N: 3, Known: true},
	}
}

// TestHeaderShowsCountsOnFirstLaunch (AC-03.4): a fresh model shows the
// formulae and casks counts from the dashboard load, before those panels open.
func TestHeaderShowsCountsOnFirstLaunch(t *testing.T) {
	m := updateModel(newTestModel(), tea.WindowSizeMsg{Width: 140, Height: 40})
	m = updateModel(m, DataLoadedMsg{PanelID: PanelStatus, Summary: dashboardSummary(),
		Items: statusRows(dashboardSummary())})

	header := plainLine(m.renderHeader())
	if !strings.Contains(header, "125 formulae") || !strings.Contains(header, "2 casks") {
		t.Fatalf("the header should show the dashboard's counts on first launch, got %q", header)
	}
	if strings.Contains(header, "outdated") {
		t.Fatalf("an unknown outdated count must be left out, got %q", header)
	}
}

// TestUnknownCountIsLeftOutNotZero (AC-03.4): a count the dashboard could not
// read shows nowhere: not in the header, not in the Status sidebar.
func TestUnknownCountIsLeftOutNotZero(t *testing.T) {
	s := dashboardSummary()
	s.Formulae = knownCount{N: 0, Known: false}
	m := updateModel(newTestModel(), tea.WindowSizeMsg{Width: 140, Height: 40})
	m = updateModel(m, DataLoadedMsg{PanelID: PanelStatus, Summary: s, Items: statusRows(s)})

	if header := plainLine(m.renderHeader()); strings.Contains(header, "formulae") {
		t.Fatalf("an unknown formulae count must not be in the header, got %q", header)
	}
	// The Formulae box always has its title; a Status row is the label and a number.
	side := sgr.ReplaceAllString(m.renderSidebar(), "")
	if statusRowFormulae.MatchString(side) {
		t.Fatalf("an unknown formulae count must not be a Status row:\n%s", side)
	}
	if !statusRowCasks.MatchString(side) {
		t.Fatalf("a known casks count should be a Status row:\n%s", side)
	}
}

// TestStatusRowsAreCompactSummary (AC-03.4): the Status sidebar lists the
// version, the prefix and the known counts, in that order.
func TestStatusRowsAreCompactSummary(t *testing.T) {
	rows := statusRows(dashboardSummary())
	want := []string{
		"Homebrew  4.5.2",
		"Prefix    /opt/homebrew",
		"Formulae  125",
		"Casks     2",
		"Taps      3",
		"Services  3",
	}
	if strings.Join(rows, "\n") != strings.Join(want, "\n") {
		t.Fatalf("statusRows =\n%s\nwant\n%s", strings.Join(rows, "\n"), strings.Join(want, "\n"))
	}
}

// TestStatusDashboardTabShowsTheFullDashboard (AC-03.1): the Dashboard tab
// draws its five cards from the typed summary, not as a list. (v2-05 AC-03:
// the expectation moved from the old text block's lines to the cards.)
func TestStatusDashboardTabShowsTheFullDashboard(t *testing.T) {
	m := updateModel(newTestModel(), tea.WindowSizeMsg{Width: 140, Height: 40})
	m = updateModel(m, DataLoadedMsg{PanelID: PanelStatus, Summary: dashboardSummary(),
		Items: statusRows(dashboardSummary())})

	out := ansi.Strip(m.renderContent(140, 30))
	for _, line := range []string{"Homebrew", "Packages", "Taps", "Services", "Health",
		"Formulae:  125 installed", "not loaded — press R", "press d to run doctor"} {
		if !strings.Contains(out, line) {
			t.Fatalf("the Dashboard tab should draw %q, got:\n%s", line, out)
		}
	}
}
