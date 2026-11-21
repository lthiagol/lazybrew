package flows_test

import (
	"bytes"
	"context"
	"encoding/base64"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/x/exp/teatest/v2"
	"github.com/lthiagol/lazybrew/internal/brew"
	"github.com/lthiagol/lazybrew/internal/brew/brewtest"
	"github.com/lthiagol/lazybrew/internal/gui/testutil"
)

// TestVulnsReportFlow: v runs the check and the report from brew's JSON shows its
// summary and the package blocks; y copies the whole report, with the full commit
// SHAs of the fixes and the names of the packages not scanned.
func TestVulnsReportFlow(t *testing.T) {
	report, err := os.ReadFile("../../brew/testdata/vulns_v1.json")
	if err != nil {
		t.Fatal(err)
	}
	r := brewtest.NewMockRunner()
	r.ExecuteFn = func(ctx context.Context, args ...string) ([]byte, error) {
		if len(args) >= 1 && args[0] == "vulns" {
			// brew exits 1 when it found vulnerabilities, and prints the report.
			return report, &brew.BrewExitError{Command: "vulns", ExitCode: 1}
		}
		return []byte{}, nil
	}

	tm := testutil.NewTestModelWithRunner(t, r, teatest.WithInitialTermSize(140, 120))
	tm.Send(logsKeyPress('v'))
	waitForAllOnScreen(t, tm, "vulnerabilities in 7 of 90 packages", "libxml2 2.15.4", "OSV-2021-777")

	tm.Send(logsKeyPress('y'))
	copied := clipboardText(t, tm)
	for _, want := range []string{
		"ead311df1037e7994a6d86c7e862d71ad59c52c7", // a full fixed version of openssl@4
		"7 packages",                      // the summary is in the copy as well
		"glib, libedit",                   // the packages not scanned, by name
		"libssh2 patched: CVE-2025-15661", // the patched IDs
	} {
		if !strings.Contains(copied, want) {
			t.Errorf("the copy lacks %q:\n%s", want, copied)
		}
	}

	tm.Send(logsKeyPress('q'))
	tm.WaitFinished(t, teatest.WithFinalTimeout(3*time.Second))
}

// waitForAllOnScreen waits until the output holds every text. A wait reads what it
// waits on, so texts drawn in one frame are waited for together, not one by one.
func waitForAllOnScreen(t *testing.T, tm *teatest.TestModel, texts ...string) {
	t.Helper()
	teatest.WaitFor(t, tm.Output(), func(bts []byte) bool {
		for _, text := range texts {
			if !strings.Contains(string(bts), text) {
				return false
			}
		}
		return true
	}, teatest.WithDuration(3*time.Second), teatest.WithCheckInterval(10*time.Millisecond))
}

// clipboardText waits for the terminal clipboard sequence (OSC 52) that y writes,
// and returns the text it carries.
func clipboardText(t *testing.T, tm *teatest.TestModel) string {
	t.Helper()
	open := []byte("\x1b]52;c;")
	var seen []byte
	teatest.WaitFor(t, tm.Output(), func(bts []byte) bool {
		seen = bts
		return bytes.Contains(bts, open)
	}, teatest.WithDuration(3*time.Second), teatest.WithCheckInterval(10*time.Millisecond))

	payload := seen[bytes.LastIndex(seen, open)+len(open):]
	if end := bytes.IndexAny(payload, "\x07\x1b"); end >= 0 {
		payload = payload[:end]
	}
	text, err := base64.StdEncoding.DecodeString(string(payload))
	if err != nil {
		t.Fatalf("the clipboard payload is not base64: %v", err)
	}
	return string(text)
}
