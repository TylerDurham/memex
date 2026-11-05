package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

var ErrRequiredConfigMissing = errors.New("missing required configuration value")

// EnvMemexConfigDir is the environment variable that overrides the default config directory.
const EnvMemexConfigDir = "MEMEX_CONFIG_DIR"
const EnvMemexAppStrategy = "MEMEX_APP_STRATEGY"

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
	cfg := &MemexConfig{}
	cfg.AbsPath = envl.required(EnvMemexConfigDir)

	// set up repo configuration
	cfg.Repo = RepoConfig{
		// should be ~/.config.memex/repos
		AbsPath: filepath.Join(cfg.AbsPath, "repos"),
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
