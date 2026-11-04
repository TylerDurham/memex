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

type Repo struct {
	fs.FileInfo
	name        string
	absPath     string
	appStrategy string
	configFilePath  string
}

func (r *Repo) Name() string {
	return r.name
}

func (r *Repo) AbsPath() string {
	return r.absPath
}

func (r *Repo) AppStrategy() string {
	return r.appStrategy
}

func (r *Repo) ConfigFilePath() string {
	return r.configFilePath
}

func InitForApp(path string, appStrategy string) (Repo, error) {

	appStrategy = strings.ToLower(appStrategy)

	switch appStrategy {
	case "obsidian":
	case "default":
	default:
		return Repo{}, fmt.Errorf("appType '%s' not supported", appStrategy)
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

	repoConfigDirPath := filepath.Join(configPath, rci.Name)

	if err := os.MkdirAll(repoConfigDirPath, 0o755); err != nil {
		if errors.Is(err, fs.ErrExist) {
			return Repo{}, fmt.Errorf("repo %q already exists in %s", rci.Name, configPath)
		}
		return Repo{}, err
	}

	data, err := yaml.Marshal(rci)
	if err != nil {
		return Repo{}, err
	}

	repoConfigFilePath := filepath.Join(repoConfigDirPath, "memex-config.yaml")
	if err := os.WriteFile(repoConfigFilePath, data, 0o644); err != nil {
		return Repo{}, err
	}

	return Repo{
		FileInfo:    fi,
		name:        rci.Name,
		absPath:     rci.AbsPath,
		appStrategy: rci.AppStrategy,
		configFilePath:  repoConfigDirPath,
	}, nil
}

func Init(path string) (Repo, error) {
	return InitForApp(path, "default")
}

// RepoConfigInfo is the contents of a repo's config.yaml.
type RepoConfigInfo struct {
	AppStrategy string `json:"strategy" yaml:"strategy"`
	Name        string `json:"name" yaml:"name"`

	// AbsPath is the absolute path to the repository directory.
	AbsPath string `json:"directory" yaml:"directory"`

	// ConfigFilePath and Database are derived by Load, not stored in config.yaml.
	ConfigFilePath string `json:"configFilePath" yaml:"-"`
	Database   string `json:"db" yaml:"-"`
}
