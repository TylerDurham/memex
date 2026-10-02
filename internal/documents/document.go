// Package documents
package documents

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/TylerDurham/memex/internal/repo"
)

// Chunk is a contiguous section of a source document, sized and bounded
// for semantic indexing. Each chunk carries its text plus enough positional
// context to trace a search hit back to where it came from in the original file.
//
// HeadingPath records the chain of Markdown headings that enclose the chunk,
// from outermost to innermost (for example, ["Setup", "Install", "macOS"]).
// It can be prepended to Text at embedding time to give the embedding
// structural context, and shown in results as a breadcrumb.
//
// StartLine and EndLine are 1-based and inclusive, and they give the chunk's
// line range in the source document.
type Chunk struct {
	// Text is the raw content of the chunk.
	Text string `json:"text"`

	// HeadingPath is the ordered list of enclosing headings, outermost first.
	// It is empty for content that appears before the first heading.
	HeadingPath []string `json:"headingPath"`

	// StartLine is the first line of the chunk in the source document (1-based).
	StartLine int `json:"startLine"`

	// EndLine is the last line of the chunk in the source document (1-based, inclusive).
	EndLine int `json:"endLine"`
}

// Properties is a set of named document properties.
type Properties map[string]any

// Document is a single source file prepared for semantic indexing. It pairs
// the file's content, split into Chunks, with the metadata needed to filter
// search results and trace them back to the original file.
type Document struct {
	// AppType identifies the application the file belongs to or opens
	// with, such as "obsidian" "acrobat", etc.
	AppType string `json:"appType"`

	// Chunks holds the file's content split into indexable sections,
	// in document order.
	Chunks []Chunk `json:"chunks"`

	// Extension is the file extension, including the leading dot (e.g. ".md").
	Extension string `json:"extension"`

	// MimeType is the detected media type of the file (e.g. "text/markdown").
	MimeType string `json:"mimeType"`

	// ModTime is the file's last modification time at the time it was indexed.
	ModTime time.Time `json:"modTime"`

	// Name is the file's base name, including its extension.
	Name string `json:"name"`

	// Path is the absolute path to the file on disk.
	Path string `json:"path"`

	// Properties holds document-level metadata, such as parsed front matter.
	Properties Properties `json:"properties"`

	// RelPath is the file's path relative to the root being indexed.
	RelPath string `json:"relPath"`

	// Repo describes the version-control repository containing the file,
	// if any.
	Repo repo.RepoInfo `json:"repo"`

	// Size is the file size in bytes.
	Size int64 `json:"size"`

	// URI is a link that opens the file in its application
	// (e.g. an obsidian:// URI). Useful for application
	// "deeplinks" that register a protocol scheme handler
	// for documents to be opened in the application.
	URI string `json:"uri"`
}

// ToJSONString marshals the Document into an indented JSON string.
func (doc *Document) ToJSONString() (string, error) {
	data, err := json.MarshalIndent(doc, "", "\t")
	if err != nil {
		return "", fmt.Errorf("could not serialize to json: %w", err)
	}

	return string(data), nil
}

// ErrNotRegularFile is returned when the path is a directory, device, or other non-regular file.
var ErrNotRegularFile = errors.New("not a regular file")

// ErrNotARepoFile is returned when the file is not within the repo directory
var ErrNotARepoFile = errors.New("file not under repo path")

// NewDocument returns a Document populated with filesystem metadata for the
// file at path, which must be a regular file inside r.Directory.
//
// Only application-agnostic fields are set. Chunks, Properties, Application,
// Name, and URI are left for the caller or an application-specific loader
// to fill in.
//
// NewDocument returns an error wrapping ErrNotRegularFile if path is not a
// regular file, and ErrNotARepoFile if path is outside the repository.
func NewDocument(r repo.RepoInfo, path string) (Document, error) {
	fInfo, err := os.Stat(path)
	if err != nil {
		return Document{}, err // os.Stat's error already includes the path
	}

	if !fInfo.Mode().IsRegular() {
		return Document{}, fmt.Errorf("%q: %w", path, ErrNotRegularFile)
	}

	// filepath.Rel does not fail for paths outside the base directory. It
	// returns "../...", so IsLocal is needed to confirm containment.
	relPath, err := filepath.Rel(r.Directory, path)
	if err != nil || !filepath.IsLocal(relPath) {
		return Document{}, fmt.Errorf("%q not in %q: %w", path, r.Directory, ErrNotARepoFile)
	}

	return Document{
		AppType: r.AppType,
		Extension: strings.ToLower(filepath.Ext(path)),
		MimeType:  MIMEType(path),
		ModTime:   fInfo.ModTime(),
		Path:      path,
		RelPath:   relPath,
		Repo:      r,
		Size:      fInfo.Size(),
	}, nil
}
