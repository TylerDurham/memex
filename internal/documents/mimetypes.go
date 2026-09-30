package documents

import (
	"mime"
	"path/filepath"
)

// MIMETypeByExt looks up the mime-type using a file's extension. If the extension is unknown,
// 'application/octet-stream' is returned as a fallback.
func MIMETypeByExt(ext string) (mimeType string) {
	mimeType = mime.TypeByExtension(ext)

	if mimeType == "" {
		mimeType = "application/octet-stream" // default/fallback
	}
	return mimeType
}

// MIMEType looks up the mime-type using a file's path. If the extension is unknown,
// 'application/octet-stream' is returned as a fallback.
func MIMEType(path string) (mimeType string) {
	return MIMETypeByExt(filepath.Ext(path))
}
