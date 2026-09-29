// Package indexers
package indexers

import (
	"fmt"
	"io/fs"
	"path/filepath"

	"github.com/TylerDurham/memex/internal/documents"
	"github.com/TylerDurham/memex/internal/documents/strategy"
	"github.com/TylerDurham/memex/internal/globals/logger"
	"github.com/TylerDurham/memex/internal/indexers/obsidian"
	"github.com/TylerDurham/memex/internal/repo"
)

func GetStrategy(application string) (strategy.IndexStrategy, error) {
	switch application {
	case "obsidian":
		return obsidian.NewObsidianIndexer(), nil

	default:
		return nil, fmt.Errorf("unknown application %q", application)
	}
}

type IndexStats struct {
	DirsSkipped  int
	DocsSkipped  int
	DocsPrepared int
}

type IndexResult struct {
	Documents []documents.Document
	Stats     IndexStats
}

func Index(repo repo.Config, i strategy.IndexStrategy) (IndexResult, error) {

	result := IndexResult{
		Stats: IndexStats{
			DocsSkipped:  0,
			DirsSkipped:  0,
			DocsPrepared: 0,
		},
	}

	docs := []documents.Document{}

	err := filepath.WalkDir(repo.Directory, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			// We had an issue loading the directory
			return err
		}

		baseName := d.Name()
		ext := filepath.Ext(path)

		if d.IsDir() {
			if _, skip := i.SkipDirectories()[baseName]; skip {
				result.Stats.DirsSkipped++
				logger.Debug("directory skipping", "dir", path)
				return filepath.SkipDir
			}
			// Process files only
			return nil
		}

		if _, process := i.Extensions()[ext]; !process {
			// Unknown extension
			result.Stats.DocsSkipped++
			logger.Debug("file skipped", "file", path)
			return nil
		}

		logger.Debug("walking", "path", path)

		doc, err := documents.NewDocument(repo, path)

		if err != nil {
			return err
		}

		docs = append(docs, doc)
		result.Stats.DocsPrepared++

		return nil
	})

	result.Documents = docs
	return result, err
}
