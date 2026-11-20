package gui

import (
	"sort"
	"strings"

	"github.com/lthiagol/lazybrew/internal/brew"
)

// sortFormulae orders outdated first, then pinned, then name A–Z (case-insensitive).
// Two separate ranks — outdated and pinned are not collapsed into one flag.
func sortFormulae(fs []brew.Formula) {
	sort.SliceStable(fs, func(i, j int) bool {
		oi, oj := 1, 1
		if fs[i].Outdated {
			oi = 0
		}
		if fs[j].Outdated {
			oj = 0
		}
		if oi != oj {
			return oi < oj
		}
		pi, pj := 1, 1
		if fs[i].Pinned {
			pi = 0
		}
		if fs[j].Pinned {
			pj = 0
		}
		if pi != pj {
			return pi < pj
		}
		return strings.ToLower(fs[i].Name) < strings.ToLower(fs[j].Name)
	})
}
