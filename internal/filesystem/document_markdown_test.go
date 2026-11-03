package filesystem_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/TylerDurham/memex/internal/testutil/obsidian"
	"github.com/stretchr/testify/assert"
)

func Test_NewDoc(t *testing.T) {

	repo := obsidian.GetTestRepo(t)
	path := obsidian.GetTestDocumentPath(t)
	doc := obsidian.GetTestDocument(t, repo)

	fi, err := os.Stat(path)
	if err != nil {
		t.Fatalf("")
	}
	assert.Equal(t, doc.Name(), filepath.Base(path))
	assert.Equal(t, doc.AbsPath(), path)
	assert.Equal(t, doc.IsDir(), fi.IsDir())
	assert.Equal(t, doc.ModTime(), fi.ModTime())
	assert.Equal(t, doc.Mode(), fi.Mode())
}
