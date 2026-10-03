// Package config
package config

import (
	"os"
	"path/filepath"
	"runtime"
)

// EnvMemexConfigDir is the environment variable that overrides the default config directory.
const EnvMemexConfigDir = "MEMEX_CONFIG_DIR"

// Dir returns the config directory: $MEMEX_CONFIG_DIR if set, otherwise
// %AppData%\memex on Windows. Everywhere else it is $XDG_CONFIG_HOME/memex
// if $XDG_CONFIG_HOME is an absolute path, falling back to $HOME/.config/memex.
func Dir() (string, error) {
	if dir := os.Getenv(EnvMemexConfigDir); dir != "" {
		return dir, nil
	}
	if runtime.GOOS == "windows" {
		appData, err := os.UserConfigDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(appData, "memex"), nil
	}
	// The XDG spec says relative paths must be ignored.
	if xdg := os.Getenv("XDG_CONFIG_HOME"); filepath.IsAbs(xdg) {
		return filepath.Join(xdg, "memex"), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "memex"), nil
}
