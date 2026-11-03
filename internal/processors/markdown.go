package processors

import "github.com/TylerDurham/memex/internal/filesystem"

type MarkdownProcessor struct {

}

func (mp *MarkdownProcessor) Process(doc *filesystem.Document) error {
	return nil
}
