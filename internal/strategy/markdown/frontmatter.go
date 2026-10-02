package markdown

import (
	"bufio"

	"github.com/TylerDurham/memex/internal/documents"
)

type MarkdownDoc struct {

}

func (m *MarkdownDoc) LoadProperties(sc *bufio.Scanner, doc *documents.Document) error {
	return nil
}

func (m *MarkdownDoc) LoadChunks(sc *bufio.Scanner, doc *documents.Document) error {
	return nil
}
