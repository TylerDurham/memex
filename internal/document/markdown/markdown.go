// Package markdown
package markdown

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/TylerDurham/memex/internal/document"
	"github.com/yuin/goldmark/v2/ast"
	"github.com/yuin/goldmark/v2/parser"
	"go.yaml.in/yaml/v3"
)

const FMDelimiter = "---"
const AltFMClosingDelim = "..."

// Document is a parsed markdown file: optional YAML frontmatter and the
// goldmark AST for the remaining body.
type Document struct {
	Frontmatter map[string]any
	Body        ast.Node
	// Source is the raw markdown body the Body AST was parsed from. AST
	// segment offsets index into it.
	Source []byte
	// BodyLine is the 1-based line number in the original file where Source
	// begins (i.e. the line after the closing frontmatter delimiter).
	BodyLine int
}

// scanFrontmatter consumes the frontmatter block, if any, from the start of
// the scanner. It returns the raw YAML bytes (nil when there is no
// frontmatter), the first line of the body when that line had to be read to
// discover there is no frontmatter, and the 1-based line number the body
// starts on.
func scanFrontmatter(scanner *bufio.Scanner) (frontmatter []byte, firstBodyLine *string, bodyLine int, err error) {
	if !scanner.Scan() {
		return nil, nil, 1, scanner.Err()
	}

	first := scanner.Text()
	if strings.TrimRight(first, " \t\r") != FMDelimiter {
		return nil, &first, 1, nil
	}

	fm := &bytes.Buffer{}
	lineNo := 1
	for scanner.Scan() {
		lineNo++
		line := scanner.Text()
		trimmed := strings.TrimRight(line, " \t\r")
		if trimmed == FMDelimiter || trimmed == AltFMClosingDelim {
			return fm.Bytes(), nil, lineNo + 1, nil
		}
		fm.WriteString(line)
		fm.WriteByte('\n')
	}

	if err := scanner.Err(); err != nil {
		return nil, nil, 0, err
	}
	return nil, nil, 0, errors.New("frontmatter: unterminated block")
}

// ReadFrontmatter reads only the frontmatter block from the start of the
// scanner and unmarshals it. It stops scanning at the closing delimiter, so
// the body is never read. Returns nil when the document has no frontmatter.
func ReadFrontmatter(scanner *bufio.Scanner) (map[string]any, error) {
	raw, _, _, err := scanFrontmatter(scanner)
	if err != nil {
		return nil, err
	}
	return unmarshalFrontmatter(raw)
}

func unmarshalFrontmatter(raw []byte) (map[string]any, error) {
	if len(raw) == 0 {
		return nil, nil
	}
	var fm map[string]any
	if err := yaml.Unmarshal(raw, &fm); err != nil {
		return nil, fmt.Errorf("could not parse frontmatter: %w", err)
	}
	return fm, nil
}

// splitDocument scans the reader once, returning the raw frontmatter YAML
// bytes (nil when there is no frontmatter), the remaining markdown body
// bytes, and the 1-based line number the body starts on.
func splitDocument(scanner *bufio.Scanner) (frontmatter, markdown []byte, bodyLine int, err error) {
	body := &bytes.Buffer{}

	frontmatter, first, bodyLine, err := scanFrontmatter(scanner)
	if err != nil {
		return nil, nil, 0, err
	}
	if first != nil {
		body.WriteString(*first)
		body.WriteByte('\n')
	}

	for scanner.Scan() {
		body.WriteString(scanner.Text())
		body.WriteByte('\n')
	}

	if err := scanner.Err(); err != nil {
		return nil, nil, 0, err
	}

	return frontmatter, body.Bytes(), bodyLine, nil
}

// Parse reads r once and returns its YAML frontmatter (nil if absent) and the
// goldmark AST for the body.
func Parse(r io.Reader) (*Document, error) {
	return ParseScanner(document.NewScanner(r))
}

// ParseScanner is Parse for a caller-supplied scanner, which must be
// positioned at the start of the file.
func ParseScanner(scanner *bufio.Scanner) (*Document, error) {
	frontmatter, markdown, bodyLine, err := splitDocument(scanner)
	if err != nil {
		return nil, err
	}

	fm, err := unmarshalFrontmatter(frontmatter)
	if err != nil {
		return nil, err
	}

	return &Document{
		Frontmatter: fm,
		Body:        parser.New().Parse(markdown),
		Source:      markdown,
		BodyLine:    bodyLine,
	}, nil
}
