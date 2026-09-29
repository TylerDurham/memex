// Package documents
package documents

import (
	"time"

	"github.com/TylerDurham/memex/internal/repo"
)

type Chunk struct {
	Text        string
	HeadingPath []string
	StartLine   int
	EndLine     int
}

// Properties is a set of named document properties.
type Properties map[string]any

type Document struct {
	Repo        repo.Config `json:"repo"`
	Chunks      []Chunk     `json:"chunks"`
	Path        string      `json:"path"`
	Extension   string      `json:"extension"`
	Application string      `json:"application"`
	MimeType    string      `json:"mimeType"`
	ModTime     time.Time   `json:"modTime"`
	Properties  Properties  `json:"properties"`
	RelPath     string      `json:"relPath"`
	Size        int64       `json:"size"`
	URI         string      `json:"uri"`
}

func NewDocument(repo repo.Config, path string) (Document, error) {
	doc := Document{
		Path: path,
	}

	return doc, nil
}
