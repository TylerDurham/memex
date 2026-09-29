package cli

import (
	"bytes"
	"io"
	"log"
	"log/slog"
	"os"
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
	logger.LogLevel.Set(slog.LevelInfo)

	var out, errOut bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&errOut)
	rootCmd.SetArgs(args)
	t.Cleanup(func() {
		rootCmd.SetOut(nil)
		rootCmd.SetErr(nil)
		rootCmd.SetArgs(nil)
		logger.LogLevel.Set(slog.LevelInfo)
	})
	err = rootCmd.Execute()
	return out.String(), errOut.String(), err
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
