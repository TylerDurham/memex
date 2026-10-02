package indexer

import (
	"bufio"
	"fmt"
	"io"
	"os"

	"github.com/TylerDurham/memex/internal/documents"
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

	req.emit(Event{Kind: EventDocIndexing, Path: path})
	defer func() {
		if err != nil {
			req.emit(Event{Kind: EventDocError, Path: path, Err: err})
		}
	}()

	doc, err := documents.NewDocument(req.Repo, path)
	if err != nil {
		return documents.Document{}, err
	}

	strat, ok := req.IdxStrategy.DocStrategy()[doc.Extension]
	if !ok {
		return documents.Document{}, fmt.Errorf("no indexing strategy for extension %q", doc.Extension)
	}

	if err = loadDoc(&doc, strat); err != nil {
		return documents.Document{}, err
	}

	req.emit(Event{Kind: EventDocIndexed, Path: path, Doc: &doc})
	return doc, nil
}

// loadDoc reads doc.Path and populates doc's properties and chunks using strat.
func loadDoc(doc *documents.Document, strat strategy.DocParser) error {
	f, err := os.ReadFile(doc.Path)
	if err != nil {
		return fmt.Errorf("open document: %w", err) // *PathError already includes the path
	}

	err = strat.Parse(f, doc)

	return err
}

const MaxScanBufferSize = 10 * 1024 * 1024 // Max buffer 10MB
const InitialScanBufferSize = 64 * 1024    // Initial buff 64KB

// NewScanner returns a line scanner over r sized for long document lines.
func NewScanner(r io.Reader) *bufio.Scanner {
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, InitialScanBufferSize), MaxScanBufferSize)
	return scanner
}
