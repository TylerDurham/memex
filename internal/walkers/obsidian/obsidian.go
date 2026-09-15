// Package obsidian
package obsidian

import (
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/TylerDurham/memex/internal/walkers"
)

func Walk(root string) (docs []walkers.Document, err error) {

	// Collection of documents
	docs = []walkers.Document{}

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

		doc, err := LoadDoc(root, path, d)
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
