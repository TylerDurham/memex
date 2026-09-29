package logger

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// readLog returns the JSON records in the log file at path.
func readLog(t *testing.T, path string) []map[string]any {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	var records []map[string]any
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		var r map[string]any
		if err := json.Unmarshal(sc.Bytes(), &r); err != nil {
			t.Fatalf("log line isn't JSON: %v\n%s", err, sc.Text())
		}
		records = append(records, r)
	}
	if err := sc.Err(); err != nil {
		t.Fatal(err)
	}
	return records
}

func TestOpenFileLogsDebugAsJSON(t *testing.T) {
	path := filepath.Join(t.TempDir(), "logs", "memex.log")
	if err := OpenFile(path); err != nil {
		t.Fatalf("OpenFile: %v", err)
	}
	t.Cleanup(func() { Close() })

	// The console stays at Info, but the file still gets Debug records.
	Debug("walking", "path", "a.md")
	Warn("careful")
	if err := Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
	Debug("after close")

	records := readLog(t, path)
	if len(records) != 2 {
		t.Fatalf("log has %d records, want 2: %v", len(records), records)
	}
	if r := records[0]; r["level"] != "DEBUG" || r["msg"] != "walking" || r["path"] != "a.md" {
		t.Errorf("first record = %v, want DEBUG walking with path a.md", r)
	}
	if r := records[1]; r["level"] != "WARN" || r["msg"] != "careful" {
		t.Errorf("second record = %v, want WARN careful", r)
	}
}

func TestOpenFileAppends(t *testing.T) {
	path := filepath.Join(t.TempDir(), "memex.log")
	t.Cleanup(func() { Close() })
	for _, msg := range []string{"first", "second"} {
		if err := OpenFile(path); err != nil {
			t.Fatalf("OpenFile: %v", err)
		}
		Warn(msg)
	}
	Close()

	records := readLog(t, path)
	if len(records) != 2 || records[0]["msg"] != "first" || records[1]["msg"] != "second" {
		t.Errorf("log records = %v, want first then second", records)
	}
}

func TestOpenFileError(t *testing.T) {
	// A path under a regular file can't be created.
	blocker := filepath.Join(t.TempDir(), "file")
	if err := os.WriteFile(blocker, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := OpenFile(filepath.Join(blocker, "memex.log")); err == nil {
		Close()
		t.Error("OpenFile under a regular file succeeded, want error")
	}
}
