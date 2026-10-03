package cli

import (
	"path/filepath"
	"strings"
	"testing"
	"weak"

	"github.com/TylerDurham/memex/internal/config"
	"github.com/TylerDurham/memex/internal/repo"
)

// runRepoInit runs "memex repo init" with args and returns its stdout.
func runRepoInit(t *testing.T, args ...string) (string, error) {
	t.Helper()
	out, _, err := runCmd(t, append([]string{"repo", "init"}, args...)...)
	return out, err
}

func TestRepoInitDefaults(t *testing.T) {
	configDir := t.TempDir()
	dir := filepath.Join(t.TempDir(), "My-Vault")

	out, err := runRepoInit(t, "--config-dir", configDir, dir)
	if err != nil {
		t.Fatalf("repo init: %v", err)
	}
	if out != "" {
		t.Errorf("repo init without -v wrote output:\n%s", out)
	}

	cfg, err := repo.LoadWithConfigDir(configDir, "My-Vault")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Name != "My-Vault" || cfg.AppType != "obsidian" || cfg.Directory != dir {
		t.Errorf("Load = %+v, want name My-Vault, application obsidian, directory %q", cfg, dir)
	}
}

func TestRepoInitFlags(t *testing.T) {
	configDir := t.TempDir()
	dir := t.TempDir()

	if _, err := runRepoInit(t, "--config-dir", configDir, "-n", "notes", "-a", "markdown", dir); err != nil {
		t.Fatalf("repo init: %v", err)
	}

	cfg, err := repo.LoadWithConfigDir(configDir, "notes")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Name != "notes" || cfg.AppType != "markdown" || cfg.Directory != dir {
		t.Errorf("Load = %+v, want name notes, application markdown, directory %q", cfg, dir)
	}
}

func TestRepoInitVerbose(t *testing.T) {

	// TODO: rewrite to be less brittle
	t.Skip("rewrite to be less brittle")
	configDir := t.TempDir()

	out, err := runRepoInit(t, "-v", "--config-dir", configDir, "-n", "notes", t.TempDir())
	if err != nil {
		t.Fatalf("repo init: %v", err)
	}

	// Verbose details are Debug logs, not command output.
	if out != "" {
		t.Errorf("repo init -v wrote output:\n%s", out)
	}
	log := readLogFile(t, configDir)
	for _, want := range []string{
		`"config dir":"` + configDir + `"`,
		`"wrote":"` + filepath.Join(configDir, repo.ReposDir, "notes", repo.ConfigFile) + `"`,
	} {
		if !strings.Contains(log, want) {
			t.Errorf("log file missing %s:\n%s", want, log)
		}
	}
}

func TestRepoInitConfigDirFromEnv(t *testing.T) {
	configDir := t.TempDir()
	t.Setenv(config.EnvConfigDir, configDir)

	if _, err := runRepoInit(t, "-n", "notes", t.TempDir()); err != nil {
		t.Fatalf("repo init: %v", err)
	}
	if _, err := repo.LoadWithConfigDir(configDir, "notes"); err != nil {
		t.Errorf("Load from $%s: %v", config.EnvConfigDir, err)
	}
}

func TestRepoInitAlreadyExists(t *testing.T) {
	configDir := t.TempDir()
	dir := t.TempDir()

	if _, err := runRepoInit(t, "--config-dir", configDir, "-n", "notes", dir); err != nil {
		t.Fatalf("first repo init: %v", err)
	}
	_, err := runRepoInit(t, "--config-dir", configDir, "-n", "notes", dir)
	if err == nil || !strings.Contains(err.Error(), "already exists") {
		t.Errorf("second repo init error = %v, want already exists", err)
	}
}

func TestRepoInitArgs(t *testing.T) {
	configDir := t.TempDir()
	for _, args := range [][]string{
		{"--config-dir", configDir},
		{"--config-dir", configDir, t.TempDir(), t.TempDir()},
	} {
		if _, err := runRepoInit(t, args...); err == nil {
			t.Errorf("repo init %v succeeded, want argument error", args)
		}
	}
}

func TestRepoInitInvalidName(t *testing.T) {
	if _, err := runRepoInit(t, "--config-dir", t.TempDir(), "-n", "a/b", t.TempDir()); err == nil {
		t.Error("repo init with name a/b succeeded, want invalid name error")
	}
}
