package globals

import (
	"log/slog"
	"os"
)

var h = slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
	Level: slog.LevelDebug,
})

var logger = *slog.New(h)

func Logger() *slog.Logger {
	return &logger
} 
