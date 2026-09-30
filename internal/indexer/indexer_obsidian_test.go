package indexers

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/TylerDurham/memex/internal/repo"
	"github.com/TylerDurham/memex/internal/strategy/obsidian"
	"github.com/stretchr/testify/assert"
)

func TestIndexFile(t *testing.T) {

	tmpDir := t.TempDir()
	repoName := "obsidian-vault"
	configDir := filepath.Join(tmpDir, ".config/memex")
	configReposDir := filepath.Join(configDir, "repos")
	wantRepoDir, _ := filepath.Abs(filepath.Join("../../testdata/repos/obsidian-vault/"))

	walkRootDir := filepath.Join(wantRepoDir, "Projects")

	_, err := repo.Init(configDir, repo.InitOptions{
		Application: "obsidian",
		Directory:   wantRepoDir,
		Name:        repoName,
	})

	if err != nil {
		t.Fatalf("could not init repo '%s' at '%s': %v", repoName, configReposDir, err)
	}

	err = filepath.WalkDir(walkRootDir, func(wantPath string, d fs.DirEntry, err error) error {

		t.Logf("found: %s", wantPath)

		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		repo, err := repo.Load(configDir, repoName)

		if err != nil {
			return err
		}

		doc, err := IndexFile(FileIndexRequest{
			FilePath:    wantPath,
			IdxStrategy: obsidian.NewObsidianIndexer(),
			Repo:        *repo,
		})

		assert.Nil(t, err, "")
		assert.NotNilf(t, doc, "doc cannot be nil")

		stat, err := os.Stat(wantPath)

		if err != nil {
			return err
		}

		wantSize := stat.Size()
		wantModTime := stat.ModTime()
		wantRel, _ := filepath.Rel(wantRepoDir, wantPath)

		assert.Equal(t, wantPath, doc.Path)
		assert.Equal(t, wantSize, doc.Size)
		assert.Equal(t, wantModTime, doc.ModTime)
		assert.Equal(t, wantRel, doc.RelPath)

		return nil
	})

	if err != nil {
		t.Fatalf("could not walk '%s': ", err)
	}
}
