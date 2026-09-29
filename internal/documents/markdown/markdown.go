// Package markdown
package markdown

import (
	"bufio"

	"github.com/TylerDurham/memex/internal/documents"
)

type MarkdownDocStrategy struct {
}

func (m *MarkdownDocStrategy) LoadProperties(sc bufio.Scanner, doc *documents.Document) error {
	return nil
}

func (m *MarkdownDocStrategy) LoadChunks(sc bufio.Scanner, doc *documents.Document) error {
	return nil
}

func NewMarkdownDocStrategy() (MarkdownDocStrategy, error) {
	return MarkdownDocStrategy{}, nil
}
