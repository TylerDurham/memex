package strategy

import (
	"github.com/TylerDurham/memex/internal/documents"
)

// DocParser extracts indexable content from one kind of file, identified by its
// extension. An Index maps extensions to DocParser strategies (see Extensions), and the indexer calls one for each matching file it walks.
//
// For each file, the indexer calls LoadProperties and then LoadChunks on the
// same scanner, so LoadChunks continues where LoadProperties stopped.
// Implementations must not keep per-file state: one instance is reused for
// every file with its extension.
type DocParser interface {
	Parse(data []byte, doc *documents.Document) error

	// // LoadProperties reads the file's metadata (for Markdown, the YAML
	// // frontmatter) and stores it in doc.Properties. It should stop reading
	// // at the end of the metadata. A file with no metadata is not an error.
	// LoadProperties(sc *bufio.Scanner, doc *documents.Document) error
	//
	// // LoadChunks reads the rest of the file and appends it to doc.Chunks,
	// // split into the sections that get embedded for semantic search. Each
	// // Chunk records its heading path and 1-based start and end lines.
	// LoadChunks(sc *bufio.Scanner, doc *documents.Document) error

	// Name returns a human-readable name for the strategy, such as
	// "Markdown", for logs and CLI output.
	Name() string

	// Ext returns the file extension this strategy handles, lowercase and
	// with the leading dot (".md"), in the form filepath.Ext returns.
	Ext() string
}
