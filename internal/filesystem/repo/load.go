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

	if strings.HasPrefix("../", name) || strings.HasPrefix("./", name) {
		return Repo{}, fmt.Errorf("name cannot contain '/': %w", ErrInvalidArgument)
	}

	configDirPath, err := config.MemexConfigDir()
	if err != nil {
		return Repo{}, err
	}

	repoConfigFilePath := filepath.Join(configDirPath, "repos", name, MemexConfigFileName)
	data, err := os.ReadFile(repoConfigFilePath)
	if err != nil {
		return Repo{}, err
	}

	var rci = RepoConfigInfo{}
	if err = yaml.Unmarshal(data, &rci); err != nil {
		return Repo{}, err
	}

	fi, err := os.Stat(rci.AbsPath)
	if err != nil {
		return Repo{}, err
	}

	return Repo{
		FileInfo: fi,
		name: rci.Name,
		absPath: rci.AbsPath,
		appStrategy: rci.AppStrategy,
		configFilePath: repoConfigFilePath,
	}, nil
}
