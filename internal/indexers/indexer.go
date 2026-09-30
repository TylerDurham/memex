// Package indexers
package indexers

import (
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/TylerDurham/memex/internal/documents"
	"github.com/TylerDurham/memex/internal/globals/logger"
	"github.com/TylerDurham/memex/internal/strategy"
	"github.com/TylerDurham/memex/internal/strategy/obsidian"
)

// ErrUnknownApplication is returned when the application's
// IdxStrategy cannot be determine.
var ErrUnknownApplication = errors.New("unknown application")

// LoadStrategy is a factory function which loads the appropriate
// IdxStrategy for the specified application.
func LoadStrategy(application string) (strategy.Index, error) {
	switch application {
	case "obsidian":
		return obsidian.NewObsidianIndexer(), nil

	default:
		return nil, fmt.Errorf("'%s': %w", application, ErrUnknownApplication)
	}
}

// IndexStats is returned after a recursive directory walk
// and contains useful statisics about index operations.
type IndexStats struct {
	DirsSkipped  int
	DocsSkipped  int
	DocsPrepared int
	DocsFailed   int
}

// IndexResult is returned after a recursive directory walk
// and contains an array of Documents and statisics about
// operations.
type IndexResult struct {
	Documents []documents.Document
	Stats     IndexStats
}

// IndexFile builds a Document for the file at req.FilePath, which must be
// inside req.Repo.
//
// It reports progress through req's event callback. It emits EventDocIndexing
// when it starts, then either EventDocIndexed with the resulting Document or
// EventDocError with the error. Each call ends with exactly one of those two
// events.
func IndexFile(req FileIndexRequest) (_ documents.Document, err error) {
	path := strings.TrimSpace(req.FilePath)

	defer func() {
		if err != nil {
			req.emit(Event{Kind: EventDocError, Path: path, Err: err})
		}
	}()

	req.emit(Event{Kind: EventDocIndexing, Path: path})

	doc, err := documents.NewDocument(req.Repo, path)
	if err != nil {
		return documents.Document{}, err
	}

	req.emit(Event{Kind: EventDocIndexed, Path: path, Doc: &doc})
	return doc, nil
}

// IndexDir walks req.Repo.Directory and builds a Document for every file
// whose extension has a strategy in req.IdxStrategy. It does not descend
// into directories named in the strategy's skip list.
//
// A file that fails to index is counted in Stats.DocsFailed and reported
// with an EventDocError, and the walk continues. IndexDir returns an error
// only if the directory tree itself cannot be read. In that case the result
// still holds the documents prepared before the failure.
func IndexDir(req DirIndexRequest) (IndexResult, error) {
	rDirPath := req.Repo.Directory
	strat := req.IdxStrategy
	skipDirs := strat.SkipDirectories()
	extensions := strat.Extensions()

	var result IndexResult

	err := filepath.WalkDir(rDirPath, func(fPath string, d fs.DirEntry, err error) error {
		if err != nil {
			return err // d may be nil here, so don't touch it
		}

		if d.IsDir() {
			if fPath == rDirPath {
				// Do not index repo root
				return nil
			}

			if _, skip := skipDirs[d.Name()]; skip {
				// strategy impl says we should skip this directory
				result.Stats.DirsSkipped++
				req.emit(Event{Kind: EventDirSkipped, Path: fPath})
				return filepath.SkipDir
			}
			return nil
		}

		// Extension keys are lowercase, so ".MD" matches the ".md" strategy.
		ext := strings.ToLower(filepath.Ext(fPath))
		docStrat, ok := extensions[ext]
		if !ok {
			result.Stats.DocsSkipped++
			logger.Debug("file skipped", "file", fPath)
			req.emit(Event{Kind: EventDocSkipped, Path: fPath})
			return nil
		}
		_ = docStrat // TODO: use docStrat to load the file

		doc, err := IndexFile(FileIndexRequest{
			IndexOptions: req.IndexOptions,
			FilePath:     fPath,
		})
		if err != nil {
			// IndexFile has already emitted EventDocError.
			result.Stats.DocsFailed++
			return nil
		}

		result.Documents = append(result.Documents, doc)
		result.Stats.DocsPrepared++
		return nil
	})

	return result, err
}
