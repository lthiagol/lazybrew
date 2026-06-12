package brew

import (
	"context"
	"encoding/json"
)

type ServicesService interface {
	List(ctx context.Context) ([]Service, error)
	Start(ctx context.Context, name string) error
	Stop(ctx context.Context, name string) error
	Restart(ctx context.Context, name string) error
	Run(ctx context.Context, name string) error
	Cleanup(ctx context.Context) error
}

type servicesService struct {
	runner Runner
	cache  *Cache
}

func NewServicesService(runner Runner, cache *Cache) ServicesService {
	return &servicesService{runner: runner, cache: cache}
}

func (s *servicesService) List(ctx context.Context) ([]Service, error) {
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

	s.cache.Set(KeyServicesList, services)
	return services, nil
}

func (s *servicesService) Start(ctx context.Context, name string) error {
	s.cache.Invalidate(KeyServicesList)
	_, err := s.runner.Execute(ctx, "services", "start", name)
	return err
}

func (s *servicesService) Stop(ctx context.Context, name string) error {
	s.cache.Invalidate(KeyServicesList)
	_, err := s.runner.Execute(ctx, "services", "stop", name)
	return err
}

func (s *servicesService) Restart(ctx context.Context, name string) error {
	s.cache.Invalidate(KeyServicesList)
	_, err := s.runner.Execute(ctx, "services", "restart", name)
	return err
}

func (s *servicesService) Run(ctx context.Context, name string) error {
	_, err := s.runner.Execute(ctx, "services", "run", name)
	return err
}

func (s *servicesService) Cleanup(ctx context.Context) error {
	s.cache.Invalidate(KeyServicesList)
	_, err := s.runner.Execute(ctx, "services", "cleanup")
	return err
}
