package brew

import "time"

// TTLSetter is an optional interface implemented by readers that support
// a per-key TTL override (currently Outdated readers). Client.SetOutdatedTTL
// uses it to avoid leaking concrete reader types. Readers that do not
// implement it are silently skipped.
type TTLSetter interface {
	SetOutdatedTTL(ttl time.Duration)
}
