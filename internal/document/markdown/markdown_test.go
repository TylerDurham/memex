package markdown

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yuin/goldmark/v2/ast"
)

// getCWD gets the working directory starting at the project root.
func getCWD() string {
	wd, _ := os.Getwd()
	return wd
}

func Test_splitDocument(t *testing.T) {
	fname := filepath.Join(getCWD(), "./testdata/good.md")
	r, err := os.Open(fname)

	if err != nil {
		t.Fatalf("could not open '%s': %+v", fname, err)
	}
	defer r.Close()

	frm, body, _, err := splitDocument(bufio.NewScanner(r))
	if err != nil {
		t.Fatalf("splitDocument() failed: %+v", err)
	}

	if len(frm) == 0 || len(body) == 0 {
		t.Fatalf("splitDocument() expected both frontmatter and body, got %d/%d bytes", len(frm), len(body))
	}
}

func Test_Parse_Frontmatter(t *testing.T) {
	const input = "---\ntitle: foo\n---\n\n# Hello\n"

	doc, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() failed: %+v", err)
	}

	if got, want := doc.Frontmatter["title"], "foo"; got != want {
		t.Errorf("Frontmatter[title] = %v, want %v", got, want)
	}

	if doc.Body == nil || doc.Body.Kind() != ast.KindDocument {
		t.Errorf("Body = %+v, want a document node", doc.Body)
	}
}

func Test_Parse_NoFrontmatter(t *testing.T) {
	const input = "# Hello\n\nWorld\n"

	doc, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() failed: %+v", err)
	}

	if doc.Frontmatter != nil {
		t.Errorf("Frontmatter = %v, want nil", doc.Frontmatter)
	}

	if doc.Body == nil || doc.Body.Kind() != ast.KindDocument {
		t.Errorf("Body = %+v, want a document node", doc.Body)
	}
}

func newTestScanner(s string) *bufio.Scanner {
	return bufio.NewScanner(strings.NewReader(s))
}
