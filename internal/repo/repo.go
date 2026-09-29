// Package repo manages memex repositories and their configuration.
package repo

import (
	"encoding/json"
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

// ReposDir is the subdirectory of the config directory that holds one directory per repo.
const ReposDir = "repos"

// reposDir returns <configDir>/repos.
func reposDir(configDir string) string {
	return filepath.Join(configDir, ReposDir)
}

// RepoInfo is the contents of a repo's config.yaml.
type RepoInfo struct {
	Application string `json:"application" yaml:"application"`
	Name        string `json:"name" yaml:"name"`

	// Directory is the absolute path to the repository directory.
	Directory  string `json:"directory" yaml:"directory"`
	ConfigFile string `json:"config" yaml:"config"`
	Database   string `json:"db" yaml:"db"`
}

// ToJSONString marshalls the Config into a JSON string.
func (doc *RepoInfo) ToJSONString() (string, error) {

	json, err := json.MarshalIndent(doc, "", "	")

	if err != nil {
		return "", fmt.Errorf("could not serialize to json: %+v", err)
	}

	return string(json), err
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

// Init creates <configDir>/repos/<name>/config.yaml for a new repo and returns the path
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

	repos := reposDir(configDir)
	if err := os.MkdirAll(repos, 0o755); err != nil {
		return "", err
	}

	repoDir := filepath.Join(repos, name)
	if err := os.Mkdir(repoDir, 0o755); err != nil {
		if errors.Is(err, fs.ErrExist) {
			return "", fmt.Errorf("repo %q already exists in %s", name, repos)
		}
		return "", err
	}

	data, err := yaml.Marshal(RepoInfo{Application: opts.Application, Name: name, Directory: directory})
	if err != nil {
		return "", err
	}
	path := filepath.Join(repoDir, ConfigFile)
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return "", err
	}
	return path, nil
}

// ErrNotFound is returned by Load when no repo with the given name exists.
var ErrNotFound = errors.New("repo not found")

// Load reads <configDir>/repos/<name>/config.yaml and returns the repo's configuration.
// It returns an error wrapping ErrNotFound if the repo doesn't exist.
func Load(configDir, name string) (*RepoInfo, error) {
	if err := validateName(name); err != nil {
		return nil, err
	}
	repos := reposDir(configDir)
	path := filepath.Join(repos, name, ConfigFile)
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, fmt.Errorf("%w: %q in %s", ErrNotFound, name, repos)
		}
		return nil, err
	}
	var cfg RepoInfo
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}

	cfg.Database = filepath.Join(repos, name, "memex.db")
	cfg.ConfigFile = path
	return &cfg, nil
}

// Names returns the names of all repos in configDir, sorted. A repo is any
// subdirectory of <configDir>/repos containing a config.yaml. A missing
// directory means no repos.
func Names(configDir string) ([]string, error) {
	repos := reposDir(configDir)
	entries, err := os.ReadDir(repos)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	var names []string
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		if _, err := os.Stat(filepath.Join(repos, e.Name(), ConfigFile)); err == nil {
			names = append(names, e.Name())
		}
	}
	return names, nil
}

// validateName rejects names that can't be used as a single directory name.
func validateName(name string) error {
	if name == "" || name == "." || name == ".." || name == string(filepath.Separator) ||
		strings.ContainsAny(name, `/\`) {
		return fmt.Errorf("invalid repo name %q", name)
	}
	return nil
}
