package gui

import (
	"context"
	"encoding/json"
	"slices"
	"testing"
	"time"

	"github.com/lthiagol/lazybrew/internal/brew"
	"github.com/lthiagol/lazybrew/internal/brew/brewtest"
	"github.com/lthiagol/lazybrew/internal/gui/style"
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

// TestSearchRowsCarryTypeAndInstalledState (lb19 AC-02, lb25 AC-02): each row
// names its type and installed state in its result, and its cells draw them:
// the mark for an installed row, the type word, and no row says unknown.
func TestSearchRowsCarryTypeAndInstalledState(t *testing.T) {
	done := searchDone(t, searchRunner(nil), "claude-code")
	if done.Err != nil {
		t.Fatal(done.Err)
	}
	want := map[string]struct {
		typ       string
		installed bool
	}{
		"claude-code-router":    {"formula", false},
		"claude-code-templates": {"formula", false},
		"claude-code":           {"cask", true},
		"claude-code@latest":    {"cask", false},
	}
	if len(done.Raw) != len(want) {
		t.Fatalf("got %d rows, want %d: %+v", len(done.Raw), len(want), done.Raw)
	}
	for i, r := range done.Raw {
		if done.Results[i] != r.Name {
			t.Errorf("item %d is %q, want the row's name %q", i, done.Results[i], r.Name)
		}
		w := want[r.Name]
		cells := searchCells(r, style.UnicodeIcons)
		if cells.right != w.typ {
			t.Errorf("row %q type = %q, want %q", r.Name, cells.right, w.typ)
		}
		if marked := cells.icon == style.UnicodeIcons.Installed; marked != w.installed {
			t.Errorf("row %q installed mark = %v, want %v", r.Name, marked, w.installed)
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
	for _, r := range done.Raw {
		if r.Installed {
			t.Errorf("row %q marked installed with no installed list", r.Name)
		}
	}
	if done.Raw[0].IsFormula != true || done.Raw[2].IsCask != true {
		t.Errorf("types lost with the installed list unread: %+v", done.Raw)
	}
}

// TestRemarkKeepsMarksOfUnreadableList (lb19 AC-02, F-r1-04): a re-mark whose
// installed list cannot be read leaves that type's marks as they were, so the
// row still installed stays installed. The other type is re-marked from its
// list as usual, and its list is empty here, so its row is unmarked.
func TestRemarkKeepsMarksOfUnreadableList(t *testing.T) {
	cases := []struct {
		unreadable                   string
		jqInstalled, claudeInstalled bool
	}{
		{unreadable: "formula", jqInstalled: true, claudeInstalled: false},
		{unreadable: "cask", jqInstalled: false, claudeInstalled: true},
	}
	for _, tc := range cases {
		mr := brewtest.NewMockRunner()
		mr.ExecuteJSONFn = func(ctx context.Context, result any, args ...string) error {
			isCask := slices.Contains(args, "--cask")
			if isCask == (tc.unreadable == "cask") {
				return brewtest.ErrTest
			}
			if isCask {
				return jsonInto(result, `{"casks": []}`)
			}
			return jsonInto(result, `{"formulae": []}`)
		}
		rows := []brew.SearchResult{
			{Name: "jq", IsFormula: true, Installed: true},
			{Name: "claude-code", IsCask: true, Installed: true},
		}
		m := searchRowsModel(&searchUninstallRunner{}, rows)
		m.client = brew.NewClient(mr)

		m = updateModel(m, m.remarkSearch()())

		if got := m.searchResults[0].Installed; got != tc.jqInstalled {
			t.Errorf("%s list unreadable: jq Installed = %v, want %v", tc.unreadable, got, tc.jqInstalled)
		}
		if got := m.searchResults[1].Installed; got != tc.claudeInstalled {
			t.Errorf("%s list unreadable: claude-code Installed = %v, want %v", tc.unreadable, got, tc.claudeInstalled)
		}
	}
}

// jsonInto decodes the JSON text into result, as the runner would.
func jsonInto(result any, text string) error {
	return json.Unmarshal([]byte(text), result)
}
