package brew

import (
	"context"
	"encoding/json"
	"time"
)

type ServicesReader interface {
	List(ctx context.Context) ([]Service, error)
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
	if cached, ok := s.cache.Get(KeyServicesList); ok {
		if services, ok := cached.([]Service); ok {
			return services, nil
		}
	}

	output, err := s.runner.Execute(ctx, "services", "list", "--json")
	if err != nil {
		return nil, err
	}

	var rawJSON []json.RawMessage
	if err := json.Unmarshal(output, &rawJSON); err != nil {
		return nil, err
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

	s.cache.SetWithTTL(KeyServicesList, services, s.servicesTTL)
	return services, nil
}

// SetCacheTTLs configures the cache TTL for `brew services list` results.
// Implements TTLSetter.
func (s *servicesReader) SetCacheTTLs(ttls CacheTTLs) {
	if ttls.Services > 0 {
		s.servicesTTL = ttls.Services
	}
}

func (s *servicesWriter) Start(ctx context.Context, name string) error {
	s.cache.Invalidate(KeyServicesList)
	_, err := s.runner.Execute(ctx, "services", "start", name)
	return err
}

func (s *servicesWriter) Stop(ctx context.Context, name string) error {
	s.cache.Invalidate(KeyServicesList)
	_, err := s.runner.Execute(ctx, "services", "stop", name)
	return err
}

func (s *servicesWriter) Restart(ctx context.Context, name string) error {
	s.cache.Invalidate(KeyServicesList)
	_, err := s.runner.Execute(ctx, "services", "restart", name)
	return err
}

func (s *servicesWriter) Run(ctx context.Context, name string) error {
	_, err := s.runner.Execute(ctx, "services", "run", name)
	return err
}

func (s *servicesWriter) Cleanup(ctx context.Context) error {
	s.cache.Invalidate(KeyServicesList)
	_, err := s.runner.Execute(ctx, "services", "cleanup")
	return err
}
