// Package vault handles discovery of markdown files within an Obsidian vault,
// skipping the .obsidian config directory and any hidden/dot directories.
package vault

import (
	"crypto/sha256"
	"encoding/hex"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// File represents a single markdown file discovered in the vault.
type File struct {
	// AbsPath is the absolute filesystem path to the file.
	AbsPath string
	// RelPath is the path relative to the vault root, used as a stable
	// identifier in the index (so the vault can be moved/synced without
	// invalidating the whole index).
	RelPath string
	// ModTime (unix seconds) is used as a cheap first-pass change check
	// before falling back to a content hash comparison.
	ModTime int64
	// Size in bytes.
	Size int64
}

// Walk returns every .md file under root, sorted by RelPath for deterministic
// ordering (which matters for --dry-run diffing and for tests).
func Walk(root string) ([]File, error) {
	var files []File

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		name := d.Name()

		// Skip Obsidian's own config dir and any hidden directories
		// (.git, .trash, .obsidian, etc). Never descend into them.
		if d.IsDir() {
			if name == ".obsidian" || name == ".trash" || strings.HasPrefix(name, ".") {
				return filepath.SkipDir
			}
			return nil
		}

		if !strings.EqualFold(filepath.Ext(name), ".md") {
			return nil
		}

		info, err := d.Info()
		if err != nil {
			return err
		}

		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}

		files = append(files, File{
			AbsPath: path,
			RelPath: filepath.ToSlash(rel),
			ModTime: info.ModTime().Unix(),
			Size:    info.Size(),
		})

		return nil
	})
	if err != nil {
		return nil, err
	}

	return files, nil
}

// HashContent returns a hex-encoded sha256 of the given bytes. Used to detect
// whether a file's content actually changed (not just its mtime — matters
// because sync tools like Syncthing/iCloud sometimes touch mtimes without
// changing content).
func HashContent(b []byte) string {
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

// ReadFile is a small wrapper kept here so callers don't need to import os
// directly just to read a vault file; also centralizes any future decoding
// concerns (e.g. BOM stripping).
func ReadFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}
