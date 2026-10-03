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

// TestRepoConfigInfo loads a repo for testing. It initializes a new instance within
// a temporary directory loads that instance.
// NOTE: If there are any errors initializing or loading the instance, it fails all tests
// and shuts the test runner down.
func TestRepoConfigInfo(t *testing.T, appType string) repo.RepoConfigInfo {
	// testrunner should not treat as a test
	t.Helper()

	// generate a real, but temporary config directory
	tmpConfigDir := filepath.Join(t.TempDir(), ".config", "memex")

	// we need a repo name; generate it
	newRepoName := appType + "-vault"

	// Calculate path to actual repo
	path, err := filepath.Abs("../../testdata/repos/" + appType)
	if err != nil {
		t.Fatalf("could not find testdata directory '%s': %v", testDataDir, err)
	}

	// initialize our repo
	_, err = repo.InitWithConfigDir(tmpConfigDir, repo.InitOptions{
		Application: appType,
		Name:        newRepoName,
		Directory:   path,
	})

	if err != nil {
		t.Fatalf("could not init repo: %v", err)
	}

	// load our repo, by name
	repoConfig, err := repo.LoadWithConfigDir(tmpConfigDir, newRepoName)

	if err != nil {
		t.Fatalf("could not load repo: %v", err)
	}
	return repoConfig
}
