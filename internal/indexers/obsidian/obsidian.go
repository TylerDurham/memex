package obsidian

import "github.com/TylerDurham/memex/internal/documents/strategy"

type ObsidianIndexer struct {
	skipDirectories strategy.SkipDirectories
}

func (oi *ObsidianIndexer) SkipDirectories() strategy.SkipDirectories {
	return oi.skipDirectories
}

func NewObsidianIndexer() *ObsidianIndexer {

	i := &ObsidianIndexer{
		skipDirectories: strategy.SkipDirectories{
			".obsidian": {},
		},
	}

	return i
}
