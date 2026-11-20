package gui

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/lthiagol/lazybrew/internal/brew"
)

// TestRemovalDetail locks AC-02.3 for uninstall and zap: the subtle line says
// what goes and how many packages depend on it, and a timed-out check says the
// count is unknown instead of zero.
func TestRemovalDetail(t *testing.T) {
	timeout := &brew.TimeoutError{Command: "uses", Timeout: 30 * time.Second}
	cases := []struct {
		name string
		zap  bool
		uses []string
		err  error
		want string
	}{
		{"no dependents", false, nil, nil, "removes bash"},
		{"one dependent", false, []string{"direnv"}, nil, "removes bash; 1 package depends on it"},
		{"two dependents", false, []string{"direnv", "bat"}, nil, "removes bash; 2 packages depend on it"},
		{"zap", true, nil, nil, "removes bash and all its files and preferences"},
		{"timed out", false, nil, timeout, "removes bash; dependents unknown (check timed out)"},
		{"other error", false, nil, fmt.Errorf("boom"), "removes bash"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := removalDetail("bash", tc.zap, tc.uses, tc.err); got != tc.want {
				t.Fatalf("removalDetail = %q, want %q", got, tc.want)
			}
		})
	}
}

// TestUntapConfirmShowsWarning locks AC-02.3 for untap: the existing warning
// text is the subtle line under the question.
func TestUntapConfirmShowsWarning(t *testing.T) {
	m := newTestModel()
	m.activePanel = PanelTaps
	m.panels[PanelTaps].items = []string{"user/tap  3 formulae"}
	m.panels[PanelTaps].selected = 0

	nm, _ := m.confirmUntap()
	view := sgr.ReplaceAllString(nm.(Model).activeModal.View().Content, "")
	// The box border is a column of ┃ on wrapped rows; it is not text.
	text := strings.Join(strings.Fields(strings.ReplaceAll(view, "┃", " ")), " ")
	for _, want := range []string{"Remove tap user/tap?", "Installed formulae from this tap will become unavailable for updates."} {
		if !strings.Contains(text, want) {
			t.Errorf("untap dialog lacks %q:\n%s", want, view)
		}
	}
}
