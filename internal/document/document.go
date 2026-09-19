// Package document provides types and interfaces for working with documents.
package document

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"time"
)

type ProcessFlags uint32

type FileExtensions map[string]struct{}

const (
	Unspecified       ProcessFlags = 0
	IncludeProperties ProcessFlags = 1 << iota
	IncludeChunks     ProcessFlags = 1 << iota
)

type Chunk struct {
	Text        string
	HeadingPath []string
	StartLine   int
	EndLine     int
}

// Properties is a set of named document properties.
type Properties map[string]any

// // Processor defines an interface for processing a Document.
// type Processor interface {
// 	Process(root string, path string, d fs.DirEntry, flags ProcessFlags) (IndexDocument, error)
// }

// IndexDocumentProvider defines an interface for processing a Document.
type IndexDocumentProvider interface {
	AppName() string
	Extensions() FileExtensions
	LoadFileMetadata(doc *IndexDocument, d fs.DirEntry) error
	LoadDocumentMetadata(doc *IndexDocument, scanner *bufio.Scanner) error
	LoadDocumentChunks(doc *IndexDocument, scanner *bufio.Scanner) error
}

type IndexDocument struct {
	provider    IndexDocumentProvider
	RepoPath    string     `json:"RepoPath"`
	Chunks      []Chunk    `json:"Chunks"`
	DocPath     string     `json:"docPath"`
	Extension   string     `json:"Extension"`
	Application string     `json:"application"`
	MimeType    string     `json:"mimeType"`
	ModTime     time.Time  `json:"modTime"`
	Properties  Properties `json:"properties"`
	RelPath     string     `json:"relPath"`
	Size        int64      `json:"size"`
	URI         string     `json:"uri"`
}

// ToJSONString marshalls the RepoDocument into a JSON string.
func (doc *IndexDocument) ToJSONString() (string, error) {

	json, err := json.MarshalIndent(doc, "", "	")

	if err != nil {
		return "", fmt.Errorf("could not serialize to json: %+v", err)
	}

	return string(json), err
}

// LoadDocumentMetadata loads the Properties for the Document with file-format
// specific properties, such as YAML frontmatter for markdown, etc.
func (doc *IndexDocument) LoadDocumentMetadata() error {
	if err := doc.checkInitialized(); err != nil {
		return err
	}
	return nil
}

func (doc *IndexDocument) init() {
	doc.provider.LoadFileMetadata(doc, nil)
}

func (doc *IndexDocument) checkInitialized() error {
	if doc.RepoPath == "" {
		return errors.New("not initialized: missing 'RepoPath'")
	} else if doc.DocPath == "" {
		return errors.New("not initialized: missing 'DocPath'")
	} else if doc.provider == nil {
		return errors.New("not initialized: missing 'provider'")
	}
	return nil
}

func (doc *IndexDocument) LoadDocumentChunks() error {
	return nil
}

func NewIndexableDocument(repoPath string, docPath string, provider IndexDocumentProvider) (IndexDocument, error) {
	var doc = IndexDocument{
		provider: provider,
		DocPath:  docPath,
		RepoPath: repoPath,
	}

	info, err := os.Stat(docPath)
	if err != nil {
		return doc, fmt.Errorf("could not load indexable document '%q': %w", docPath, err)
	}

	doc.ModTime = info.ModTime()
	doc.RelPath, err = filepath.Rel(repoPath, docPath)
	doc.Size = info.Size()
	doc.Extension = filepath.Ext(docPath)
	
	if err != nil {
		return doc, fmt.Errorf("could not determine relative path between '%q' and '%q': %w", repoPath, docPath, err)
	}

	return doc, nil
}
