package indexers

import (
	"github.com/TylerDurham/memex/internal/documents"
	"github.com/TylerDurham/memex/internal/repo"
	"github.com/TylerDurham/memex/internal/strategy"
)

// EventKind identifies what happened in an Event.
type EventKind string

const (
	EventDocError EventKind = "doc-error"
	// EventDocIndexing is sent just before a document is indexed.
	EventDocIndexing EventKind = "doc-indexing"
	// EventDocIndexed is sent after a document is indexed successfully.
	EventDocIndexed EventKind = "doc-indexed"
	// EventDocSkipped is sent for a file whose extension has no strategy.Doc.
	EventDocSkipped EventKind = "doc-skipped"
	// EventDirSkipped is sent for a directory listed in SkipDirectories.
	EventDirSkipped EventKind = "dir-skipped"
)

// Event reports progress during indexing.
type Event struct {
	Kind EventKind
	// Path is the file or directory the event is about.
	Path string
	// Doc is the indexed document. It is set only for EventDocIndexed.
	Doc *documents.Document

	Err error
}

type IndexRequest struct {
	Repo        repo.RepoInfo
	IdxStrategy strategy.Index
	FilePath    string

	// OnEvent, if set, is called for each Event as indexing progresses. It
	// runs synchronously on the indexing goroutine, so a slow handler slows
	// indexing down.
	OnEvent func(Event)
}

// emit sends e to req.OnEvent, if one is set.
func (req IndexRequest) emit(e Event) {
	if req.OnEvent != nil {
		req.OnEvent(e)
	}
}
