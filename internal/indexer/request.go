package indexer

import (
	"github.com/TylerDurham/memex/internal/documents"
	"github.com/TylerDurham/memex/internal/repo"
	"github.com/TylerDurham/memex/internal/strategy"
)

// EventKind identifies what happened in an Event.
type EventKind string

const (
	// EventFileIndexing is sent just before a document is indexed.
	EvenFileEmbedded EventKind = "file-embedded"

	EventDirIndexed EventKind = "dir-indexed"
	EventDirIndexing EventKind = "dir-indexing"

	// EventDirSkipped is sent for a directory listed in SkipDirectories.
	EventDirSkipped EventKind = "dir-skipped"

	EventFileEmbedding EventKind = "file-embedding"

	// EventFileError is sent when a document fails to index. Event.Err holds
	// the error
	EventFileError EventKind = "file-error"

	// EventFileIndexed is sent after a document is indexed successfully.
	EventFileIndexed EventKind = "file-indexed"
	EventFileIndexing EventKind = "file-indexing"
	EventFileParsed EventKind = "file-parsed"
	EventFileParsing EventKind = "file-parsing"

	// EventFileSkipped is sent for a file whose extension has no strategy.Doc.
	EventFileSkipped EventKind = "file-skipped"
	EventFileStored EventKind = "file-stored"
	EventFileStoring EventKind = "file-storing"
)

// Event reports progress during indexing.
type Event struct {
	Kind EventKind
	// Path is the file or directory the event is about.
	Path string
	// Doc is the indexed document. It is set only for EventDocIndexed.
	Doc *documents.Document
	// Err is the indexing error. It is set only for EventDocError.
	Err error
}

// IndexOptions holds the settings shared by every index request.
type IndexOptions struct {
	// Repo is the repository being indexed.
	Repo repo.RepoInfo
	// RepoStrategy decides which files are indexed and which directories are
	// skipped.
	RepoStrategy strategy.Index

	// OnEvent, if set, is called for each Event as indexing progresses. It
	// runs synchronously on the indexing goroutine, so a slow handler slows
	// indexing down.
	OnEvent func(Event)
}

// DirIndexRequest asks IndexDir to index every eligible file under
// Repo.Directory.
type DirIndexRequest struct {
	IndexOptions
}

// FileIndexRequest asks IndexFile to index the single file at FilePath,
// which must be inside Repo.Directory.
type FileIndexRequest struct {
	IndexOptions
	// FilePath is the path of the file to index.
	FilePath string

}

// emit sends e to opts.OnEvent, if one is set.
func (opts IndexOptions) emit(e Event) {
	if opts.OnEvent != nil {
		opts.OnEvent(e)
	}
}
