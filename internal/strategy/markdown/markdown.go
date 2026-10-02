// Package markdown
package markdown

import (
	"bytes"
	"errors"

	"github.com/TylerDurham/memex/internal/documents"
	"go.yaml.in/yaml/v3"
)

func splitMarkdownDoc(data []byte) (fm, body []byte, err error) {

const delim = "---\n"

	data = bytes.ReplaceAll(data, []byte("\r\n"), []byte("\n")) // normalize CRLF
	if !bytes.HasPrefix(data, []byte(delim)) {
		return nil, data, nil // no frontmatter
	}

	rest := data[len(delim):]
	if bytes.HasPrefix(rest, []byte(delim)) { // empty frontmatter: ---\n---\n
		return nil, rest[len(delim):], nil
	}

	fm, body, ok := bytes.Cut(rest, []byte("\n"+delim))
	if !ok {
		return nil, nil, errors.New("unterminated frontmatter")
	}
	return fm, body, nil	
}

func readMarkdownBody(body []byte, doc *documents.Document) error {
	return nil
}

func readMarkdownFrontmatter(fm []byte, doc *documents.Document) error {
	if len(fm) == 0 {
		return nil
	}

	err := yaml.Unmarshal(fm, &doc.Properties)
	if err != nil {
		return err
	}
	return nil
}

type MarkdownDocStrategy struct {
}

func (m *MarkdownDocStrategy) Parse(data []byte, doc *documents.Document) error {
	fm, body, err := splitMarkdownDoc(data)
	if err != nil {
		return err
	}

	if err = readMarkdownFrontmatter(fm, doc); err != nil {
		return err
	}

	if err = readMarkdownBody(body, doc); err != nil {
		return err
	}

	return nil
}

func (m *MarkdownDocStrategy) Name() string {
	return "Markdown"
}

func (m *MarkdownDocStrategy) Ext() string {
	return ".md"
}

func NewMarkdownDocStrategy() (MarkdownDocStrategy, error) {
	return MarkdownDocStrategy{}, nil
}
