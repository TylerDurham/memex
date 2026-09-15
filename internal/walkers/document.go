// Package walkers
package walkers

import (
	"encoding/json"
	"fmt"
	"mime"
	"path/filepath"
	"time"
)

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

// DocumentProperties is a set of named document properties.
type DocumentProperties map[string]any

type Document struct {
	AbsPath     string             `json:"absPath"`
	Application string             `json:"application"`
	ModTime     time.Time          `json:"modTime"`
	Properties  DocumentProperties `json:"properties"`
	RelPath     string             `json:"relPath"`
	Size        int64              `json:"size"`
	URI         string             `json:"uri"`
}

func (doc *Document) ToJSONString() (string, error) {

	json, err := json.MarshalIndent(doc, "", "	")

	if err != nil {
		return "", fmt.Errorf("could not serialize to json: %+v", err)
	}

	return string(json), err
}
