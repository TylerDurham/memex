// Package strategy defines the interfaces that adapt indexing to a repo's
// application and file types.
package strategy

import (
	"github.com/TylerDurham/memex/internal/documents"
)

// SkipDirectories is a set of directory base names (such as ".obsidian")
// that the indexer does not descend into, wherever they appear in the repo.
type SkipDirectories map[string]struct{}

// Extensions maps a file extension, in the form DocStrategy.Ext returns, to
// the DocStrategy that handles files with that extension. Files whose
// extension has no entry are skipped.
type Extensions map[string]DocStrategy

// IdxStrategy adapts indexing to the application that owns a repo, such as
// Obsidian. It decides which parts of the repo are walked and which
// DocStrategy reads each file. GetStrategy in the indexers package returns
// the IdxStrategy for a repo's application.
//
// One instance is used for a whole indexing run, so the maps it returns must
// not change during the run, and callers must not modify them.
type IdxStrategy interface {
	// SkipDirectories returns the directories to leave out of the walk,
	// such as the application's own settings folder.
	SkipDirectories() SkipDirectories

	// Extensions returns the file types to index and the DocStrategy that
	// reads each one.
	Extensions() Extensions

	// Load fills in the document fields that depend on the application
	// rather than the file format, such as doc.URI. The indexer calls it
	// after the DocStrategy has loaded the file's properties and chunks.
	Load(doc *documents.Document) error
}
