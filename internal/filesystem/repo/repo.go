// Package repo
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

var ErrNotADirectory = errors.New("not a directory")
var ErrRepoAlreadyExists = errors.New("repo already exists")
var ErrInvalidArgument = errors.New("invalid argument")

const (
	// AppStrategyDefault is the default strategy used for indexing a repo
	AppStrategyDefault = "default"

	// AppStrategyObsidian is the strategy used for indexing an Obsidian Notes repo
	AppStrategyObsidian = "obsidian"

	MemexConfigFileName = "memex-config.yaml"
)

type Repo struct {
	fs.FileInfo
	name           string
	absPath        string
	appStrategy    string
	configFilePath string
}

// Name returns the name of the repo. The base name of the absolute path is used for the name.
func (r *Repo) Name() string {
	return r.name
}

// AbsPath returns the absolute path to repo on the filesystem.
func (r *Repo) AbsPath() string {
	return r.absPath
}

// AppStrategy returns the strategy used for indexing the repo.
func (r *Repo) AppStrategy() string {
	return r.appStrategy
}

// ConfigFilePath returns the absolute path to the config file for the repo
func (r *Repo) ConfigFilePath() string {
	return r.configFilePath
}



func Save(repo Repo) error {
	return nil
}

func InitForApp(path string, appStrategy string) (Repo, error) {

	appStrategy = strings.ToLower(appStrategy)

	switch appStrategy {
	case AppStrategyObsidian:
	case AppStrategyDefault:
	default:
		return Repo{}, fmt.Errorf("app '%s' not supported: %w", appStrategy, ErrInvalidArgument)
	}

	fi, err := os.Stat(path)
	if err != nil {
		return Repo{}, err
	}

	if !fi.IsDir() {
		return Repo{}, fmt.Errorf("invalid repo path '%s': '%w'", path, ErrNotADirectory)
	}

	rci := RepoConfigInfo{
		Name:        filepath.Base(path),
		AbsPath:     filepath.Clean(path),
		AppStrategy: appStrategy,
	}

	configPath, err := config.MemexConfigDir()
	if err != nil {
		return Repo{}, err
	}

	repoConfigDirPath := filepath.Join(configPath, "repos", rci.Name)

	// MkdirAll succeeds when the directory already exists, so the
	// "already exists" check happens on the config file below.
	if err := os.MkdirAll(repoConfigDirPath, 0o755); err != nil {
		return Repo{}, err
	}

	data, err := yaml.Marshal(rci)
	if err != nil {
		return Repo{}, err
	}

	// O_EXCL makes creation fail with fs.ErrExist if the repo has already
	// been initialized, instead of silently overwriting its config.
	repoConfigFilePath := filepath.Join(repoConfigDirPath, MemexConfigFileName)
	f, err := os.OpenFile(repoConfigFilePath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		if errors.Is(err, fs.ErrExist) {
			return Repo{}, fmt.Errorf("repo %q already exists in %s", rci.Name, configPath)
		}
		return Repo{}, err
	}
	if _, err := f.Write(data); err != nil {
		f.Close()
		return Repo{}, err
	}
	if err := f.Close(); err != nil {
		return Repo{}, err
	}

	return Repo{
		FileInfo:       fi,
		name:           rci.Name,
		absPath:        rci.AbsPath,
		appStrategy:    rci.AppStrategy,
		configFilePath: repoConfigFilePath,
	}, nil
}

func Init(path string) (Repo, error) {
	return InitForApp(path, AppStrategyDefault)
}

var Empty = Repo{}

// RepoConfigInfo is the contents of a repo's config.yaml. It is used to persist repo
// settings to the filesystem.
type RepoConfigInfo struct {
	AppStrategy string `json:"strategy" yaml:"strategy"`
	Name        string `json:"name" yaml:"name"`

	// AbsPath is the absolute path to the repository directory.
	AbsPath string `json:"directory" yaml:"directory"`

	// ConfigFilePath and Database are derived by Load, not stored in config.yaml.
	ConfigFilePath string `json:"configFilePath" yaml:"-"`
	Database       string `json:"db" yaml:"-"`
}

