package globals

import (
	"log/slog"
	"os"
)

var h = slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
	Level: slog.LevelDebug,
})

var sl = slog.New(h)

func newLogger() (*slog.Logger) {
	return sl
} 
