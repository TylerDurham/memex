package indexer

import (
	"fmt"
	"os"

	"github.com/TylerDurham/memex/internal/documents"
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

	err = FileParse(req, &doc)
	if err != nil {
		return documents.Document{}, fmt.Errorf("could not parse '%s': %v", doc.Abs, err)
	}

	req.emit(Event{Kind: EventFileIndexed, Path: path, Doc: &doc})
	return doc, nil
}

func FileParse(req FileIndexRequest, doc *documents.Document) error {

	req.emit(Event{Kind: EventFileParsing, Path: req.FilePath})

	data, err := os.ReadFile(doc.Abs)
	if err != nil {
		return fmt.Errorf("could not read file for indexing: %w", err) // *PathError already includes the path
	}

	p := req.RepoStrategy.DocStrategy(doc.Extension)
	if p == nil {
		return fmt.Errorf("doc parser is nil")
	}
	p.Parse(data, doc)

	req.emit(Event{Kind: EventFileParsed, Path: req.FilePath})
	return nil
}

func FileEmbed() {

}

func FileStore() {

}
