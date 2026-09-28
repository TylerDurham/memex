// Package walker
package walker

import (
	"bufio"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/TylerDurham/memex/internal/document"
	"github.com/TylerDurham/memex/internal/document/obsidian"
	"github.com/TylerDurham/memex/internal/globals/logger"
)

type Walker struct {
	application string
	flags       document.ProcessFlags
	handler     document.IndexDocumentProvider
}

type RegistryInfo map[string]document.IndexDocumentProvider

var registry = RegistryInfo{
	"obsidian": obsidian.NewObsidianIndexDocumentProvider(),
}

func Registry() RegistryInfo {
	return registry
}

func NewWalker(application string, flags document.ProcessFlags) (w Walker, err error) {
	p, ok := registry[application]
	if !ok {
		return w, fmt.Errorf("application '%s' not supported", application)
	}

	w = Walker{
		application: application,
		flags:       flags,
		handler:     p,
	}

	return w, nil
}

func (w *Walker) Walk(root string) (docs []document.IndexDocument, err error) {
	// Collection of documents
	docs = []document.IndexDocument{}

	err = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip hidden files and folders (.obsidian, .trash, .git, ...): they
		// hold app state, not documents. The root is exempt so a repo can
		// itself live at a hidden path such as ~/.notes.
		if path != root && isHidden(d.Name()) {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// process files only
		if d.IsDir() {
			return nil
		}

		// process only extensions the provider specifies
		var ext = filepath.Ext(path)
		if _, ok := w.handler.Extensions()[ext]; !ok {
			return nil
		}

		doc, err := document.NewIndexableDocument(root, path, w.handler)

		if err != nil {
			return err
		}

		if err := w.handler.LoadFileMetadata(&doc, d); err != nil {
			return err
		}

		// A malformed document shouldn't abort the whole walk: keep its file
		// metadata and log what couldn't be loaded.
		if err := w.loadContent(&doc); err != nil {
			logger.Warn("could not load document content", "path", path, "err", err)
		}

		docs = append(docs, doc)
		return nil
	})

	if err != nil {
		return nil, err
	}

	return docs, nil
}

// loadContent runs the provider's content loaders requested by the walker's
// flags. Each loader gets a fresh scanner positioned at the start of the file.
func (w *Walker) loadContent(doc *document.IndexDocument) error {
	loaders := []func(*document.IndexDocument, *bufio.Scanner) error{}
	if w.flags&document.IncludeProperties != 0 {
		loaders = append(loaders, w.handler.LoadDocumentMetadata)
	}
	if w.flags&document.IncludeChunks != 0 {
		loaders = append(loaders, w.handler.LoadDocumentChunks)
	}
	if len(loaders) == 0 {
		return nil
	}

	f, err := os.Open(doc.DocPath)
	if err != nil {
		return err
	}
	defer f.Close()

	for _, load := range loaders {
		if _, err := f.Seek(0, io.SeekStart); err != nil {
			return err
		}
		if err := load(doc, document.NewScanner(f)); err != nil {
			return err
		}
	}
	return nil
}

// isHidden reports whether a file or folder name is hidden by the Unix
// dot-prefix convention, which Obsidian also follows.
func isHidden(name string) bool {
	return strings.HasPrefix(name, ".")
}
