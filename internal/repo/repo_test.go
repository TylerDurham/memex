package repo

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/TylerDurham/memex/internal/config"
)

func TestInitThenLoad(t *testing.T) {
	const wantVaultName = "My-Vault"
	const wantAppType = "obsidian"
	wantConfigDir := t.TempDir()
	dir := filepath.Join(t.TempDir(), wantVaultName)

	t.Setenv(config.EnvConfigDir, wantConfigDir)
	repoConfig, err := InitWithConfigDir(wantConfigDir, InitOptions{Name: wantVaultName, Directory: dir, Application: wantAppType})
	if err != nil {
		t.Fatalf("Init: %v", err)
	}
	if wantConfigFile := filepath.Join(wantConfigDir, ReposDir, wantVaultName, ConfigFile); repoConfig.ConfigFile != wantConfigFile {
		t.Errorf("Init path = %q, want %q", repoConfig, wantConfigFile)
	}

	cfg, err := LoadWithConfigDir(wantConfigDir, "My-Vault")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	want := RepoConfigInfo{
		AppType: "obsidian",
		Name:        "My-Vault",
		Directory:   dir,
		ConfigFile:  repoConfig.ConfigFile,
		Database:    filepath.Join(wantConfigDir, ReposDir, wantVaultName, "memex.db"),
	}
	if cfg != want {
		t.Errorf("Load = %+v, want %+v", cfg, want)
	}

	// Derived fields must not be written to config.yaml.
	data, err := os.ReadFile(repoConfig.ConfigFile)
	if err != nil {
		t.Fatal(err)
	}
	if s := string(data); strings.Contains(s, "config:") || strings.Contains(s, "db:") {
		t.Errorf("config.yaml contains derived fields:\n%s", s)
	}
}

func TestLoadNotFound(t *testing.T) {
	_, err := LoadWithConfigDir(t.TempDir(), "missing")
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("Load error = %v, want ErrNotFound", err)
	}
}

func TestLoadInvalidName(t *testing.T) {
	for _, name := range []string{"", ".", "..", "a/b", `a\b`} {
		if _, err := LoadWithConfigDir(t.TempDir(), name); err == nil || errors.Is(err, ErrNotFound) {
			t.Errorf("Load(%q) error = %v, want invalid name error", name, err)
		}
	}
}

func TestNames(t *testing.T) {
	configDir := t.TempDir()
	for _, name := range []string{"zeta", "alpha"} {
		if _, err := InitWithConfigDir(configDir, InitOptions{Directory: t.TempDir(), Name: name}); err != nil {
			t.Fatalf("Init(%q): %v", name, err)
		}
	}
	// Neither a stray file nor a directory without config.yaml is a repo.
	if err := os.WriteFile(filepath.Join(configDir, ReposDir, "stray.txt"), nil, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(configDir, ReposDir, "empty"), 0o755); err != nil {
		t.Fatal(err)
	}

	names, err := Names(configDir)
	if err != nil {
		t.Fatalf("Names: %v", err)
	}
	if len(names) != 2 || names[0] != "alpha" || names[1] != "zeta" {
		t.Errorf("Names = %v, want [alpha zeta]", names)
	}
}

func TestNamesMissingConfigDir(t *testing.T) {
	names, err := Names(filepath.Join(t.TempDir(), "does-not-exist"))
	if err != nil || len(names) != 0 {
		t.Errorf("Names = %v, %v; want no names and no error", names, err)
	}
}

func TestLoadBadYAML(t *testing.T) {
	configDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(configDir, ReposDir, "broken"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(configDir, ReposDir, "broken", ConfigFile), []byte("name: [unclosed"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadWithConfigDir(configDir, "broken"); err == nil {
		t.Error("Load of malformed config.yaml succeeded, want error")
	}
}
