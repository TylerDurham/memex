// Package walkers
package walkers

import (
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
	AbsPath     string
	Application string
	ModTime     time.Time
	Properties  DocumentProperties
	RelPath     string
	Size        int64
	URI         string
}
