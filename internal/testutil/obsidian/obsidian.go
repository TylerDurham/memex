// Package obsidian provides test helpers that load the sample Obsidian vault
// and markdown fixtures stored under testdata/.
//
// The fixture paths are resolved relative to the working directory of the
// running test. "go test" sets that to the directory of the package under
// test, so these helpers only find testdata/ when called from a package two
// levels below the module root (for example internal/processors).
//
// This package imports internal/filesystem. Tests inside internal/filesystem
// that use these helpers must therefore be declared as package
// filesystem_test; an in-package test (package filesystem) importing this
// package creates an import cycle.
//
// Every helper calls t.Fatalf on error, which marks the calling test as failed
// and stops it immediately. Helpers return only valid values, so callers do
// not need to check for errors.
package obsidian

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/TylerDurham/memex/internal/filesystem/document"
	"github.com/TylerDurham/memex/internal/filesystem/repo"
	"github.com/TylerDurham/memex/internal/testutil"
)

// Fixture locations, relative to the working directory of the running test.
const (
	testRepoRelPath     = "../../../testdata/repos/obsidian/"
	testDocumentRelPath = "../../../testdata/processor/markdown.md"
)

// workingDir returns the test's working directory for use in failure
// messages, or a placeholder if it cannot be determined.
func workingDir() string {
	wd, err := os.Getwd()
	if err != nil {
		return "<unknown: " + err.Error() + ">"
	}
	return wd
}

// InitTestRepo returns a filesystem.Repo for the sample Obsidian vault at
// testdata/repos/obsidian. It fails the test if the path cannot be resolved
// or the vault cannot be opened, for example because it does not exist.
func InitTestRepo(t *testing.T, path string, strategy string) repo.Repo {
	t.Helper()
	testutil.IsolateConfig(t)

	repo, _ := repo.Init(path, strategy)

	return repo
}

func InitTestRepoWithErr(t *testing.T, path string, strategy string) (repo.Repo, error) {
	t.Helper()
	testutil.IsolateConfig(t)

	repo, err := repo.Init(path, strategy)

	return repo, err
}

func GetTestRepoPath(t *testing.T) string {
	t.Helper()
	path, err := filepath.Abs(testRepoRelPath)
	if err != nil {
		t.Fatalf("obsidian.GetTestRepoPath: could not make sample vault path %q absolute (working dir %q): %v",
			testRepoRelPath, workingDir(), err)
	}

	return path
}

// GetTestDocumentPath returns the absolute path to the sample markdown
// document, testdata/processor/markdown.md. That file is a symlink to a diary
// entry in the sample vault. The returned path is the symlink itself, not its
// resolved target. Like GetTestRepoPath, it does not check that the file
// exists.
func GetTestDocumentPath(t *testing.T) string {
	t.Helper()
	path, err := filepath.Abs(testDocumentRelPath)
	if err != nil {
		t.Fatalf("obsidian.GetTestDocumentPath: could not make sample document path %q absolute (working dir %q): %v",
			testDocumentRelPath, workingDir(), err)
	}

	return path
}


// GetTestDocument returns the file at path as a filesystem.Document
// belonging to repo. Use it when a test needs a document other than the
// default sample document, such as another note in the sample vault or a file
// written to a t.TempDir(). path is passed to filesystem.NewDocument
// unchanged and must be absolute: NewDocument computes the document's
// RelPath against repo.AbsPath(), which fails for a relative path. The test
// fails if the file cannot be opened or its path cannot be made relative to
// the repo root.
func GetTestDocument(t *testing.T, repo repo.Repo, path string) document.Document {
	t.Helper()
	doc, err := document.NewDocument(repo, path)
	if err != nil {
		t.Fatalf("obsidian.GetTestDocumentWithPath: could not load document %q into repo %q (working dir %q): %v\n"+
			"hint: path must be absolute; build it with filepath.Join(repo.AbsPath(), ...) or under t.TempDir()",
			path, repo.AbsPath(), workingDir(), err)
	}

	return doc
}
