package brew

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"
)

type ServicesReader interface {
	List(ctx context.Context) ([]Service, error)
	Info(ctx context.Context) ([]ServiceInfo, error)
	SetCacheTTLs(ttls CacheTTLs)
}

type ServicesWriter interface {
	Start(ctx context.Context, name string) error
	Stop(ctx context.Context, name string) error
	Restart(ctx context.Context, name string) error
	Run(ctx context.Context, name string) error
	Cleanup(ctx context.Context) error
}

type servicesReader struct {
	runner      Runner
	cache       *Cache
	servicesTTL time.Duration
}

type servicesWriter struct {
	runner Runner
	cache  *Cache
}

func NewServicesReader(runner Runner, cache *Cache) ServicesReader {
	return &servicesReader{runner: runner, cache: cache}
}

func NewServicesWriter(runner Runner, cache *Cache) ServicesWriter {
	return &servicesWriter{runner: runner, cache: cache}
}

func (s *servicesReader) List(ctx context.Context) ([]Service, error) {
	return cached(s.cache, KeyServicesList, s.servicesTTL, func() ([]Service, bool, error) {
		output, err := s.runner.Execute(ctx, "services", "list", "--json")
		if err != nil {
			return nil, false, err
		}

		var rawJSON []json.RawMessage
		if err := json.Unmarshal(output, &rawJSON); err != nil {
			return nil, false, err
		}

		services := make([]Service, 0, len(rawJSON))
		for _, raw := range rawJSON {
			var srv Service
			if err := json.Unmarshal(raw, &srv); err != nil {
				continue
			}
			if srv.Name != "" {
				services = append(services, srv)
			}
		}
		return services, true, nil
	})
}

// Info is `brew services info --all --json`: the detail of every service in one
// call. An entry that does not parse is skipped, as List skips one.
func (s *servicesReader) Info(ctx context.Context) ([]ServiceInfo, error) {
	return cached(s.cache, KeyServicesInfo, s.servicesTTL, func() ([]ServiceInfo, bool, error) {
		output, err := s.runner.Execute(ctx, "services", "info", "--all", "--json")
		if err != nil {
			return nil, false, err
		}

		var rawJSON []json.RawMessage
		if err := json.Unmarshal(output, &rawJSON); err != nil {
			return nil, false, err
		}

		info := make([]ServiceInfo, 0, len(rawJSON))
		for _, raw := range rawJSON {
			var entry ServiceInfo
			if err := json.Unmarshal(raw, &entry); err != nil {
				continue
			}
			if entry.Name != "" {
				info = append(info, entry)
			}
		}
		return info, true, nil
	})
}

// SetCacheTTLs configures the cache TTL for `brew services list` results.
// Implements TTLSetter.
func (s *servicesReader) SetCacheTTLs(ttls CacheTTLs) {
	if ttls.Services > 0 {
		s.servicesTTL = ttls.Services
	}
}

func (s *servicesWriter) Start(ctx context.Context, name string) error {
	s.cache.Invalidate(KeyServicesList, KeyServicesInfo)
	_, err := s.runner.Execute(ctx, "services", "start", name)
	return serviceWriteError("start", name, err)
}

func (s *servicesWriter) Stop(ctx context.Context, name string) error {
	s.cache.Invalidate(KeyServicesList, KeyServicesInfo)
	_, err := s.runner.Execute(ctx, "services", "stop", name)
	return serviceWriteError("stop", name, err)
}

func (s *servicesWriter) Restart(ctx context.Context, name string) error {
	s.cache.Invalidate(KeyServicesList, KeyServicesInfo)
	_, err := s.runner.Execute(ctx, "services", "restart", name)
	return serviceWriteError("restart", name, err)
}

func (s *servicesWriter) Run(ctx context.Context, name string) error {
	// Invalidate like every other services write: `services run` changes
	// service state just as start/stop do (r1 finding 19).
	s.cache.Invalidate(KeyServicesList, KeyServicesInfo)
	_, err := s.runner.Execute(ctx, "services", "run", name)
	return serviceWriteError("run", name, err)
}

func (s *servicesWriter) Cleanup(ctx context.Context) error {
	s.cache.Invalidate(KeyServicesList, KeyServicesInfo)
	_, err := s.runner.Execute(ctx, "services", "cleanup")
	return err
}

// rootMarkers are the words of brew's refusal for want of root: its own sudo
// hint, or the system's permission error.
var rootMarkers = []string{"sudo", "requires root", "permission denied"}

// serviceWriteError says that a services write brew refused for want of root
// needs sudo, naming the command to run (AC-04.1). Any other error passes
// through unchanged.
func serviceWriteError(action, name string, err error) error {
	if err == nil || !refusedForRoot(err) {
		return err
	}
	return &ServiceNeedsRootError{Action: action, Name: name}
}

// refusedForRoot reads brew's words for a root refusal: the error text and the
// stderr brew prints its hint to.
func refusedForRoot(err error) bool {
	text := err.Error()
	var exit *BrewExitError
	if errors.As(err, &exit) {
		text += "\n" + exit.Stderr
	}
	text = strings.ToLower(text)
	for _, marker := range rootMarkers {
		if strings.Contains(text, marker) {
			return true
		}
	}
	return false
}
