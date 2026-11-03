package filesystem

import (
	"io/fs"
	"os"
)

type Repo struct {
	FileInfo fs.FileInfo
	name     string
	absPath  string
}

func (r *Repo) Name() string {
	return r.name
}

func (r *Repo) AbsPath() string {
	return r.absPath
}

func NewRepo(name string, path string) (Repo, error) {

	fi, err := os.Stat(path)
	if err != nil {
		return Repo{}, err
	}

	return Repo{
		FileInfo: fi,
		name:     name,
		absPath:  path,
	}, nil
}
