package logtail

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// writeFile puts content in a new file under the test's temp directory and
// returns its path.
func writeFile(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "log.txt")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func mustRead(t *testing.T, path string, n int) Tail {
	t.Helper()
	tail, err := Read(path, n)
	if err != nil {
		t.Fatalf("Read(%q, %d): %v", path, n, err)
	}
	return tail
}

func TestReadShortFile(t *testing.T) {
	content := "a\nb\nc\n"
	tail := mustRead(t, writeFile(t, content), 10)
	if want := []string{"a", "b", "c"}; !reflect.DeepEqual(tail.Lines, want) {
		t.Errorf("Lines = %q, want %q", tail.Lines, want)
	}
	if tail.Size != int64(len(content)) {
		t.Errorf("Size = %d, want %d", tail.Size, len(content))
	}
	if tail.ModTime.IsZero() {
		t.Error("ModTime is zero, want the file's modification time")
	}
}

func TestReadKeepsLastNLines(t *testing.T) {
	tail := mustRead(t, writeFile(t, "a\nb\nc\nd\ne\n"), 2)
	if want := []string{"d", "e"}; !reflect.DeepEqual(tail.Lines, want) {
		t.Errorf("Lines = %q, want %q", tail.Lines, want)
	}
}

// A file bigger than the byte bound is read from its end only: the first,
// partial line is dropped, and the lines kept are whole and are the file's last.
func TestReadBoundedByBytes(t *testing.T) {
	var b strings.Builder
	for i := range 100 {
		fmt.Fprintf(&b, "line-%03d\n", i)
	}
	path := writeFile(t, b.String())

	// An 8-byte bound leaves the tail of the last line, then the lines whole.
	tail, err := read(path, 1000, 40)
	if err != nil {
		t.Fatal(err)
	}
	if len(tail.Lines) == 0 || len(tail.Lines) >= 100 {
		t.Fatalf("got %d lines, want a bounded tail", len(tail.Lines))
	}
	if last := tail.Lines[len(tail.Lines)-1]; last != "line-099" {
		t.Errorf("last line = %q, want line-099", last)
	}
	for _, line := range tail.Lines {
		if len(line) != len("line-000") {
			t.Errorf("a partial line was kept: %q", line)
		}
	}
	if tail.Size != int64(b.Len()) {
		t.Errorf("Size = %d, want the whole file's %d", tail.Size, b.Len())
	}
}

// The public Read bounds at MaxBytes: a file several times that size keeps
// only its end, and the first line is gone.
func TestReadNeverReadsWholeFile(t *testing.T) {
	big := "first\n" + strings.Repeat("x", MaxBytes*3) + "\nlast\n"
	tail := mustRead(t, writeFile(t, big), 100000)
	for _, line := range tail.Lines {
		if line == "first" {
			t.Fatal("the first line of a file past the bound was read")
		}
	}
	if last := tail.Lines[len(tail.Lines)-1]; last != "last" {
		t.Errorf("last line = %q, want last", last)
	}
	if tail.Size != int64(len(big)) {
		t.Errorf("Size = %d, want %d", tail.Size, len(big))
	}
}

// A read that starts exactly after a newline starts a line: that line is kept
// whole. Here the bound leaves "bbb" first, with the newline before it.
func TestReadKeepsLineAtBoundaryAfterNewline(t *testing.T) {
	content := "aaa\nbbb\nccc\n"
	tail, err := read(writeFile(t, content), 10, 8)
	if err != nil {
		t.Fatal(err)
	}
	if want := []string{"bbb", "ccc"}; !reflect.DeepEqual(tail.Lines, want) {
		t.Errorf("Lines = %q, want %q", tail.Lines, want)
	}
}

// A file that is exactly the bound starts at its first byte, so nothing is dropped.
func TestReadExactlyTheBoundDropsNothing(t *testing.T) {
	content := "one\ntwo\nthree\n"
	tail, err := read(writeFile(t, content), 10, int64(len(content)))
	if err != nil {
		t.Fatal(err)
	}
	if want := []string{"one", "two", "three"}; !reflect.DeepEqual(tail.Lines, want) {
		t.Errorf("Lines = %q, want %q", tail.Lines, want)
	}
}

func TestReadNoTrailingNewline(t *testing.T) {
	tail := mustRead(t, writeFile(t, "a\nb"), 10)
	if want := []string{"a", "b"}; !reflect.DeepEqual(tail.Lines, want) {
		t.Errorf("Lines = %q, want %q", tail.Lines, want)
	}
}

func TestReadEmptyFile(t *testing.T) {
	tail := mustRead(t, writeFile(t, ""), 10)
	if len(tail.Lines) != 0 {
		t.Errorf("Lines = %q, want none", tail.Lines)
	}
	if tail.Size != 0 {
		t.Errorf("Size = %d, want 0", tail.Size)
	}
}

// A file of one newline has one empty line, unlike an empty file.
func TestReadOnlyNewlineIsOneEmptyLine(t *testing.T) {
	tail := mustRead(t, writeFile(t, "\n"), 10)
	if want := []string{""}; !reflect.DeepEqual(tail.Lines, want) {
		t.Errorf("Lines = %q, want %q", tail.Lines, want)
	}
}

func TestReadMissingFile(t *testing.T) {
	_, err := Read(filepath.Join(t.TempDir(), "absent.log"), 10)
	if !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("err = %v, want fs.ErrNotExist", err)
	}
}

func TestReadEmptyPath(t *testing.T) {
	_, err := Read("", 10)
	if !errors.Is(err, ErrNoPath) {
		t.Errorf("err = %v, want ErrNoPath", err)
	}
}

// A read error that is not a missing file is its own error, never ErrNoPath or
// ErrNotExist. A directory cannot be read as a file.
func TestReadOtherErrorIsNeither(t *testing.T) {
	_, err := Read(t.TempDir(), 10)
	if err == nil {
		t.Fatal("reading a directory gave no error")
	}
	if errors.Is(err, fs.ErrNotExist) || errors.Is(err, ErrNoPath) {
		t.Errorf("err = %v, want an error of its own", err)
	}
}

func TestReadCRLF(t *testing.T) {
	tail := mustRead(t, writeFile(t, "a\r\nb\r\n"), 10)
	if want := []string{"a", "b"}; !reflect.DeepEqual(tail.Lines, want) {
		t.Errorf("Lines = %q, want %q", tail.Lines, want)
	}
}

func TestReadZeroLinesWanted(t *testing.T) {
	tail := mustRead(t, writeFile(t, "a\nb\n"), 0)
	if len(tail.Lines) != 0 {
		t.Errorf("Lines = %q, want none", tail.Lines)
	}
}

// Escape sequences, other control bytes and invalid UTF-8 never reach the
// lines raw: the escape byte and controls become U+FFFD, a tab a space, and a
// run of invalid bytes one U+FFFD.
func TestReadMakesLinesSafeToDraw(t *testing.T) {
	tail := mustRead(t, writeFile(t, "ok\x1b[31mred\x07\xff\xfe\tend\x7f\n"), 10)
	want := []string{"ok�[31mred�� end�"}
	if !reflect.DeepEqual(tail.Lines, want) {
		t.Fatalf("Lines = %q, want %q", tail.Lines, want)
	}
	for _, r := range tail.Lines[0] {
		if r < 0x20 || r == 0x7f {
			t.Errorf("a control rune %U reached the line", r)
		}
	}
}
