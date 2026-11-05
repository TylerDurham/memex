package repo

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/TylerDurham/memex/internal/config"
	"go.yaml.in/yaml/v3"
)

// Init registers the directory at path as a memex repo and writes its
// configuration file into the memex config directory.
//
// Arguments:
//   - path: path to an existing directory to register as the repo. The
//     repo name is the base name of path, so two directories with the same
//     base name cannot both be initialized.
//   - strategy: the application that manages the notes in the repo,
//     such as [AppStrategyObsidian] or [AppStrategyGeneric]. Matching is
//     case-insensitive.
//
// The configuration is written to repos/<name>/[MemexConfigFileName] under
// the directory returned by [config.MemexConfigDir]. Init never overwrites
// an existing repo configuration.
//
// Init returns an error wrapping:
//   - [ErrInvalidArgument] if appStrategy is not supported.
//   - [ErrNotADirectory] if path exists but is not a directory.
//   - [ErrRepoAlreadyExists] if a repo with the same name is already
//     initialized.
//
// Errors from [os.Stat] are returned unwrapped, so callers can check for
// [fs.ErrNotExist] when path does not exist.
func Init(path string, strategy string) (Repo, error) {

	strategy = strings.ToLower(strategy)

	//TODO: Need to move to table driven strategy lookup at some point
	switch strategy {
	case AppStrategyObsidian:
	case AppStrategyGeneric:
	default:
		return Repo{}, fmt.Errorf("app '%s' not supported: %w", strategy, ErrInvalidArgumentStrategy)
	}

	// did the caller pass us something that actually exists on the filesystem?
	fi, err := os.Stat(path)
	if err != nil {
		return Repo{}, err
	}

	// only accept a path to a directory
	if !fi.IsDir() {
		return Repo{}, fmt.Errorf("invalid repo path '%s': '%w'", path, ErrNotADirectory)
	}

	// set up struct that will be marshalled to filesystem
	rci := RepoConfigInfo{
		Name:        filepath.Base(path),
		AbsPath:     filepath.Clean(path),
		AppStrategy: strategy,
	}

	// should be ~/.config/memex
	cfg, err := config.Load()
	if err != nil {
		return Empty, err
	}
	configPath := cfg.AbsPath
	if err != nil {
		return Empty, err
	}

	// should be ~/.config/memex/repos/<repo-name>
	repoConfigDirPath :=  filepath.Join(cfg.Repo.AbsPath, rci.Name)

	// MkdirAll succeeds when the directory already exists, so the
	// "already exists" check happens on the config file below.
	if err := os.MkdirAll(repoConfigDirPath, 0o755); err != nil {
		return Empty, err
	}

	data, err := yaml.Marshal(rci)
	if err != nil {
		return Empty, err
	}

	// O_EXCL makes creation fail with fs.ErrExist if the repo has already
	// been initialized, instead of silently overwriting its config.
	repoConfigFilePath := filepath.Join(repoConfigDirPath, MemexConfigFileName)
	f, err := os.OpenFile(repoConfigFilePath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		if errors.Is(err, fs.ErrExist) {
			return Empty, fmt.Errorf("repo %q already exists in %s: %w", rci.Name, configPath, ErrRepoAlreadyExists)
		}
		return Empty, err
	}

	// persist to filesystem
	if _, err := f.Write(data); err != nil {
		f.Close()
		return Empty, err
	}

	if err := f.Close(); err != nil {
		return Empty, err
	}

	return Repo{
		FileInfo:       fi,
		name:           rci.Name,
		absPath:        rci.AbsPath,
		appStrategy:    rci.AppStrategy,
		configFilePath: repoConfigFilePath,
	}, nil
}
