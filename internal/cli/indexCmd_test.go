package cli

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/TylerDurham/memex/internal/repo"
)

// vaultFixture is the sample Obsidian vault in testdata.
var vaultFixture = filepath.Join("..", "..", "testdata", "repos", "obsidian")

// initVault creates a repo named vault in a new config directory, pointing at the
// fixture vault, and returns the config directory.
func initVault(t *testing.T, application string) string {
	t.Helper()
	configDir := t.TempDir()
	if _, err := repo.Init(configDir, repo.InitOptions{Directory: vaultFixture, Name: "vault", Application: application}); err != nil {
		t.Fatalf("Init: %v", err)
	}
	return configDir
}

// runIndex runs "memex index" with args and returns what it printed to stdout and logged.
func runIndex(t *testing.T, args ...string) (stdout, logged string, err error) {
	t.Helper()
	stdout, logged = captureOutput(t, func() {
		_, _, err = runCmd(t, append([]string{"index"}, args...)...)
	})
	return stdout, logged, err
}

func TestIndexVault(t *testing.T) {
	// TODO: These tests are too brittle
	t.Skip("skip and make less brittle")
	configDir := initVault(t, "obsidian")

	out, logged, err := runIndex(t, "--config-dir", configDir, "vault")
	if err != nil {
		t.Fatalf("index: %v", err)
	}

	vault, err := filepath.Abs(vaultFixture)
	if err != nil {
		t.Fatal(err)
	}

	// Extensions match case-insensitively, so SHOUTING.MD is indexed too.
	for _, indexed := range []string{"Welcome.md", filepath.Join("Inbox", "SHOUTING.MD")} {
		if want := "file-indexed " + filepath.Join(vault, indexed) + "\n"; !strings.Contains(out, want) {
			t.Errorf("index output doesn't report %s as indexed:\n%s", indexed, out)
		}
	}

	for _, skipped := range []string{"Should Not Be Indexed.md", "notes.txt", "architecture.canvas"} {
		if strings.Contains(out, skipped) {
			t.Errorf("index output reports %s, which should be skipped:\n%s", skipped, out)
		}
	}

	if !strings.Contains(logged, "DocsPrepared:") {
		t.Errorf("index didn't log its stats:\n%s", logged)
	}
	if strings.Contains(out, "{") {
		t.Errorf("index without -v printed document JSON:\n%s", out)
	}
}

func TestIndexVerbosePrintsDocJSON(t *testing.T) {
	// TODO: Test needs to be re-writing
	t.Skip("needs to be re-written")
	configDir := initVault(t, "obsidian")

	out, _, err := runIndex(t, "-v", "--config-dir", configDir, "vault")
	if err != nil {
		t.Fatalf("index -v: %v", err)
	}
	if !strings.Contains(out, `"path":`) || !strings.Contains(out, "Welcome.md") {
		t.Errorf("index -v output has no document JSON:\n%s", out)
	}
}

func TestIndexWritesLogFile(t *testing.T) {
	// TODO: Test needs to be re-writing
	t.Skip("needs to be re-written")
	configDir := initVault(t, "obsidian")

	// Debug records reach the log file even without -v.
	if _, _, err := runIndex(t, "--config-dir", configDir, "vault"); err != nil {
		t.Fatalf("index: %v", err)
	}
	data, err := os.ReadFile(filepath.Join(configDir, LogFile))
	if err != nil {
		t.Fatalf("read log file: %v", err)
	}
	if !strings.Contains(string(data), `"level":"DEBUG"`) {
		t.Errorf("log file has no debug records:\n%s", data)
	}
}

func TestIndexRepoNotFound(t *testing.T) {
	_, _, err := runIndex(t, "--config-dir", t.TempDir(), "missing")
	if !errors.Is(err, repo.ErrNotFound) {
		t.Errorf("index error = %v, want ErrNotFound", err)
	}
}

func TestIndexUnknownApplication(t *testing.T) {
	configDir := initVault(t, "emacs")

	_, _, err := runIndex(t, "--config-dir", configDir, "vault")
	if err == nil || !strings.Contains(err.Error(), "unknown application") {
		t.Errorf("index error = %v, want unknown application", err)
	}
}

func TestIndexArgs(t *testing.T) {
	configDir := t.TempDir()
	for _, args := range [][]string{
		{"--config-dir", configDir},
		{"--config-dir", configDir, "a", "b"},
	} {
		if _, _, err := runIndex(t, args...); err == nil {
			t.Errorf("index %v succeeded, want argument error", args)
		}
	}
}
