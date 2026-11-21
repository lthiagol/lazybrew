package brew

import (
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
)

var logger atomic.Value

// logMu guards logFile, the debug log file EnableFileLogging opened.
var (
	logMu   sync.Mutex
	logFile *os.File
)

func init() {
	ResetLogger()
}

func Logger() *slog.Logger {
	return logger.Load().(*slog.Logger)
}

// ResetLogger puts the brew logger back to its default, which discards
// everything below warnings, and closes the debug log file EnableFileLogging
// opened. init calls it for the default; main calls it when the TUI exits, so
// the file is closed before the process ends. Tests call it when they are done.
func ResetLogger() {
	logMu.Lock()
	defer logMu.Unlock()
	logger.Store(slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{
		Level: slog.LevelWarn,
	})))
	if logFile != nil {
		logFile.Close()
		logFile = nil
	}
}

// EnableFileLogging sends the brew logger's debug output to the file at path,
// creating it if needed. A file opened by an earlier call is closed.
func EnableFileLogging(path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	logMu.Lock()
	defer logMu.Unlock()
	if logFile != nil {
		logFile.Close()
	}
	logFile = f
	logger.Store(slog.New(slog.NewTextHandler(f, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})))
	return nil
}
