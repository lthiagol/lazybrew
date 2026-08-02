package brew

import (
	"time"
)

type Client struct {
	Runner           Runner
	Formulae         FormulaeReader
	FormulaeWrite    FormulaeWriter
	Casks            CasksReader
	CasksWrite       CasksWriter
	Taps             TapsReader
	TapsWrite        TapsWriter
	Services         ServicesReader
	ServicesWrite    ServicesWriter
	Search           SearchService
	Trust            TrustReader
	TrustWrite       TrustWriter
	Diagnostics      DiagnosticsReader
	DiagnosticsWrite DiagnosticsWriter
	Cache            *Cache
}

func NewClient(runner Runner) *Client {
	cache := NewCache(30 * time.Second)

	return &Client{
		Runner:           runner,
		Formulae:         NewFormulaeReader(runner, cache),
		FormulaeWrite:    NewFormulaeWriter(runner, cache),
		Casks:            NewCasksReader(runner, cache),
		CasksWrite:       NewCasksWriter(runner, cache),
		Taps:             NewTapsReader(runner, cache),
		TapsWrite:        NewTapsWriter(runner, cache),
		Services:         NewServicesReader(runner, cache),
		ServicesWrite:    NewServicesWriter(runner, cache),
		Search:           NewSearchService(runner),
		Trust:            NewTrustReader(runner, cache),
		TrustWrite:       NewTrustWriter(runner, cache),
		Diagnostics:      NewDiagnosticsReader(runner, cache),
		DiagnosticsWrite: NewDiagnosticsWriter(runner, cache),
		Cache:            cache,
	}
}

// SetCacheTTLs configures per-class cache TTLs on all readers that
// participate in tiered refresh (M12). A value <= 0 in any field keeps
// the cache default TTL (30s) for that class, preserving pre-M12
// behavior. Readers that do not implement the optional TTLSetter
// interface (e.g. test doubles) are silently skipped.
func (c *Client) SetCacheTTLs(ttls CacheTTLs) {
	for _, r := range []any{c.Formulae, c.Casks, c.Taps, c.Services, c.Diagnostics} {
		if t, ok := r.(TTLSetter); ok {
			t.SetCacheTTLs(ttls)
		}
	}
}
