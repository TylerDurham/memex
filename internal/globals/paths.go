// Package globals
package globals

import (
	"fmt"
	"os"
)

const AppName = "memex"

// EnsureMemexConfigDir ensures that the Memex configuration directory exists.
// It retrieves the configuration directory using GetConfigDir() and checks if it already exists using os.Stat().
// If the directory does not exist, it creates the directory along with any necessary parent directories using os.MkdirAll().
// The function logs informational messages about the process using the slog package.
func EnsureMemexConfigDir() error {
	hd := GetConfigDir()

	Logger().Debug("ensuring memex directory", "directory", hd)
	_, err := os.Stat(hd)
	if err != nil {

		Logger().Debug("memex directory does not exist... creating", "directory", hd)
		err := os.MkdirAll(hd, 0750)
		if err != nil {
			return err
		}
	}
	return nil
}

// GetConfigDir returns the configuration directory for Memex.
// It first checks if the MEMEX_CONFIG_DIR environment variable is set and returns its value if found.
// If not, it uses the user's home directory to construct the path in a cross-platform compatible manner (e.g., ~/.config/memex).
func GetConfigDir() string {

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
func GetMemexLogDir() string {
	return fmt.Sprintf("%s/logs", GetConfigDir())
}
