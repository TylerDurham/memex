package globals

import (
	"errors"
	"fmt"
	"os"
)

// EnvConfigDir is the environment variable that overrides the default config directory.
const EnvConfigDir = "MEMEX_CONFIG_DIR"

// ConfigInfo holds the resolved config and log directory paths for the
// application. Created via NewConfig, which also ensures both directories
// exist on disk.
type ConfigInfo struct {
	configDirectory string
	logDirectory    string
}

// ConfigDirectory returns the application's config directory.
func (c *ConfigInfo) ConfigDirectory() string {
	return c.configDirectory
}

// LogDirectory returns the application's log directory.
func (c *ConfigInfo) LogDirectory() string {
	return c.logDirectory
}

// NewConfig creates a new instance of the ConfigInfo struct. Also ensures the config
// and log directories exist on the local file system. Returns the newly created instance of
// ConfigInfo and an error if one occurred.
func NewConfig() (cfg ConfigInfo, err error) {
	cfgDir := GetConfigDir()
	if _, err := EnsureDirectory(cfgDir); err != nil {
		return cfg, err
	}

	logDir := GetLogDir()
	if _, err := EnsureDirectory(logDir); err != nil {
		return cfg, err
	}

	cfg.configDirectory = cfgDir
	cfg.logDirectory = logDir

	return cfg, nil
}

// ErrNotADirectory is returned when a path exists but is not a directory.
var ErrNotADirectory = errors.New("path exists but is not a directory")

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

// GetConfigDir returns the configuration directory for Memex.
// It first checks if the EnvConfigDir environment variable is set and returns its value if found.
// If not, it uses the user's home directory to construct the path in a cross-platform compatible manner (e.g., ~/.config/memex).
func GetConfigDir() string {

	home, ok := os.LookupEnv(EnvConfigDir)
	if !ok {
		home, _ = os.UserHomeDir()
		// TODO: Need to ensure works cross-platform (win, macos, etc.)
		return fmt.Sprintf("%s/.config/%s", home, AppName)

	}

	return home
}

// GetLogDir returns the log directory for Memex.
// It constructs the path by appending "/logs" to the configuration directory obtained from GetConfigDir().
func GetLogDir() string {
	return fmt.Sprintf("%s/logs", GetConfigDir())
}
