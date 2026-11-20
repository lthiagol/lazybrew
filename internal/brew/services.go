package brew

import (
	"context"
	"encoding/json"
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
		services, err := decodeEach(output, func(srv Service) bool { return srv.Name != "" })
		return services, err == nil, err
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
		info, err := decodeEach(output, func(entry ServiceInfo) bool { return entry.Name != "" })
		return info, err == nil, err
	})
}

// decodeEach decodes each entry of a JSON array as a T and keeps the ones ok
// accepts. An entry that does not decode is skipped; output that is not an
// array is an error.
func decodeEach[T any](data []byte, ok func(T) bool) ([]T, error) {
	var rawJSON []json.RawMessage
	if err := json.Unmarshal(data, &rawJSON); err != nil {
		return nil, err
	}
	out := make([]T, 0, len(rawJSON))
	for _, raw := range rawJSON {
		var v T
		if err := json.Unmarshal(raw, &v); err != nil {
			continue
		}
		if ok(v) {
			out = append(out, v)
		}
	}
	return out, nil
}

// SetCacheTTLs configures the cache TTL for `brew services` results. A zero TTL
// is the cache default.
func (s *servicesReader) SetCacheTTLs(ttls CacheTTLs) {
	s.servicesTTL = ttls.Services
}

func (s *servicesWriter) Start(ctx context.Context, name string) error {
	defer s.cache.InvalidateFor("services")
	_, err := s.runner.Execute(ctx, "services", "start", name)
	return serviceWriteError("start", name, err)
}

func (s *servicesWriter) Stop(ctx context.Context, name string) error {
	defer s.cache.InvalidateFor("services")
	_, err := s.runner.Execute(ctx, "services", "stop", name)
	return serviceWriteError("stop", name, err)
}

func (s *servicesWriter) Restart(ctx context.Context, name string) error {
	defer s.cache.InvalidateFor("services")
	_, err := s.runner.Execute(ctx, "services", "restart", name)
	return serviceWriteError("restart", name, err)
}

func (s *servicesWriter) Run(ctx context.Context, name string) error {
	// Invalidate like every other services write: `services run` changes
	// service state just as start/stop do (r1 finding 19).
	defer s.cache.InvalidateFor("services")
	_, err := s.runner.Execute(ctx, "services", "run", name)
	return serviceWriteError("run", name, err)
}

func (s *servicesWriter) Cleanup(ctx context.Context) error {
	defer s.cache.InvalidateFor("services")
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
	if err == nil || !errorMentions(err, rootMarkers) {
		return err
	}
	return &ServiceNeedsRootError{Action: action, Name: name}
}
