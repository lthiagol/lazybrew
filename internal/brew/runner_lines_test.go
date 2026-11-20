package brew

import (
	"bytes"
	"context"
	"fmt"
	"testing"
)

// writeLines feeds chunks to a lineWriter, then flushes it when flush is set,
// and returns the lines it sent.
func writeLines(t *testing.T, flush bool, chunks ...string) []string {
	t.Helper()
	out := make(chan string, 1000)
	w := &lineWriter{ctx: context.Background(), out: out}
	for _, c := range chunks {
		if _, err := w.Write([]byte(c)); err != nil {
			t.Fatal(err)
		}
	}
	if flush {
		w.flush()
	}
	close(out)
	var lines []string
	for line := range out {
		lines = append(lines, line)
	}
	return lines
}

func TestLineWriterSplitsLines(t *testing.T) {
	cases := []struct {
		name   string
		chunks []string
		flush  bool
		want   []string
	}{
		{"lines split across writes", []string{"ab", "c\nd", "ef\n"}, true, []string{"abc", "def"}},
		{"carriage return ends a line", []string{"one\rtwo\rthree"}, true, []string{"one", "two", "three"}},
		{"crlf is one break", []string{"a\r\nb\r\n"}, true, []string{"a", "b"}},
		{"crlf split across writes", []string{"a\r", "\nb\n"}, true, []string{"a", "b"}},
		{"empty line between newlines", []string{"a\n\nb\n"}, true, []string{"a", "", "b"}},
		{"progress redraws become lines", []string{"10%\r20%\r30%\rdone\n"}, true, []string{"10%", "20%", "30%", "done"}},
		{"final partial line is flushed", []string{"x\ny"}, true, []string{"x", "y"}},
		{"final partial line waits for flush", []string{"x\ny"}, false, []string{"x"}},
		{"no output", nil, true, nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := writeLines(t, tc.flush, tc.chunks...)
			if fmt.Sprintf("%q", got) != fmt.Sprintf("%q", tc.want) {
				t.Errorf("lines = %q, want %q", got, tc.want)
			}
		})
	}
}

// TestLineWriterIndependentOfChunking: the lines do not depend on where the
// writes happen to split the output.
func TestLineWriterIndependentOfChunking(t *testing.T) {
	text := "alpha\r\nbeta\rgamma\n\ndelta\r\n\r\nepsilon"
	whole := fmt.Sprintf("%q", writeLines(t, true, text))
	for size := 1; size <= len(text); size++ {
		var chunks []string
		for start := 0; start < len(text); start += size {
			end := min(start+size, len(text))
			chunks = append(chunks, text[start:end])
		}
		if got := fmt.Sprintf("%q", writeLines(t, true, chunks...)); got != whole {
			t.Fatalf("chunks of %d bytes: lines = %s, want %s", size, got, whole)
		}
	}
}

// benchLineWriter writes 1 MiB to a lineWriter in chunks of the given bytes,
// then flushes it. The lines are drained by another goroutine, as the runner's
// consumer does.
func benchLineWriter(b *testing.B, chunk []byte) {
	const total = 1 << 20
	out := make(chan string, 100)
	drained := make(chan struct{})
	go func() {
		for range out {
		}
		close(drained)
	}()

	b.SetBytes(total)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := &lineWriter{ctx: context.Background(), out: out}
		for written := 0; written < total; written += len(chunk) {
			if _, err := w.Write(chunk); err != nil {
				b.Fatal(err)
			}
		}
		w.flush()
	}
	b.StopTimer()
	close(out)
	<-drained
}

func BenchmarkLineWriterNoNewlines(b *testing.B) {
	benchLineWriter(b, bytes.Repeat([]byte("x"), 64))
}

func BenchmarkLineWriterNewlines(b *testing.B) {
	benchLineWriter(b, append(bytes.Repeat([]byte("x"), 63), '\n'))
}
