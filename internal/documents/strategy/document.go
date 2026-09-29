package strategy

import (
	"bufio"

	"github.com/TylerDurham/memex/internal/documents"
)

// DocumentStrategy is an interface that defines the operations necessary 
// to load properties and semantic indexing chunks from a file.
type DocumentStrategy interface {
	// LoadProperties reads the properties from the scanner and loads them
	// into the document.
	LoadProperties(sc bufio.Scanner, doc *documents.Document) error

	// LoadChunks reads the chunks from the scanner and loads them 
	// into the document.
	LoadChunks(sc bufio.Scanner, doc *documents.Document) error
}
