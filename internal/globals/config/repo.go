package config

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"go.yaml.in/yaml/v3"
)

// RepoConfigFile is the name of a repo's config file inside its directory
// under the config dir, next to its store.
const RepoConfigFile = "config.yaml"

// ErrRepoNotFound is returned when a repo has no config file, i.e. it was
// never initialized.
var ErrRepoNotFound = errors.New("repo not found")

// Repo is a repository's persisted settings, written by `memex init` and read
// by every command that works on that repository.
type Repo struct {
	Name      string `yaml:"-"`          // from the directory name, not stored
	Directory string `yaml:"directory"`  // absolute path to the repository (e.g. Obsidian vault)
	App       string `yaml:"app"`        // application/provider, e.g. "obsidian"
	OllamaURL string `yaml:"ollama_url"` // embedding server
	Model     string `yaml:"model"`      // embedding model; index and search must match
}

// RepoDir returns the directory holding a repo's config and store.
func RepoDir(base, name string) string {
	return filepath.Join(base, name)
}

// ValidateRepoName rejects names that aren't a single, plain directory name.
func ValidateRepoName(name string) error {
	if name == "" || name == "." || name == ".." || strings.ContainsAny(name, `/\`) {
		return fmt.Errorf("invalid repo name %q: must be a plain name without path separators", name)
	}
	return nil
}

// LoadRepo reads the named repo's config from the config dir.
func LoadRepo(name string) (*Repo, error) {
	return LoadRepoFrom(ConfigDir(), name)
}

// LoadRepoFrom reads the named repo's config from base. It returns an error
// wrapping ErrRepoNotFound if the repo was never initialized.
func LoadRepoFrom(base, name string) (*Repo, error) {
	if err := ValidateRepoName(name); err != nil {
		return nil, err
	}

	path := filepath.Join(RepoDir(base, name), RepoConfigFile)
	data, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, fmt.Errorf("%w: %q (run 'memex init %s --directory <path>')", ErrRepoNotFound, name, name)
	}
	if err != nil {
		return nil, fmt.Errorf("could not read repo config %q: %w", path, err)
	}

	r := &Repo{Name: name}
	if err := yaml.Unmarshal(data, r); err != nil {
		return nil, fmt.Errorf("could not parse repo config %q: %w", path, err)
	}
	return r, nil
}

// Save writes the repo's config into the config dir.
func (r *Repo) Save() error {
	return r.SaveTo(ConfigDir())
}

// SaveTo writes the repo's config under base, creating its directory. The
// file is written to a temp file and renamed, so a failed write never leaves
// a truncated config behind.
func (r *Repo) SaveTo(base string) error {
	if err := ValidateRepoName(r.Name); err != nil {
		return err
	}

	dir := RepoDir(base, r.Name)
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return fmt.Errorf("could not create repo dir %q: %w", dir, err)
	}

	data, err := yaml.Marshal(r)
	if err != nil {
		return fmt.Errorf("could not encode repo config: %w", err)
	}

	tmp, err := os.CreateTemp(dir, RepoConfigFile+".*")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name()) // no-op once renamed

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmp.Name(), filepath.Join(dir, RepoConfigFile))
}
