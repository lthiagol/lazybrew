package brew

import (
	"context"
	"testing"
	"time"
)

func TestCasksReaderList(t *testing.T) {
	r := NewMockRunner()
	r.ExecuteFn = func(ctx context.Context, args ...string) ([]byte, error) {
		return []byte(`{
			"casks": [
				{
					"name": "google-chrome",
					"full_name": "google-chrome",
					"tap": "homebrew/cask",
					"version": "132.0.6834.160",
					"desc": "Web browser",
					"homepage": "https://www.google.com/chrome/",
					"auto_updates": true,
					"pinned": false
				},
				{
					"name": "firefox",
					"full_name": "firefox",
					"tap": "homebrew/cask",
					"version": "135.0",
					"desc": "Mozilla Firefox",
					"homepage": "https://www.mozilla.org/firefox/",
					"auto_updates": false,
					"pinned": true
				}
			]
		}`), nil
	}
	cache := NewCache(time.Minute)
	reader := NewCasksReader(r, cache)

	casks, err := reader.List(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	if len(casks) != 2 {
		t.Fatalf("expected 2 casks, got %d", len(casks))
	}

	chrome := casks[0]
	if chrome.Name != "google-chrome" {
		t.Errorf("Name = %q, want google-chrome", chrome.Name)
	}
	if !chrome.AutoUpdates {
		t.Error("google-chrome should have auto_updates")
	}
	if chrome.Pinned {
		t.Error("google-chrome should not be pinned")
	}

	firefox := casks[1]
	if firefox.Name != "firefox" {
		t.Errorf("Name = %q, want firefox", firefox.Name)
	}
	if !firefox.Pinned {
		t.Error("firefox should be pinned (6.0.0)")
	}
}

func TestCasksReaderListCached(t *testing.T) {
	callCount := 0
	r := NewMockRunner()
	r.ExecuteFn = func(ctx context.Context, args ...string) ([]byte, error) {
		callCount++
		return []byte(`{"casks":[]}`), nil
	}
	cache := NewCache(time.Minute)
	reader := NewCasksReader(r, cache)

	reader.List(context.Background())
	reader.List(context.Background())

	if callCount != 1 {
		t.Errorf("expected 1 brew call, got %d", callCount)
	}
}

func TestCasksReaderGet(t *testing.T) {
	r := NewMockRunner()
	r.ExecuteFn = func(ctx context.Context, args ...string) ([]byte, error) {
		return []byte(`{
			"casks": [{
				"name": "alfred",
				"full_name": "alfred",
				"tap": "homebrew/cask",
				"version": "5.5",
				"desc": "Productivity app"
			}]
		}`), nil
	}
	cache := NewCache(time.Minute)
	reader := NewCasksReader(r, cache)

	cask, err := reader.Get(context.Background(), "alfred")
	if err != nil {
		t.Fatal(err)
	}
	if cask == nil {
		t.Fatal("expected cask, got nil")
	}
	if cask.Name != "alfred" {
		t.Errorf("Name = %q, want alfred", cask.Name)
	}
}

func TestCasksReaderGetMissing(t *testing.T) {
	r := NewMockRunner()
	r.ExecuteFn = func(ctx context.Context, args ...string) ([]byte, error) {
		return []byte(`{"casks":[]}`), nil
	}
	cache := NewCache(time.Minute)
	reader := NewCasksReader(r, cache)

	cask, err := reader.Get(context.Background(), "nonexistent")
	if err != nil {
		t.Fatal(err)
	}
	if cask != nil {
		t.Error("expected nil for missing cask")
	}
}

func TestCasksReaderOutdated(t *testing.T) {
	r := NewMockRunner()
	r.ExecuteFn = func(ctx context.Context, args ...string) ([]byte, error) {
		return []byte(`{"casks":[{"name":"firefox","full_name":"firefox","installed_versions":["134.0"],"current_version":"135.0"}]}`), nil
	}
	cache := NewCache(time.Minute)
	reader := NewCasksReader(r, cache)

	outdated, err := reader.Outdated(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	if len(outdated) != 1 {
		t.Fatalf("expected 1, got %d", len(outdated))
	}
	if outdated[0].Name != "firefox" || !outdated[0].Outdated || outdated[0].NewVersion != "135.0" {
		t.Errorf("got %+v", outdated[0])
	}
}

func TestCasksWriterInstall(t *testing.T) {
	r := NewMockRunner()
	r.ExecuteStreamFn = func(ctx context.Context, args ...string) (<-chan string, <-chan error) {
		ch := make(chan string, 1)
		errCh := make(chan error, 1)
		ch <- "installed"
		close(ch)
		close(errCh)
		return ch, errCh
	}
	cache := NewCache(time.Minute)
	writer := NewCasksWriter(r, cache)

	ch, errCh := writer.Install(context.Background(), "firefox")
	if line := <-ch; line != "installed" {
		t.Errorf("got %q, want installed", line)
	}
	if err := <-errCh; err != nil {
		t.Fatal(err)
	}
}

func TestCasksWriterZap(t *testing.T) {
	r := NewMockRunner()
	r.ExecuteStreamFn = func(ctx context.Context, args ...string) (<-chan string, <-chan error) {
		ch := make(chan string, 1)
		errCh := make(chan error, 1)
		ch <- "zapped"
		close(ch)
		close(errCh)
		return ch, errCh
	}
	cache := NewCache(time.Minute)
	writer := NewCasksWriter(r, cache)

	ch, errCh := writer.Zap(context.Background(), "firefox")
	if line := <-ch; line != "zapped" {
		t.Errorf("got %q, want zapped", line)
	}
	if err := <-errCh; err != nil {
		t.Fatal(err)
	}
}

func TestCasksWriterPin(t *testing.T) {
	r := NewMockRunner()
	r.ExecuteFn = func(ctx context.Context, args ...string) ([]byte, error) {
		return []byte(""), nil
	}
	cache := NewCache(time.Minute)
	writer := NewCasksWriter(r, cache)

	if err := writer.Pin(context.Background(), "firefox"); err != nil {
		t.Fatal(err)
	}
}

func TestCasksWriterUnpin(t *testing.T) {
	r := NewMockRunner()
	r.ExecuteFn = func(ctx context.Context, args ...string) ([]byte, error) {
		return []byte(""), nil
	}
	cache := NewCache(time.Minute)
	writer := NewCasksWriter(r, cache)

	if err := writer.Unpin(context.Background(), "firefox"); err != nil {
		t.Fatal(err)
	}
}
