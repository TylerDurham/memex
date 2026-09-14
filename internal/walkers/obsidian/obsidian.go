// Package obsidian
package obsidian

import (
	"fmt"
	"io/fs"
	"net/url"
	"os"
	"path/filepath"

	"github.com/TylerDurham/memex/internal/walkers"
)

func formatObsidianURL(vault string, path string) string {
	// obsidian://open?vault=Tech-Kasten&file=development%2Fgo%2FGo%20%60fmt%60%20Formatting%20Verbs
	return fmt.Sprintf("obsidian://open?vault=%s&file=%s", url.PathEscape(vault), url.PathEscape(path))
}

func Walk(root string) (docs []walkers.Document, err error) {

	docs = []walkers.Document{}
	err = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		vault := filepath.Base(filepath.Dir(filepath.Dir(path)))
		rel, _ := filepath.Rel(root, path)

		if !d.IsDir() {
			info, _ := os.Stat(path)
			doc := walkers.Document{
				AbsPath:     path,
				RelPath:     rel,
				Application: "obsidian",
				Size:        info.Size(),
				ModTime:     info.ModTime(),
				URI:         formatObsidianURL(vault, rel),
			}

			docs = append(docs, doc)
		}
		return err
	})

	if err != nil {
		return nil, err
	}
	return docs, nil
}
