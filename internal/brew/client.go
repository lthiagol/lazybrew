package brew

import (
	"time"
)

// CacheTTLs bundles the per-class cache TTLs. A value of 0 is the cache default
// TTL (30s) for that class.
type CacheTTLs struct {
	Formulae time.Duration
	Casks    time.Duration
	Outdated time.Duration
	Taps     time.Duration
	Services time.Duration
	Doctor   time.Duration
}

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
	runner = sudoRunner{inner: runner}

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

// SetCacheTTLs configures per-class cache TTLs on the readers that take them.
func (c *Client) SetCacheTTLs(ttls CacheTTLs) {
	c.Formulae.SetCacheTTLs(ttls)
	c.Casks.SetCacheTTLs(ttls)
	c.Taps.SetCacheTTLs(ttls)
	c.Services.SetCacheTTLs(ttls)
	c.Diagnostics.SetCacheTTLs(ttls)
}
