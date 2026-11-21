package brew

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"

	"github.com/lthiagol/lazybrew/internal/brew/brewtest"
)

var errWriteBoom = errors.New("boom")

// TestStreamingWriteInvalidatesAfterItFinishes: a list read that runs during a
// streaming install caches the data from before the install. The cache must
// drop it once the install has finished, so the next read shows the new data.
func TestStreamingWriteInvalidatesAfterItFinishes(t *testing.T) {
	ctx := context.Background()
	var mu sync.Mutex
	reads := 0
	runner := &brewtest.MockRunner{
		ExecuteFn: func(ctx context.Context, args ...string) ([]byte, error) {
			mu.Lock()
			reads++
			n := reads
			mu.Unlock()
			if n == 1 {
				return []byte(`{"formulae":[{"name":"before-install"}]}`), nil
			}
			return []byte(`{"formulae":[{"name":"after-install"}]}`), nil
		},
	}
	lines := make(chan string)
	errs := make(chan error, 1)
	runner.ExecuteStreamFn = func(ctx context.Context, args ...string) (<-chan string, <-chan error) {
		return lines, errs
	}
	client := NewClient(runner)

	_, done := client.FormulaeWrite.Install(ctx, "ripgrep")

	during, err := client.Formulae.List(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(during) != 1 || during[0].Name != "before-install" {
		t.Fatalf("read during the install got %+v, want before-install", during)
	}

	close(errs)
	if err := <-done; err != nil {
		t.Fatalf("install finished with %v", err)
	}

	after, err := client.Formulae.List(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(after) != 1 || after[0].Name != "after-install" {
		t.Fatalf("read after the install got %+v, want after-install", after)
	}
}

// TestAutoremoveClearsFormulaeList: autoremove removes formulae, so the cached
// formulae list must go.
func TestAutoremoveClearsFormulaeList(t *testing.T) {
	client := NewClient(brewtest.NewMockRunner())
	client.Cache.Set(KeyFormulaeList, []Formula{{Name: "orphan"}})

	_, errs := client.DiagnosticsWrite.Autoremove(context.Background(), false)
	<-errs

	if _, ok := client.Cache.Get(KeyFormulaeList); ok {
		t.Fatal("autoremove left the formulae list cached")
	}
}

// writeCase is one writer: the group whose keys it must drop, the one brew
// command it runs, and the call itself. Stream writers return the stream's
// error; the others return the command's error.
type writeCase struct {
	name    string
	group   string
	command string
	run     func(c *Client) error
}

func writeCases() []writeCase {
	ctx := context.Background()
	drain := func(errs <-chan error) error { return <-errs }
	return []writeCase{
		{"formulae install", "install", "install x", func(c *Client) error {
			_, errs := c.FormulaeWrite.Install(ctx, "x")
			return drain(errs)
		}},
		{"formulae uninstall", "uninstall", "uninstall x", func(c *Client) error {
			_, errs := c.FormulaeWrite.Uninstall(ctx, "x")
			return drain(errs)
		}},
		{"formulae reinstall", "reinstall", "reinstall x", func(c *Client) error {
			_, errs := c.FormulaeWrite.Reinstall(ctx, "x")
			return drain(errs)
		}},
		{"formulae upgrade", "upgrade", "upgrade x", func(c *Client) error {
			_, errs := c.FormulaeWrite.Upgrade(ctx, "x")
			return drain(errs)
		}},
		{"formulae pin", "pin", "pin x", func(c *Client) error {
			return c.FormulaeWrite.Pin(ctx, "x")
		}},
		{"formulae unpin", "unpin", "unpin x", func(c *Client) error {
			return c.FormulaeWrite.Unpin(ctx, "x")
		}},
		{"cask install", "install", "install --cask x", func(c *Client) error {
			_, errs := c.CasksWrite.Install(ctx, "x")
			return drain(errs)
		}},
		{"cask uninstall", "uninstall", "uninstall --cask x", func(c *Client) error {
			_, errs := c.CasksWrite.Uninstall(ctx, "x")
			return drain(errs)
		}},
		{"cask reinstall", "reinstall", "reinstall --cask x", func(c *Client) error {
			_, errs := c.CasksWrite.Reinstall(ctx, "x")
			return drain(errs)
		}},
		{"cask zap", "uninstall", "uninstall --zap --cask x", func(c *Client) error {
			_, errs := c.CasksWrite.Zap(ctx, "x")
			return drain(errs)
		}},
		{"cask upgrade", "upgrade", "upgrade --cask x", func(c *Client) error {
			_, errs := c.CasksWrite.Upgrade(ctx, "x")
			return drain(errs)
		}},
		{"tap", "tap", "tap x", func(c *Client) error {
			return c.TapsWrite.Tap(ctx, "x")
		}},
		{"untap", "untap", "untap x", func(c *Client) error {
			return c.TapsWrite.Untap(ctx, "x")
		}},
		{"repair", "tap", "tap --repair x", func(c *Client) error {
			return c.TapsWrite.Repair(ctx, "x")
		}},
		{"trust tap", "trust", "trust x", func(c *Client) error {
			return c.TrustWrite.TrustTap(ctx, "x")
		}},
		{"trust formula", "trust", "trust --formula x", func(c *Client) error {
			return c.TrustWrite.TrustFormula(ctx, "x")
		}},
		{"trust cask", "trust", "trust --cask x", func(c *Client) error {
			return c.TrustWrite.TrustCask(ctx, "x")
		}},
		{"untrust tap", "untrust", "untrust x", func(c *Client) error {
			return c.TrustWrite.UntrustTap(ctx, "x")
		}},
		{"untrust formula", "untrust", "untrust --formula x", func(c *Client) error {
			return c.TrustWrite.UntrustFormula(ctx, "x")
		}},
		{"untrust cask", "untrust", "untrust --cask x", func(c *Client) error {
			return c.TrustWrite.UntrustCask(ctx, "x")
		}},
		{"untrust command", "untrust", "untrust --command x", func(c *Client) error {
			return c.TrustWrite.UntrustCommand(ctx, "x")
		}},
		{"services start", "services", "services start x", func(c *Client) error {
			return c.ServicesWrite.Start(ctx, "x")
		}},
		{"services stop", "services", "services stop x", func(c *Client) error {
			return c.ServicesWrite.Stop(ctx, "x")
		}},
		{"services restart", "services", "services restart x", func(c *Client) error {
			return c.ServicesWrite.Restart(ctx, "x")
		}},
		{"services run", "services", "services run x", func(c *Client) error {
			return c.ServicesWrite.Run(ctx, "x")
		}},
		{"services cleanup", "services", "services cleanup", func(c *Client) error {
			return c.ServicesWrite.Cleanup(ctx)
		}},
		{"brew update", "update", "update", func(c *Client) error {
			_, errs := c.DiagnosticsWrite.Update(ctx)
			return drain(errs)
		}},
		{"cleanup", "cleanup", "cleanup", func(c *Client) error {
			_, errs := c.DiagnosticsWrite.Cleanup(ctx, false)
			return drain(errs)
		}},
		{"autoremove", "autoremove", "autoremove", func(c *Client) error {
			_, errs := c.DiagnosticsWrite.Autoremove(ctx, false)
			return drain(errs)
		}},
	}
}

// TestWritesInvalidateTheirGroup: every writer, when its command succeeds or
// fails, drops every key of its group and runs exactly the one brew command.
func TestWritesInvalidateTheirGroup(t *testing.T) {
	for _, tc := range writeCases() {
		for _, fail := range []bool{false, true} {
			name := tc.name + "/succeeds"
			if fail {
				name = tc.name + "/fails"
			}
			t.Run(name, func(t *testing.T) {
				var ran []string
				runner := &brewtest.MockRunner{
					ExecuteFn: func(ctx context.Context, args ...string) ([]byte, error) {
						ran = append(ran, strings.Join(args, " "))
						if fail {
							return nil, errWriteBoom
						}
						return nil, nil
					},
					ExecuteStreamFn: func(ctx context.Context, args ...string) (<-chan string, <-chan error) {
						ran = append(ran, strings.Join(args, " "))
						lines := make(chan string)
						close(lines)
						errs := make(chan error, 1)
						if fail {
							errs <- errWriteBoom
						}
						close(errs)
						return lines, errs
					},
				}
				client := NewClient(runner)

				keys := invalidateGroups[tc.group]
				if len(keys) == 0 {
					t.Fatalf("group %q has no keys to invalidate", tc.group)
				}
				for _, key := range keys {
					client.Cache.Set(key, "seeded")
				}

				if err := tc.run(client); (err != nil) != fail {
					t.Errorf("error = %v, want failure=%v", err, fail)
				}
				if len(ran) != 1 || ran[0] != tc.command {
					t.Errorf("ran %q, want [%q]", ran, tc.command)
				}
				for _, key := range keys {
					if _, ok := client.Cache.Get(key); ok {
						t.Errorf("%s still cached after %s", key, tc.name)
					}
				}
			})
		}
	}
}
