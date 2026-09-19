// Package obsidian
package obsidian

import (
	"bufio"
	"fmt"
	"io/fs"
	"net/url"

	"github.com/TylerDurham/memex/internal/document"
)

const AppName = "obsidian"

// ObsidianIndexDocumentProvider is a map that exposes the file extensions the provider supports.
type ObsidianIndexDocumentProvider struct {
	appName    string
	canLaunch  bool
	extensions document.FileExtensions
}

// AppName gets the name of the provider.
func (p *ObsidianIndexDocumentProvider) AppName() string {
	return p.appName
}

// CanLaunch returns whether or not the indexable document can be launched on the desktop
// Default is true for Obsidian Notes.
func (p *ObsidianIndexDocumentProvider) CanLaunch(doc *document.IndexDocument) bool {
	return p.canLaunch
}

// LaunchURL provides an xdg-open compatible URL for the document.
func (p *ObsidianIndexDocumentProvider) LaunchURL(doc *document.IndexDocument) (string, error) {
	if !p.CanLaunch(doc) {
		return "", document.ErrorGeneratingLauncURL("can launch: false", doc)
	}

	// Format the URL for Obsidian notes.
	url := fmt.Sprintf("obsidian://open?vault=%s&file=%s", url.PathEscape(doc.RepoPath), url.PathEscape(doc.DocPath))

	return url, nil
}

// Extensions returns a map of extensions the provider supports.
func (p *ObsidianIndexDocumentProvider) Extensions() document.FileExtensions {
	return p.extensions
}

// LoadFileMetadata loads additional file metadata, if any.
func (p *ObsidianIndexDocumentProvider) LoadFileMetadata(doc *document.IndexDocument, d fs.DirEntry) error {
	doc.Application = p.AppName()
	doc.URI = FormatObsidianURL(doc.RepoPath, doc.DocPath)
	return nil
}

// LoadDocumentMetadata loads document/format metadata for the document.
// NOTE: This provider only supports metadata found in Markdown frontmatter.
func (p *ObsidianIndexDocumentProvider) LoadDocumentMetadata(doc *document.IndexDocument, scanner *bufio.Scanner) error {
	return nil
}

func (p *ObsidianIndexDocumentProvider) LoadDocumentChunks(doc *document.IndexDocument, scanner *bufio.Scanner) error {
	return nil
}

// NewObsidianIndexDocumentProvider Returns a provider that can handle Markdown files found in Obsidian notes.
func NewObsidianIndexDocumentProvider() *ObsidianIndexDocumentProvider {
	ext := document.FileExtensions{
		".md": {},
	}
	return &ObsidianIndexDocumentProvider{
		appName:    "obsidian",
		canLaunch:  true,
		extensions: ext,
	}
}

//	type ObsidianProcessor struct {
//		extensions map[string]struct{}
//	}
//
//	func NewObsidianProcessor() *ObsidianProcessor {
//		ext := map[string]struct{}{
//			".md": {},
//		}
//		op := &ObsidianProcessor{
//			extensions: ext,
//		}
//
//		return op
//	}
//
//	func (op *ObsidianProcessor) Process(root string, path string, file fs.DirEntry, flags document.ProcessFlags) (document.IndexDocument, error) {
//		var doc = document.IndexDocument{}
//
//		if file.IsDir() {
//			return doc, fmt.Errorf("%q: %w", file.Name(), globals.ErrNotAFile)
//		}
//
//		vault := filepath.Base(filepath.Dir(filepath.Dir(path)))
//		rel, err := filepath.Rel(root, path)
//		if err != nil {
//			return doc, err
//		}
//
//		doc.DocPath = path
//		doc.Application = "obsidian"
//		info, err := file.Info()
//		if err != nil {
//			return doc, err
//		}
//
//		doc.ModTime = info.ModTime()
//		doc.RelPath = rel
//		doc.Size = info.Size()
//		doc.URI = FormatObsidianURL(vault, rel)
//
// //	ScanDoc(&doc)
//
//		return doc, nil
//	}

// FormatObsidianURL formats a local file path into a canonical Obsidian URL.
// Example: obsidian://open?vault=Tech-Kasten&file=development%2Fgo%2FGo%20%60fmt%60%20Formatting%20Verbs
func FormatObsidianURL(vault string, path string) string {
	return fmt.Sprintf("obsidian://open?vault=%s&file=%s", url.PathEscape(vault), url.PathEscape(path))
}
