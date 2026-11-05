package repo

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/TylerDurham/memex/internal/config"
	"go.yaml.in/yaml/v3"
)

func Load(name string) (Repo, error) {

	// load config
	cfg, err := config.Load()
	if err != nil {
		return Empty, err
	}

	// sanitize name
	name = strings.TrimSpace(name)
	if name == "" || name == "." || name == ".." || strings.ContainsRune(name, filepath.Separator) {
		return Empty, fmt.Errorf("invalid repo name %q: %w", name, ErrInvalidArgument)
	}

	repoConfigFilePath := filepath.Join(cfg.Repo.AbsPath, name, MemexConfigFileName)
	data, err := os.ReadFile(repoConfigFilePath)
	if err != nil {
		return Empty, err
	}

	var rci = RepoConfigInfo{}
	if err = yaml.Unmarshal(data, &rci); err != nil {
		return Empty, err
	}

	fi, err := os.Stat(rci.AbsPath)
	if err != nil {
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
