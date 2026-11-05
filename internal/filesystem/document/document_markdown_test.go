package document_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/TylerDurham/memex/internal/filesystem/repo"
	"github.com/TylerDurham/memex/internal/testutil/obsidian"
	"github.com/stretchr/testify/assert"
)

func Test_NewDoc(t *testing.T) {
	repo := obsidian.InitTestRepo(t, obsidian.GetTestRepoPath(t), repo.AppStrategyObsidian)
	path := obsidian.GetTestDocumentPath(t)
	doc := obsidian.GetTestDocument(t, repo, path)

	fi, err := os.Stat(path)
	assert.NoErrorf(t, err, "could not stat sample document %q to compare against: %v", path, err)
	assert.Equal(t, doc.Name(), filepath.Base(path))
	assert.Equal(t, doc.AbsPath(), path)
	assert.Equal(t, doc.IsDir(), fi.IsDir())
	assert.Equal(t, doc.ModTime(), fi.ModTime())
	assert.Equal(t, doc.Mode(), fi.Mode())
}
