package indexers

import (
	"fmt"
	"io/fs"
	"path/filepath"

	"github.com/TylerDurham/memex/internal/documents/strategy"
	"github.com/TylerDurham/memex/internal/indexers/obsidian"
	"github.com/TylerDurham/memex/internal/repo"
)

func GetStrategy(application string) (strategy.IndexStrategy, error) {
	switch application {
	case "obsidian":
		return obsidian.NewObsidianIndexer(), nil

	default:
		return nil, fmt.Errorf("unknown application %q", application)
	}
}

func Index(cfg repo.Config, i strategy.IndexStrategy) error {

	err := filepath.WalkDir(cfg.Directory, func(path string, d fs.DirEntry, err error) error {
		return nil
	})

	return err
}
