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
	repo := obsidian.InitTestRepo(t, path, repo.AppStrategyDefault)
	fi, err := os.Stat(path)
	if err != nil {
		t.Fatalf("could not stat '%s': %v", path, err)
	}

	assert.Equal(t, repo.AbsPath(), path)
	assert.Equal(t, repo.IsDir(), fi.IsDir())
	assert.Equal(t, repo.Name(), fi.Name())
}

func Test_LoadRepo(t *testing.T) {
	path := obsidian.GetTestRepoPath(t)
	expected := obsidian.InitTestRepo(t, path, repo.AppStrategyObsidian)
	actual, err := repo.Load(expected.Name())
	assert.NoErrorf(t, err, "should not return error")

	assert.Equal(t, expected.Name(), actual.Name())
	assert.Equal(t, expected.FileInfo, actual.FileInfo)
	assert.Equal(t, expected.AbsPath(), actual.AbsPath())
	assert.Equal(t, expected.AppStrategy(), actual.AppStrategy())
}

func Test_NewRepo_InvalidStrategy(t *testing.T) {
	name := "BAD_STRATEGY"
	path := obsidian.GetTestRepoPath(t)
	if _, err := obsidian.InitTestRepoWithErr(t, path, name); err == nil || !errors.Is(err, repo.ErrInvalidArgument) {
		t.Errorf("InitWithApp(%q, %q) error = %v, want ErrInvalidArgument", path, name, err)
	}
}

func Test_InitRepo_ErrNotADirectory(t *testing.T) {
	path := "../../../testdata/processor/markdown.md"
	if _, err := obsidian.InitTestRepoWithErr(t, path, repo.AppStrategyDefault); err == nil || !errors.Is(err, repo.ErrNotADirectory) {
		t.Errorf("InitWithApp(%q, %q) error = %v, want ErrNotADirectory", path, repo.AppStrategyDefault, err)
	}
}
