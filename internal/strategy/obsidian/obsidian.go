// Package obsidian
package obsidian

import (
	"strings"

	"github.com/TylerDurham/memex/internal/documents"
	"github.com/TylerDurham/memex/internal/strategy"
	"github.com/TylerDurham/memex/internal/strategy/markdown"
)

type ObsidianIndexer struct {
	skipDirectories strategy.SkipDirectories
	extensions      strategy.Extensions
}

func (oi *ObsidianIndexer) SkipDirectories() strategy.SkipDirectories {
	return oi.skipDirectories
}

func (oi *ObsidianIndexer) DocStrategy(ext string) strategy.DocParser {
	switch strings.ToLower(ext) {
	case ".md":
		return markdown.NewMarkdownDocStrategy()
	}

	return nil

}

func (oi *ObsidianIndexer) GetURI(doc *documents.Document) (string, error) {
	return "", nil
}

func (oi *ObsidianIndexer) Load(doc *documents.Document) error {
	return nil
}

func NewObsidianIndexer() *ObsidianIndexer {

	strat := markdown.NewMarkdownDocStrategy()

	return &ObsidianIndexer{
		skipDirectories: strategy.SkipDirectories{
			".obsidian": {},
		},
		extensions: strategy.Extensions{
			".md": strat,
		},
	}
}
