package gui

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/lthiagol/lazybrew/internal/brew"
	"github.com/lthiagol/lazybrew/internal/brew/brewtest"
)

// servicesRunner answers the two reads the Services panel makes: `services
// list` with list, and `services info` with info or infoErr.
func servicesRunner(list, info string, infoErr error) *brewtest.MockRunner {
	r := brewtest.NewMockRunner()
	r.ExecuteFn = func(ctx context.Context, args ...string) ([]byte, error) {
		switch strings.Join(args, " ") {
		case "services list --json":
			return []byte(list), nil
		case "services info --all --json":
			if infoErr != nil {
				return nil, infoErr
			}
			return []byte(info), nil
		}
		return nil, errors.New("unexpected brew call: " + strings.Join(args, " "))
	}
	return r
}

const servicesListJSON = `[
	{"name":"dbus","status":"none","user":"","file":"","exit_code":0},
	{"name":"moshi-hook","status":"started","user":"agent","file":"/x/moshi-hook.service","exit_code":0}
]`

const servicesInfoJSON = `[
	{"name":"dbus","service_name":"sh.brew.dbus","running":false,"loaded":false,"schedulable":false,
	 "pid":null,"exit_code":null,"user":null,"status":"none","file":"","registered":false,"loaded_file":null},
	{"name":"moshi-hook","service_name":"sh.brew.moshi-hook","running":true,"loaded":true,"schedulable":false,
	 "pid":800,"exit_code":null,"user":"agent","status":"started","file":"/x/moshi-hook.service","registered":true,
	 "loaded_file":"/x/moshi-hook.service; enabled","command":"moshi-hook serve","working_dir":null,
	 "root_dir":null,"log_path":"/var/log/moshi-hook.log","error_log_path":"/var/log/moshi-hook.log",
	 "interval":null,"cron":null}
]`

func TestServicesPanelLoadsInfoWithList(t *testing.T) {
	client := brew.NewClient(servicesRunner(servicesListJSON, servicesInfoJSON, nil))

	msg, ok := awaitMsg(t, fetchPanelData(client, PanelServices), 5*time.Second).(DataLoadedMsg)
	if !ok {
		t.Fatal("fetchPanelData(PanelServices) did not return a DataLoadedMsg")
	}
	if msg.Err != nil {
		t.Fatalf("Err = %v, want nil", msg.Err)
	}
	if len(msg.Items) != 2 || len(msg.Services) != 2 {
		t.Fatalf("list: %d items, %d services, want 2 and 2", len(msg.Items), len(msg.Services))
	}
	if len(msg.ServiceInfo) != 2 {
		t.Fatalf("ServiceInfo has %d entries, want 2", len(msg.ServiceInfo))
	}
	if msg.ServiceInfo[1].Name != "moshi-hook" || msg.ServiceInfo[1].PID == nil || *msg.ServiceInfo[1].PID != 800 {
		t.Errorf("ServiceInfo[1] = %+v, want moshi-hook with pid 800", msg.ServiceInfo[1])
	}
}

// A failed `services info` leaves the panel loaded from the list: the Info
// view falls back to the list's fields instead of showing an error.
func TestServicesPanelInfoFailureKeepsList(t *testing.T) {
	client := brew.NewClient(servicesRunner(servicesListJSON, "", errors.New("brew services info timed out")))

	msg, ok := awaitMsg(t, fetchPanelData(client, PanelServices), 5*time.Second).(DataLoadedMsg)
	if !ok {
		t.Fatal("fetchPanelData(PanelServices) did not return a DataLoadedMsg")
	}
	if msg.Err != nil {
		t.Fatalf("Err = %v, want nil: a failed info read must not fail the panel", msg.Err)
	}
	if len(msg.Services) != 2 || len(msg.ServiceInfo) != 0 {
		t.Errorf("services %d, info %d, want 2 and 0", len(msg.Services), len(msg.ServiceInfo))
	}
}
