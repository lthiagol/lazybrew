package gui

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/lthiagol/lazybrew/internal/brew"
	"github.com/lthiagol/lazybrew/internal/gui/style"
)

// TestDashboardCardsSideBySideWhenTheWidthAllowsIt (AC-03.1): two cards share a
// row when the main area is at least twice a card's width, and stack otherwise.
func TestDashboardCardsSideBySideWhenTheWidthAllowsIt(t *testing.T) {
	m := newTestModel()
	m.status = statusSummary{Formulae: knownCount{N: 1, Known: true}}
	wide := strings.Split(ansi.Strip(m.renderDashboard(2*dashboardCardWidth)), "\n")
	// Health is the second card, so the first row is Homebrew and Health (fix F-r1-02).
	if !strings.Contains(wide[0], "Homebrew") || !strings.Contains(wide[0], "Health") {
		t.Fatalf("at twice a card's width the first two cards should share a row, got %q", wide[0])
	}
	narrow := strings.Split(ansi.Strip(m.renderDashboard(2*dashboardCardWidth-1)), "\n")
	if strings.Contains(narrow[0], "Health") {
		t.Fatalf("below twice a card's width the cards should stack, got %q", narrow[0])
	}
	for _, row := range narrow {
		if w := len([]rune(row)); w > dashboardCardWidth {
			t.Fatalf("a stacked card is %d cells, want at most %d: %q", w, dashboardCardWidth, row)
		}
	}
}

// TestOutdatedGaugeShowsTheCurrentShare (AC-03.1): the solid run is the current
// share of the installed packages, the dashed tail the outdated ones; with
// nothing outdated the whole bar is solid.
func TestOutdatedGaugeShowsTheCurrentShare(t *testing.T) {
	if got := ansi.Strip(outdatedGauge(0, 10, 20)); got != strings.Repeat("━", 20) {
		t.Fatalf("nothing outdated should be a full solid bar, got %q", got)
	}
	got := ansi.Strip(outdatedGauge(2, 10, 20))
	if want := strings.Repeat("━", 16) + strings.Repeat("╌", 4); got != want {
		t.Fatalf("2 of 10 outdated: got %q, want %q", got, want)
	}
}

// TestOutdatedUnknownIsNotLoadedNeverZero: the Packages card says so when the
// outdated count has not loaded, and draws no gauge for it.
func TestOutdatedUnknownIsNotLoadedNeverZero(t *testing.T) {
	m := newTestModel()
	m.status = statusSummary{Formulae: knownCount{N: 3, Known: true}, Casks: knownCount{N: 1, Known: true}}
	out := ansi.Strip(m.packagesBody(34))
	if !strings.Contains(out, "Outdated:  not loaded — press R") {
		t.Fatalf("an unloaded outdated count should read not loaded, got:\n%s", out)
	}
	if strings.Contains(out, "━") || strings.Contains(out, "╌") {
		t.Fatalf("an unloaded count draws no gauge, got:\n%s", out)
	}
}

// TestHealthReadsTheDoctorSummary (AC-03.1): a clean run is ready to brew, a
// run with warnings says how many, and no run says to run it.
func TestHealthReadsTheDoctorSummary(t *testing.T) {
	m := newTestModel()
	m.icons = style.UnicodeIcons
	cases := []struct {
		name   string
		doctor knownCount
		want   string
	}{
		{"clean", knownCount{N: 0, Known: true}, "✓ ready to brew"},
		{"warnings", knownCount{N: 2, Known: true}, "⚠ 2 warnings"},
		{"one warning", knownCount{N: 1, Known: true}, "⚠ 1 warning"},
		{"never run", knownCount{}, "press d to run doctor"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			m.status = statusSummary{Doctor: c.doctor}
			if out := ansi.Strip(m.healthBody()); !strings.Contains(out, c.want) {
				t.Fatalf("want %q, got:\n%s", c.want, out)
			}
		})
	}
}

// TestServicesShowRunningAndStoppedWithIcons (AC-03.1).
func TestServicesShowRunningAndStoppedWithIcons(t *testing.T) {
	m := newTestModel()
	m.icons = style.UnicodeIcons
	m.status = statusSummary{Services: knownCount{N: 4, Known: true}, ServicesStarted: 1}
	out := ansi.Strip(m.servicesBody(34))
	for _, want := range []string{"▶ 1", "■ 3"} {
		if !strings.Contains(out, want) {
			t.Fatalf("want %q in:\n%s", want, out)
		}
	}
}

// TestDoctorTabBlocksEachWarning (AC-03.2): each warning is a heading line with
// the warning icon, its details below, and blocks are separated by a blank line.
func TestDoctorTabBlocksEachWarning(t *testing.T) {
	warnings := []brew.DoctorWarning{
		{Title: "Outdated Xcode", Details: "Update Xcode."},
		{Title: "Unlinked kegs", Details: "Run brew link."},
	}
	out := ansi.Strip(doctorView(warnings, style.UnicodeIcons))
	want := "⚠ Outdated Xcode\nUpdate Xcode.\n\n⚠ Unlinked kegs\nRun brew link."
	if out != want {
		t.Fatalf("got:\n%s\nwant:\n%s", out, want)
	}
	clean := ansi.Strip(doctorView(nil, style.UnicodeIcons))
	if clean != "✓ Your system is ready to brew." {
		t.Fatalf("a clean run should say so in Success, got %q", clean)
	}
}

// TestDashboardTabWaitsForTheLoad: the Dashboard tab says it is not loaded
// before the first dashboard load, and draws the cards after it.
func TestDashboardTabWaitsForTheLoad(t *testing.T) {
	m := updateModel(newTestModel(), tea.WindowSizeMsg{Width: 140, Height: 40})
	if out := ansi.Strip(m.renderContent(80, 20)); !strings.Contains(out, "not loaded yet") {
		t.Fatalf("before the load the tab should say so, got:\n%s", out)
	}
}

// dashboardCards is a status load as the user reads it: the cards, drawn plain
// at a width where every card shows (v2-05 fix F-r1-02: the checks that read
// the dashboard now read these cards, not a text block no user sees).
func dashboardCards(msg DataLoadedMsg) string {
	m := updateModel(newTestModel(), tea.WindowSizeMsg{Width: 140, Height: 40})
	m = updateModel(m, msg)
	return ansi.Strip(m.renderDashboard(140))
}

// healthWords is the Health card's text with its line breaks joined, so a line
// the card wraps still reads as one phrase.
func healthWords(msg DataLoadedMsg) string {
	m := newTestModel()
	m.status = msg.Summary
	return strings.Join(strings.Fields(ansi.Strip(m.healthBody())), " ")
}
