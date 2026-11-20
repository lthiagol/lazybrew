package gui

import (
	"testing"

	tea "charm.land/bubbletea/v2"
)

// benchMsg is a message no handler knows: Update passes it through, so the
// benchmark measures what every message costs outside the handlers.
type benchMsg struct{}

// BenchmarkUpdateUnhandled measures one unhandled message at 140x40, with the
// animations off (LAZYBREW_ANIMATIONS=0) and on (=1). It is not part of the
// gate: run it with -bench UpdateUnhandled.
func BenchmarkUpdateUnhandled(b *testing.B) {
	for _, motion := range []string{"0", "1"} {
		b.Run("animations="+motion, func(b *testing.B) {
			b.Setenv("LAZYBREW_ANIMATIONS", motion)
			m := updateModel(newTestModel(), tea.WindowSizeMsg{Width: 140, Height: 40})
			var msg tea.Msg = benchMsg{}
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				m.Update(msg)
			}
		})
	}
}
