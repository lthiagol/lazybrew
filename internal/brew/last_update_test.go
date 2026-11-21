package brew

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/lthiagol/lazybrew/internal/brew/brewtest"
)

// repoRunner is a runner whose `brew --repository` prints repo and counts the
// calls it gets.
func repoRunner(repo string, calls *int) *brewtest.MockRunner {
	r := brewtest.NewMockRunner()
	r.ExecuteFn = func(ctx context.Context, args ...string) ([]byte, error) {
		if len(args) == 1 && args[0] == "--repository" {
			*calls++
			return []byte(repo + "\n"), nil
		}
		return nil, brewtest.ErrTest
	}
	return r
}

func TestLastUpdateReadsFetchHeadMtime(t *testing.T) {
	repo := t.TempDir()
	if err := os.MkdirAll(filepath.Join(repo, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	fetchHead := filepath.Join(repo, ".git", "FETCH_HEAD")
	if err := os.WriteFile(fetchHead, []byte("abc\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	want := time.Date(2026, 11, 20, 9, 30, 0, 0, time.Local)
	if err := os.Chtimes(fetchHead, want, want); err != nil {
		t.Fatal(err)
	}

	var calls int
	diag := NewDiagnosticsReader(repoRunner(repo, &calls), NewCache(time.Minute))
	got, err := diag.LastUpdate(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if d := got.Sub(want); d > time.Second || d < -time.Second {
		t.Errorf("LastUpdate = %s, want %s", got, want)
	}
}

func TestLastUpdateWithoutFetchHeadIsZero(t *testing.T) {
	var calls int
	diag := NewDiagnosticsReader(repoRunner(t.TempDir(), &calls), NewCache(time.Minute))
	got, err := diag.LastUpdate(context.Background())
	if err != nil {
		t.Fatalf("a brew with no FETCH_HEAD is not an error: %v", err)
	}
	if !got.IsZero() {
		t.Errorf("LastUpdate = %s, want the zero time", got)
	}
}

func TestLastUpdateRepositoryErrorIsReturned(t *testing.T) {
	r := brewtest.NewMockRunner()
	r.ExecuteFn = func(ctx context.Context, args ...string) ([]byte, error) {
		return nil, brewtest.ErrTest
	}
	diag := NewDiagnosticsReader(r, NewCache(time.Minute))
	if _, err := diag.LastUpdate(context.Background()); err == nil {
		t.Fatal("LastUpdate with brew --repository failing returned no error")
	}
}

func TestLastUpdateReadsRepositoryOnce(t *testing.T) {
	var calls int
	diag := NewDiagnosticsReader(repoRunner(t.TempDir(), &calls), NewCache(time.Minute))
	for range 3 {
		if _, err := diag.LastUpdate(context.Background()); err != nil {
			t.Fatal(err)
		}
	}
	if calls != 1 {
		t.Errorf("brew --repository ran %d times, want 1", calls)
	}
}
