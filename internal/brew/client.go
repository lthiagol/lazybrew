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

// SetOutdatedTTL configures the cache TTL for `brew outdated` results on
// both formulae and casks readers. A value <= 0 keeps the cache default
// (30s) so callers that want the previous behavior can pass 0. M9 wires
// this from `brew.outdated_ttl` (default 30m).
//
// Readers that do not implement the optional TTLSetter interface (e.g. test
// doubles) are silently skipped so production code does not need to type
// switch.
func (c *Client) SetOutdatedTTL(ttl time.Duration) {
	if r, ok := c.Formulae.(TTLSetter); ok {
		r.SetOutdatedTTL(ttl)
	}
	if r, ok := c.Casks.(TTLSetter); ok {
		r.SetOutdatedTTL(ttl)
	}
}
