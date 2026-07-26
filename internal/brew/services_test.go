package brew

import (
	"context"
	"testing"
	"time"
)

func TestServicesServiceList(t *testing.T) {
	r := NewMockRunner()
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
	r := NewMockRunner()
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
	r := NewMockRunner()
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
	r := NewMockRunner()
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
	r := NewMockRunner()
	r.ExecuteFn = func(ctx context.Context, args ...string) ([]byte, error) {
		return []byte(""), nil
	}
	cache := NewCache(time.Minute)
	svc := NewServicesWriter(r, cache)

	if err := svc.Run(context.Background(), "redis"); err != nil {
		t.Fatal(err)
	}
}
