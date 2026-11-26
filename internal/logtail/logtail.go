// Package logtail reads the end of a log file for the Services Logs tab: the
// last lines, from a bounded read at the end of the file, made safe to draw.
package logtail

import (
	"errors"
	"io"
	"os"
	"strings"
	"time"
	"unicode"
)

// MaxBytes bounds one read: the tail never reads more than this from the end
// of the file, however large the file is.
const MaxBytes = 256 << 10

// ErrNoPath is returned when there is no path to read: brew reports no log
// file for the stream.
var ErrNoPath = errors.New("no log path")

// Tail is the end of a log file: its last lines, and the file's size and
// modification time at the read, so a later read can tell growth, truncation
// or rotation.
type Tail struct {
	Lines   []string
	Size    int64
	ModTime time.Time
}

// Read returns the last n lines of the file at path, reading at most MaxBytes
// from its end. A file that does not exist is an error for which
// errors.Is(err, fs.ErrNotExist) holds; an empty path is ErrNoPath.
func Read(path string, n int) (Tail, error) {
	return read(path, n, MaxBytes)
}

func read(path string, n int, maxBytes int64) (Tail, error) {
	if path == "" {
		return Tail{}, ErrNoPath
	}
	f, err := os.Open(path)
	if err != nil {
		return Tail{}, err
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		return Tail{}, err
	}
	size := info.Size()
	start := max(0, size-maxBytes)
	// The read also takes the byte before start, to tell whether start begins a
	// line: after a newline it does, and its line is kept whole.
	off := start
	if start > 0 {
		off = start - 1
	}
	if _, err := f.Seek(off, io.SeekStart); err != nil {
		return Tail{}, err
	}
	buf := make([]byte, size-off)
	got, err := io.ReadFull(f, buf)
	// A file that shrank between the stat and the read leaves a short read:
	// the bytes it did give are the tail.
	if err != nil && !errors.Is(err, io.ErrUnexpectedEOF) && !errors.Is(err, io.EOF) {
		return Tail{}, err
	}
	text := string(buf[:got])
	if start > 0 && got > 0 {
		before := text[0]
		text = text[1:]
		if before != '\n' {
			// The read began inside a line, so that partial line is dropped.
			i := strings.IndexByte(text, '\n')
			if i < 0 {
				text = ""
			} else {
				text = text[i+1:]
			}
		}
	}
	return Tail{
		Lines:   lastLines(text, n),
		Size:    size,
		ModTime: info.ModTime(),
	}, nil
}

// lastLines splits text into lines, drops the trailing newline's empty line,
// and keeps the last n lines, each cleaned for the terminal. A text that is
// empty has no lines; one that is only a newline has one empty line.
func lastLines(text string, n int) []string {
	if text == "" || n <= 0 {
		return nil
	}
	lines := strings.Split(strings.TrimSuffix(text, "\n"), "\n")
	if len(lines) > n {
		lines = lines[len(lines)-n:]
	}
	for i, line := range lines {
		lines[i] = sanitize(strings.TrimSuffix(line, "\r"))
	}
	return lines
}

// sanitize makes one line safe to draw. Invalid UTF-8 becomes U+FFFD, a tab
// becomes a space, and any other control character (an escape byte, DEL, a C1
// control) becomes U+FFFD, so no raw escape sequence reaches the terminal.
func sanitize(s string) string {
	s = strings.ToValidUTF8(s, "�")
	return strings.Map(func(r rune) rune {
		switch {
		case r == '\t':
			return ' '
		case unicode.IsControl(r):
			return '�'
		}
		return r
	}, s)
}
