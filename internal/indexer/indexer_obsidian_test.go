package indexer

import (
	"path/filepath"
	"testing"

	"github.com/TylerDurham/memex/internal/strategy/obsidian"
	"github.com/TylerDurham/memex/internal/testutil"
	"github.com/stretchr/testify/assert"
)

// func createTestRepo(t *testing.T, application string, directory string) (*repo.RepoInfo, string, error) {
//
// 	tmpDir := t.TempDir()
// 	configDir := filepath.Join(tmpDir, ".config", "memex")
// 	repoName := filepath.Base(directory)
// 	path, err := repo.Init(configDir, repo.InitOptions{
// 		Application: application,
// 		Name: repoName,
// 		Directory: directory,
// 	})
//
// 	if err != nil {
// 		return nil, "", err
// 	}
//
// 	r, err := repo.Load(configDir, repoName)
// 	return r, path, err
// }
//

func TestIndexFile(t *testing.T) {
	wantAppType := "obsidian"
	r := testutil.LoadTestRepoInfo(t, wantAppType)

	wantAbs := filepath.Join(r.Directory, "Projects/memex.md")

	doc, err := IndexFile(FileIndexRequest{
		Repo:        *r,
		IdxStrategy: obsidian.NewObsidianIndexer(),
		FilePath:    wantAbs,
	})

	if err != nil {
		t.Fatalf("could not index document '%s': %v", wantAbs, err)
	}

	assert.NotNilf(t, doc, "doc should not be nil")
	assert.Equal(t, wantAbs, doc.Path)
	assert.Equal(t, wantAppType, doc.AppType)

}

// func TestIndexFiles(t *testing.T) {
//
// 	tmpDir := t.TempDir()
// 	repoName := "obsidian-vault"
// 	configDir := filepath.Join(tmpDir, ".config/memex")
// 	configReposDir := filepath.Join(configDir, "repos")
// 	wantRepoDir, _ := filepath.Abs(filepath.Join("../../testdata/repos/obsidian-vault/"))
//
// 	walkRootDir := filepath.Join(wantRepoDir, "Projects")
//
// 	_, err := repo.Init(configDir, repo.InitOptions{
// 		Application: "obsidian",
// 		Directory:   wantRepoDir,
// 		Name:        repoName,
// 	})
//
// 	if err != nil {
// 		t.Fatalf("could not init repo '%s' at '%s': %v", repoName, configReposDir, err)
// 	}
//
// 	err = filepath.WalkDir(walkRootDir, func(wantPath string, d fs.DirEntry, err error) error {
//
// 		t.Logf("found: %s", wantPath)
//
// 		if err != nil {
// 			return err
// 		}
//
// 		if d.IsDir() {
// 			return nil
// 		}
//
// 		repo, err := repo.Load(configDir, repoName)
//
// 		if err != nil {
// 			return err
// 		}
//
// 		doc, err := IndexFile(FileIndexRequest{
// 			FilePath:    wantPath,
// 			IdxStrategy: obsidian.NewObsidianIndexer(),
// 			Repo:        *repo,
// 		})
//
// 		assert.Nil(t, err, "")
// 		assert.NotNilf(t, doc, "doc cannot be nil")
//
// 		stat, err := os.Stat(wantPath)
// 		_ = stat
//
// 		if err != nil {
// 			return err
// 		}
//
// 		// TODO: Add this back
// 		// wantSize := stat.Size()
// 		// wantModTime := stat.ModTime()
// 		// wantRel, _ := filepath.Rel(wantRepoDir, wantPath)
// 		// _ = wantRel
// 		//
// 		// assert.Equal(t, wantPath, doc.Path)
// 		// // assert.Equal(t, wantSize, doc.Size)
// 		// // assert.Equal(t, wantModTime, doc.ModTime)
// 		// assert.Equal(t, wantRel, doc.RelPath)
//
// 		assert.NotNilf(t, doc.Properties, "properties is nil")
// 		assert.GreaterOrEqual(t, len(doc.Properties), 3, "properties are empty")
//
// 		return nil
// 	})
//
// 	if err != nil {
// 		t.Fatalf("could not walk '%s': ", err)
// 	}
// }
