// Package repo manages memex repositories and their configuration.
package repo

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"go.yaml.in/yaml/v3"
)

// ConfigFile is the name of the configuration file inside each repo's config subdirectory.
const ConfigFile = "config.yaml"

// Config is the contents of a repo's config.yaml.
type Config struct {
	Application string `yaml:"application"`
	Name        string `yaml:"name"`
	// Directory is the absolute path to the repository directory.
	Directory string `yaml:"directory"`
}

// InitOptions are the inputs to Init, as given to the repo init command.
type InitOptions struct {
	// Directory is the repository directory (e.g. an Obsidian vault).
	Directory string
	// Name is the repo name. If empty, the base name of Directory is used.
	Name string
	// Application is the application the repo belongs to.
	Application string
}

// Init creates <configDir>/<name>/config.yaml for a new repo and returns the path
// to the file. It fails if a repo with that name already exists.
func Init(configDir string, opts InitOptions) (string, error) {
	directory, err := filepath.Abs(opts.Directory)
	if err != nil {
		return "", err
	}
	name := opts.Name
	if name == "" {
		name = filepath.Base(directory)
	}
	if err := validateName(name); err != nil {
		return "", err
	}

	if err := os.MkdirAll(configDir, 0o755); err != nil {
		return "", err
	}
	repoDir := filepath.Join(configDir, name)
	if err := os.Mkdir(repoDir, 0o755); err != nil {
		if errors.Is(err, fs.ErrExist) {
			return "", fmt.Errorf("repo %q already exists in %s", name, configDir)
		}
		return "", err
	}

	data, err := yaml.Marshal(Config{Application: opts.Application, Name: name, Directory: directory})
	if err != nil {
		return "", err
	}
	path := filepath.Join(repoDir, ConfigFile)
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return "", err
	}
	return path, nil
}

// validateName rejects names that can't be used as a single directory name.
func validateName(name string) error {
	if name == "" || name == "." || name == ".." || name == string(filepath.Separator) ||
		strings.ContainsAny(name, `/\`) {
		return fmt.Errorf("invalid repo name %q", name)
	}
	return nil
}
