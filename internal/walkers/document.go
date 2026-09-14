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

// DocumentProperty is any primitive or slice-of-primitive that may appear as a
// map entry: string, int, float64, bool, or a slice of one of those.
type DocumentProperty interface {
	isValue()
}

type (
	String  string
	Int     int
	Float   float64
	Bool    bool
	Strings []string
	Ints    []int
	Floats  []float64
	Bools   []bool
)

func (String) isValue()  {}
func (Int) isValue()     {}
func (Float) isValue()   {}
func (Bool) isValue()    {}
func (Strings) isValue() {}
func (Ints) isValue()    {}
func (Floats) isValue()  {}
func (Bools) isValue()   {}

// DocumentProperties is a set of named document properties.
type DocumentProperties map[string]DocumentProperty

type Document struct {
	AbsPath     string
	Application string
	ModTime     time.Time
	Properties  DocumentProperties
	RelPath     string
	Size        int64
	URI         string
}
