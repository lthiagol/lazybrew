package gui

import (
	"fmt"
	"reflect"
	"testing"
)

// logWith is a command log of the given size holding n commands, numbered from
// 0, appended in order.
func logWith(size, n int) *CommandLog {
	cl := NewCommandLog(size)
	for i := 0; i < n; i++ {
		cl.Append(fmt.Sprint(i))
	}
	return cl
}

// commandsOf is the command text of each entry, in order.
func commandsOf(entries []LogEntry) []string {
	out := make([]string, 0, len(entries))
	for _, e := range entries {
		out = append(out, e.Command)
	}
	return out
}

// TestCommandLogLenAndLast (cr3 AC-05): Len counts the entries held, and Last(n)
// is the newest n of them, oldest first, for an empty ring, a partly full one,
// and one that has wrapped.
func TestCommandLogLenAndLast(t *testing.T) {
	cases := []struct {
		name     string
		size, n  int
		wantLen  int
		last     int
		wantLast []string
	}{
		{name: "empty", size: 5, n: 0, wantLen: 0, last: 3, wantLast: nil},
		{name: "partly full, fewer than asked", size: 5, n: 3, wantLen: 3, last: 10, wantLast: []string{"0", "1", "2"}},
		{name: "partly full, the newest two", size: 5, n: 3, wantLen: 3, last: 2, wantLast: []string{"1", "2"}},
		{name: "wrapped, all of it", size: 5, n: 7, wantLen: 5, last: 5, wantLast: []string{"2", "3", "4", "5", "6"}},
		{name: "wrapped, the newest two", size: 5, n: 7, wantLen: 5, last: 2, wantLast: []string{"5", "6"}},
		{name: "asked for none", size: 5, n: 7, wantLen: 5, last: 0, wantLast: nil},
		{name: "asked for a negative count", size: 5, n: 7, wantLen: 5, last: -1, wantLast: nil},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			cl := logWith(c.size, c.n)
			if got := cl.Len(); got != c.wantLen {
				t.Fatalf("Len() = %d, want %d", got, c.wantLen)
			}
			got := commandsOf(cl.Last(c.last))
			if !reflect.DeepEqual(got, c.wantLast) && !(len(got) == 0 && len(c.wantLast) == 0) {
				t.Fatalf("Last(%d) = %q, want %q", c.last, got, c.wantLast)
			}
			// Last is the tail of Entries, which copies the whole ring.
			all := commandsOf(cl.Entries())
			if tail := all[max(0, len(all)-max(0, c.last)):]; len(got) != 0 && !reflect.DeepEqual(got, tail) {
				t.Fatalf("Last(%d) = %q, the tail of Entries is %q", c.last, got, tail)
			}
		})
	}
}
