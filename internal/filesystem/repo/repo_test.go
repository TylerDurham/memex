package repo_test

import (
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/TylerDurham/memex/internal/filesystem/repo"
	"github.com/TylerDurham/memex/internal/testutil"
	"github.com/TylerDurham/memex/internal/testutil/obsidian"
	"github.com/stretchr/testify/assert"
)

func Test_ObsidianInitRepo(t *testing.T) {

	path := obsidian.GetTestRepoPath(t)
	repo := obsidian.InitTestRepo(t, path, repo.AppStrategyObsidian)
	fi, err := os.Stat(path)
	if err != nil {
		t.Fatalf("could not stat '%s': %v", path, err)
	}

	assert.Equal(t, repo.AbsPath(), path)
	assert.Equal(t, repo.IsDir(), fi.IsDir())
	assert.Equal(t, repo.Name(), fi.Name())
}

func Test_LoadRepoErrors(t *testing.T) {

	testutil.IsolateConfig(t)

	tests := []struct {
		// TODO: path is never read by the test loop; remove it.
		path        string
		name        string
		expected    error
		description string
	}{
		{path: obsidian.GetTestRepoPath(t), name: "", expected: repo.ErrInvalidArgument, description: "empty name"},
		{path: obsidian.GetTestRepoPath(t), name: " ", expected: repo.ErrInvalidArgument, description: "empty name"},
		{path: obsidian.GetTestRepoPath(t), name: ".", expected: repo.ErrInvalidArgument, description: "name cannot begin with '.'"},
		{path: obsidian.GetTestRepoPath(t), name: "..", expected: repo.ErrInvalidArgument, description: "name cannot begin with '..'"},
		{path: obsidian.GetTestRepoPath(t), name: "/", expected: repo.ErrInvalidArgument, description: "name cannot contain '/'"},
		{path: obsidian.GetTestRepoPath(t), name: "../..", expected: repo.ErrInvalidArgument, description: "name cannot contain '../..'"},
	}

	for i, test := range tests {
		empty, err := repo.Load(test.name)

		assert.Equal(t, repo.Empty, empty, fmt.Sprintf("test %d: %s", i+1, "returned repo should be empty"))
		assert.ErrorIs(t, err, test.expected, fmt.Sprintf("test %d: %s", i+1, test.description))
	}
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

func Test_InitRepo_Errors(t *testing.T) {

	path := obsidian.GetTestRepoPath(t)
	r := obsidian.InitTestRepo(t, path, repo.AppStrategyObsidian)

	tests := []struct {
		path        string
		strategy    string
		description string
		expected    error
	}{
		{path: r.AbsPath(), strategy: r.AppStrategy(), expected: repo.ErrRepoAlreadyExists, description: "dir exists"},
		{path: r.AbsPath(), strategy: "INVALID", expected: repo.ErrInvalidArgumentStrategy, description: "app strategy not supported"},
	}

	for i, test := range tests {
		empty, err := repo.Init(test.path, test.strategy)

		assert.Equal(t, repo.Empty, empty, fmt.Sprintf("test %d: %s", i+1, "returned repo should be empty"))
		assert.ErrorIs(t, err, test.expected, fmt.Sprintf("test %d: %s", i+1, test.description))
	}

}

func Test_InitRepo_ErrNotADirectory(t *testing.T) {
	path := "../../../testdata/processor/markdown.md"
	if _, err := obsidian.InitTestRepoWithErr(t, path, repo.AppStrategyObsidian); err == nil || !errors.Is(err, repo.ErrNotADirectory) {
		t.Errorf("InitWithApp(%q, %q) error = %v, want ErrNotADirectory", path, repo.AppStrategyGeneric, err)
	}
}
