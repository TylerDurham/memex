package indexer

import (
	"fmt"
	"os"

	"github.com/TylerDurham/memex/internal/documents"
	"github.com/TylerDurham/memex/internal/repo"
	"github.com/TylerDurham/memex/internal/strategy"
)

// IndexFile builds a Document for the file at req.FilePath, which must be
// inside req.Repo.
//
// It reports progress through req's event callback. It emits EventDocIndexing
// when it starts, then either EventDocIndexed with the resulting Document or
// EventDocError with the error. Each call ends with exactly one of those two
// events.
func IndexFile(req FileIndexRequest) (_ documents.Document, err error) {
	path := req.FilePath

	req.emit(Event{Kind: EventFileIndexing, Path: path})
	defer func() {
		if err != nil {
			req.emit(Event{Kind: EventFileError, Path: path, Err: err})
		}
	}()

	doc, err := documents.NewDocument(req.Repo, path)
	if err != nil {
		return documents.Document{}, err
	}

	strat, ok := req.RepoStrategy.DocStrategy()[doc.Extension]
	if !ok {
		return documents.Document{}, fmt.Errorf("no indexing strategy for extension %q", doc.Extension)
	}

	if err = loadDoc(&doc, strat); err != nil {
		return documents.Document{}, err
	}

	req.emit(Event{Kind: EventFileIndexed, Path: path, Doc: &doc})
	return doc, nil
}

// loadDoc reads doc.Path and populates doc's properties and chunks using strat.
func loadDoc(doc *documents.Document, strat strategy.DocParser) error {
	f, err := os.ReadFile(doc.Abs)
	if err != nil {
		return fmt.Errorf("open document: %w", err) // *PathError already includes the path
	}

	err = strat.Parse(f, doc)

	return err
}

func FileParse(repo repo.RepoInfo, path string, parser strategy.DocParser) {

}

func FileEmbed() {

}

func FileStore() {

}
