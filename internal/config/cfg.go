package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

var ErrRequiredConfigMissing = errors.New("missing required configuration value")

// EnvMemexConfigHome is the environment variable that overrides the default config directory.
const EnvMemexConfigHome = "MEMEX_CONFIG_DIR"
const EnvMemexAppStrategy = "MEMEX_APP_STRATEGY"
const EnvXdgConfigHome = "XDG_CONFIG_HOME"

type RepoConfig struct {
	AbsPath  string
	Strategy string
}

type MemexConfig struct {
	AbsPath string
	Repo    RepoConfig
}

func Load() (*MemexConfig, error) {
	var envl envLoader

	cp, err := getConfigDir(envl)
	if err != nil {
		return nil, err
	}

	cfg := &MemexConfig{}

	cfg.AbsPath = cp

	// set up repo configuration
	cfg.Repo = RepoConfig{
		// should be ~/.config.memex/repos
		AbsPath:  filepath.Join(cfg.AbsPath, "repos"),
		Strategy: envl.str(EnvMemexAppStrategy, "obsidian"),
	}

	if err := errors.Join(envl.errs...); err != nil {
		return nil, err
	}

	return cfg, nil
}

// envLoader is a simple struct that providers utility functions
// for getting configuration values from ENV
type envLoader struct{ errs []error }

func (l *envLoader) required(key string) string {
	v := os.Getenv(key)
	if v == "" {
		l.errs = append(l.errs, fmt.Errorf("'%s': %w", key, ErrRequiredConfigMissing))
	}

	return v
}

func (l *envLoader) str(key string, def string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}

	return def
}

// getConfigDir figures out where the user's config directory lives. It tries to get from
// the environment variable [EnvMemexConfigHome] first
func getConfigDir(envl envLoader) (string, error) {

	if dir := envl.str(EnvMemexConfigHome, ""); dir != "" {
		return dir, nil
	}

	if runtime.GOOS == "windows" {
		win, err := os.UserConfigDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(win, "memex"), nil
	}

	// The XDG spec says relative paths must be ignored.
	if xdg := envl.str(EnvXdgConfigHome, ""); filepath.IsAbs(xdg) {
		return filepath.Join(xdg, "memex"), nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "memex"), nil
}
