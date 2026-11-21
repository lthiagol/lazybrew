package gui

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"sync"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/lthiagol/lazybrew/internal/brew"
	"github.com/lthiagol/lazybrew/internal/brew/brewtest"
	"github.com/lthiagol/lazybrew/internal/gui/modal"
)

// trustFixture is `brew trust --json=v1` with one tap, formulae, a cask and a
// command. octo/tools is a tap, so it is a tap row only, not an item row.
const trustFixture = `{"taps":["octo/tools"],"formulae":["octo/tools/zeta","acme/fonts/alpha"],"casks":["acme/fonts/gamma-app"],"commands":["octo/tools/cmd"]}`

// trustedFixtureItems is the Taps list the fixture makes: the two taps, then the
// items in brew's kind order, alphabetical within each kind.
var trustedFixtureItems = []string{
	"octo/tools", "acme/fonts",
	"acme/fonts/alpha", "octo/tools/zeta",
	"acme/fonts/gamma-app",
	"octo/tools/cmd",
}

// callMu guards the call logs of the helpers below: the Taps read runs its taps
// and trust reads at the same time, so both append to a log at once.
var callMu sync.Mutex

// tapsReadModel is the Taps panel read through fetchPanelData, so the test covers
// the real read: `brew tap`, `brew tap-info` and `brew trust --json=v1`, answered by
// the mock. tapInfo is the tap-info JSON; trust is what the trust read returns.
// Every brew call is recorded in calls.
func tapsReadModel(t *testing.T, tapInfo, trust string, trustErr error, calls *[][]string) *Model {
	t.Helper()
	r := brewtest.NewMockRunner()
	r.ExecuteFn = func(ctx context.Context, args ...string) ([]byte, error) {
		callMu.Lock()
		*calls = append(*calls, args)
		callMu.Unlock()
		switch args[0] {
		case "tap":
			return []byte("octo/tools\nacme/fonts"), nil
		case "tap-info":
			if tapInfo == "" {
				return nil, errors.New("tap-info timed out")
			}
			return []byte(tapInfo), nil
		case "trust":
			if len(args) == 2 && args[1] == "--json=v1" {
				return []byte(trust), trustErr
			}
			return []byte(""), nil
		}
		return []byte(""), nil
	}
	m := updateModel(New(brew.NewClient(r), testConfig()), tea.WindowSizeMsg{Width: 140, Height: 40})
	m.switchPanel(PanelTaps)
	return updateModel(m, fetchPanelData(m.client, PanelTaps)())
}

// tapsTrustedModel is the Taps panel loaded from the given taps and trusted
// entries, with the rows the fetch would build. It is for tests that need a
// reload with a different list.
func tapsTrustedModel(t *testing.T, taps []brew.Tap, trusted []brew.TrustEntry) (*Model, *[][]string) {
	t.Helper()
	calls := &[][]string{}
	m := updateModel(newTestModel(), tea.WindowSizeMsg{Width: 140, Height: 40})
	m.client = brew.NewClient(recordingRunner(calls))
	m.switchPanel(PanelTaps)
	m = updateModel(m, tapsMsg(taps, trusted))
	return m, calls
}

// tapsMsg is the DataLoadedMsg the Taps read lands: tap names, then the trusted
// rows, with the tap-info of each tap in taps.
func tapsMsg(taps []brew.Tap, trusted []brew.TrustEntry) DataLoadedMsg {
	rows := trustedRows(trusted)
	items := make([]string, 0, len(taps)+len(rows))
	for _, tap := range taps {
		items = append(items, tap.Name)
	}
	for _, e := range rows {
		items = append(items, e.Name)
	}
	return DataLoadedMsg{PanelID: PanelTaps, Items: items, Taps: taps, Trusted: rows}
}

// recordingRunner answers every brew call with nothing and records its args.
func recordingRunner(calls *[][]string) *brewtest.MockRunner {
	r := brewtest.NewMockRunner()
	r.ExecuteFn = func(ctx context.Context, args ...string) ([]byte, error) {
		callMu.Lock()
		*calls = append(*calls, args)
		callMu.Unlock()
		return []byte(""), nil
	}
	return r
}

func trustedAt(name string, typ brew.TrustType) brew.TrustEntry {
	return brew.TrustEntry{Name: name, Type: typ, Tap: tapOfName(name)}
}

// tapOfName is the tap part of a full name, for the test fixtures.
func tapOfName(name string) string {
	if i := strings.LastIndex(name, "/"); i > 0 && strings.Count(name, "/") == 2 {
		return name[:i]
	}
	return ""
}

