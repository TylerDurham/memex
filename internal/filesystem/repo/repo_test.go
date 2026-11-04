package repo_test

import (
	"errors"
	"os"
	"testing"

	"github.com/TylerDurham/memex/internal/filesystem/repo"
	"github.com/TylerDurham/memex/internal/testutil/obsidian"
	"github.com/stretchr/testify/assert"
)

// func ExampleInit() {
// 	path := "testdata/"
//
// 	repo, err := repo.Init(path)
// 	if err != nil {
// 		_ = fmt.Errorf("could not initialize repo: %w", err)
// 	}
//
// 	msg := fmt.Sprintf("Name: '%s'", repo.Name())
// 	fmt.Println(msg)
//
// 	// Output:
// 	// Name: 'testdata'
// }

func Test_InitRepo(t *testing.T) {

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

func Test_NewRepo_InvalidStrategy(t *testing.T) {
	name := "BAD_STRATEGY"
	path := obsidian.GetTestRepoPath(t)
	if _, err := obsidian.InitTestRepo(t, path, name); err == nil || !errors.Is(err, repo.ErrInvalidArgument) {
		t.Errorf("InitWithApp(%q, %q) error = %v, want ErrInvalidArgument", path, name, err)
	}
}

func Test_InitRepo_ErrNotADirectory(t *testing.T) {
	path := "../../../testdata/processor/markdown.md"
	if _, err := obsidian.InitTestRepo(t, path, repo.AppStrategyDefault); err == nil || !errors.Is(err, repo.ErrNotADirectory) {
		t.Errorf("InitWithApp(%q, %q) error = %v, want ErrNotADirectory", path, repo.AppStrategyDefault, err)
	}
}
