package store

import (
	"fmt"
	"path/filepath"

	"github.com/TylerDurham/memex/internal/globals"
)

// Init initializes the store (creating a new one if it doesn't exist) and returns a
// a reference to the store.
func Init(ctx globals.App, name string) (store *Store, err error) {

	// build a path starting with config dir, the name of the repo, and the default db name
	storePath := filepath.Join(ctx.Config.ConfigDirectory(), name, globals.DBName)

	created, err := globals.EnsureDirectory(filepath.Dir(storePath))

	if err != nil {
		return nil, fmt.Errorf("could not ensure directory: %q: %w", storePath, err)
	}

	if created {
		ctx.Logger.Debug("store directory created at %q", "path", storePath)
	} else {
		ctx.Logger.Debug("store will be loaded from %q", "path", storePath)
	}

	// Open creates if it doesn't exit
	store, err = Open(storePath)

	if err != nil {
		return nil, fmt.Errorf("could not open store from %q: %w", storePath, err) 
	}

	return store, nil
}
