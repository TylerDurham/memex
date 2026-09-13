package globals

import (
	"log/slog"
	"os"
)

// LogLevel controls the verbosity of the package logger at runtime. It
// defaults to Info, which hides Debug output; the -v/--verbose flag on the
// root command lowers it to Debug via LogLevel.Set(slog.LevelDebug). Being a
// *slog.LevelVar (rather than a plain slog.Level) means the change takes
// effect immediately for the already-constructed logger below — no need to
// rebuild the handler.
var LogLevel = new(slog.LevelVar)

var h = slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
	Level: LogLevel,
})

var sl = slog.New(h)

func newLogger() *slog.Logger {
	return sl
}
