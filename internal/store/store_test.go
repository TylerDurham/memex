package store

import (
	"path/filepath"
	"testing"

	"github.com/TylerDurham/memex/internal/globals"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Tests use InitInPath with t.TempDir() rather than Init: Init resolves the
// config directory once at program start, so pointing MEMEX_CONFIG_DIR at a
// temp dir from inside a test has no effect, and cleaning up
// config.ConfigDir() would delete the user's real config.
func Test_Store_Init(t *testing.T) {
	dir := t.TempDir()

	store, err := InitInPath(dir, "foo-test")
	require.NoError(t, err)
	defer store.Close()

	assert.FileExists(t, filepath.Join(dir, "foo-test", globals.App().DBName()))

	chunks, files, err := store.Count(t.Context())
	require.NoError(t, err)
	assert.Zero(t, chunks)
	assert.Zero(t, files)
}
