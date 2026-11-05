// Package testutil
package testutil

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/TylerDurham/memex/internal/config"
)

// IsolateConfig points MEMEX_CONFIG_DIR at a fresh temporary directory for
// the rest of the test, so repo.Init never writes to the user's real memex
// config. t.Setenv restores the variable and t.TempDir is removed when the
// test ends. Because it uses t.Setenv, callers cannot use t.Parallel().
func IsolateConfig(t *testing.T) {
	t.Helper()
	t.Setenv(config.EnvMemexConfigDir, filepath.Join(t.TempDir(), ".config"))
	t.Logf("config dir isolated to '%s'", os.Getenv(config.EnvMemexConfigDir))
}
