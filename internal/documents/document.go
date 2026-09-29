// Package documents
package documents

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/TylerDurham/memex/internal/repo"
)

type Chunk struct {
	Text        string   `json:"text"`
	HeadingPath []string `json:"headingPath"`
	StartLine   int      `json:"startLine"`
	EndLine     int      `json:"endLine"`
}

// Properties is a set of named document properties.
type Properties map[string]any

type Document struct {
	Repo        repo.RepoInfo `json:"repo"`
	Chunks      []Chunk       `json:"chunks"`
	Path        string        `json:"path"`
	Extension   string        `json:"extension"`
	Application string        `json:"application"`
	MimeType    string        `json:"mimeType"`
	ModTime     time.Time     `json:"modTime"`
	Properties  Properties    `json:"properties"`
	RelPath     string        `json:"relPath"`
	Size        int64         `json:"size"`
	URI         string        `json:"uri"`
}

// ToJSONString marshals the Document into an indented JSON string.
func (doc *Document) ToJSONString() (string, error) {
	data, err := json.MarshalIndent(doc, "", "\t")
	if err != nil {
		return "", fmt.Errorf("could not serialize to json: %w", err)
	}

	return string(data), nil
}

func NewDocument(repo repo.RepoInfo, path string) (Document, error) {
	doc := Document{
		Path: path,
	}

	return doc, nil
}
