// Package obsidian
package obsidian

import (
	"io/fs"
	"os"
	"path/filepath"

	"github.com/TylerDurham/memex/internal/walkers"
)

func Walk(path string) (docs []walkers.Document, err error) {

	docs = []walkers.Document{}
	err = filepath.WalkDir(path, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() {
			info, _ := os.Stat(path)
			doc := walkers.Document{
				AbsPath:     path,
				RelPath:     filepath.Base(path),
				Application: "obsidian",
				Size:        info.Size(),
				ModTime:     info.ModTime(),
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
