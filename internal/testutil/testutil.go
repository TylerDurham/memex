// Package testutil
package testutil

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/TylerDurham/memex/internal/repo"
)

const testDataDir = "../../testdata/repos"

func MustStat(t *testing.T, path string) os.FileInfo {
	stat, err := os.Stat(path)
	if err != nil {
		t.Fatalf("could not stat '%s': %s", path, err)
	}

	return stat
} 

// MustLoadRepoInfo loads a repo for testing.
func MustLoadRepoInfo(t *testing.T, appType string) (r *repo.RepoInfo) {
	t.Helper()
	name := appType + "-vault"
	configDir := filepath.Join(t.TempDir(), ".config", "memex")
	path, err := filepath.Abs("../../testdata/repos/" + appType)
	if err != nil {
		t.Fatalf("could not find testdata directory '%s': %v", testDataDir, err)
	}

	_, err = repo.Init(configDir, repo.InitOptions{
		Application: appType,
		Name:        name,
		Directory:   path,
	})

	if err != nil {
		t.Fatalf("could not init repo: %v", err)
	}

	r, err = repo.Load(configDir, name)

	if err != nil {
		t.Fatalf("could not load repo: %v", err)
	}
	return r
}
