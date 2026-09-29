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
)

func GetStrategy(application string) (strategy.IdxStrategy, error) {
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

func IndexFile(req IndexRequest) (documents.Document, error) {
	repo := req.Repo
	path := req.FilePath

	req.emit(Event{Kind: EventDocIndexing, Path: path})

	doc, err := documents.NewDocument(repo, path)
	if err != nil {
		return doc, err
	}

	req.emit(Event{Kind: EventDocIndexed, Path: path, Doc: &doc})
	return doc, nil
}

func IndexDir(req IndexRequest) (IndexResult, error) {

	repo := req.Repo
	idxStrat := req.IdxStrategy

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
			if _, skip := idxStrat.SkipDirectories()[baseName]; skip {
				result.Stats.DirsSkipped++
				logger.Debug("directory skipping", "dir", path)
				req.emit(Event{Kind: EventDirSkipped, Path: path})
				return filepath.SkipDir
			}
			// Process files only
			return nil
		}

		docStrat, ok := idxStrat.Extensions()[ext]

		_ = docStrat

		if !ok {
			// Unknown extension
			result.Stats.DocsSkipped++
			logger.Debug("file skipped", "file", path)
			req.emit(Event{Kind: EventDocSkipped, Path: path})
			return nil
		}

		logger.Debug("walking", "path", path)

		fileReq := req
		fileReq.FilePath = path
		doc, err := IndexFile(fileReq)

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
