package logger

import (
	"context"
	"io"
	"log/slog"
	"os"
	"slices"
	"strconv"
	"strings"
	"sync"
	"unicode"
)

// ANSI escape codes used by consoleHandler.
const (
	ansiReset  = "\x1b[0m"
	ansiDim    = "\x1b[2m"
	ansiGray   = "\x1b[1;90m"
	ansiCyan   = "\x1b[1;36m"
	ansiYellow = "\x1b[1;33m"
	ansiRed    = "\x1b[1;31m"
)

// consoleHandler is a slog.Handler for people reading a terminal. It writes one
// line per record:
//
//	WARN  could not parse frontmatter path=Inbox/a.md line=3
//
// with the level colored by severity and attribute keys dimmed when color is on.
// It leaves out the time; the log file has it.
type consoleHandler struct {
	w     io.Writer
	mu    *sync.Mutex // shared by handlers derived with WithAttrs and WithGroup
	level slog.Leveler
	color bool

	// prefix is the group path ("a.b.") that qualifies keys added from now on.
	prefix string
	// attrs holds attributes from WithAttrs, already formatted.
	attrs []byte
}

func newConsoleHandler(w io.Writer, level slog.Leveler, color bool) *consoleHandler {
	return &consoleHandler{w: w, mu: &sync.Mutex{}, level: level, color: color}
}

// useColor reports whether output to f should be colored: f must be a terminal,
// and neither NO_COLOR (https://no-color.org) nor TERM=dumb may be set.
func useColor(f *os.File) bool {
	if os.Getenv("NO_COLOR") != "" || os.Getenv("TERM") == "dumb" {
		return false
	}
	fi, err := f.Stat()
	return err == nil && fi.Mode()&os.ModeCharDevice != 0
}

func (h *consoleHandler) Enabled(_ context.Context, l slog.Level) bool {
	return l >= h.level.Level()
}

func (h *consoleHandler) Handle(_ context.Context, r slog.Record) error {
	buf := make([]byte, 0, 256)
	buf = h.appendLevel(buf, r.Level)
	buf = append(buf, ' ')
	buf = append(buf, r.Message...)
	buf = append(buf, h.attrs...)
	r.Attrs(func(a slog.Attr) bool {
		buf = h.appendAttr(buf, h.prefix, a)
		return true
	})
	buf = append(buf, '\n')

	h.mu.Lock()
	defer h.mu.Unlock()
	_, err := h.w.Write(buf)
	return err
}

func (h *consoleHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	h2 := *h
	h2.attrs = slices.Clone(h.attrs)
	for _, a := range attrs {
		h2.attrs = h.appendAttr(h2.attrs, h.prefix, a)
	}
	return &h2
}

func (h *consoleHandler) WithGroup(name string) slog.Handler {
	if name == "" {
		return h
	}
	h2 := *h
	h2.prefix = h.prefix + name + "."
	return &h2
}

// appendLevel appends the level, padded to line up the messages that follow it.
func (h *consoleHandler) appendLevel(buf []byte, l slog.Level) []byte {
	s := l.String()
	if !h.color {
		return append(buf, padLevel(s)...)
	}
	var code string
	switch {
	case l < slog.LevelInfo:
		code = ansiGray
	case l < slog.LevelWarn:
		code = ansiCyan
	case l < slog.LevelError:
		code = ansiYellow
	default:
		code = ansiRed
	}
	// Pad outside the escape codes so the padding isn't counted as color.
	buf = append(buf, code...)
	buf = append(buf, s...)
	buf = append(buf, ansiReset...)
	return append(buf, strings.Repeat(" ", len(padLevel(s))-len(s))...)
}

// padLevel pads s to the width of the longest standard level name, DEBUG and ERROR.
func padLevel(s string) string {
	const width = len("ERROR")
	if len(s) >= width {
		return s
	}
	return s + strings.Repeat(" ", width-len(s))
}

// appendAttr appends " key=value" for a, with group attributes flattened into
// dotted keys. It skips empty attributes and empty groups, as slog.Handler requires.
func (h *consoleHandler) appendAttr(buf []byte, prefix string, a slog.Attr) []byte {
	a.Value = a.Value.Resolve()
	if a.Equal(slog.Attr{}) {
		return buf
	}
	if a.Value.Kind() == slog.KindGroup {
		if a.Key != "" {
			prefix += a.Key + "."
		}
		for _, ga := range a.Value.Group() {
			buf = h.appendAttr(buf, prefix, ga)
		}
		return buf
	}

	buf = append(buf, ' ')
	if h.color {
		buf = append(buf, ansiDim...)
	}
	buf = appendQuoted(buf, prefix+a.Key)
	buf = append(buf, '=')
	if h.color {
		buf = append(buf, ansiReset...)
	}
	return appendQuoted(buf, a.Value.String())
}

// appendQuoted appends a key or value, quoted if it's empty or has spaces, '=',
// '"', or unprintable characters. Quoting also keeps escape codes in keys and
// values from reaching the terminal.
func appendQuoted(buf []byte, s string) []byte {
	needsQuote := s == "" || strings.ContainsFunc(s, func(r rune) bool {
		return unicode.IsSpace(r) || r == '=' || r == '"' || !unicode.IsPrint(r)
	})
	if needsQuote {
		return strconv.AppendQuote(buf, s)
	}
	return append(buf, s...)
}
