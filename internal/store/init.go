package store

import (
	"fmt"
	"path/filepath"

	"github.com/TylerDurham/memex/internal/globals"
	"github.com/TylerDurham/memex/internal/globals/config"
	"github.com/TylerDurham/memex/internal/globals/logger"
)

// Init initializes the store (creating a new one if it doesn't exist) and returns a
// a reference to the store.
func Init(name string) (store *Store, err error) {
	// TODO: Need to make this overridable to a custom directory
	return InitInPath(config.ReposDir(config.ConfigDir()), name)
}

// InitInPath is Init with an explicit base directory instead of the config
// directory. The store lives at <path>/<name>/<db name>.
func InitInPath(path string, name string) (store *Store, err error) {

	// build a path starting with the base dir, the name of the repo, and the default db name
	storePath := filepath.Join(path, name, globals.App().DBName())

	created, err := globals.EnsureDirectory(filepath.Dir(storePath))

	if err != nil {
		return nil, fmt.Errorf("could not ensure directory: %q: %w", storePath, err)
	}

	if created {
		logger.Debug("store directory created at %q", "path", storePath)
	} else {
		logger.Debug("store will be loaded from %q", "path", storePath)
	}

	// Open creates if it doesn't exit
	store, err = Open(storePath)

	if err != nil {
		return nil, fmt.Errorf("could not open store from %q: %w", storePath, err)
	}

	return store, nil
}
