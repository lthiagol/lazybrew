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
	logger.Store(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelWarn,
	})))
}

func Logger() *slog.Logger {
	return logger.Load().(*slog.Logger)
}

func SetDebug(enabled bool) {
	if enabled {
		logger.Store(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})))
	} else {
		logger.Store(slog.New(slog.NewTextHandler(io.Discard, nil)))
	}
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
	w := io.MultiWriter(os.Stderr, f)
	logger.Store(slog.New(slog.NewTextHandler(w, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})))
	return nil
}
