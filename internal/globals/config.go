package globals

import (
	"errors"
	"fmt"
	"os"
)

type ConfigInfo struct {
	configDirectory string
	logDirectory    string
}

func (c *ConfigInfo) ConfigDirectory() string {
	return c.configDirectory
}

func (c *ConfigInfo) LogDirectory() string {
	return c.logDirectory
}

func newConfig() (cfg ConfigInfo, err error) {
	cfgDir := calculateConfigDir()
	if _, err := EnsureDirectory(cfgDir); err != nil {
		return cfg, err
	}

	logDir := calculateLogDir() 
	if _, err := EnsureDirectory(logDir); err != nil {
		return cfg, err
	}

	cfg.configDirectory = cfgDir
	cfg.logDirectory = logDir

	return cfg, nil
}

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

// // EnsureMemexConfigDir ensures that the Memex configuration directory exists.
// // It retrieves the configuration directory using GetConfigDir() and checks if it already exists using os.Stat().
// // If the directory does not exist, it creates the directory along with any necessary parent directories using os.MkdirAll().
// // The function logs informational messages about the process using the slog package.
// func ensureConfigDir() (string, error) {
// 	app, _ := InitApp()
// 	log := app.Logger
// 	hd := GetConfigDir()
//
// 	log.Debug("ensuring memex directory", "directory", hd)
// 	_, err := os.Stat(hd)
// 	if err != nil {
//
// 		log.Debug("memex directory does not exist... creating", "directory", hd)
// 		err := os.MkdirAll(hd, 0750)
// 		if err != nil {
// 			return "", err
// 		}
// 	}
// 	return hd, nil
// }
//
// GetConfigDir returns the configuration directory for Memex.
// It first checks if the MEMEX_CONFIG_DIR environment variable is set and returns its value if found.
// If not, it uses the user's home directory to construct the path in a cross-platform compatible manner (e.g., ~/.config/memex).
func calculateConfigDir() string {

	home, ok := os.LookupEnv("MEMEX_CONFIG_DIR")
	if !ok {
		home, _ = os.UserHomeDir()
		// TODO: Need to ensure works cross-platform (win, macos, etc.)
		return fmt.Sprintf("%s/.config/%s", home, AppName)

	}

	return home
}

// GetMemexLogDir returns the log directory for Memex.
// It constructs the path by appending "/logs" to the configuration directory obtained from GetConfigDir().
func calculateLogDir() string {
	return fmt.Sprintf("%s/logs", calculateConfigDir())
}
