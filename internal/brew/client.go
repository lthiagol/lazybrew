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
