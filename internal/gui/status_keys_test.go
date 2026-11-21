package gui

import (
	"context"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/lthiagol/lazybrew/internal/brew"
	"github.com/lthiagol/lazybrew/internal/brew/brewtest"
)

// countingRunner answers every brew read with no output and counts each verb.
func countingRunner(calls map[string]int) *brewtest.MockRunner {
	r := brewtest.NewMockRunner()
	r.ExecuteFn = func(ctx context.Context, args ...string) ([]byte, error) {
		if len(args) >= 1 {
			calls[args[0]]++
		}
		return []byte{}, nil
	}
	return r
}

// TestStatusKeysOpenTheirTabs: d, m and v switch the Status panel to the Doctor,
// Missing and Vulns tabs and load them, with no Operation overlay. Only v runs
// brew vulns.
func TestStatusKeysOpenTheirTabs(t *testing.T) {
	cases := []struct {
		key  string
		tab  int
		verb string
	}{
		{"d", statusDoctorTab, "doctor"},
		{"m", statusMissingTab, "missing"},
		{"v", statusVulnsTab, "vulns"},
	}
	for _, c := range cases {
		t.Run(c.key, func(t *testing.T) {
			calls := map[string]int{}
			m := loadedStatus(140, 40)
			m.client = brew.NewClient(countingRunner(calls))

			nm, cmd := m.Update(tea.KeyPressMsg{Code: []rune(c.key)[0], Text: c.key})
			m = modelOf(nm)
			if m.activeTab != c.tab {
				t.Fatalf("%s: active tab %d, want %d", c.key, m.activeTab, c.tab)
			}
			if m.opState != nil {
				t.Fatalf("%s opened an Operation overlay", c.key)
			}
			m = landCmd(m, cmd)
			if calls[c.verb] != 1 {
				t.Fatalf("%s ran brew %s %d times, want 1", c.key, c.verb, calls[c.verb])
			}
			for _, other := range []string{"doctor", "missing", "vulns"} {
				if other != c.verb && calls[other] != 0 {
					t.Errorf("%s also ran brew %s", c.key, other)
				}
			}
			if m.opState != nil {
				t.Fatalf("%s left an Operation overlay open", c.key)
			}
		})
	}
}

// TestStatusDoctorKeyRefreshes: d on the Doctor tab runs brew doctor again, so
// the tab shows what brew says now rather than the cached read.
func TestStatusDoctorKeyRefreshes(t *testing.T) {
	calls := map[string]int{}
	m := loadedStatus(140, 40)
	m.client = brew.NewClient(countingRunner(calls))
	for range 2 {
		nm, cmd := m.Update(tea.KeyPressMsg{Code: 'd', Text: "d"})
		m = landCmd(modelOf(nm), cmd)
	}
	if calls["doctor"] != 2 {
		t.Fatalf("two d presses ran brew doctor %d times, want 2", calls["doctor"])
	}
}
