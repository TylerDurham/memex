package filesystem

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_NewRepo(t *testing.T) {
	path, err := filepath.Abs("../../testdata/processor/markdown.md")
	if err != nil {
		t.Fatalf("could not load sample file: %v", err)
	}

	fi, err := os.Stat(path)
	if err != nil {
		t.Fatalf("could not stat '%s': %v", path, err)
	}

	repo, err := NewRepo(path)

	if err != nil {
		t.Fatalf("could not load repo: %v", err)
	}

	assert.Equal(t, repo.Name(), filepath.Base(path))
	assert.Equal(t,repo.absPath, path)
	assert.Equal(t, repo.IsDir(), fi.IsDir())
	assert.Equal(t, repo.ModTime(), fi.ModTime())
	assert.Equal(t, repo.Mode(), fi.Mode())
}
