package gui

import (
	tea "charm.land/bubbletea/v2"
	"github.com/lthiagol/lazybrew/internal/brew"
)

// preloadOrder is the order the panels are read in once the Status dashboard
// has loaded: the panels a user most often opens come first. Search is never
// preloaded; it is driven by the user's query.
var preloadOrder = []PanelID{PanelFormulae, PanelOutdated, PanelCasks, PanelTaps, PanelServices}

// preloadWidth is how many preload reads may be in flight at once.
const preloadWidth = 2

// preloadState is the background load that follows the Status dashboard. It is
// driven from Update: each read that lands starts the next one, so no goroutine
// is started outside a tea.Cmd.
type preloadState struct {
	// started is set once the first Status load has landed with preload on, so
	// a later refresh never starts a second preload.
	started bool
	// queue holds the panels still to read, in the order they are read.
	queue []PanelID
	// inFlight counts the preload reads started and not yet landed.
	inFlight int
}

// preloadedMsg is the result of a preload read. It carries the same data as a
// DataLoadedMsg, and it frees a read slot. It is a message of its own so that a
// refresh, which counts the DataLoadedMsg it is waiting for, is not disturbed by
// a preload read landing in the middle of it.
type preloadedMsg DataLoadedMsg

// preloadRead is one panel's read as a preloadedMsg. It is fetchPanelData, so
// the read is bounded like every other panel read.
func preloadRead(client *brew.Client, id PanelID) tea.Cmd {
	read := fetchPanelData(client, id)
	return func() tea.Msg {
		return preloadedMsg(read().(DataLoadedMsg))
	}
}

// startPreload queues every panel that is neither loaded nor loading, in
// preloadOrder, and returns the first reads. A panel the user opened before the
// Status dashboard landed is already loading, so it is not read twice.
func (m *Model) startPreload() []tea.Cmd {
	m.preload.started = true
	for _, id := range preloadOrder {
		p := m.panels[id]
		if p.loaded || p.loading {
			continue
		}
		// loading is set while the panel is queued too, so a switch to it
		// is taken as a read in progress, not a second one.
		p.loading = true
		m.preload.queue = append(m.preload.queue, id)
	}
	return m.nextPreloadReads()
}

// nextPreloadReads starts queued reads until preloadWidth are in flight. A
// panel that has loaded since it was queued is dropped from the queue.
func (m *Model) nextPreloadReads() []tea.Cmd {
	var cmds []tea.Cmd
	for m.preload.inFlight < preloadWidth && len(m.preload.queue) > 0 {
		id := m.preload.queue[0]
		m.preload.queue = m.preload.queue[1:]
		if m.panels[id].loaded {
			continue
		}
		m.preload.inFlight++
		cmds = append(cmds, preloadRead(m.client, id))
	}
	return cmds
}

// absorbLanded puts a panel's landed read into the Status dashboard, and the
// Status sidebar's rows with it, for every landing with the preload on: a
// preloaded read or one the user's switch started (F-r1-01). The count is also
// kept in m.landed, so a Status load that lands later keeps it (fillUnknown).
//
// A failed read forgets its source's count instead (F-r2-02): the dashboard and
// m.landed must not keep a count older than the failure. Its cache entry goes
// too, so a Status load that lands later cannot read the older cached value
// either. The error shows on its own panel alone (AC-02.2).
func (m *Model) absorbLanded(msg DataLoadedMsg) {
	if !m.cfg.GUI.Preload {
		return
	}
	if msg.Err != nil {
		m.landed.forgetPanel(msg.PanelID)
		m.status.forgetPanel(msg.PanelID)
		m.invalidatePanelCache(msg.PanelID)
	} else {
		m.landed.absorbPanel(msg.PanelID, msg)
		m.status.absorbPanel(msg.PanelID, msg)
	}
	if status := m.panels[PanelStatus]; status.loaded {
		status.items = statusRows(m.status)
		status.normalize()
	}
}

// promotePreload moves a queued panel to the front of the queue, so the read the
// user's switch needs starts as soon as a slot frees. It reports whether the
// panel was queued; a panel in flight or not queued is left alone.
func (m *Model) promotePreload(id PanelID) bool {
	for i, q := range m.preload.queue {
		if q != id {
			continue
		}
		rest := append(m.preload.queue[:i:i], m.preload.queue[i+1:]...)
		m.preload.queue = append([]PanelID{id}, rest...)
		return true
	}
	return false
}
