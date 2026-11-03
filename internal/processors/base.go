// Package processors
package processors

import (
	doc "github.com/TylerDurham/memex/internal/filesystem"
)

type DocumentProcessor interface {
	Process(doc *doc.Document) error
}

func Load(path string) (DocumentProcessor, error) {

	d := MarkdownProcessor{}

	return &d, nil

}
