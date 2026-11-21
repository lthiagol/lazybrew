package brew

import (
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"sync/atomic"
)

var logger atomic.Value

func init() {
	ResetLogger()
}

func Logger() *slog.Logger {
	return logger.Load().(*slog.Logger)
}

// ResetLogger puts the brew logger back to its default, which discards
// everything below warnings. It undoes EnableFileLogging; tests call it when
// they are done. Nothing in production calls it.
func ResetLogger() {
	logger.Store(slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{
		Level: slog.LevelWarn,
	})))
}

func EnableFileLogging(path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	logger.Store(slog.New(slog.NewTextHandler(f, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})))
	return nil
}
