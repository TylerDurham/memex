// Package obsidian
package obsidian

import (
	"fmt"
	"io/fs"
	"net/url"
	"path/filepath"

	"github.com/TylerDurham/memex/internal/document"
	"github.com/TylerDurham/memex/internal/globals"
)

type ObsidianProcessor struct {
	extensions map[string]struct{}
}

func NewObsidianProcessor() *ObsidianProcessor {
	ext := map[string]struct{}{
		".md": {},
	}
	op := &ObsidianProcessor{
		extensions: ext,
	}

	return op
}

func (op *ObsidianProcessor) Process(root string, path string, file fs.DirEntry, flags document.ProcessFlags) (document.Document, error) {
	var doc = document.Document{}

	if file.IsDir() {
		return doc, fmt.Errorf("%q: %w", file.Name(), globals.ErrNotAFile)
	}

	vault := filepath.Base(filepath.Dir(filepath.Dir(path)))
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return doc, err
	}

	doc.AbsPath = path
	doc.Application = "obsidian"
	info, err := file.Info()
	if err != nil {
		return doc, err
	}

	doc.ModTime = info.ModTime()
	doc.RelPath = rel
	doc.Size = info.Size()
	doc.URI = FormatObsidianURL(vault, rel)

//	ScanDoc(&doc)

	return doc, nil
}

func FormatObsidianURL(vault string, path string) string {
	// obsidian://open?vault=Tech-Kasten&file=development%2Fgo%2FGo%20%60fmt%60%20Formatting%20Verbs
	return fmt.Sprintf("obsidian://open?vault=%s&file=%s", url.PathEscape(vault), url.PathEscape(path))
}

// func Walk(root string) (docs []walkers.Document, err error) {
//
// 	// Collection of documents
// 	docs = []walkers.Document{}
//
// 	err = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
// 		if err != nil {
// 			return err
// 		}
//
// 		// process files only
// 		if d.IsDir() {
// 			return nil
// 		}
//
// 		// process markdown files only
// 		if !strings.EqualFold(filepath.Ext(path), ".md") {
// 			return nil
// 		}
//
// 		doc, err := Process(root, path, d)
// 		if err != nil {
// 			return err
// 		}
//
// 		docs = append(docs, doc)
// 		return nil
// 	})
//
// 	if err != nil {
// 		return nil, err
// 	}
// 	return docs, nil
// }
