// Package config
package config

import (
	"fmt"
	"log"
	"os"

	"path/filepath"

	"github.com/TylerDurham/memex/internal/globals"
)

// EnvConfigDir is the environment variable that overrides the default config directory.
const EnvConfigDir = "MEMEX_CONFIG_DIR"

// EnvLogDir is the environment variable that overrides the default config directory.
const EnvLogDir = "MEMEX_LOG_DIR"

var (
	configDir string
	logDir    string
)

func ConfigDir() string {
	return configDir
}

func LogDir() string {
	return logDir
}

func init() {
	reset()
}

func reset() {
	err := initConfigDir(&configDir)
	if err != nil {
		log.Fatalf("could not initialize config dir: %+v", err)
	}

	err = initLogDir(configDir, &logDir)
	if err != nil {
		log.Fatalf("could not initialize log dir: %+v", err)
	}
}

// initConfigDir determines the configuration directory for Memex.
// It first checks if the MEMEX_CONFIG_DIR environment variable is
// set and returns its value if found.
//
// If not, it uses the user's home directory to construct the path
// in a cross-platform compatible manner (e.g., ~/.config/memex).
func initConfigDir(dir *string) error {
	d, isSet := os.LookupEnv(EnvConfigDir)

	if isSet {
		*dir = d
		return nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("could not detect user home directory: %w", err)
	}

	// TODO: Need to ensure works cross-platform (win, macos, etc.)
	*dir = filepath.Join(home, ".config", globals.App().Name())

	// TODO: Log whether config dir was created or already existed
	if _, err := globals.EnsureDirectory(*dir); err != nil {
		return err
	}

	return nil
}

func initLogDir(cfgDir string, dir *string) error {
	var isSet bool
	*dir, isSet = os.LookupEnv(EnvLogDir)

	if !isSet {
		// env not set... put in config dir
		*dir = filepath.Join(cfgDir, "logs")
	}

	// TODO: Log whether config dir was created or already existed
	if _, err := globals.EnsureDirectory(*dir); err != nil {
		return err
	}

	return nil
}
