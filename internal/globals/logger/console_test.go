package logger

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"
)

// newTestConsole returns a logger writing through a consoleHandler into the returned buffer.
func newTestConsole(level slog.Level, color bool) (*slog.Logger, *bytes.Buffer) {
	var buf bytes.Buffer
	return slog.New(newConsoleHandler(&buf, level, color)), &buf
}

func TestConsoleFormat(t *testing.T) {
	l, buf := newTestConsole(slog.LevelDebug, false)
	l.Debug("walking", "path", "a.md")
	l.Info("indexed", "docs", 3)
	l.Warn("careful")
	l.Error("failed", "err", "boom")

	want := "DEBUG walking path=a.md\n" +
		"INFO  indexed docs=3\n" +
		"WARN  careful\n" +
		"ERROR failed err=boom\n"
	if got := buf.String(); got != want {
		t.Errorf("output =\n%s\nwant\n%s", got, want)
	}
}

func TestConsoleLevel(t *testing.T) {
	var level slog.LevelVar
	var buf bytes.Buffer
	l := slog.New(newConsoleHandler(&buf, &level, false))

	l.Debug("hidden")
	level.Set(slog.LevelDebug)
	l.Debug("shown")

	if got := buf.String(); got != "DEBUG shown\n" {
		t.Errorf("output = %q, want only the Debug record logged after lowering the level", got)
	}
}

func TestConsoleQuotesKeysAndValues(t *testing.T) {
	l, buf := newTestConsole(slog.LevelInfo, false)
	l.Info("m", "space", "Error Handling.md", "empty", "", "eq", "a=b", "esc", "\x1b[31mred", "config dir", "/tmp")

	want := `INFO  m space="Error Handling.md" empty="" eq="a=b" esc="\x1b[31mred" "config dir"=/tmp` + "\n"
	if got := buf.String(); got != want {
		t.Errorf("output = %q, want %q", got, want)
	}
}

func TestConsoleGroupsAndWithAttrs(t *testing.T) {
	l, buf := newTestConsole(slog.LevelInfo, false)
	l = l.With("repo", "vault").WithGroup("doc").With("ext", ".md")
	l.Info("indexed",
		"path", "a.md",
		slog.Group("size", "bytes", 10),
		slog.Group("empty"),
		slog.Group("", "inline", true),
		slog.Attr{},
	)

	want := "INFO  indexed repo=vault doc.ext=.md doc.path=a.md doc.size.bytes=10 doc.inline=true\n"
	if got := buf.String(); got != want {
		t.Errorf("output = %q, want %q", got, want)
	}
}

func TestConsoleWithAttrsDoesNotLeak(t *testing.T) {
	l, buf := newTestConsole(slog.LevelInfo, false)
	base := l.With("a", 1)
	base.With("b", 2).Info("first")
	base.With("c", 3).Info("second")

	want := "INFO  first a=1 b=2\nINFO  second a=1 c=3\n"
	if got := buf.String(); got != want {
		t.Errorf("output = %q, want %q", got, want)
	}
}

func TestConsoleColor(t *testing.T) {
	l, buf := newTestConsole(slog.LevelDebug, true)
	l.Warn("careful", "path", "a.md")

	want := ansiYellow + "WARN" + ansiReset + "  careful " + ansiDim + "path=" + ansiReset + "a.md\n"
	if got := buf.String(); got != want {
		t.Errorf("output = %q, want %q", got, want)
	}

	for level, code := range map[slog.Level]string{
		slog.LevelDebug: ansiGray,
		slog.LevelInfo:  ansiCyan,
		slog.LevelError: ansiRed,
	} {
		buf.Reset()
		l.Log(t.Context(), level, "m")
		if !strings.HasPrefix(buf.String(), code+level.String()+ansiReset) {
			t.Errorf("%s output = %q, want it colored with %q", level, buf.String(), code)
		}
	}
}

func TestUseColorNoColor(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	if useColor(nil) {
		t.Error("useColor with NO_COLOR set = true, want false")
	}
}
