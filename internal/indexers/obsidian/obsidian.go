package obsidian

import "github.com/TylerDurham/memex/internal/documents/strategy"

type ObsidianIndexer struct {
	skipDirectories strategy.SkipDirectories
	extensions strategy.Extensions
}

func (oi *ObsidianIndexer) SkipDirectories() strategy.SkipDirectories {
	return oi.skipDirectories
}

func (oi *ObsidianIndexer) Extensions() strategy.Extensions {
	return oi.extensions
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
