// Package walker
package walker

import (
	"fmt"
	"io/fs"
	"path/filepath"

	"github.com/TylerDurham/memex/internal/document"
	"github.com/TylerDurham/memex/internal/document/obsidian"
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
		flags:       document.Unspecified,
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

		w.handler.LoadFileMetadata(&doc, d)

		docs = append(docs, doc)
		return nil
	})

	if err != nil {
		return nil, err
	}

	return docs, nil
}

