package brew

import (
	"time"
)

type Client struct {
	Formulae    FormulaeReader
	FormulaeWrite FormulaeWriter
	Casks       CasksReader
	CasksWrite  CasksWriter
	Taps        TapsService
	Services    ServicesService
	Search      SearchService
	Trust       TrustService
	Diagnostics DiagnosticsReader
	DiagnosticsWrite DiagnosticsWriter
	Cache       *Cache
}

func NewClient(runner Runner) *Client {
	cache := NewCache(30 * time.Second)

	return &Client{
		Formulae:         NewFormulaeReader(runner, cache),
		FormulaeWrite:    NewFormulaeWriter(runner, cache),
		Casks:            NewCasksReader(runner, cache),
		CasksWrite:       NewCasksWriter(runner, cache),
		Taps:             NewTapsService(runner, cache),
		Services:         NewServicesService(runner, cache),
		Search:           NewSearchService(runner),
		Trust:            NewTrustService(runner, cache),
		Diagnostics:      NewDiagnosticsReader(runner, cache),
		DiagnosticsWrite: NewDiagnosticsWriter(runner, cache),
		Cache:            cache,
	}
}
