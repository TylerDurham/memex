// Package logger
package logger

import (
	"log/slog"
	"os"
	"path/filepath"
)

// ConsoleLevel controls the verbosity of console output at runtime. It
// defaults to Info, which hides Debug output; the -v/--verbose flag on the
// root command lowers it to Debug via ConsoleLevel.Set(slog.LevelDebug). Being a
// *slog.LevelVar (rather than a plain slog.Level) means the change takes
// effect immediately for the already-constructed logger below — no need to
// rebuild the handler.
var ConsoleLevel = new(slog.LevelVar)

// FileLevel controls the verbosity of the log file opened by OpenFile. It
// defaults to Debug, so the file has the full detail whether or not -v was given.
var FileLevel = new(slog.LevelVar)

func init() {
	FileLevel.Set(slog.LevelDebug)
}

// console writes human-readable, colored lines to stderr; see consoleHandler.
var console slog.Handler = newConsoleHandler(os.Stderr, ConsoleLevel, useColor(os.Stderr))

var sl = slog.New(console)

// file is the log file opened by OpenFile, or nil if there is none.
var file *os.File

// OpenFile appends logs to path as JSON lines, as well as writing them to the
// console. It creates path's directory if needed and closes any log file opened
// before. It isn't safe to call while other goroutines are logging.
func OpenFile(path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return err
	}
	Close()
	file = f
	sl = slog.New(slog.NewMultiHandler(
		console,
		slog.NewJSONHandler(f, &slog.HandlerOptions{Level: FileLevel}),
	))
	return nil
}

// Close stops logging to the file opened by OpenFile, if any, and closes it.
// Logs still go to the console.
func Close() error {
	sl = slog.New(console)
	if file == nil {
		return nil
	}
	err := file.Close()
	file = nil
	return err
}

func Debug(msg string, args ...any) {
	sl.Debug(msg, args...)
}

func Info(msg string, args ...any) {
	sl.Info(msg, args...)
}

func Warn(msg string, args ...any) {
	sl.Warn(msg, args...)
}

func Error(msg string, args ...any) {
	sl.Error(msg, args...)
}