// selectRow moves the Taps cursor to the row called name, with j presses as a user
// does, so the tab bar follows the cursor through Update.
func selectRow(t *testing.T, m *Model, name string) *Model {
	t.Helper()
	for range m.panels[PanelTaps].items {
		if m.panels[PanelTaps].selectedName() == name {
			return m
		}
		m = sendKey(m, "j")
	}
	t.Fatalf("no Taps row called %q in %v", name, m.panels[PanelTaps].items)
	return m
}

func TestTapsPanelReadListsTrustedItemsBelowTheTaps(t *testing.T) {
	var calls [][]string
	m := tapsReadModel(t, `[{"name":"acme/fonts","trusted":true}]`, trustFixture, nil, &calls)

	if got := m.panels[PanelTaps].items; !reflect.DeepEqual(got, trustedFixtureItems) {
		t.Fatalf("Taps rows = %v, want %v", got, trustedFixtureItems)
	}
	if out := ansi.Strip(m.View().Content); !strings.Contains(out, "Taps & Trust") {
		t.Fatalf("the Taps box should be titled Taps & Trust:\n%s", out)
	}
	if !hasCall(calls, "trust", "--json=v1") {
		t.Fatalf("the Taps read should read the trust list, calls = %v", calls)
	}
}

func TestTapsReadFailureKeepsTheTapsAndToastsOnce(t *testing.T) {
	var calls [][]string
	m := tapsReadModel(t, `[]`, "", errors.New("trust read failed"), &calls)

	p := m.panels[PanelTaps]
	if want := []string{"octo/tools", "acme/fonts"}; !reflect.DeepEqual(p.items, want) {
		t.Fatalf("with the trust read failed, Taps rows = %v, want the taps only %v", p.items, want)
	}
	if len(m.toasts) != 1 || !strings.Contains(m.toasts[0].Message, "trust") {
		t.Fatalf("expected one toast naming the trust read, got %+v", m.toasts)
	}

	// A refresh that fails the same way says nothing new.
	m = updateModel(m, DataLoadedMsg{PanelID: PanelTaps, Items: []string{"octo/tools"}, Taps: []brew.Tap{{Name: "octo/tools"}}, TrustErr: errors.New("again")})
	if len(m.toasts) != 1 {
		t.Fatalf("a repeated trust failure toasted again: %d toasts", len(m.toasts))
	}
}

func TestTrustedRowsKeepKindOrderThenName(t *testing.T) {
	got := trustedRows([]brew.TrustEntry{
		trustedAt("octo/tools/cmd", brew.TrustTypeCommand),
		{Name: "octo/tools", Type: brew.TrustTypeTap, Tap: "octo/tools"},
		trustedAt("octo/tools/zeta", brew.TrustTypeFormula),
		trustedAt("acme/fonts/gamma-app", brew.TrustTypeCask),
		trustedAt("acme/fonts/alpha", brew.TrustTypeFormula),
	})
	want := []brew.TrustEntry{
		trustedAt("acme/fonts/alpha", brew.TrustTypeFormula),
		trustedAt("octo/tools/zeta", brew.TrustTypeFormula),
		trustedAt("acme/fonts/gamma-app", brew.TrustTypeCask),
		trustedAt("octo/tools/cmd", brew.TrustTypeCommand),
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("trustedRows = %+v\nwant %+v", got, want)
	}
}

func TestTrustedItemRowHasTheInfoTabAloneAndItsFacts(t *testing.T) {
	var calls [][]string
	m := tapsReadModel(t, `[{"name":"acme/fonts","trusted":true}]`, trustFixture, nil, &calls)
	m = selectRow(t, m, "acme/fonts/alpha")

	if len(m.tabs) != 1 || m.tabs[0].name != "Info" {
		t.Fatalf("a trusted item's tab bar = %+v, want the Info tab alone", m.tabs)
	}
	m = sendKey(m, "h")
	m = sendKey(m, "l")
	if len(m.tabs) != 1 || m.activeTab != 0 {
		t.Fatalf("h and l on a trusted item moved the tabs: tabs %d, active %d", len(m.tabs), m.activeTab)
	}
	out := ansi.Strip(m.View().Content)
	for _, want := range []string{"acme/fonts/alpha", "formula", "acme/fonts", "Tap trusted", "yes", "Trusted on its own: Homebrew may load it even if its tap is not trusted"} {
		if !strings.Contains(out, want) {
			t.Errorf("Info for the trusted item lacks %q:\n%s", want, out)
		}
	}

	m = sendKey(m, "k")
	if len(m.tabs) != 3 {
		t.Fatalf("back on a tap row the tab bar has %d tabs, want 3", len(m.tabs))
	}
}

