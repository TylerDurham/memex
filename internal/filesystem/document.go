// Package filesystem
package filesystem

import (
	"io/fs"
	"os"
	"path/filepath"
)

type Document struct {
	absPath string
	ext     string
	fs.FileInfo
	name       string
	properties Properties
	relPath    string
	repo       Repo
}

// AbsPath returns the absolute path to the document on the file system
func (d *Document) AbsPath() string {
	return d.absPath
}

func (d *Document) Ext() string {
	return d.ext
}

func (d *Document) Name() string {
	return d.name
}

// Properties returns parsed document properties
func (d *Document) Properties() Properties {
	return d.properties
}

// RelPath returns the path of the document relative to the repo for the document
func (d *Document) RelPath() string {
	return d.relPath
}

// Repo returns the parent repository for the document
func (d *Document) Repo() Repo {
	return d.repo
}

// NewDocument constructs a new Document instance
func NewDocument(r Repo, path string) (Document, error) {
	fi, err := os.Stat(path)
	if err != nil {
		return Document{}, err
	}

	rel, err := filepath.Rel(r.AbsPath(), path)
	if err != nil {
		return Document{}, err
	}

	return Document{
		FileInfo:   fi,
		absPath:    filepath.Clean(path),
		ext:        filepath.Ext(path),
		name:       filepath.Base(path),
		properties: Properties{},
		relPath:    rel,
		repo:       r,
	}, nil

}
