package document

import (
	"mime"
	"path/filepath"
)

// GetMimeTypeByExt looks up the mime-type using a file's extension. If the extension is unknown,
// 'application/octet-stream' is returned as a fallback.
func GetMimeTypeByExt(ext string) (mimeType string) {
	mimeType = mime.TypeByExtension(ext)

	if mimeType == "" {
		mimeType = "application/octet-stream" // default/fallback
	}
	return mimeType
}

// GetMimeType looks up the mime-type using a file's path. If the extension is unknown,
// 'application/octet-stream' is returned as a fallback.
func GetMimeType(path string) (mimeType string) {
	return GetMimeTypeByExt(filepath.Ext(path))
}