func TestTrustedItemInfoSaysUnknownTapTrustWhenTapInfoFailed(t *testing.T) {
	var calls [][]string
	m := tapsReadModel(t, "", trustFixture, nil, &calls)
	m = selectRow(t, m, "acme/fonts/alpha")
	if out := ansi.Strip(m.View().Content); !strings.Contains(out, "Tap trusted") || !strings.Contains(out, "unknown") {
		t.Fatalf("a tap whose trust is not read should say unknown:\n%s", out)
	}
}

func TestTrustedItemCursorStaysOnItsNameAcrossReloads(t *testing.T) {
	taps := []brew.Tap{{Name: "octo/tools"}, {Name: "acme/fonts"}}
	m, _ := tapsTrustedModel(t, taps, []brew.TrustEntry{
		trustedAt("acme/fonts/alpha", brew.TrustTypeFormula),
		trustedAt("octo/tools/zeta", brew.TrustTypeFormula),
	})
	m = selectRow(t, m, "octo/tools/zeta")

	// A new formula sorts ahead of zeta, so the row moves down one.
	m = updateModel(m, tapsMsg(taps, []brew.TrustEntry{
		trustedAt("acme/fonts/alpha", brew.TrustTypeFormula),
		trustedAt("acme/fonts/beta", brew.TrustTypeFormula),
		trustedAt("octo/tools/zeta", brew.TrustTypeFormula),
	}))
	if got := m.panels[PanelTaps].selectedName(); got != "octo/tools/zeta" {
		t.Fatalf("after a reload the cursor is on %q, want octo/tools/zeta", got)
	}
}

func TestTrustedItemCursorClampsWhenItsRowIsGone(t *testing.T) {
	taps := []brew.Tap{{Name: "octo/tools"}}
	m, _ := tapsTrustedModel(t, taps, []brew.TrustEntry{
		trustedAt("acme/fonts/alpha", brew.TrustTypeFormula),
		trustedAt("octo/tools/zeta", brew.TrustTypeFormula),
	})
	m = selectRow(t, m, "octo/tools/zeta")

	m = updateModel(m, tapsMsg(taps, []brew.TrustEntry{trustedAt("acme/fonts/alpha", brew.TrustTypeFormula)}))
	p := m.panels[PanelTaps]
	if p.selected < 0 || p.selected >= len(p.items) {
		t.Fatalf("cursor %d is outside the %d rows", p.selected, len(p.items))
	}
	if p.selectedTrusted() == nil && p.selectedTap() == nil {
		t.Fatal("cursor is on no row after the item went")
	}
}

func TestTKeyOnTrustedItemAsksThenUntrustsItsKind(t *testing.T) {
	cases := []struct {
		name string
		typ  brew.TrustType
		want []string
	}{
		{"acme/fonts/alpha", brew.TrustTypeFormula, []string{"untrust", "--formula", "acme/fonts/alpha"}},
		{"acme/fonts/gamma-app", brew.TrustTypeCask, []string{"untrust", "--cask", "acme/fonts/gamma-app"}},
		{"octo/tools/cmd", brew.TrustTypeCommand, []string{"untrust", "--command", "octo/tools/cmd"}},
	}
	for _, tc := range cases {
		t.Run(string(tc.typ), func(t *testing.T) {
			m, calls := tapsTrustedModel(t, []brew.Tap{{Name: "octo/tools"}}, []brew.TrustEntry{trustedAt(tc.name, tc.typ)})
			m = selectRow(t, m, tc.name)

			m, cmd := pressKeyOn(m, "t")
			if cmd != nil {
				if _, ok := cmd().(TaskStartedMsg); ok {
					t.Fatal("t on a trusted item started a task before the confirm")
				}
			}
			c, ok := m.activeModal.(*modal.ConfirmModal)
			if !ok {
				t.Fatalf("t on a trusted item should open the untrust confirm, got %T", m.activeModal)
			}
			if want := "Homebrew will not load " + tc.name + " until it is trusted again"; !strings.Contains(c.Detail(), want) {
				t.Errorf("confirm detail = %q, want it to say %q", c.Detail(), want)
			}
			if len(*calls) != 0 {
				t.Fatalf("the confirm ran brew before it was answered: %v", *calls)
			}

			confirmAndRun(t, m)
			if !hasCall(*calls, tc.want...) {
				t.Fatalf("a confirmed untrust should run `brew %s`, got %v", strings.Join(tc.want, " "), *calls)
			}
		})
	}
}

