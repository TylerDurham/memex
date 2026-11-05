// Package repo
package repo

import (
	"errors"
	"io/fs"
)

var ErrNotADirectory = errors.New("not a directory")
var ErrRepoAlreadyExists = errors.New("repo already exists")
var ErrInvalidArgument = errors.New("invalid argument")
var ErrInvalidArgumentStrategy = errors.New("invalid argument: strategy")

const (
	// AppStrategyGeneric is the default strategy used for indexing a repo
	AppStrategyGeneric = "default"

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

// Empty is a nil repo
var Empty = (Repo{})

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
