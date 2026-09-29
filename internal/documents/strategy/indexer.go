// Package strategy
package strategy

import (
	"github.com/TylerDurham/memex/internal/documents"
)

type SkipDirectories map[string]struct{}
type Extensions map[string]DocStrategy

type IdxStrategy interface {
	SkipDirectories() SkipDirectories
	Extensions() Extensions
	Load(doc *documents.Document) error
}
