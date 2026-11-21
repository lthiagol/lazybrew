package gui

import (
	"strings"

	"github.com/lthiagol/lazybrew/internal/brew"
)

// canFilter is whether the `/` filter applies to the panel. Outdated, Taps and
// Services can be added here later.
func (p *panelData) canFilter() bool {
	return p.id == PanelFormulae || p.id == PanelCasks
}

// narrowed is whether a filter narrows the list: the leaves filter or the query.
// While it does, unfiltered* hold the full list.
func (p *panelData) narrowed() bool {
	return p.leavesActive || p.query != ""
}

// fullItems, fullFormulae and fullCasks are the list before any narrowing. The
// installed index and the untap count read them, so a filtered-out package
// still counts.
func (p *panelData) fullItems() []string {
	if p.narrowed() {
		return p.unfilteredItems
	}
	return p.items
}

func (p *panelData) fullFormulae() []brew.Formula {
	if p.narrowed() {
		return p.unfilteredFormulae
	}
	return p.formulae
}

func (p *panelData) fullCasks() []brew.Cask {
	if p.narrowed() {
		return p.unfilteredCasks
	}
	return p.casks
}

// fullName is the name of row i of the full list, "" when there is none.
func (p *panelData) fullName(i int) string {
	switch p.id {
	case PanelFormulae:
		if i < len(p.unfilteredFormulae) {
			return p.unfilteredFormulae[i].Name
		}
	case PanelCasks:
		if i < len(p.unfilteredCasks) {
			return p.unfilteredCasks[i].Name
		}
	}
	return ""
}

// totalRows is how many rows the query narrows: the full list less what the
// leaves filter hides.
func (p *panelData) totalRows() int {
	if !p.leavesActive {
		return len(p.fullItems())
	}
	n := 0
	for i := range p.fullItems() {
		if p.leaves[p.fullName(i)] {
			n++
		}
	}
	return n
}

// keepFull holds the full list while a filter narrows it.
func (p *panelData) keepFull() {
	p.unfilteredItems = p.items
	p.unfilteredFormulae = p.formulae
	p.unfilteredCasks = p.casks
}

// narrow rebuilds items, formulae and casks from the full list: the rows the
// leaves filter keeps and the query matches. The cursor stays on the row called
// name when it is still listed, and is clamped otherwise.
func (p *panelData) narrow(name string) {
	query := strings.ToLower(p.query)
	items := make([]string, 0, len(p.unfilteredItems))
	formulae := make([]brew.Formula, 0, len(p.unfilteredFormulae))
	casks := make([]brew.Cask, 0, len(p.unfilteredCasks))
	for i, item := range p.unfilteredItems {
		n := p.fullName(i)
		if p.leavesActive && !p.leaves[n] {
			continue
		}
		if query != "" && !strings.Contains(strings.ToLower(n), query) {
			continue
		}
		items = append(items, item)
		if i < len(p.unfilteredFormulae) {
			formulae = append(formulae, p.unfilteredFormulae[i])
		}
		if i < len(p.unfilteredCasks) {
			casks = append(casks, p.unfilteredCasks[i])
		}
	}
	p.items = items
	p.formulae = formulae
	p.casks = casks
	p.indexInstalled()
	p.selectName(name)
	p.normalize()
}

// restore gives the full list back once no filter narrows it, and keeps the
// cursor on the row called name.
func (p *panelData) restore(name string) {
	p.items = p.unfilteredItems
	p.formulae = p.unfilteredFormulae
	p.casks = p.unfilteredCasks
	p.unfilteredItems = nil
	p.unfilteredFormulae = nil
	p.unfilteredCasks = nil
	p.indexInstalled()
	p.selectName(name)
	p.normalize()
}

// reapply narrows the list again, or restores it when no filter is on.
func (p *panelData) reapply(name string) {
	if p.narrowed() {
		p.narrow(name)
		return
	}
	p.restore(name)
}

// selectName moves the cursor to the row called name, when the list has it.
// An empty name, or one not listed, leaves the cursor where it is.
func (p *panelData) selectName(name string) {
	if name == "" {
		return
	}
	for i := range p.items {
		if p.nameAt(i) == name {
			p.selected = i
			return
		}
	}
}

// toggleLeaves narrows a Formulae panel to the leaves (the formulae that no
// other formula needs), or lifts that narrowing. The query, if any, stays on.
func (p *panelData) toggleLeaves(leaves []string) {
	name := p.selectedName()
	if p.leavesActive {
		p.leavesActive = false
		p.leaves = nil
	} else {
		if !p.narrowed() {
			p.keepFull()
		}
		p.leaves = make(map[string]bool, len(leaves))
		for _, l := range leaves {
			p.leaves[l] = true
		}
		p.leavesActive = true
	}
	p.reapply(name)
}

// setQuery sets the `/` filter's query. A query moves the cursor to the first
// match; clearing it keeps the cursor on the same row.
func (p *panelData) setQuery(q string) {
	if q == p.query {
		return
	}
	name := p.selectedName()
	if q != "" {
		name = ""
		p.selected, p.offset = 0, 0
	}
	if !p.narrowed() {
		p.keepFull()
	}
	p.query = q
	p.reapply(name)
}
