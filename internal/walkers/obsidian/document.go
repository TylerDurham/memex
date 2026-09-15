package obsidian

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/TylerDurham/memex/internal/globals"
	"github.com/TylerDurham/memex/internal/walkers"
	"github.com/stretchr/testify/assert/yaml"
)

func FormatObsidianURL(vault string, path string) string {
	// obsidian://open?vault=Tech-Kasten&file=development%2Fgo%2FGo%20%60fmt%60%20Formatting%20Verbs
	return fmt.Sprintf("obsidian://open?vault=%s&file=%s", url.PathEscape(vault), url.PathEscape(path))
}

func Chunk(scanner *bufio.Scanner) (properties walkers.DocumentProperties, err error) {
	return properties, nil
}

const FMDelim = "---"

func ScanFM(sc *bufio.Scanner, doc *walkers.Document) (err error) {

	first := strings.TrimRight(sc.Text(), "\r")

	if first != FMDelim {
		return nil
	}

	var block bytes.Buffer
	closed := false

	for sc.Scan() {
		line := strings.TrimRight(sc.Text(), "\r")
		if line == FMDelim || line == "..." {
			closed = true
			break
		}

		block.WriteString(line)
		block.WriteByte('\n')
	}

	if !closed {
		return errors.New(("frontmatter: unterminated block"))
	}
	var fm map[string]any
	yaml.Unmarshal(block.Bytes(), &fm)

	doc.Properties = fm

	return nil
}

func ScanDoc(doc *walkers.Document) error {
	f, err := os.Open(doc.AbsPath)

	if err != nil {
		return err
	}

	defer f.Close()

	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 64*1024), 10*1024*1024)

	if !sc.Scan() {
		return sc.Err()
	}

	ScanFM(sc, doc)

	return nil

}

func LoadDoc(root string, path string, file fs.DirEntry) (walkers.Document, error) {

	var doc = walkers.Document{}

	if file.IsDir() {
		return doc, fmt.Errorf("%q: %w", file.Name(), globals.ErrNotAFile)
	}

	vault := filepath.Base(filepath.Dir(filepath.Dir(path)))
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return doc, err
	}
	info, err := os.Stat(path)
	if err != nil {
		return doc, err
	}

	doc.AbsPath = path
	doc.Application = "obsidian"
	doc.ModTime = info.ModTime()
	doc.RelPath = rel
	doc.Size = info.Size()
	doc.URI = FormatObsidianURL(vault, rel)

	ScanDoc(&doc)

	return doc, nil
}
