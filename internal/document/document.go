// Package document provides types and interfaces for working with documents.
package document

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"mime"
	"path/filepath"
	"time"
)

type ProcessFlags uint32

const (
	Unspecified  ProcessFlags = 0
	NoProperties ProcessFlags = 1 << iota
)

// Properties is a set of named document properties.
type Properties map[string]any

// Processor defines an interface for processing a Document.
type Processor interface {
	Process(root string, path string, d fs.DirEntry, flags ProcessFlags) (Document, error)
}

type Document struct {
	AbsPath     string     `json:"absPath"`
	Application string     `json:"application"`
	ModTime     time.Time  `json:"modTime"`
	Properties  Properties `json:"properties"`
	RelPath     string     `json:"relPath"`
	Size        int64      `json:"size"`
	URI         string     `json:"uri"`
}

func (doc *Document) ToJSONString() (string, error) {

	json, err := json.MarshalIndent(doc, "", "	")

	if err != nil {
		return "", fmt.Errorf("could not serialize to json: %+v", err)
	}

	return string(json), err
}

// GetMimeType looks up the mime-type using a file's path. If the extension is unknown,
// 'application/octet-stream' is returned as a fallback.
func GetMimeType(path string) (mimeType string) {
	return GetMimeTypeByExt(filepath.Ext(path))
}

// GetMimeTypeByExt looks up the mime-type using a file's extension. If the extension is unknown,
// 'application/octet-stream' is returned as a fallback.
func GetMimeTypeByExt(ext string) (mimeType string) {
	mimeType = mime.TypeByExtension(ext)

	if mimeType == "" {
		mimeType = "application/octet-stream" // default/fallback
	}
	return mimeType
}
