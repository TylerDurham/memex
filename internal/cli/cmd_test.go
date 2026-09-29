package cli

import (
	"bytes"
	"io"
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/TylerDurham/memex/internal/globals/logger"
)

// runCmd runs memex with args and returns what the command wrote to its stdout and
// stderr. Cobra doesn't reset flag variables between runs, so they're restored to their
// defaults first.
func runCmd(t *testing.T, args ...string) (stdout, stderr string, err error) {
	t.Helper()
	verbose = false
	formatJSON = false
	configDirFlag = ""
	repoInitName = ""
	repoInitApplication = "obsidian"
	logger.ConsoleLevel.Set(slog.LevelInfo)

	var out, errOut bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&errOut)
	rootCmd.SetArgs(args)
	t.Cleanup(func() {
		rootCmd.SetOut(nil)
		rootCmd.SetErr(nil)
		rootCmd.SetArgs(nil)
		logger.ConsoleLevel.Set(slog.LevelInfo)
		logger.Close()
	})
	err = rootCmd.Execute()
	return out.String(), errOut.String(), err
}

// readLogFile returns the contents of the log file in configDir.
func readLogFile(t *testing.T, configDir string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(configDir, LogFile))
	if err != nil {
		t.Fatalf("read log file: %v", err)
	}
	return string(data)
}

// captureOutput runs f and returns what it wrote to os.Stdout and to the standard
// logger, for commands that don't write through cmd.OutOrStdout.
func captureOutput(t *testing.T, f func()) (stdout, logged string) {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	var logBuf bytes.Buffer
	origStdout, origLog := os.Stdout, log.Writer()
	os.Stdout = w
	log.SetOutput(&logBuf)

	done := make(chan string)
	go func() {
		data, _ := io.ReadAll(r)
		done <- string(data)
	}()

	defer func() {
		os.Stdout = origStdout
		log.SetOutput(origLog)
	}()
	f()
	w.Close()
	return <-done, logBuf.String()
}
