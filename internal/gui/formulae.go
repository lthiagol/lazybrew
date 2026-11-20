package gui

import (
	"cmp"
	"slices"
	"strings"

	"github.com/lthiagol/lazybrew/internal/brew"
)

// sortFormulae orders outdated first, then pinned, then name A–Z (case-insensitive).
// Two separate ranks — outdated and pinned are not collapsed into one flag. Each
// name is lowercased once, not on every comparison, and equal keys keep their
// order.
func sortFormulae(fs []brew.Formula) {
	type keyed struct {
		f    brew.Formula
		name string
	}
	keys := make([]keyed, len(fs))
	for i, f := range fs {
		keys[i] = keyed{f: f, name: strings.ToLower(f.Name)}
	}
	slices.SortStableFunc(keys, func(a, b keyed) int {
		if c := cmp.Compare(flagRank(a.f.Outdated), flagRank(b.f.Outdated)); c != 0 {
			return c
		}
		if c := cmp.Compare(flagRank(a.f.Pinned), flagRank(b.f.Pinned)); c != 0 {
			return c
		}
		return strings.Compare(a.name, b.name)
	})
	for i := range keys {
		fs[i] = keys[i].f
	}
}

// flagRank puts a set flag first: 0 when set, 1 when not.
func flagRank(set bool) int {
	if set {
		return 0
	}
	return 1
}
