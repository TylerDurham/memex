package store

import (
	"path/filepath"

	"github.com/TylerDurham/memex/internal/globals"
)

func Init(ctx globals.App, name string) (store *Store, err error) {

	// build a path starting with config dir, the name of the repo, and the default db name
	storePath := filepath.Join(ctx.Config.ConfigDirectory(), name, globals.DBName)

	globals.EnsureDirectory(filepath.Dir(storePath))

	// Open creates if it doesn't exit
	store, err = Open(storePath)

	if err != nil {
		return nil, err
	}

	return store, nil
}
