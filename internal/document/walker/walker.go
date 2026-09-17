// Package walker
package walker

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/TylerDurham/memex/internal/document"
	"github.com/TylerDurham/memex/internal/document/obsidian"
)

type Walker struct {
	application string
	flags       document.ProcessFlags
	handler     document.Processor
}

type RegistryInfo map[string]document.Processor

var registry = RegistryInfo{
	"obsidian": obsidian.NewObsidianProcessor(),
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

func (w *Walker) Walk(root string) (docs []document.Document, err error) {
	// Collection of documents
	docs = []document.Document{}

	err = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// process files only
		if d.IsDir() {
			return nil
		}

		// process markdown files only
		if !strings.EqualFold(filepath.Ext(path), ".md") {
			return nil
		}

		doc, err := w.handler.Process(root, path, d, w.flags)

		if err != nil {
			return err
		}

		docs = append(docs, doc)
		return nil
	})

	if err != nil {
		return nil, err
	}

	return docs, nil
}

// func (w *Walker) AddExtension(ext string) {
// 	if _, exists := w.extensions[ext]; !exists {
// 		w.extensions[ext] = struct{}{}
// 	}
// }
