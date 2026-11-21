package gui

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/lthiagol/lazybrew/internal/brew"
	"github.com/lthiagol/lazybrew/internal/brew/brewtest"
)

// searchRunner answers the two typed search reads with the headerless rows
// brew prints without a TTY, and the installed lists with installedJSON. A nil
// installedErr is a readable list; otherwise the installed reads fail with it.
func searchRunner(installedErr error) *brewtest.MockRunner {
	r := brewtest.NewMockRunner()
	r.ExecuteFn = func(ctx context.Context, args ...string) ([]byte, error) {
		if len(args) >= 2 && args[0] == "search" {
			if args[1] == "--formula" {
				return []byte("claude-code-router\nclaude-code-templates\n"), nil
			}
			return []byte("claude-code\nclaude-code@latest\n"), nil
		}
		return []byte{}, nil
	}
	r.ExecuteJSONFn = func(ctx context.Context, result any, args ...string) error {
		if installedErr != nil {
			return installedErr
		}
		if len(args) > 0 && args[len(args)-1] == "--cask" {
			return jsonInto(result, `{"casks": [{"token": "claude-code", "name": ["Claude Code"]}]}`)
		}
		return jsonInto(result, `{"formulae": []}`)
	}
	return r
}

// searchDone runs the search for query on a model over r and returns the msg.
func searchDone(t *testing.T, r *brewtest.MockRunner, query string) SearchDoneMsg {
	t.Helper()
	m := newTestModel()
	m.client = brew.NewClient(r)
	msg := awaitMsg(t, m.executeSearch(query), 5*time.Second)
	done, ok := msg.(SearchDoneMsg)
	if !ok {
		t.Fatalf("expected SearchDoneMsg, got %T", msg)
	}
	return done
}

// TestSearchRowsCarryTypeAndInstalledState (lb19 AC-02): each row names its
// type, a row for an installed cask says installed, and no row says unknown.
func TestSearchRowsCarryTypeAndInstalledState(t *testing.T) {
	done := searchDone(t, searchRunner(nil), "claude-code")
	if done.Err != nil {
		t.Fatal(done.Err)
	}
	want := map[string]string{
		"claude-code-router":    "claude-code-router  formula  ",
		"claude-code-templates": "claude-code-templates  formula  ",
		"claude-code":           "claude-code  cask  installed  ",
		"claude-code@latest":    "claude-code@latest  cask  ",
	}
	if len(done.Results) != len(want) {
		t.Fatalf("got %d rows, want %d: %q", len(done.Results), len(want), done.Results)
	}
	for _, row := range done.Results {
		name, _, _ := strings.Cut(row, "  ")
		if !strings.HasPrefix(row, want[name]) {
			t.Errorf("row %q, want prefix %q", row, want[name])
		}
		if strings.Contains(row, "unknown") {
			t.Errorf("row %q still says unknown", row)
		}
	}
}

// TestSearchInstalledListUnreadableKeepsTypes (lb19 AC-02 point 3): with the
// installed lists unreadable, the rows keep their type, carry no installed
// mark, and the search does not fail.
func TestSearchInstalledListUnreadableKeepsTypes(t *testing.T) {
	done := searchDone(t, searchRunner(brewtest.ErrTest), "claude-code")
	if done.Err != nil {
		t.Fatalf("search failed on an unreadable installed list: %v", done.Err)
	}
	for _, row := range done.Results {
		if strings.Contains(row, "installed") {
			t.Errorf("row %q marked installed with no installed list", row)
		}
	}
	if done.Raw[0].IsFormula != true || done.Raw[2].IsCask != true {
		t.Errorf("types lost with the installed list unread: %+v", done.Raw)
	}
}

// jsonInto decodes the JSON text into result, as the runner would.
func jsonInto(result any, text string) error {
	return json.Unmarshal([]byte(text), result)
}
