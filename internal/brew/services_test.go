package brew

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/lthiagol/lazybrew/internal/brew/brewtest"
)

func TestServicesServiceList(t *testing.T) {
	r := brewtest.NewMockRunner()
	r.ExecuteFn = func(ctx context.Context, args ...string) ([]byte, error) {
		return []byte(`[
			{"name":"postgresql@16","status":"started","user":"thiago","file":"~/Library/LaunchAgents/homebrew.mxcl.postgresql@16.plist","exit_code":0},
			{"name":"redis","status":"stopped","user":"","file":"","exit_code":0}
		]`), nil
	}
	cache := NewCache(time.Minute)
	svc := NewServicesReader(r, cache)

	list, err := svc.List(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	if len(list) != 2 {
		t.Fatalf("expected 2 services, got %d", len(list))
	}

	pg := list[0]
	if pg.Name != "postgresql@16" {
		t.Errorf("Name = %q, want postgresql@16", pg.Name)
	}
	if pg.Status != ServiceStarted {
		t.Errorf("Status = %q, want started", pg.Status)
	}

	redis := list[1]
	if redis.Name != "redis" {
		t.Errorf("Name = %q, want redis", redis.Name)
	}
	if redis.Status != ServiceStopped {
		t.Errorf("Status = %q, want stopped", redis.Status)
	}
}

func TestServicesServiceStart(t *testing.T) {
	r := brewtest.NewMockRunner()
	r.ExecuteFn = func(ctx context.Context, args ...string) ([]byte, error) {
		return []byte(""), nil
	}
	cache := NewCache(time.Minute)
	svc := NewServicesWriter(r, cache)

	if err := svc.Start(context.Background(), "redis"); err != nil {
		t.Fatal(err)
	}
}

func TestServicesServiceStop(t *testing.T) {
	r := brewtest.NewMockRunner()
	r.ExecuteFn = func(ctx context.Context, args ...string) ([]byte, error) {
		return []byte(""), nil
	}
	cache := NewCache(time.Minute)
	svc := NewServicesWriter(r, cache)

	if err := svc.Stop(context.Background(), "redis"); err != nil {
		t.Fatal(err)
	}
}

func TestServicesServiceRestart(t *testing.T) {
	r := brewtest.NewMockRunner()
	r.ExecuteFn = func(ctx context.Context, args ...string) ([]byte, error) {
		return []byte(""), nil
	}
	cache := NewCache(time.Minute)
	svc := NewServicesWriter(r, cache)

	if err := svc.Restart(context.Background(), "postgresql@16"); err != nil {
		t.Fatal(err)
	}
}

func TestServicesServiceRun(t *testing.T) {
	r := brewtest.NewMockRunner()
	r.ExecuteFn = func(ctx context.Context, args ...string) ([]byte, error) {
		return []byte(""), nil
	}
	cache := NewCache(time.Minute)
	cache.Set(KeyServicesList, []Service{{Name: "redis"}})
	svc := NewServicesWriter(r, cache)

	if err := svc.Run(context.Background(), "redis"); err != nil {
		t.Fatal(err)
	}

	// `services run` changes state like start/stop, so it must invalidate
	// the services cache like every other write (r1 finding 19).
	if _, ok := cache.Get(KeyServicesList); ok {
		t.Error("expected KeyServicesList to be invalidated by Run")
	}
}

func TestServicesServiceCleanupInvalidatesCache(t *testing.T) {
	r := brewtest.NewMockRunner()
	r.ExecuteFn = func(ctx context.Context, args ...string) ([]byte, error) {
		return []byte(""), nil
	}
	cache := NewCache(time.Minute)
	cache.Set(KeyServicesList, []Service{{Name: "redis"}})
	svc := NewServicesWriter(r, cache)

	if err := svc.Cleanup(context.Background()); err != nil {
		t.Fatal(err)
	}
	if _, ok := cache.Get(KeyServicesList); ok {
		t.Error("expected KeyServicesList to be invalidated by Cleanup")
	}
}

// servicesInfoFixture is the real `brew services info --all --json` output of
// this machine (testdata/services_info_all.json): a stopped service with nulls,
// and the running moshi-hook with a pid and a log path.
func servicesInfoFixture(t *testing.T) []byte {
	t.Helper()
	raw, err := os.ReadFile("testdata/services_info_all.json")
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

func TestServicesServiceInfoParsesRealShape(t *testing.T) {
	r := brewtest.NewMockRunner()
	var gotArgs []string
	r.ExecuteFn = func(ctx context.Context, args ...string) ([]byte, error) {
		gotArgs = args
		return servicesInfoFixture(t), nil
	}
	svc := NewServicesReader(r, NewCache(time.Minute))

	info, err := svc.Info(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if want := "services info --all --json"; strings.Join(gotArgs, " ") != want {
		t.Errorf("args = %q, want %q", strings.Join(gotArgs, " "), want)
	}
	if len(info) != 3 {
		t.Fatalf("expected 3 services, got %d", len(info))
	}

	dbus := info[0]
	if dbus.Name != "dbus" || dbus.Running || dbus.Loaded {
		t.Errorf("dbus: name %q running %v loaded %v, want stopped and not loaded", dbus.Name, dbus.Running, dbus.Loaded)
	}
	if dbus.PID != nil || dbus.ExitCode != nil {
		t.Errorf("dbus: pid and exit code must be nil (null in brew), got %v and %v", dbus.PID, dbus.ExitCode)
	}
	if dbus.User != "" || dbus.Status != ServiceNone || dbus.Interval != nil || dbus.Cron != nil {
		t.Errorf("dbus: user %q status %q, want empty user, none, and no schedule", dbus.User, dbus.Status)
	}

	moshi := info[1]
	if moshi.Name != "moshi-hook" || moshi.ServiceName != "sh.brew.moshi-hook" {
		t.Errorf("moshi-hook: name %q service_name %q", moshi.Name, moshi.ServiceName)
	}
	if !moshi.Running || !moshi.Loaded || !moshi.Registered || moshi.Schedulable {
		t.Errorf("moshi-hook: running %v loaded %v registered %v schedulable %v", moshi.Running, moshi.Loaded, moshi.Registered, moshi.Schedulable)
	}
	if moshi.PID == nil || *moshi.PID != 800 {
		t.Errorf("moshi-hook: PID = %v, want 800", moshi.PID)
	}
	if moshi.Status != ServiceStarted || moshi.User != "agent" {
		t.Errorf("moshi-hook: status %q user %q, want started and agent", moshi.Status, moshi.User)
	}
	if moshi.Command != "/home/linuxbrew/.linuxbrew/opt/moshi-hook/bin/moshi-hook serve" {
		t.Errorf("moshi-hook: command = %q", moshi.Command)
	}
	if moshi.LogPath != "/home/linuxbrew/.linuxbrew/var/log/moshi-hook.log" {
		t.Errorf("moshi-hook: log path = %q", moshi.LogPath)
	}
	if moshi.LoadedFile != "/home/agent/.config/systemd/user/sh.brew.moshi-hook.service; enabled" {
		t.Errorf("moshi-hook: loaded file = %q", moshi.LoadedFile)
	}
	if moshi.RootDir != "" || moshi.Interval != nil || moshi.Cron != nil {
		t.Errorf("moshi-hook: root dir %q, interval and cron must be empty (null in brew)", moshi.RootDir)
	}
}

func TestServicesServiceInfoParsesSchedule(t *testing.T) {
	r := brewtest.NewMockRunner()
	r.ExecuteFn = func(ctx context.Context, args ...string) ([]byte, error) {
		return []byte(`[
			{"name":"backup","service_name":"sh.brew.backup","running":false,"loaded":true,"schedulable":true,
			 "pid":null,"exit_code":0,"user":"agent","status":"scheduled","file":"","registered":true,
			 "loaded_file":null,"command":"/bin/backup","working_dir":null,"root_dir":null,
			 "log_path":null,"error_log_path":null,"interval":3600,"cron":{"hour":3,"minute":0}}
		]`), nil
	}
	svc := NewServicesReader(r, NewCache(time.Minute))

	info, err := svc.Info(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(info) != 1 {
		t.Fatalf("expected 1 service, got %d", len(info))
	}
	b := info[0]
	if b.Status != ServiceScheduled {
		t.Errorf("Status = %q, want scheduled", b.Status)
	}
	if b.ExitCode == nil || *b.ExitCode != 0 {
		t.Errorf("ExitCode = %v, want 0 (present, not null)", b.ExitCode)
	}
	if b.Interval == nil || *b.Interval != 3600 {
		t.Errorf("Interval = %v, want 3600", b.Interval)
	}
	if b.Cron["hour"] != float64(3) || b.Cron["minute"] != float64(0) {
		t.Errorf("Cron = %v, want hour 3 and minute 0", b.Cron)
	}
}

func TestServicesServiceInfoIsCached(t *testing.T) {
	r := brewtest.NewMockRunner()
	calls := 0
	r.ExecuteFn = func(ctx context.Context, args ...string) ([]byte, error) {
		calls++
		return servicesInfoFixture(t), nil
	}
	svc := NewServicesReader(r, NewCache(time.Minute))

	for i := 0; i < 2; i++ {
		if _, err := svc.Info(context.Background()); err != nil {
			t.Fatal(err)
		}
	}
	if calls != 1 {
		t.Errorf("brew ran %d times for two reads, want 1 (cache hit)", calls)
	}
}

func TestServicesWritesInvalidateServicesInfo(t *testing.T) {
	writes := map[string]func(w ServicesWriter) error{
		"start":   func(w ServicesWriter) error { return w.Start(context.Background(), "redis") },
		"stop":    func(w ServicesWriter) error { return w.Stop(context.Background(), "redis") },
		"restart": func(w ServicesWriter) error { return w.Restart(context.Background(), "redis") },
		"run":     func(w ServicesWriter) error { return w.Run(context.Background(), "redis") },
		"cleanup": func(w ServicesWriter) error { return w.Cleanup(context.Background()) },
	}
	for name, write := range writes {
		t.Run(name, func(t *testing.T) {
			r := brewtest.NewMockRunner()
			r.ExecuteFn = func(ctx context.Context, args ...string) ([]byte, error) {
				return []byte(""), nil
			}
			cache := NewCache(time.Minute)
			cache.Set(KeyServicesInfo, []ServiceInfo{{Name: "redis"}})
			cache.Set(KeyServicesList, []Service{{Name: "redis"}})

			if err := write(NewServicesWriter(r, cache)); err != nil {
				t.Fatal(err)
			}
			if _, ok := cache.Get(KeyServicesInfo); ok {
				t.Error("expected KeyServicesInfo to be invalidated")
			}
			if _, ok := cache.Get(KeyServicesList); ok {
				t.Error("expected KeyServicesList to be invalidated")
			}
		})
	}
}

// A services write that brew refuses for want of root names the sudo command
// instead of brew's raw failure (AC-04.1). The refusal is brew's own: its sudo
// hint on stderr, or the system's permission error.
func TestServicesWriteNeedsSudoIsNamed(t *testing.T) {
	const want = `moshi-hook: needs sudo — run "sudo brew services %s moshi-hook" in a terminal`
	refusals := map[string]string{
		"sudo hint":         "Error: Service `moshi-hook` is started as `root`. Try:\n  sudo brew services stop moshi-hook",
		"requires root":     "Error: moshi-hook requires root",
		"permission denied": "Error: Permission denied @ rb_sysopen",
	}
	writes := map[string]func(w ServicesWriter) error{
		"start":   func(w ServicesWriter) error { return w.Start(context.Background(), "moshi-hook") },
		"stop":    func(w ServicesWriter) error { return w.Stop(context.Background(), "moshi-hook") },
		"restart": func(w ServicesWriter) error { return w.Restart(context.Background(), "moshi-hook") },
		"run":     func(w ServicesWriter) error { return w.Run(context.Background(), "moshi-hook") },
	}
	for refusal, stderr := range refusals {
		for action, write := range writes {
			t.Run(refusal+"/"+action, func(t *testing.T) {
				r := brewtest.NewMockRunner()
				r.ExecuteFn = func(ctx context.Context, args ...string) ([]byte, error) {
					return nil, &BrewExitError{Command: "services " + action, ExitCode: 1, Stderr: stderr}
				}
				err := write(NewServicesWriter(r, NewCache(time.Minute)))

				var needsRoot *ServiceNeedsRootError
				if !errors.As(err, &needsRoot) {
					t.Fatalf("err = %T %v, want *ServiceNeedsRootError", err, err)
				}
				if needsRoot.Action != action || needsRoot.Name != "moshi-hook" {
					t.Errorf("action %q name %q, want %q and moshi-hook", needsRoot.Action, needsRoot.Name, action)
				}
				if got := err.Error(); got != fmt.Sprintf(want, action) {
					t.Errorf("Error() = %q, want %q", got, fmt.Sprintf(want, action))
				}
			})
		}
	}
}

// Any other failure keeps brew's own error, unchanged (AC-04.2).
func TestServicesWriteOtherErrorsUnchanged(t *testing.T) {
	brewErr := &BrewExitError{Command: "services stop", ExitCode: 1, Stderr: "Error: Failed to stop moshi-hook: timed out"}
	for _, cause := range []error{brewErr, errors.New("boom")} {
		r := brewtest.NewMockRunner()
		r.ExecuteFn = func(ctx context.Context, args ...string) ([]byte, error) {
			return nil, cause
		}
		err := NewServicesWriter(r, NewCache(time.Minute)).Stop(context.Background(), "moshi-hook")

		var needsRoot *ServiceNeedsRootError
		if errors.As(err, &needsRoot) {
			t.Errorf("%v became a needs-sudo error", cause)
		}
		if err != cause {
			t.Errorf("err = %v, want the runner's error %v unchanged", err, cause)
		}
	}
}
