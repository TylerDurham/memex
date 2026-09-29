package obsidian

import (
	"github.com/TylerDurham/memex/internal/documents"
	"github.com/TylerDurham/memex/internal/documents/strategy"
)

type ObsidianIndexer struct {
	skipDirectories strategy.SkipDirectories
	extensions      strategy.Extensions
}

func (oi *ObsidianIndexer) SkipDirectories() strategy.SkipDirectories {
	return oi.skipDirectories
}

func (oi *ObsidianIndexer) Extensions() strategy.Extensions {
	return oi.extensions
}

func (oi *ObsidianIndexer) GetURI(doc *documents.Document) (string, error) {
	return "", nil
}

func (oi *ObsidianIndexer) Load(doc *documents.Document) error {
	return nil
}

func NewObsidianIndexer() *ObsidianIndexer {

	i := &ObsidianIndexer{
		skipDirectories: strategy.SkipDirectories{
			".obsidian": {},
		},
		extensions: strategy.Extensions{
			".md": {},
		},
	}

	return i
}
