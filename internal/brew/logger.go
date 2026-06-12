package brew

import (
	"io"
	"log/slog"
	"os"
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
