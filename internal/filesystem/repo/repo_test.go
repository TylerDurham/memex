package repo_test

import (
	"os"
	"testing"

	"github.com/TylerDurham/memex/internal/testutil/obsidian"
	"github.com/stretchr/testify/assert"
)

func Test_NewRepo(t *testing.T) {

	path := obsidian.GetTestRepoPath(t)
	repo := obsidian.GetTestRepo(t)
	fi, err := os.Stat(path)
	if err != nil {
		t.Fatalf("could not stat '%s': %v", path, err)
	}

	assert.Equal(t, repo.AbsPath(), path)
	assert.Equal(t, repo.IsDir(), fi.IsDir())
	assert.Equal(t, repo.Name(), fi.Name())
}
