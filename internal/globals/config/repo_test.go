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
	assert.FileExists(t, filepath.Join(base, "tech-kasten", RepoConfigFile))

	got, err := LoadRepoFrom(base, "tech-kasten")
	require.NoError(t, err)
	assert.Equal(t, want, got)

	// Saving again overwrites, and leaves no temp files behind.
	want.Model = "mxbai-embed-large"
	require.NoError(t, want.SaveTo(base))
	got, err = LoadRepoFrom(base, "tech-kasten")
	require.NoError(t, err)
	assert.Equal(t, "mxbai-embed-large", got.Model)

	entries, err := os.ReadDir(filepath.Join(base, "tech-kasten"))
	require.NoError(t, err)
	assert.Len(t, entries, 1)
}

func Test_LoadRepo_NotFound(t *testing.T) {
	_, err := LoadRepoFrom(t.TempDir(), "nope")
	assert.ErrorIs(t, err, ErrRepoNotFound)
	assert.Contains(t, err.Error(), "memex init nope")
}

func Test_ValidateRepoName(t *testing.T) {
	for _, bad := range []string{"", ".", "..", "a/b", `a\b`, "../escape"} {
		assert.Error(t, ValidateRepoName(bad), bad)
		_, err := LoadRepoFrom(t.TempDir(), bad)
		assert.Error(t, err, bad)
	}
	assert.NoError(t, ValidateRepoName("tyler-kasten"))
}
