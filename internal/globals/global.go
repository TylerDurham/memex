// Package globals
package globals

import (
	"errors"
	"fmt"
	"os"
)

// const AppName = "memex"
// const DBName = AppName + "_index.db"

// ErrNotADirectory is returned when a path exists but is not a directory.
var ErrNotADirectory = errors.New("path exists but is not a directory")

// ErrNotAFile is returned when a path exists but is not a file.
var ErrNotAFile = errors.New("path exists but is not a file")

// EnsureDirectory ensures that path exists as a directory, creating it (and any
// necessary parents) with permissions 0750 if it does not already exist.
//
// created reports whether the directory was newly created. If path already
// exists but is not a directory, EnsureDirectory returns created=false and an
// error wrapping ErrNotADirectory. Any other stat or creation failure is
// also returned wrapped with context about the failing path.
func EnsureDirectory(path string) (created bool, err error) {
	dir, err := os.Stat(path)

	switch {
	case err == nil:
		if dir.IsDir() {
			// Directory existed; no error
			return false, nil
		}
		// Path existed, but was not a directory
		return false, fmt.Errorf("ensure dir: %q: %w", path, ErrNotADirectory)
	case !errors.Is(err, os.ErrNotExist):
		// No idea what happened if we get here...
		return false, fmt.Errorf("ensure dir: %q: %w", path, err)
	}

	// Attempt to make the path
	if err := os.MkdirAll(path, 0750); err != nil {
		return false, fmt.Errorf("ensure dir: %q: %w", path, err)
	}

	return true, nil
}