func TestTKeyOnTrustedItemIsNotBlockedByUnknownTapState(t *testing.T) {
	m, _ := tapsTrustedModel(t, []brew.Tap{{Name: "octo/tools"}}, []brew.TrustEntry{trustedAt("octo/tools/zeta", brew.TrustTypeFormula)})
	m = selectRow(t, m, "octo/tools/zeta")
	m.panels[PanelTaps].trustUnknown = true

	m, _ = pressKeyOn(m, "t")
	if _, ok := m.activeModal.(*modal.ConfirmModal); !ok {
		t.Fatalf("t on a trusted item should not depend on the tap-info read; got %T", m.activeModal)
	}
}

func TestXAndRRefuseTrustedItemsWithAToastNamingT(t *testing.T) {
	m, calls := tapsTrustedModel(t, []brew.Tap{{Name: "octo/tools"}}, []brew.TrustEntry{trustedAt("octo/tools/zeta", brew.TrustTypeFormula)})
	m = selectRow(t, m, "octo/tools/zeta")

	for _, key := range []string{"x", "r"} {
		var cmd tea.Cmd
		m, cmd = pressKeyOn(m, key)
		if m.activeModal != nil {
			t.Errorf("%s on a trusted item opened %T", key, m.activeModal)
		}
		if cmd != nil {
			if _, ok := cmd().(TaskStartedMsg); ok {
				t.Errorf("%s on a trusted item started a task", key)
			}
		}
		if len(m.toasts) == 0 || !strings.Contains(m.toasts[len(m.toasts)-1].Message, "t untrusts") {
			t.Errorf("%s on a trusted item: toast %+v, want one naming t", key, m.toasts)
		}
	}
	if len(*calls) != 0 {
		t.Fatalf("x and r on a trusted item ran brew: %v", *calls)
	}
}

func TestTAndAOnTrustedItem(t *testing.T) {
	m, calls := tapsTrustedModel(t, []brew.Tap{{Name: "octo/tools"}}, []brew.TrustEntry{trustedAt("octo/tools/zeta", brew.TrustTypeFormula)})
	m = selectRow(t, m, "octo/tools/zeta")

	m, _ = pressKeyOn(m, "T")
	if m.activeModal != nil {
		t.Fatalf("T on a trusted item opened %T; it needs a tap", m.activeModal)
	}
	if len(m.toasts) == 0 || !strings.Contains(m.toasts[len(m.toasts)-1].Message, "tap") {
		t.Fatalf("T on a trusted item should say it needs a tap, toasts %+v", m.toasts)
	}

	m, _ = pressKeyOn(m, "a")
	if _, ok := m.activeModal.(*modal.InputModal); !ok {
		t.Fatalf("a on a trusted item should open Add tap, got %T", m.activeModal)
	}
	if len(*calls) != 0 {
		t.Fatalf("T and a on a trusted item ran brew: %v", *calls)
	}
}

// TestTapsFooterRowOnTrustedItemLeadsWithUntrust (lb30 r2 F-r1-01): the footer row,
// the last row of View, leads its panel group with t untrust and then a and y. No
// x, r or T hint is on it: x and r are refused on an item and T needs a tap.
func TestTapsFooterRowOnTrustedItemLeadsWithUntrust(t *testing.T) {
	m, _ := tapsTrustedModel(t, []brew.Tap{{Name: "octo/tools"}}, []brew.TrustEntry{trustedAt("octo/tools/zeta", brew.TrustTypeFormula)})
	m = selectRow(t, m, "octo/tools/zeta")

	rows := strings.Split(ansi.Strip(m.View().Content), "\n")
	panel, _, _ := strings.Cut(rows[len(rows)-1], "│")
	if got, want := strings.TrimSpace(panel), "t untrust · a add tap · y copy"; got != want {
		t.Fatalf("the footer row's panel group on a trusted item = %q, want %q", got, want)
	}
}

func TestTapsRowsFitAtEightyByTwentyFour(t *testing.T) {
	long := "acme/fonts/" + strings.Repeat("very-long-formula-", 4) + "end"
	m, _ := tapsTrustedModel(t, []brew.Tap{{Name: "octo/tools"}}, []brew.TrustEntry{trustedAt(long, brew.TrustTypeFormula)})
	m = updateModel(m, tea.WindowSizeMsg{Width: 80, Height: 24})

	out := ansi.Strip(m.View().Content)
	lines := strings.Split(out, "\n")
	if got := len(lines); got != 24 {
		t.Fatalf("the frame is %d rows at 80x24", got)
	}
	var title, row string
	for _, line := range lines {
		if strings.Contains(line, "Taps & Trust") {
			title = line
		}
		if strings.Contains(line, "…") && strings.Contains(line, "formula") {
			row = line
		}
	}
	if title == "" {
		t.Fatalf("the Taps title does not fit at 80x24:\n%s", out)
	}
	if row == "" {
		t.Fatalf("the long item row should be cut with … and keep its type:\n%s", out)
	}
}
