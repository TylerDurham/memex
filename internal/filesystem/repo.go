package filesystem

import (
	"io/fs"
	"os"
	"path/filepath"
)

type Repo struct {
	fs.FileInfo
	name     string
	absPath  string
}

func (r *Repo) Name() string {
	return r.name
}

func (r *Repo) AbsPath() string {
	return r.absPath
}

func NewRepo(path string) (Repo, error) {

	fi, err := os.Stat(path)
	if err != nil {
		return Repo{}, err
	}

	return Repo{
		FileInfo: fi,
		name:     filepath.Base(path),
		absPath:  filepath.Clean(path),
	}, nil
}

