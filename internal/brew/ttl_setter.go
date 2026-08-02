package brew

import "time"

// CacheTTLs bundles per-class cache TTLs (M12). Each field is honored
// by the matching reader when it writes its cache entry; values <= 0
// keep the cache default TTL in effect for that class.
type CacheTTLs struct {
	Formulae time.Duration
	Casks    time.Duration
	Outdated time.Duration
	Taps     time.Duration
	Services time.Duration
	Doctor   time.Duration
}

// TTLSetter configures per-class cache TTLs. Readers that participate
// in tiered refresh (formulae, casks, taps, services, doctor, outdated)
// implement this so a single Client.SetCacheTTLs(ttls) call wires the
// config to all readers at once.
//
// Readers that do not implement TTLSetter are silently skipped (test
// doubles can opt out without breaking the call).
type TTLSetter interface {
	SetCacheTTLs(ttls CacheTTLs)
}
