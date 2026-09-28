package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_Repo_SaveLoad(t *testing.T) {
	base := t.TempDir()
	want := &Repo{
		Name:      "tech-kasten",
		Directory: "/home/me/vaults/Tech-Kasten",
		App:       "obsidian",
		OllamaURL: "https://ollama.example.com",
		Model:     "nomic-embed-text",
	}
	require.NoError(t, want.SaveTo(base))
	assert.FileExists(t, filepath.Join(base, "repos", "tech-kasten", RepoConfigFile))

	got, err := LoadRepoFrom(base, "tech-kasten")
	require.NoError(t, err)
	assert.Equal(t, want, got)

	// Saving again overwrites, and leaves no temp files behind.
	want.Model = "mxbai-embed-large"
	require.NoError(t, want.SaveTo(base))
	got, err = LoadRepoFrom(base, "tech-kasten")
	require.NoError(t, err)
	assert.Equal(t, "mxbai-embed-large", got.Model)

	entries, err := os.ReadDir(filepath.Join(base, "repos", "tech-kasten"))
	require.NoError(t, err)
	assert.Len(t, entries, 1)
}

func Test_LoadRepo_NotFound(t *testing.T) {
	_, err := LoadRepoFrom(t.TempDir(), "nope")
	assert.ErrorIs(t, err, ErrRepoNotFound)
	assert.Contains(t, err.Error(), "memex repo init nope")
}

func Test_ValidateRepoName(t *testing.T) {
	for _, bad := range []string{"", ".", "..", "a/b", `a\b`, "../escape"} {
		assert.Error(t, ValidateRepoName(bad), bad)
		_, err := LoadRepoFrom(t.TempDir(), bad)
		assert.Error(t, err, bad)
	}
	assert.NoError(t, ValidateRepoName("tyler-kasten"))
}

func Test_ListRepos(t *testing.T) {
	base := t.TempDir()

	repos, err := ListReposFrom(base)
	require.NoError(t, err, "no repos dir yet")
	assert.Empty(t, repos)

	for _, name := range []string{"zeta", "alpha"} {
		require.NoError(t, (&Repo{Name: name, Directory: "/v/" + name, App: "obsidian"}).SaveTo(base))
	}
	// Not repos: a dir with no config, and a stray file.
	require.NoError(t, os.MkdirAll(filepath.Join(base, "repos", "empty"), 0o750))
	require.NoError(t, os.WriteFile(filepath.Join(base, "repos", "stray.txt"), nil, 0o644))
	// A repo with a broken config.
	require.NoError(t, os.MkdirAll(filepath.Join(base, "repos", "broken"), 0o750))
	require.NoError(t, os.WriteFile(filepath.Join(base, "repos", "broken", RepoConfigFile), []byte("directory: [\n"), 0o644))

	repos, err = ListReposFrom(base)
	assert.ErrorContains(t, err, "broken")
	require.Len(t, repos, 2)
	assert.Equal(t, "alpha", repos[0].Name)
	assert.Equal(t, "/v/alpha", repos[0].Directory)
	assert.Equal(t, "zeta", repos[1].Name)
}

func Test_RemoveRepo(t *testing.T) {
	base := t.TempDir()
	vault := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(vault, "note.md"), []byte("# Note\n"), 0o644))

	for _, name := range []string{"keep", "gone"} {
		require.NoError(t, (&Repo{Name: name, Directory: vault, App: "obsidian"}).SaveTo(base))
	}
	// An index database next to the config, as store.Init creates it.
	require.NoError(t, os.WriteFile(filepath.Join(base, "repos", "gone", "memex_store.db"), []byte("db"), 0o644))

	require.NoError(t, RemoveRepoFrom(base, "gone"))

	assert.NoDirExists(t, filepath.Join(base, "repos", "gone"))
	assert.DirExists(t, filepath.Join(base, "repos", "keep"), "other repos untouched")
	assert.FileExists(t, filepath.Join(vault, "note.md"), "vault untouched")

	_, err := LoadRepoFrom(base, "gone")
	assert.ErrorIs(t, err, ErrRepoNotFound)
	assert.ErrorIs(t, RemoveRepoFrom(base, "gone"), ErrRepoNotFound, "second remove")
}

func Test_RemoveRepo_LeftoverWithoutConfig(t *testing.T) {
	base := t.TempDir()
	dir := filepath.Join(base, "repos", "old")
	require.NoError(t, os.MkdirAll(dir, 0o750))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "memex_store.db"), []byte("db"), 0o644))

	require.NoError(t, RemoveRepoFrom(base, "old"))
	assert.NoDirExists(t, dir)
}

func Test_RemoveRepo_RejectsUnsafeNames(t *testing.T) {
	base := t.TempDir()
	require.NoError(t, (&Repo{Name: "keep", Directory: "/v", App: "obsidian"}).SaveTo(base))

	for _, bad := range []string{"", ".", "..", "../repos", "keep/..", `a\b`} {
		assert.Error(t, RemoveRepoFrom(base, bad), bad)
	}
	assert.DirExists(t, filepath.Join(base, "repos", "keep"))
	assert.DirExists(t, filepath.Join(base, "repos"))
}
